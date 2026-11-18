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
	if e.Paused {
		e.World.LastTickLog = []string{"Мир на паузе (админ)"}
		return
	}
	e.clearDirty()
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
	e.payRentsAndCredit()
	e.restockAllAutoShops(&log)
	e.aiAct(&log)
	e.refreshCityMarket()
	e.runAIChat(&log)
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
	e.autosave()
}

func (e *Engine) autosave() {
	if e.Store == nil {
		return
	}
	if err := e.Store.SaveWorld(e.World); err != nil {
		e.World.Events = append(e.World.Events, "autosave error: "+err.Error())
	}
}

// Persist saves world immediately (register/actions) so progress survives restart before the next tick.
func (e *Engine) Persist() {
	e.autosave()
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
		own := c.UseOwnFleet && buyer.Trucks > 0
		ship := e.createShipment(c.BuyerID, c.ItemID, c.CityFrom, c.CityTo, moved, avgQ, own)
		ticks := 1
		dist := 0.0
		if ship != nil {
			ticks = ship.TicksLeft
			dist = ship.Distance
		}
		c.TicksLeft--
		*log = append(*log, fmt.Sprintf("Контракт %s: %d %s → %s · %d тик(ов), дист. %.0f", c.ID[:8], moved, c.ItemID, c.CityTo, ticks, dist))
	}
}

func (e *Engine) advanceShipments(log *[]string) {
	for id, s := range e.World.Shipments {
		s.TicksLeft--
		if s.TicksLeft > 0 {
			continue
		}
		if s.ToBuildingID != "" {
			b := e.World.Buildings[s.ToBuildingID]
			if b == nil || b.FirmID != s.FirmID || b.CityID != s.ToCity {
				// Keep the cargo in transit if the recipient is unavailable in a save.
				s.TicksLeft = 1
				continue
			}
			deliverCargo(b, s)
			*log = append(*log, fmt.Sprintf("Доставка %d %s в %s (дист. %.0f)", s.Qty, s.ItemID, s.ToCity, s.Distance))
			delete(e.World.Shipments, id)
			continue
		}
		// deliver into any warehouse/building of firm in destination
		delivered := false
		for _, b := range e.World.Buildings {
			if b.FirmID == s.FirmID && b.CityID == s.ToCity {
				deliverCargo(b, s)
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
		*log = append(*log, fmt.Sprintf("Доставка %d %s в %s (дист. %.0f)", s.Qty, s.ItemID, s.ToCity, s.Distance))
		delete(e.World.Shipments, id)
	}
}

func deliverCargo(b *model.BuildingInst, s *model.Shipment) {
	if b.Inventory == nil {
		b.Inventory = map[string]int{}
	}
	b.Inventory[s.ItemID] += s.Qty
	if s.Quality > 0 {
		b.Quality = (b.Quality + s.Quality) / 2
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
		if !hasLabor(city, recipe.Labor) || e.RecipeFitsBuilding(b.Type, recipe.ID) != nil {
			continue
		}
		extract := isExtractRecipe(recipe)
		utility := isUtilityExtract(recipe)
		ww, we, wm := e.wageRates()
		cost := e.opsCost(recipe, extract || utility) +
			float64(recipe.Labor.Worker)*ww + float64(recipe.Labor.Engineer)*we + float64(recipe.Labor.Manager)*wm
		if firm.Cash < cost {
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
		quality := clamp(b.Quality*0.7+2.0*0.3, 1, 5)
		switch {
		case utility:
			quality = clamp(b.Quality*0.5+2.5*0.5, 1, 5)
		case extract:
			qty, depositQuality := e.consumeDeposit(city, b.X, b.Y, outQty, recipe.Output.ItemID, log)
			if qty == 0 {
				continue
			}
			outQty = qty
			quality = clamp(depositQuality*0.7+b.Quality*0.3, 1, 5)
		default:
			e.procureInputs(b, recipe, log)
			if firm.Cash < cost || !hasInputs(b, recipe) {
				continue
			}
			for _, in := range recipe.Inputs {
				b.Inventory[in.ItemID] -= in.Qty
			}
			quality = clamp(b.Quality*0.7+2.0*0.3, 1, 5)
		}
		city.LaborUsed.Worker += recipe.Labor.Worker
		city.LaborUsed.Engineer += recipe.Labor.Engineer
		city.LaborUsed.Manager += recipe.Labor.Manager
		firm.Cash -= cost
		b.Inventory[recipe.Output.ItemID] += outQty
		b.Quality = quality
	}
}

func hasLabor(city *model.CityState, need model.Labor) bool {
	return city.LaborUsed.Worker+need.Worker <= city.LaborPool.Worker &&
		city.LaborUsed.Engineer+need.Engineer <= city.LaborPool.Engineer &&
		city.LaborUsed.Manager+need.Manager <= city.LaborPool.Manager
}

func hasInputs(b *model.BuildingInst, recipe model.RecipeDef) bool {
	for _, in := range recipe.Inputs {
		if b.Inventory[in.ItemID] < in.Qty {
			return false
		}
	}
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
			external := s.FirmID != b.FirmID
			taken := e.takeInventoryRespectingSales(s.FirmID, s.CityID, s.ItemID, qty, external)
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
			own := buyer.Trucks > 0
			// Доставка всегда по расстоянию, минимум 1 тик — даже в том же городе.
			ship := e.createShipment(b.FirmID, b.ItemID, s.CityID, b.CityID, taken, s.Quality, own)
			ticks, dist := 1, 0.0
			if ship != nil {
				ticks = ship.TicksLeft
				dist = ship.Distance
			}
			s.Filled += taken
			b.Filled += taken
			*log = append(*log, fmt.Sprintf("Биржа: %d %s @ %.1f · доставка %d тик (%.0f)", taken, k.item, price, ticks, dist))
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
	// CapLab-ish demand: residential cells + class mix + quality/brand/corridor price.
	e.decayBrands()
	cfg := e.demandCfg()

	for _, city := range e.World.Cities {
		type demandCell struct {
			x, y int
			pool map[string]float64
		}
		var cells []demandCell
		resCount := 0
		for _, p := range city.Parcels {
			if p.Zone != "residential" && p.Zone != "center" {
				continue
			}
			resCount++
			cells = append(cells, demandCell{p.X, p.Y, map[string]float64{}})
		}
		if resCount == 0 {
			continue
		}
		cats := []string{"food", "home", "apparel", "fuel", "tech", "textile"}
		cityBudget := map[string]float64{}
		for _, cat := range cats {
			pool := e.categoryDemandPool(city, cat)
			cityBudget[cat] = pool
			share := pool / float64(resCount)
			for i := range cells {
				cells[i].pool[cat] = share
			}
		}

		var cityShops []*model.ShopLease
		for _, s := range e.World.Shops {
			if s.CityID == city.ID {
				cityShops = append(cityShops, s)
			}
		}
		// Higher quality first so premium shelves capture class demand before junk.
		sort.SliceStable(cityShops, func(i, j int) bool {
			qi := e.shopShelfQuality(cityShops[i])
			qj := e.shopShelfQuality(cityShops[j])
			if qi != qj {
				return qi > qj
			}
			return cityShops[i].ID < cityShops[j].ID
		})

		for _, shop := range cityShops {
			item, ok := e.Catalog.ItemByID[shop.ItemID]
			if !ok {
				continue
			}
			firm := e.World.Firms[shop.FirmID]
			if firm == nil {
				continue
			}
			if shop.Stock <= 0 {
				*log = append(*log, fmt.Sprintf("Розница %s: stockout %s", firm.Name, shop.ItemID))
				continue
			}
			radius := shop.Radius
			if radius <= 0 {
				radius = 2
			}
			brand := 0.0
			if firm.BrandAwareness != nil {
				brand = firm.BrandAwareness[item.Category]
			}
			quality := e.shopShelfQuality(shop)
			mid := e.corridorMid(shop.ItemID)
			sold := 0
			tooExpensiveHits := 0
			maxSold := 8
			if shop.Stock < 20 {
				maxSold = 5
			}
			cat := item.Category
			if cat == "" {
				cat = "food"
			}
			for i := range cells {
				if sold >= maxSold || cityBudget[cat] <= 0 {
					break
				}
				c := &cells[i]
				dist := abs(c.x-shop.X) + abs(c.y-shop.Y)
				if dist > radius {
					continue
				}
				local := c.pool[cat]
				if local <= 0 {
					local = (c.pool["food"] + c.pool["home"]) * 0.2
				}
				if local <= 0 {
					continue
				}
				// Contested city budget: better shops (sorted earlier) drain demand first.
				local = math.Min(local, cityBudget[cat])
				distFactor := 1.0 - float64(dist)/float64(radius+1)
				attract, tooExp := e.retailAttract(quality, brand, shop.Price, mid, distFactor)
				if tooExp {
					tooExpensiveHits++
					continue
				}
				take := softTake(local, attract, shop.Stock-sold, maxSold-sold)
				if take <= 0 {
					continue
				}
				c.pool[cat] = maxf(0, c.pool[cat]-float64(take))
				cityBudget[cat] = maxf(0, cityBudget[cat]-float64(take))
				sold += take
			}
			if sold <= 0 {
				if tooExpensiveHits > 0 {
					*log = append(*log, fmt.Sprintf("Розница %s: too_expensive %s @%.0f (mid %.0f)", firm.Name, shop.ItemID, shop.Price, mid))
				}
				continue
			}
			revenue := float64(sold) * shop.Price
			rentShare := revenue * shop.RevenueShare
			firm.Cash += revenue - rentShare
			shop.Stock -= sold
			city.DemandSeg[cat] = maxf(0, city.DemandSeg[cat]-float64(sold))
			if firm.BrandAwareness == nil {
				firm.BrandAwareness = map[string]float64{}
			}
			firm.BrandAwareness[cat] = clamp(brand+cfg.BrandGainOnSale, 0, 1)
			note := fmt.Sprintf("Розница %s q=%.1f: sold %d %s, +%.0f₽", firm.Name, quality, sold, shop.ItemID, revenue-rentShare)
			if shop.Stock == 0 {
				note += " · stockout"
			}
			*log = append(*log, note)
		}
	}
}

func (e *Engine) payRentsAndCredit() {
	for _, shop := range e.World.Shops {
		firm := e.World.Firms[shop.FirmID]
		if firm == nil {
			continue
		}
		firm.Cash -= shop.FixedRent / 24 // почасовая доля месячной/суточной аренды
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
}

// restockAllAutoShops fills AutoRestock shops for every firm (players + AI) from own inventory.
// AI may fall back to city wholesale; players only pull from their buildings unless StockSource=city.
func (e *Engine) restockAllAutoShops(log *[]string) {
	for _, shop := range e.World.Shops {
		if !shop.AutoRestock || shop.Stock >= 25 {
			continue
		}
		firm := e.World.Firms[shop.FirmID]
		if firm == nil {
			continue
		}
		if e.ShopAllowsItem(shop.ShopType, shop.ItemID) != nil {
			continue
		}
		need := 25 - shop.Stock
		taken := 0
		switch shop.StockSource {
		case "city":
			for _, b := range e.World.Buildings {
				if b.FirmID != firm.ID {
					continue
				}
				n := e.buyCityIntoBuilding(b, shop.ItemID, need, log)
				if n > 0 {
					b.Inventory[shop.ItemID] -= n
					taken += n
				}
				break
			}
		default:
			taken = e.takeInventory(firm.ID, shop.CityID, shop.ItemID, need)
			if taken < need && firm.IsAI {
				for _, b := range e.World.Buildings {
					if b.FirmID != firm.ID {
						continue
					}
					n := e.buyCityIntoBuilding(b, shop.ItemID, need-taken, log)
					if n > 0 {
						b.Inventory[shop.ItemID] -= n
						taken += n
					}
					if taken >= need {
						break
					}
				}
			}
		}
		if taken > 0 {
			shop.Stock += taken
			*log = append(*log, fmt.Sprintf("%s: автопополнение %s +%d", firm.Name, shop.ItemID, taken))
		}
	}
}

func (e *Engine) aiAct(log *[]string) {
	for _, f := range e.World.Firms {
		if !f.IsAI {
			continue
		}
		// Items reserved for own shopfronts — don't dump them on the exchange first.
		shopNeed := map[string]int{}
		for _, shop := range e.World.Shops {
			if shop.FirmID != f.ID {
				continue
			}
			if shop.Stock < 25 {
				shopNeed[shop.ItemID] += 25 - shop.Stock
			}
		}
		for _, b := range e.World.Buildings {
			if b.FirmID != f.ID || !b.SellExternal {
				continue
			}
			for item, qty := range b.Inventory {
				idef := e.Catalog.ItemByID[item]
				if idef.Tier != "finished" && idef.Tier != "intermediate" {
					continue
				}
				reserve := shopNeed[item]
				avail := qty - reserve
				if avail < 18 {
					continue
				}
				sellQty := avail / 3
				b.Inventory[item] -= sellQty
				id := uuid.NewString()
				price := 80.0
				if c, ok := e.Catalog.Economy.PriceCorridors[item]; ok {
					price = (c.Min + c.Max) / 2
				} else if idef.Tier != "finished" {
					price = 40
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

func (e *Engine) energyRate(extract bool) float64 {
	b := e.Catalog.Economy.Balance
	if extract {
		if b.EnergyCostExtract > 0 {
			return b.EnergyCostExtract
		}
		return 8
	}
	if b.EnergyCostFactory > 0 {
		return b.EnergyCostFactory
	}
	return 15
}

func (e *Engine) wageRates() (worker, engineer, manager float64) {
	w := e.Catalog.Economy.Balance.WagePerTick
	worker, engineer, manager = 12, 24, 36
	if w == nil {
		return
	}
	if v, ok := w["worker"]; ok && v > 0 {
		worker = v
	}
	if v, ok := w["engineer"]; ok && v > 0 {
		engineer = v
	}
	if v, ok := w["manager"]; ok && v > 0 {
		manager = v
	}
	return
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
