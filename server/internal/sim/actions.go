package sim

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

func (e *Engine) Build(firmID, cityID, buildingType, recipeID string, x, y int) (*model.BuildingInst, error) {
	if recipeID == "" {
		// pick first recipe for type
		for _, r := range e.Catalog.Recipes {
			if r.BuildingType == buildingType {
				recipeID = r.ID
				break
			}
		}
	}
	id, err := e.placeBuilding(firmID, cityID, buildingType, recipeID, x, y, false)
	if err != nil {
		return nil, err
	}
	return e.World.Buildings[id], nil
}

func (e *Engine) SetRecipe(firmID, buildingID, recipeID string) error {
	b := e.World.Buildings[buildingID]
	if b == nil || b.FirmID != firmID {
		return fmt.Errorf("здание не найдено")
	}
	r, ok := e.Catalog.RecipeByID[recipeID]
	if !ok || r.BuildingType != b.Type {
		return fmt.Errorf("рецепт не подходит зданию")
	}
	firm := e.World.Firms[firmID]
	if r.ResearchCost > 0 && !firm.Researched[recipeID] {
		return fmt.Errorf("рецепт не исследован")
	}
	if b.RecipeID != recipeID {
		b.RecipeID = recipeID
		b.RetoolWeeks = 1
	}
	return nil
}

func (e *Engine) RentShop(firmID, cityID string, x, y int, itemID string, price float64) (*model.ShopLease, error) {
	city := e.World.Cities[cityID]
	if city == nil {
		return nil, fmt.Errorf("город не найден")
	}
	firm := e.World.Firms[firmID]
	if firm == nil {
		return nil, fmt.Errorf("фирма не найдена")
	}
	var parcel *model.Parcel
	for i := range city.Parcels {
		p := &city.Parcels[i]
		if p.X == x && p.Y == y {
			parcel = p
			break
		}
	}
	if parcel == nil || parcel.Slots <= 0 {
		return nil, fmt.Errorf("нет коммерческого слота")
	}
	if parcel.RentedBy != "" {
		return nil, fmt.Errorf("слот уже арендован")
	}
	if _, ok := e.Catalog.ItemByID[itemID]; !ok {
		return nil, fmt.Errorf("неизвестный товар")
	}
	if price <= 0 {
		price = 50
	}
	// Магазин тоже должен быть на дороге / с подъездом
	if _, _, err := e.ensureRoadAccess(firm, city, x, y, false); err != nil {
		return nil, err
	}
	id := uuid.NewString()
	shop := &model.ShopLease{
		ID: id, FirmID: firmID, CityID: cityID, X: x, Y: y,
		ItemID: itemID, Price: price, Stock: 0, FixedRent: 400 + parcel.Traffic*400, RevenueShare: 0.08,
		Radius: shopRadiusForZone(parcel.Zone, parcel.Slots),
	}
	e.World.Shops[id] = shop
	parcel.RentedBy = firmID
	parcel.HasRoad = true
	return shop, nil
}

func shopRadiusForZone(zone string, slots int) int {
	if zone == "center" || slots >= 3 {
		return 4 // якорь ТЦ / центр
	}
	if slots >= 2 {
		return 3
	}
	return 2 // витрина в доме — только соседние кварталы
}

func (e *Engine) StockShop(firmID, shopID string, qty int) error {
	shop := e.World.Shops[shopID]
	if shop == nil || shop.FirmID != firmID {
		return fmt.Errorf("магазин не найден")
	}
	taken := e.takeInventory(firmID, shop.CityID, shop.ItemID, qty)
	if taken == 0 {
		return fmt.Errorf("нет товара на складе в городе")
	}
	shop.Stock += taken
	return nil
}

func (e *Engine) PlaceOrder(firmID, cityID, itemID, side string, qty int, price float64) (*model.MarketOrder, error) {
	side = strings.ToLower(side)
	if side != "buy" && side != "sell" {
		return nil, fmt.Errorf("side buy|sell")
	}
	if qty <= 0 || price <= 0 {
		return nil, fmt.Errorf("qty/price")
	}
	if side == "sell" {
		taken := e.takeInventory(firmID, cityID, itemID, qty)
		if taken < qty {
			e.putInventory(firmID, cityID, itemID, taken)
			return nil, fmt.Errorf("недостаточно товара")
		}
	}
	id := uuid.NewString()
	o := &model.MarketOrder{ID: id, FirmID: firmID, CityID: cityID, ItemID: itemID, Side: side, Qty: qty, Price: price, Quality: 3}
	e.World.Orders[id] = o
	return o, nil
}

func (e *Engine) CreateContract(sellerID, buyerID, itemID, from, to string, qty int, price float64, ticks int, ownFleet bool) (*model.Contract, error) {
	if ticks <= 0 {
		ticks = 24 // 1 game month
	}
	id := uuid.NewString()
	c := &model.Contract{
		ID: id, SellerID: sellerID, BuyerID: buyerID, ItemID: itemID,
		CityFrom: from, CityTo: to, QtyPerTick: qty, Price: price, Quality: 3,
		TicksLeft: ticks, UseOwnFleet: ownFleet,
	}
	e.World.Contracts[id] = c
	return c, nil
}

func (e *Engine) Research(firmID, recipeID string) error {
	firm := e.World.Firms[firmID]
	r, ok := e.Catalog.RecipeByID[recipeID]
	if firm == nil || !ok {
		return fmt.Errorf("не найдено")
	}
	if firm.Researched[recipeID] {
		return nil
	}
	if firm.Cash < float64(r.ResearchCost) {
		return fmt.Errorf("недостаточно средств")
	}
	firm.Cash -= float64(r.ResearchCost)
	firm.Researched[recipeID] = true
	return nil
}

func (e *Engine) BuyTruck(firmID string) error {
	firm := e.World.Firms[firmID]
	if firm == nil {
		return fmt.Errorf("фирма не найдена")
	}
	cost := 25000.0
	if firm.Cash < cost {
		return fmt.Errorf("недостаточно средств")
	}
	firm.Cash -= cost
	firm.Trucks++
	return nil
}

func (e *Engine) TakeCredit(firmID string, amount float64) error {
	firm := e.World.Firms[firmID]
	if firm == nil {
		return fmt.Errorf("фирма не найдена")
	}
	if amount <= 0 || amount > 200000 {
		return fmt.Errorf("сумма кредита 1..200000")
	}
	// simple collateral: buildings
	collateral := 0.0
	for _, b := range e.World.Buildings {
		if b.FirmID == firmID {
			if def, ok := e.Catalog.BuildingByID[b.Type]; ok {
				collateral += float64(def.BuildCost) * 0.5
			}
		}
	}
	if firm.Credit+amount > collateral+50000 {
		return fmt.Errorf("не хватает залога")
	}
	firm.Credit += amount
	firm.Cash += amount
	return nil
}

func (e *Engine) BuyParcel(firmID, cityID string, x, y int) error {
	city := e.World.Cities[cityID]
	firm := e.World.Firms[firmID]
	if city == nil || firm == nil {
		return fmt.Errorf("не найдено")
	}
	for i := range city.Parcels {
		p := &city.Parcels[i]
		if p.X == x && p.Y == y {
			if p.OwnerID != "" {
				return fmt.Errorf("уже занято")
			}
			price := 8000.0 + float64(abs(x)+abs(y))*500
			if firm.Cash < price {
				return fmt.Errorf("недостаточно средств")
			}
			firm.Cash -= price
			p.OwnerID = firmID
			return nil
		}
	}
	return fmt.Errorf("участок не найден")
}

func (e *Engine) DevelopHousing(firmID, cityID string, x, y int) error {
	city := e.World.Cities[cityID]
	firm := e.World.Firms[firmID]
	if city == nil || firm == nil {
		return fmt.Errorf("не найдено")
	}
	cost := 20000.0
	if firm.Cash < cost {
		return fmt.Errorf("недостаточно средств")
	}
	for i := range city.Parcels {
		p := &city.Parcels[i]
		if p.X == x && p.Y == y {
			if p.OwnerID != firmID {
				return fmt.Errorf("нужна собственность на участок")
			}
			if p.Building != "" {
				return fmt.Errorf("занято зданием")
			}
			if _, _, err := e.ensureRoadAccess(firm, city, x, y, false); err != nil {
				return err
			}
			if firm.Cash < cost {
				return fmt.Errorf("недостаточно средств")
			}
			firm.Cash -= cost
			p.Zone = "residential"
			p.Slots = 2
			p.Traffic = 0.75
			p.HasRoad = true
			return nil
		}
	}
	return fmt.Errorf("участок не найден")
}

func (e *Engine) Advertise(firmID, category string, budget float64) error {
	firm := e.World.Firms[firmID]
	if firm == nil || budget <= 0 {
		return fmt.Errorf("ошибка")
	}
	if firm.Cash < budget {
		return fmt.Errorf("недостаточно средств")
	}
	firm.Cash -= budget
	gain := budget / 50000.0
	firm.BrandAwareness[category] = clamp(firm.BrandAwareness[category]+gain, 0, 1)
	return nil
}
