package sim

import (
	"fmt"
	"math"
	"sort"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

const ticksPerMonth = 24 // 1 game month = 24 real hours = 24 ticks

func (e *Engine) Tick() {
	log := []string{}
	e.World.Tick++
	// 1 tick = 1 real hour = 1/24 game month
	hourOfMonth := ((e.World.Tick - 1) % ticksPerMonth) + 1
	if e.World.Tick > 1 && hourOfMonth == 1 {
		e.World.Month++
		if e.World.Month > 12 {
			e.World.Month = 1
			e.World.Year++
		}
	}
	log = append(log, fmt.Sprintf("Тик %d — час %d/24 · %02d.%d", e.World.Tick, hourOfMonth, e.World.Month, e.World.Year))

	e.resetLabor()
	e.runDeputies(&log)
	e.resolveContracts(&log)
	e.advanceShipments(&log)
	e.runProduction(&log)
	e.matchMarket(&log)
	e.runRetail(&log)
	e.payRentsAndWages(&log)
	e.aiAct(&log)
	e.growCities(&log)
	e.updateRankings()
	if hourOfMonth == ticksPerMonth {
		e.monthlyEvent(&log)
		e.discoverResources(&log)
	}
	e.World.LastTickLog = log
	e.World.Events = append(e.World.Events, log...)
	if len(e.World.Events) > 200 {
		e.World.Events = e.World.Events[len(e.World.Events)-200:]
	}
}

func (e *Engine) resetLabor() {
	for _, c := range e.World.Cities {
		c.LaborUsed = model.Labor{}
	}
}

func (e *Engine) resolveContracts(log *[]string) {
	for _, c := range e.World.Contracts {
		if c.TicksLeft <= 0 {
			continue
		}
		sellerBuildings := e.firmBuildingsInCity(c.SellerID, c.CityFrom)
		need := c.QtyPerTick
		moved := 0
		var qualitySum float64
		for _, b := range sellerBuildings {
			have := b.Inventory[c.ItemID]
			if have <= 0 {
				continue
			}
			take := min(have, need-moved)
			b.Inventory[c.ItemID] -= take
			moved += take
			qualitySum += b.Quality * float64(take)
			if moved >= need {
				break
			}
		}
		if moved == 0 {
			continue
		}
		buyer := e.World.Firms[c.BuyerID]
		seller := e.World.Firms[c.SellerID]
		pay := float64(moved) * c.Price
		if buyer.Cash < pay {
			continue
		}
		buyer.Cash -= pay
		seller.Cash += pay
		avgQ := qualitySum / float64(moved)
		routeTicks, cost := e.routeInfo(c.CityFrom, c.CityTo)
		own := c.UseOwnFleet && buyer.Trucks > 0
		shipCost := float64(moved) * cost
		if own {
			shipCost *= 0.55
		} else {
			shipCost *= 1.0
		}
		buyer.Cash -= shipCost
		id := uuid.NewString()
		e.World.Shipments[id] = &model.Shipment{
			ID: id, FirmID: c.BuyerID, ItemID: c.ItemID, Qty: moved, Quality: avgQ,
			FromCity: c.CityFrom, ToCity: c.CityTo, TicksLeft: max(1, routeTicks), CostPaid: shipCost, OwnFleet: own,
		}
		c.TicksLeft--
		*log = append(*log, fmt.Sprintf("Контракт %s: %d %s → %s", c.ID[:8], moved, c.ItemID, c.CityTo))
	}
}

func (e *Engine) advanceShipments(log *[]string) {
	for id, s := range e.World.Shipments {
		s.TicksLeft--
		if s.TicksLeft > 0 {
			continue
		}
		// deliver into any warehouse/building of firm in destination
		delivered := false
		for _, b := range e.World.Buildings {
			if b.FirmID == s.FirmID && b.CityID == s.ToCity {
				b.Inventory[s.ItemID] += s.Qty
				if s.Quality > 0 {
					b.Quality = (b.Quality + s.Quality) / 2
				}
				delivered = true
				break
			}
		}
		if !delivered {
			// create virtual holding on first building of firm
			for _, b := range e.World.Buildings {
				if b.FirmID == s.FirmID {
					b.Inventory[s.ItemID] += s.Qty
					delivered = true
					break
				}
			}
		}
		*log = append(*log, fmt.Sprintf("Доставка %d %s в %s", s.Qty, s.ItemID, s.ToCity))
		delete(e.World.Shipments, id)
	}
}

func (e *Engine) runProduction(log *[]string) {
	for _, b := range e.World.Buildings {
		if !b.Active {
			continue
		}
		if b.RetoolWeeks > 0 {
			b.RetoolWeeks--
			*log = append(*log, fmt.Sprintf("Переналадка %s (%d)", b.Type, b.RetoolWeeks))
			continue
		}
		recipe, ok := e.Catalog.RecipeByID[b.RecipeID]
		if !ok {
			continue
		}
		firm := e.World.Firms[b.FirmID]
		if firm == nil {
			continue
		}
		if recipe.ResearchCost > 0 && !firm.Researched[recipe.ID] {
			continue
		}
		city := e.World.Cities[b.CityID]
		if city == nil {
			continue
		}
		if !e.reserveLabor(city, recipe.Labor) {
			continue
		}
		outQty := recipe.Output.Qty
		// season bonus
		for _, m := range recipe.SeasonBonusMonths {
			if m == e.World.Month {
				outQty = int(math.Ceil(float64(outQty) * 1.25))
				break
			}
		}
		// Добыча: списываем запас залежи на клетке (не бесконечно)
		if isExtractRecipe(recipe) {
			depQ, ok := e.consumeDeposit(city, b.X, b.Y, outQty, recipe.Output.ItemID, log)
			if !ok {
				continue // залежь иссякла / нет ресурса
			}
			energyCost := float64(recipe.Energy) * 8
			firm.Cash -= energyCost
			b.Inventory[recipe.Output.ItemID] += outQty
			b.Quality = clamp((depQ*0.7)+(b.Quality*0.3), 1, 5)
			continue
		}
		// check inputs
		can := true
		for _, in := range recipe.Inputs {
			if b.Inventory[in.ItemID] < in.Qty {
				can = false
				break
			}
		}
		if !can {
			continue
		}
		inputQ := b.Quality
		for _, in := range recipe.Inputs {
			b.Inventory[in.ItemID] -= in.Qty
		}
		energyCost := float64(recipe.Energy) * 15
		firm.Cash -= energyCost
		b.Inventory[recipe.Output.ItemID] += outQty
		b.Quality = clamp((inputQ*0.7)+2.0*0.3, 1, 5)
	}
}

func (e *Engine) reserveLabor(city *model.CityState, need model.Labor) bool {
	if city.LaborUsed.Worker+need.Worker > city.LaborPool.Worker {
		return false
	}
	if city.LaborUsed.Engineer+need.Engineer > city.LaborPool.Engineer {
		return false
	}
	if city.LaborUsed.Manager+need.Manager > city.LaborPool.Manager {
		return false
	}
	city.LaborUsed.Worker += need.Worker
	city.LaborUsed.Engineer += need.Engineer
	city.LaborUsed.Manager += need.Manager
	return true
}

func (e *Engine) matchMarket(log *[]string) {
	type key struct{ city, item string }
	buys := map[key][]*model.MarketOrder{}
	sells := map[key][]*model.MarketOrder{}
	for _, o := range e.World.Orders {
		remain := o.Qty - o.Filled
		if remain <= 0 {
			continue
		}
		k := key{o.CityID, o.ItemID}
		if o.Side == "buy" {
			buys[k] = append(buys[k], o)
		} else {
			sells[k] = append(sells[k], o)
		}
	}
	for k, sellList := range sells {
		buyList := buys[k]
		sort.Slice(sellList, func(i, j int) bool { return sellList[i].Price < sellList[j].Price })
		sort.Slice(buyList, func(i, j int) bool { return buyList[i].Price > buyList[j].Price })
		si, bi := 0, 0
		for si < len(sellList) && bi < len(buyList) {
			s, b := sellList[si], buyList[bi]
			if b.Price < s.Price {
				bi++
				continue
			}
			qty := min(s.Qty-s.Filled, b.Qty-b.Filled)
			if qty <= 0 {
				if s.Qty-s.Filled <= 0 {
					si++
				}
				if b.Qty-b.Filled <= 0 {
					bi++
				}
				continue
			}
			// take from seller inventory
			taken := e.takeInventory(s.FirmID, s.CityID, s.ItemID, qty)
			if taken == 0 {
				si++
				continue
			}
			price := (s.Price + b.Price) / 2
			buyer, seller := e.World.Firms[b.FirmID], e.World.Firms[s.FirmID]
			pay := float64(taken) * price
			if buyer.Cash < pay {
				e.putInventory(s.FirmID, s.CityID, s.ItemID, taken)
				bi++
				continue
			}
			buyer.Cash -= pay
			seller.Cash += pay
			e.putInventory(b.FirmID, b.CityID, b.ItemID, taken)
			s.Filled += taken
			b.Filled += taken
			*log = append(*log, fmt.Sprintf("Биржа: %d %s @ %.1f", taken, k.item, price))
			if s.Qty-s.Filled <= 0 {
				si++
			}
			if b.Qty-b.Filled <= 0 {
				bi++
			}
		}
	}
}

func (e *Engine) takeInventory(firmID, cityID, itemID string, qty int) int {
	left := qty
	for _, b := range e.World.Buildings {
		if b.FirmID != firmID || b.CityID != cityID {
			continue
		}
		have := b.Inventory[itemID]
		take := min(have, left)
		b.Inventory[itemID] -= take
		left -= take
		if left == 0 {
			break
		}
	}
	return qty - left
}

func (e *Engine) putInventory(firmID, cityID, itemID string, qty int) {
	for _, b := range e.World.Buildings {
		if b.FirmID == firmID && b.CityID == cityID {
			b.Inventory[itemID] += qty
			return
		}
	}
	for _, b := range e.World.Buildings {
		if b.FirmID == firmID {
			b.Inventory[itemID] += qty
			return
		}
	}
}

func (e *Engine) runRetail(log *[]string) {
	// Спрос привязан к жилым клеткам; магазин берёт только зону охвата (Radius).
	for _, city := range e.World.Cities {
		type demandCell struct {
			x, y int
			food, home float64
		}
		var cells []demandCell
		resCount := 0
		for _, p := range city.Parcels {
			if p.Zone != "residential" && p.Zone != "center" {
				continue
			}
			resCount++
			cells = append(cells, demandCell{p.X, p.Y, 0, 0})
		}
		if resCount == 0 {
			continue
		}
		foodShare := (city.DemandSeg["food"] + float64(city.Population)*0.02/24) / float64(resCount)
		homeShare := (city.DemandSeg["home"] + float64(city.Population)*0.005/24) / float64(resCount)
		for i := range cells {
			cells[i].food = foodShare
			cells[i].home = homeShare
		}

		var cityShops []*model.ShopLease
		for _, s := range e.World.Shops {
			if s.CityID == city.ID {
				cityShops = append(cityShops, s)
			}
		}
		for _, shop := range cityShops {
			if shop.Stock <= 0 {
				continue
			}
			item, ok := e.Catalog.ItemByID[shop.ItemID]
			if !ok {
				continue
			}
			firm := e.World.Firms[shop.FirmID]
			radius := shop.Radius
			if radius <= 0 {
				radius = 2
			}
			brand := firm.BrandAwareness[item.Category]
			quality := 3.0
			for _, b := range e.World.Buildings {
				if b.FirmID == shop.FirmID {
					quality = b.Quality
					break
				}
			}
			sold := 0
			for i := range cells {
				c := &cells[i]
				dist := abs(c.x-shop.X) + abs(c.y-shop.Y)
				if dist > radius {
					continue // дальняя часть города просто не едет
				}
				local := c.food
				if item.Category == "home" {
					local = c.home
				} else if item.Category != "food" {
					local = (c.food + c.home) * 0.25
				}
				if local <= 0 {
					continue
				}
				distFactor := 1.0 - float64(dist)/float64(radius+1)
				attract := (quality/5.0)*0.35 + (1.0-clamp(shop.Price/300.0, 0, 1))*0.3 + brand*0.15 + distFactor*0.2
				take := int(math.Min(float64(shop.Stock-sold), local*attract*0.5))
				if take <= 0 {
					continue
				}
				if item.Category == "home" {
					c.home = maxf(0, c.home-float64(take))
				} else {
					c.food = maxf(0, c.food-float64(take))
				}
				sold += take
			}
			if sold <= 0 {
				continue
			}
			revenue := float64(sold) * shop.Price
			rentShare := revenue * shop.RevenueShare
			firm.Cash += revenue - rentShare
			shop.Stock -= sold
			city.DemandSeg[item.Category] = maxf(0, city.DemandSeg[item.Category]-float64(sold))
			firm.BrandAwareness[item.Category] = clamp(brand+0.005, 0, 1)
			*log = append(*log, fmt.Sprintf("Розница %s r=%d: −%d %s, +%.0f₽", firm.Name, radius, sold, shop.ItemID, revenue-rentShare))
		}
	}
}

func (e *Engine) payRentsAndWages(log *[]string) {
	for _, shop := range e.World.Shops {
		firm := e.World.Firms[shop.FirmID]
		if firm == nil {
			continue
		}
		firm.Cash -= shop.FixedRent / 24 // почасовая доля месячной/суточной аренды
	}
	for _, city := range e.World.Cities {
		// ставки за тик-час (~1/24 месячного фонда)
		wage := float64(city.LaborUsed.Worker)*12 + float64(city.LaborUsed.Engineer)*24 + float64(city.LaborUsed.Manager)*36
		// split wage across firms with buildings in city proportional to buildings
		counts := map[string]int{}
		total := 0
		for _, b := range e.World.Buildings {
			if b.CityID == city.ID {
				counts[b.FirmID]++
				total++
			}
		}
		if total == 0 || wage == 0 {
			continue
		}
		for fid, n := range counts {
			share := wage * float64(n) / float64(total)
			e.World.Firms[fid].Cash -= share
		}
	}
	// credit interest
	for _, f := range e.World.Firms {
		if f.Credit > 0 {
			interest := f.Credit * 0.01
			f.Cash -= interest
			payment := math.Min(f.Cash*0.05, f.Credit*0.05)
			if payment > 0 {
				f.Cash -= payment
				f.Credit -= payment
			}
		}
	}
	_ = log
}

func (e *Engine) aiAct(log *[]string) {
	for _, f := range e.World.Firms {
		if !f.IsAI {
			continue
		}
		// list excess finished goods on market / restock shops
		for _, shop := range e.World.Shops {
			if shop.FirmID != f.ID || shop.Stock >= 15 {
				continue
			}
			need := 15 - shop.Stock
			taken := e.takeInventory(f.ID, shop.CityID, shop.ItemID, need)
			shop.Stock += taken
		}
		for _, b := range e.World.Buildings {
			if b.FirmID != f.ID {
				continue
			}
			for item, qty := range b.Inventory {
				idef := e.Catalog.ItemByID[item]
				if idef.Tier != "finished" && idef.Tier != "intermediate" {
					continue
				}
				if qty < 12 {
					continue
				}
				sellQty := qty / 3
				b.Inventory[item] -= sellQty
				id := uuid.NewString()
				price := 30.0
				if idef.Tier == "finished" {
					price = 80
				}
				e.World.Orders[id] = &model.MarketOrder{
					ID: id, FirmID: f.ID, CityID: b.CityID, ItemID: item, Side: "sell",
					Qty: sellQty, Price: price, Quality: b.Quality,
				}
			}
		}
		// research if can
		for _, r := range e.Catalog.Recipes {
			if r.ResearchCost > 0 && !f.Researched[r.ID] && f.Cash > float64(r.ResearchCost)*2 {
				f.Cash -= float64(r.ResearchCost)
				f.Researched[r.ID] = true
				*log = append(*log, fmt.Sprintf("AI %s исследовал %s", f.Name, r.Name))
				break
			}
		}
	}
}

func (e *Engine) growCities(log *[]string) {
	for _, city := range e.World.Cities {
		jobs := city.LaborUsed.Worker + city.LaborUsed.Engineer + city.LaborUsed.Manager
		// почасовой рост населения (раньше было за «неделю»)
		if jobs > 0 && e.World.Tick%6 == 0 {
			city.Population += max(1, jobs/40)
		}
		resCount := 0
		for _, p := range city.Parcels {
			if p.Zone == "residential" || p.Zone == "center" {
				resCount++
			}
		}
		city.HousingOcc = float64(city.Population) / float64(max(1, resCount*400))
		// расширение не чаще раза в сутки игрового месяца и с потолком радиуса
		if city.HousingOcc > 1.05 && e.World.Tick%24 == 0 && city.Radius < 10 {
			e.expandCity(city)
			*log = append(*log, fmt.Sprintf("%s расширяется (радиус %d)", city.ID, city.Radius))
		}
		if e.World.Tick%12 == 0 {
			for i := range city.Parcels {
				p := &city.Parcels[i]
				dist := abs(p.X) + abs(p.Y)
				if dist <= 2 && p.Zone != "center" && p.Slots == 0 && p.Building == "" && p.Zone != "empty" {
					p.Zone = "residential"
					p.Slots = 1
					p.Traffic = 0.65
				}
			}
		}
		if jobs > city.LaborPool.Worker/2 && e.World.Tick%24 == 0 {
			city.LaborPool.Worker += 5
			city.LaborPool.Engineer += 1
		}
	}
}

func (e *Engine) expandCity(city *model.CityState) {
	city.Radius++
	r := city.Radius
	existing := map[[2]int]bool{}
	for _, p := range city.Parcels {
		existing[[2]int{p.X, p.Y}] = true
	}
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if abs(x)+abs(y) != r {
				continue
			}
			if existing[[2]int{x, y}] {
				continue
			}
			zone := "empty"
			if x > 0 {
				zone = "industrial"
			} else {
				zone = "rural"
			}
			city.Parcels = append(city.Parcels, model.Parcel{X: x, Y: y, Zone: zone, Traffic: 0.25})
		}
	}
}

func (e *Engine) updateRankings() {
	for _, f := range e.World.Firms {
		assets := f.Cash - f.Credit
		for _, b := range e.World.Buildings {
			if b.FirmID != f.ID {
				continue
			}
			if def, ok := e.Catalog.BuildingByID[b.Type]; ok {
				assets += float64(def.BuildCost) * 0.7
			}
			for _, qty := range b.Inventory {
				assets += float64(qty) * 10
			}
		}
		f.CapitalSnapshot = assets
	}
}

func (e *Engine) monthlyEvent(log *[]string) {
	msgs := []string{
		"Администрация РБ: субсидия переработчикам молока (+маленький бонус спроса на food).",
		"Ярмарка в Уфе: рост brand awareness у розницы.",
		"Сезонная сводка: проверьте вишню и урожай.",
	}
	msg := msgs[e.World.Tick%len(msgs)]
	*log = append(*log, msg)
	e.World.Events = append(e.World.Events, "Событие месяца: "+msg)
	for _, city := range e.World.Cities {
		city.DemandSeg["food"] *= 1.05
	}
}

func (e *Engine) firmBuildingsInCity(firmID, cityID string) []*model.BuildingInst {
	var out []*model.BuildingInst
	for _, b := range e.World.Buildings {
		if b.FirmID == firmID && b.CityID == cityID {
			out = append(out, b)
		}
	}
	return out
}

func (e *Engine) routeInfo(from, to string) (int, float64) {
	if from == to {
		return 0, 0
	}
	for _, r := range e.Catalog.Routes {
		if (r.From == from && r.To == to) || (r.From == to && r.To == from) {
			return r.Ticks, r.CostPerUnit
		}
	}
	return 2, 4
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
