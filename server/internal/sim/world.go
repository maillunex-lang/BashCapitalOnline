package sim

import (
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

type Engine struct {
	Catalog *model.Catalog
	World   *model.World
}

func NewEngine(cat *model.Catalog) *Engine {
	e := &Engine{Catalog: cat, World: &model.World{
		Tick: 0, Month: 1, Year: 2026,
		Cities:    map[string]*model.CityState{},
		Firms:     map[string]*model.Firm{},
		Buildings: map[string]*model.BuildingInst{},
		Shops:     map[string]*model.ShopLease{},
		Orders:    map[string]*model.MarketOrder{},
		Contracts: map[string]*model.Contract{},
		Shipments: map[string]*model.Shipment{},
		Events:    []string{},
		Chat:      []model.ChatMessage{},
	}}
	e.seedCities()
	e.seedAI()
	e.initCityMarket()
	e.World.Events = append(e.World.Events, "Мир Башкортостана инициализирован. Активны Уфа и Белорецк · каталог как в Capitalism Lab.")
	return e
}

func (e *Engine) seedCities() {
	for _, def := range e.Catalog.Cities {
		if !def.Enabled {
			continue
		}
		cs := &model.CityState{
			ID:         def.ID,
			Population: def.StartPopulation,
			Radius:     def.StartRadius,
			LaborPool:  def.Labor,
			LaborUsed:  model.Labor{},
			Parcels:    generateParcels(def.StartRadius),
			DemandSeg: map[string]float64{
				"food":    float64(def.StartPopulation) * 0.08,
				"home":    float64(def.StartPopulation) * 0.02,
				"apparel": float64(def.StartPopulation) * 0.015,
				"fuel":    float64(def.StartPopulation) * 0.01,
				"tech":    float64(def.StartPopulation) * 0.008,
				"textile": float64(def.StartPopulation) * 0.01,
			},
			HousingOcc: 0.75,
		}
		// seed some residential with retail slots near center
		for i := range cs.Parcels {
			p := &cs.Parcels[i]
			dist := math.Abs(float64(p.X)) + math.Abs(float64(p.Y))
			if dist == 0 {
				p.Zone = "center"
				p.Traffic = 1.0
				p.Slots = 4
			} else if dist <= 2 {
				p.Zone = "residential"
				p.Slots = 1
				p.Traffic = 0.7
			} else if dist <= float64(def.StartRadius)-1 {
				if p.X > 0 {
					p.Zone = "industrial"
					p.Traffic = 0.3
				} else {
					p.Zone = "rural"
					p.Traffic = 0.2
				}
			} else {
				p.Zone = "empty"
			}
		}
		seedCityRoads(cs)
		e.World.Cities[def.ID] = cs
	}
	e.seedResourceDeposits()
}

func generateParcels(radius int) []model.Parcel {
	var out []model.Parcel
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if abs(x)+abs(y) > radius {
				continue
			}
			out = append(out, model.Parcel{X: x, Y: y, Zone: "empty"})
		}
	}
	return out
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (e *Engine) seedAI() {
	profiles := []struct {
		id, name, industry, persona string
		cash                        float64
		city                        string
		shopItem                    string
		shopPrice                   float64
		builds                      []struct{ typ, recipe string; x, y int }
	}{
		{"ai_dairy", "Агидель Молоко", "dairy", "ai_dairy", 180000, "ufa", "bottled_milk", 45,
			[]struct{ typ, recipe string; x, y int }{
				{"livestock_farm", "farm_milk", -2, 0},
				{"orchard", "farm_cherry", -3, 0},
				{"quarry", "extract_sand", 3, 1},
				{"glass_plant", "make_bottle", 2, 1},
				{"dairy_plant", "bottle_milk", 2, 0},
			}},
		{"ai_bakery", "Хлебный Урал", "food", "ai_rookie", 150000, "ufa", "bread", 35,
			[]struct{ typ, recipe string; x, y int }{
				{"farm", "grow_wheat", -4, 1},
				{"farm", "grow_potato", -4, 0},
				{"mill", "mill_flour", 1, -1},
				{"plastics_plant", "make_packaging", 2, -1},
				{"bakery", "bake_bread", 1, 0},
			}},
		{"ai_furniture", "УралМебель", "furniture", "ai_furniture", 160000, "ufa", "furniture", 220,
			[]struct{ typ, recipe string; x, y int }{
				{"farm", "grow_flax", -3, 1},
				{"logging_camp", "extract_wood", -4, 2},
				{"textile_mill", "weave_fabric", 1, 2},
				{"board_plant", "make_particle_board", 2, 2},
				{"furniture_plant", "make_furniture", 3, 2},
			}},
		{"ai_apparel", "БашТекстиль", "apparel", "ai_trader", 140000, "ufa", "clothing", 95,
			[]struct{ typ, recipe string; x, y int }{
				{"farm", "grow_flax", -1, 3},
				{"livestock_farm", "raise_cattle", -2, 3},
				{"textile_mill", "weave_fabric", 0, 3},
				{"apparel_factory", "make_clothing", 1, 3},
			}},
		{"ai_steel", "БелорецкСталь", "steel", "ai_steel", 200000, "beloretsk", "steel_goods", 90,
			[]struct{ typ, recipe string; x, y int }{
				{"mine", "mine_iron", 2, 0},
				{"mine", "mine_coal", 3, -1},
				{"logging_camp", "extract_wood", -2, 1},
				{"steel_mill", "smelt_steel", 2, 1},
				{"steel_mill", "steel_goods", 3, 1},
			}},
		{"ai_oil", "УралНефть", "oil", "ai_trader", 220000, "beloretsk", "fuel_can", 85,
			[]struct{ typ, recipe string; x, y int }{
				{"oil_well", "extract_oil", -1, 2},
				{"chem_plant", "extract_chemicals", 0, 2},
				{"plastics_plant", "make_plastic", 1, 2},
				{"refinery", "refine_gasoline", 2, 2},
				{"refinery", "pack_fuel", 3, 2},
			}},
	}

	for _, p := range profiles {
		f := &model.Firm{
			ID: p.id, Name: p.name, IsAI: true, Industry: p.industry, PersonaID: p.persona, Cash: p.cash,
			Researched: map[string]bool{}, BrandAwareness: map[string]float64{
				"food": 0.25, "home": 0.2, "apparel": 0.15, "fuel": 0.15, "tech": 0.1,
			},
			Trucks: 2,
		}
		for _, r := range e.Catalog.Recipes {
			if r.ResearchCost == 0 {
				f.Researched[r.ID] = true
			}
		}
		// unlock a few mid-tier recipes so chains aren't stuck
		for _, rid := range []string{"make_packaging", "make_plastic", "pack_fuel", "make_clothing", "make_leather", "steel_goods"} {
			f.Researched[rid] = true
		}
		e.World.Firms[f.ID] = f
		for _, b := range p.builds {
			e.placeBuilding(f.ID, p.city, b.typ, b.recipe, b.x, b.y, true)
		}
		e.seedAIInventory(f.ID, p.industry)
		e.openAIShop(f, p.city, p.shopItem, p.shopPrice)
	}
}

func (e *Engine) seedAIInventory(firmID, industry string) {
	for _, b := range e.World.Buildings {
		if b.FirmID != firmID {
			continue
		}
		switch industry {
		case "dairy":
			b.Inventory["sand"] += 40
			b.Inventory["quartz"] += 20
			b.Inventory["bottle"] += 20
			b.Inventory["milk"] += 30
			b.Inventory["cherry"] += 15
			b.Inventory["packaging"] += 10
		case "food":
			b.Inventory["wheat"] += 40
			b.Inventory["flour"] += 20
			b.Inventory["plastic"] += 15
			b.Inventory["paper"] += 15
			b.Inventory["packaging"] += 20
		case "furniture":
			b.Inventory["flax"] += 40
			b.Inventory["wood"] += 40
			b.Inventory["chemicals"] += 30
			b.Inventory["fabric"] += 10
			b.Inventory["particle_board"] += 10
		case "apparel":
			b.Inventory["flax"] += 30
			b.Inventory["fabric"] += 15
			b.Inventory["cotton_fabric"] += 10
			b.Inventory["packaging"] += 15
			b.Inventory["cattle"] += 8
			b.Inventory["leather"] += 8
		case "steel":
			b.Inventory["iron_ore"] += 50
			b.Inventory["coal"] += 40
			b.Inventory["steel"] += 15
		case "oil":
			b.Inventory["oil"] += 40
			b.Inventory["chemicals"] += 25
			b.Inventory["plastic"] += 15
			b.Inventory["gasoline"] += 12
		}
	}
}

func (e *Engine) openAIShop(f *model.Firm, cityID, item string, price float64) {
	if item == "" {
		item = "bottled_milk"
	}
	if price <= 0 {
		price = 45
	}
	shopType := e.defaultShopTypeForItem(item)
	city := e.World.Cities[cityID]
	for i := range city.Parcels {
		p := &city.Parcels[i]
		if p.Zone == "residential" && p.Slots > 0 && p.RentedBy == "" {
			id := uuid.NewString()
			e.World.Shops[id] = &model.ShopLease{
				ID: id, FirmID: f.ID, CityID: cityID, X: p.X, Y: p.Y,
				ShopType: shopType, ItemID: item, Price: price, Stock: 20, FixedRent: 500, RevenueShare: 0.08,
				Radius: shopRadiusForZone(p.Zone, p.Slots),
				StockSource: "own", AutoRestock: true,
			}
			p.RentedBy = f.ID
			return
		}
	}
}

func (e *Engine) placeBuilding(firmID, cityID, typ, recipeID string, x, y int, free bool) (string, error) {
	city, ok := e.World.Cities[cityID]
	if !ok {
		return "", fmt.Errorf("unknown city")
	}
	bdef, ok := e.Catalog.BuildingByID[typ]
	if !ok {
		return "", fmt.Errorf("unknown building type")
	}
	firm, ok := e.World.Firms[firmID]
	if !ok {
		return "", fmt.Errorf("unknown firm")
	}
	if !free && firm.Cash < float64(bdef.BuildCost) {
		return "", fmt.Errorf("недостаточно средств")
	}
	var parcel *model.Parcel
	for i := range city.Parcels {
		p := &city.Parcels[i]
		if p.X == x && p.Y == y {
			parcel = p
			break
		}
	}
	if parcel == nil {
		return "", fmt.Errorf("участок вне города")
	}
	if parcel.Building != "" {
		return "", fmt.Errorf("участок занят")
	}
	needZone := bdef.Zone
	if needZone == "industrial" && parcel.Zone != "industrial" && parcel.Zone != "empty" {
		return "", fmt.Errorf("промка только в промзоне/пустых окраинах")
	}
	if needZone == "rural" && parcel.Zone != "rural" && parcel.Zone != "empty" {
		return "", fmt.Errorf("фермы только в сельской зоне")
	}
	if recipeID == "" {
		recipeID = e.defaultRecipeForBuilding(firm, typ)
	}
	if recipeID != "" {
		if err := e.RecipeFitsBuilding(typ, recipeID); err != nil {
			return "", err
		}
	}
	needRes := e.depositItemForBuild(typ, recipeID)
	if extractBuildingResources[typ] || needRes != "" {
		if err := e.requireDepositForBuild(city, typ, recipeID, x, y); err != nil {
			if free && needRes != "" {
				if depositAt(city, x, y) == nil {
					city.Deposits = append(city.Deposits, model.ResourceDeposit{
						ID: uuid.NewString(), ItemID: needRes, X: x, Y: y,
						Amount: 200, MaxAmount: 200, Quality: 3.5,
					})
				}
			} else {
				return "", err
			}
		}
	}
	// Сначала дорога до сети (за счёт строителя), потом здание
	roadTiles, roadCost, err := e.ensureRoadAccess(firm, city, x, y, free)
	if err != nil {
		return "", err
	}
	total := float64(bdef.BuildCost)
	if !free && firm.Cash < total {
		// road already charged inside ensureRoadAccess; refund road if can't afford building
		if roadCost > 0 {
			firm.Cash += roadCost
			// do not unpave — keep network growth simple; only refund money
		}
		return "", fmt.Errorf("недостаточно средств на здание")
	}
	id := uuid.NewString()
	inst := &model.BuildingInst{
		ID: id, Type: typ, FirmID: firmID, CityID: cityID, X: x, Y: y,
		RecipeID: recipeID, Inventory: map[string]int{}, Quality: 3, Active: true,
		AutoProcure: true, SellExternal: true,
		DefaultInput: model.InputSource{Mode: "own"},
	}
	e.World.Buildings[id] = inst
	parcel.Building = id
	parcel.OwnerID = firmID
	parcel.HasRoad = true
	if parcel.Zone == "empty" {
		parcel.Zone = needZone
	}
	if !free {
		firm.Cash -= total
	}
	if roadTiles > 0 {
		e.World.Events = append(e.World.Events, fmt.Sprintf("%s проложил дорогу (%d кл., %.0f ₽) к (%d,%d) в %s", firm.Name, roadTiles, roadCost, x, y, cityID))
	}
	return id, nil
}

func (e *Engine) defaultRecipeForBuilding(firm *model.Firm, typ string) string {
	var fallback string
	for _, r := range e.Catalog.Recipes {
		if r.BuildingType != typ {
			continue
		}
		if fallback == "" {
			fallback = r.ID
		}
		if r.ResearchCost == 0 || (firm != nil && firm.Researched[r.ID]) {
			return r.ID
		}
	}
	return fallback
}

func (e *Engine) CreatePlayer(name, industry string) (*model.Firm, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("нужно имя фирмы")
	}
	for _, f := range e.World.Firms {
		if !f.IsAI && strings.EqualFold(f.Name, name) {
			return nil, fmt.Errorf("имя занято")
		}
	}
	bonus := 100000.0
	researched := map[string]bool{}
	for _, r := range e.Catalog.Recipes {
		if r.ResearchCost == 0 {
			researched[r.ID] = true
		}
	}
	switch industry {
	case "dairy":
		bonus = 120000
	case "food":
		bonus = 115000
	case "furniture":
		bonus = 110000
	case "apparel":
		bonus = 110000
	case "steel":
		bonus = 130000
	case "oil":
		bonus = 140000
	default:
		industry = "general"
	}
	id := uuid.NewString()
	f := &model.Firm{
		ID: id, Name: name, IsAI: false, Industry: industry, Cash: bonus,
		Researched: researched, BrandAwareness: map[string]float64{"food": 0.05, "home": 0.05, "apparel": 0.05, "fuel": 0.05},
		Trucks: 0,
	}
	e.World.Firms[id] = f
	e.World.Events = append(e.World.Events, fmt.Sprintf("Новая фирма: %s (%s)", name, industry))
	return f, nil
}
