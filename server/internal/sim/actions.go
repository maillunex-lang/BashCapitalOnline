package sim

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

func (e *Engine) Build(firmID, cityID, buildingType, recipeID string, x, y int) (*model.BuildingInst, error) {
	if recipeID == "" {
		for _, r := range e.Catalog.Recipes {
			if r.BuildingType == buildingType {
				recipeID = r.ID
				break
			}
		}
	}
	if recipeID != "" {
		if err := e.RecipeFitsBuilding(buildingType, recipeID); err != nil {
			return nil, err
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
	if err := e.RecipeFitsBuilding(b.Type, recipeID); err != nil {
		return err
	}
	r := e.Catalog.RecipeByID[recipeID]
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

func (e *Engine) ConfigureBuilding(firmID, buildingID, recipeID string, autoProcure bool, defaultMode, preferFirm string, sellExternal bool, inputMap map[string]any) (*model.BuildingInst, error) {
	b := e.World.Buildings[buildingID]
	if b == nil || b.FirmID != firmID {
		return nil, fmt.Errorf("здание не найдено")
	}
	if recipeID != "" {
		if err := e.SetRecipe(firmID, buildingID, recipeID); err != nil {
			return nil, err
		}
	}
	b.AutoProcure = autoProcure
	b.SellExternal = sellExternal
	b.DefaultInput = model.InputSource{Mode: normalizeInputMode(defaultMode), FirmID: preferFirm}
	if b.DefaultInput.Mode == "firm" && preferFirm == "" {
		return nil, fmt.Errorf("для закупки у фирмы укажите firm_id")
	}
	if inputMap != nil {
		b.InputSources = map[string]model.InputSource{}
		for itemID, raw := range inputMap {
			m, _ := raw.(map[string]any)
			mode, _ := m["mode"].(string)
			fid, _ := m["firm_id"].(string)
			b.InputSources[itemID] = model.InputSource{Mode: normalizeInputMode(mode), FirmID: fid}
		}
	}
	return b, nil
}

func (e *Engine) RentShop(firmID, cityID string, x, y int, shopType, itemID string, price float64) (*model.ShopLease, error) {
	city := e.World.Cities[cityID]
	if city == nil {
		return nil, fmt.Errorf("город не найден")
	}
	firm := e.World.Firms[firmID]
	if firm == nil {
		return nil, fmt.Errorf("фирма не найдена")
	}
	if shopType == "" {
		shopType = e.defaultShopTypeForItem(itemID)
	}
	if err := e.ShopAllowsItem(shopType, itemID); err != nil {
		return nil, err
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
	if price <= 0 {
		price = 50
	}
	var err error
	price, err = e.ClampPrice(itemID, price)
	if err != nil {
		return nil, err
	}
	if _, _, err := e.ensureRoadAccess(firm, city, x, y, false); err != nil {
		return nil, err
	}
	sdef := e.Catalog.ShopByID[shopType]
	rent := (400 + parcel.Traffic*400) * sdef.RentMult
	if rent <= 0 {
		rent = 400
	}
	id := uuid.NewString()
	shop := &model.ShopLease{
		ID: id, FirmID: firmID, CityID: cityID, X: x, Y: y,
		ShopType: shopType, ItemID: itemID, Price: price, Stock: 0,
		FixedRent: rent, RevenueShare: 0.08,
		Radius: shopRadiusForZone(parcel.Zone, parcel.Slots),
		StockSource: "own", AutoRestock: true,
	}
	e.World.Shops[id] = shop
	parcel.RentedBy = firmID
	parcel.HasRoad = true
	e.markDirty(cityID, x, y, "shop")
	return shop, nil
}

func (e *Engine) defaultShopTypeForItem(itemID string) string {
	prefer := []string{}
	if it, ok := e.Catalog.ItemByID[itemID]; ok {
		switch it.Category {
		case "fuel":
			prefer = []string{"fuel_station"}
		case "food", "dairy", "grain", "veg", "fruit":
			prefer = []string{"grocery", "bakery_shop", "dairy_shop"}
		case "apparel", "textile":
			prefer = []string{"apparel_shop", "fashion_boutique"}
		case "home", "timber":
			prefer = []string{"furniture_shop", "home_decor", "hardware"}
		case "metal", "mineral", "packaging":
			prefer = []string{"hardware", "building_market"}
		case "tech":
			prefer = []string{"electronics_shop"}
		}
	}
	for _, id := range prefer {
		if e.ShopAllowsItem(id, itemID) == nil {
			return id
		}
	}
	for _, s := range e.Catalog.Shops {
		if e.ShopAllowsItem(s.ID, itemID) == nil {
			return s.ID
		}
	}
	return "department"
}

func (e *Engine) ConfigureShop(firmID, shopID, shopType, itemID string, price float64, stockSource string, autoRestock bool) (*model.ShopLease, error) {
	shop := e.World.Shops[shopID]
	if shop == nil || shop.FirmID != firmID {
		return nil, fmt.Errorf("магазин не найден")
	}
	if shopType == "" {
		shopType = shop.ShopType
	}
	if itemID == "" {
		itemID = shop.ItemID
	}
	if err := e.ShopAllowsItem(shopType, itemID); err != nil {
		return nil, err
	}
	if price > 0 {
		p, err := e.ClampPrice(itemID, price)
		if err != nil {
			return nil, err
		}
		shop.Price = p
	}
	shop.ShopType = shopType
	shop.ItemID = itemID
	shop.StockSource = normalizeInputMode(stockSource)
	if shop.StockSource == "firm" {
		shop.StockSource = "market"
	}
	shop.AutoRestock = autoRestock
	return shop, nil
}

func shopRadiusForZone(zone string, slots int) int {
	if zone == "center" || slots >= 3 {
		return 4
	}
	if slots >= 2 {
		return 3
	}
	return 2
}

func (e *Engine) StockShop(firmID, shopID string, qty int) error {
	shop := e.World.Shops[shopID]
	if shop == nil || shop.FirmID != firmID {
		return fmt.Errorf("магазин не найден")
	}
	if err := e.ShopAllowsItem(shop.ShopType, shop.ItemID); err != nil {
		return err
	}
	taken := 0
	switch shop.StockSource {
	case "city":
		// temporary building-like buy into firm then to shop
		for _, b := range e.World.Buildings {
			if b.FirmID == firmID {
				taken = e.buyCityIntoBuilding(b, shop.ItemID, qty, nil)
				if taken > 0 {
					b.Inventory[shop.ItemID] -= taken
					shop.Stock += taken
				}
				break
			}
		}
	default:
		taken = e.takeInventory(firmID, shop.CityID, shop.ItemID, qty)
		shop.Stock += taken
	}
	if taken == 0 {
		return fmt.Errorf("нет товара для витрины (проверьте источник и склад)")
	}
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
	var err error
	price, err = e.ClampPrice(itemID, price)
	if err != nil {
		return nil, err
	}
	if side == "sell" {
		if !e.firmAllowsExternalSale(firmID, itemID) {
			return nil, fmt.Errorf("продажа другим игрокам запрещена на ваших заводах (режим «только себе»)")
		}
		taken := e.takeInventoryRespectingSales(firmID, cityID, itemID, qty, true)
		if taken < qty {
			e.putInventory(firmID, cityID, itemID, taken)
			return nil, fmt.Errorf("недостаточно товара для внешней продажи")
		}
		// hold inventory: put back as reserved via order — restore for matching to take again
		e.putInventory(firmID, cityID, itemID, taken)
	}
	id := uuid.NewString()
	o := &model.MarketOrder{ID: id, FirmID: firmID, CityID: cityID, ItemID: itemID, Side: side, Qty: qty, Price: price, Quality: 3}
	e.World.Orders[id] = o
	return o, nil
}

func (e *Engine) CreateContract(sellerID, buyerID, itemID, from, to string, qty int, price float64, ticks int, ownFleet bool) (*model.Contract, error) {
	if ticks <= 0 {
		ticks = 24
	}
	if sellerID != buyerID && !e.firmAllowsExternalSale(sellerID, itemID) {
		return nil, fmt.Errorf("продавец запретил продажи другим фирмам")
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
	cost := float64(r.ResearchCost)
	if firm.Cash < cost {
		return fmt.Errorf("недостаточно средств на R&D")
	}
	firm.Cash -= cost
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
	if firm == nil || amount <= 0 || amount > 500000 {
		return fmt.Errorf("кредит 1…500000")
	}
	firm.Cash += amount
	firm.Credit += amount * 1.08
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
				return fmt.Errorf("участок занят")
			}
			cost := 5000.0
			if firm.Cash < cost {
				return fmt.Errorf("недостаточно средств")
			}
			firm.Cash -= cost
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
	cost := 18000.0
	if firm.Cash < cost {
		return fmt.Errorf("недостаточно средств")
	}
	for i := range city.Parcels {
		p := &city.Parcels[i]
		if p.X == x && p.Y == y {
			if p.Building != "" {
				return fmt.Errorf("занято")
			}
			firm.Cash -= cost
			p.Zone = "residential"
			p.Slots = 1
			p.OwnerID = firmID
			city.Population += 40
			return nil
		}
	}
	return fmt.Errorf("участок не найден")
}

func (e *Engine) Advertise(firmID, category string, budget float64) error {
	firm := e.World.Firms[firmID]
	if firm == nil || budget <= 0 {
		return fmt.Errorf("бюджет")
	}
	if firm.Cash < budget {
		return fmt.Errorf("недостаточно средств")
	}
	firm.Cash -= budget
	gain := budget / 50000.0
	firm.BrandAwareness[category] = clamp(firm.BrandAwareness[category]+gain, 0, 1)
	return nil
}
