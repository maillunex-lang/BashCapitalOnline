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
	Paused  bool
	DataDir string
	Store   WorldStore
}

// WorldStore — optional SQLite persistence (RM-PERSIST).
type WorldStore interface {
	SaveWorld(w *model.World) error
	HasWorld() bool
}

func NewEngine(cat *model.Catalog) *Engine {
	ver := cat.Version
	if ver == 0 {
		ver = 1
	}
	e := &Engine{Catalog: cat, World: &model.World{
		Tick: 0, Month: 1, Year: 2026, CatalogVersion: ver,
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
	e.ensureOpenExtractDeposits()
	e.initCityMarket()
	msg := "Мир Башкортостана инициализирован. Активны Уфа, Белорецк и Стерлитамак."
	if ver >= 2 {
		msg += " Каталог экономики V2."
	}
	e.World.Events = append(e.World.Events, msg)
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
		for i := range cs.Parcels {
			zoneParcel(&cs.Parcels[i], def.StartRadius)
		}
		seedCityRoads(cs)
		e.World.Cities[def.ID] = cs
	}
	e.seedResourceDeposits()
}

func zoneParcel(p *model.Parcel, radius int) {
	distance := abs(p.X) + abs(p.Y)
	switch {
	case distance == 0:
		p.Zone, p.Traffic, p.Slots = "center", 1, 4
	case distance <= 2:
		p.Zone, p.Traffic, p.Slots = "residential", 0.7, 1
	case distance < radius && p.X > 0:
		p.Zone, p.Traffic = "industrial", 0.3
	case distance < radius:
		p.Zone, p.Traffic = "rural", 0.2
	}
}

// ExpandCitiesToCatalog extends saved cities with new parcels while preserving the
// existing coordinates, deposits, buildings and parcel ownership.
func (e *Engine) ExpandCitiesToCatalog() bool {
	expanded := false
	for _, def := range e.Catalog.Cities {
		city := e.World.Cities[def.ID]
		if !def.Enabled || city == nil || city.Radius >= def.StartRadius {
			continue
		}
		type point struct{ X, Y int }
		existing := make(map[point]bool, len(city.Parcels))
		for _, parcel := range city.Parcels {
			existing[point{parcel.X, parcel.Y}] = true
		}
		outer := &model.CityState{ID: city.ID}
		for _, parcel := range generateParcels(def.StartRadius) {
			if existing[point{parcel.X, parcel.Y}] {
				continue
			}
			zoneParcel(&parcel, def.StartRadius)
			outer.Parcels = append(outer.Parcels, parcel)
		}
		seedCityRoads(outer)
		city.Parcels = append(city.Parcels, outer.Parcels...)
		city.Radius = def.StartRadius
		e.World.MapRevision++
		expanded = true
	}
	return expanded
}

func generateParcels(radius int) []model.Parcel {
	var out []model.Parcel
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if abs(x)+abs(y) > radius {
				continue
			}
			// Mild hills for map_3d preview (Urals-ish east rise).
			elev := 0.15*math.Sin(float64(x)*0.7) + 0.12*math.Cos(float64(y)*0.55) + 0.04*float64(x)
			out = append(out, model.Parcel{X: x, Y: y, Zone: "empty", Elevation: elev})
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
	type buildSpec struct {
		typ, recipe string
		x, y        int
	}
	type profile struct {
		id, name, industry, persona string
		cash                        float64
		city                        string
		shopItem                    string
		shopPrice                   float64
		builds                      []buildSpec
	}
	var profiles []profile
	if e.Catalog != nil && e.Catalog.Version >= 2 {
		profiles = []profile{
			{"ai_dairy", "Агидель Молоко", "dairy", "ai_dairy", 180000, "ufa", "milk_bottle", 55,
				[]buildSpec{{"dairy_farm", "extract_milk", -2, 0}, {"dairy_plant", "food_22", 2, 0}}},
			{"ai_bakery", "Хлебный Урал", "food", "ai_rookie", 150000, "ufa", "wheat_bread", 48,
				[]buildSpec{{"grain_farm", "extract_wheat", -4, 1}, {"flour_mill", "food_01", 1, -1}, {"bakery", "food_11", 1, 0}}},
			{"ai_furniture", "УралМебель", "furniture", "ai_furniture", 160000, "ufa", "table", 280,
				[]buildSpec{{"logging_camp", "extract_softwood", -4, 2}, {"sawmill", "wood_01", 1, 2}, {"furniture_plant", "wood_06", 2, 2}, {"furniture_plant", "home_01", 3, 2}}},
			{"ai_apparel", "БашТекстиль", "apparel", "ai_trader", 140000, "ufa", "tshirt", 120,
				[]buildSpec{{"fiber_farm", "extract_flax", -1, 3}, {"sheep_farm", "extract_wool", -2, 3}, {"fiber_plant", "tex_01", 0, 3}, {"spinning_mill", "tex_02", 1, 3}, {"knitwear", "tex_04", 2, 3}, {"garment_factory", "app_01", 3, 3}}},
			{"ai_steel", "БелорецкСталь", "steel", "ai_steel", 200000, "beloretsk", "pipe_set", 160,
				[]buildSpec{{"mine", "extract_iron_ore", 2, 0}, {"mine", "extract_coal", 3, -1}, {"blast_furnace", "met_01", 2, 1}, {"steel_plant", "met_02", 3, 1}, {"pipe_plant", "met_06", 3, 2}}},
			{"ai_oil", "УфаНефть", "oil", "ai_trader", 220000, "ufa", "ai92", 55,
				[]buildSpec{{"oil_well", "extract_crude_oil", -1, 2}, {"refinery", "oil_01", 2, 2}, {"refinery", "fuel_01", 3, 2}}},
			{"ai_chem", "СодаСтерлитамак", "oil", "ai_trader", 210000, "sterlitamak", "detergent", 70,
				[]buildSpec{{"mine", "extract_salt", -1, -2}, {"quarry", "extract_limestone", -1, 2}, {"chemical_plant", "oil_08", 1, 2}, {"chemical_plant", "home_06", 2, 1}}},
		}
	} else {
		profiles = []profile{
			{"ai_dairy", "Агидель Молоко", "dairy", "ai_dairy", 180000, "ufa", "bottled_milk", 45,
				[]buildSpec{{"livestock_farm", "milk_cattle", -2, 0}, {"dairy_plant", "bottle_milk", 2, 0}}},
			{"ai_bakery", "Хлебный Урал", "food", "ai_rookie", 150000, "ufa", "bread_wheat", 35,
				[]buildSpec{{"farm", "grow_wheat", -4, 1}, {"bakery", "bake_bread_wheat", 1, 0}}},
			{"ai_steel", "БелорецкСталь", "steel", "ai_steel", 200000, "beloretsk", "goods_steel", 90,
				[]buildSpec{{"mine", "extract_iron_ore", 2, 0}, {"mine", "extract_coal", 3, -1}}},
		}
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
		for _, rid := range []string{"make_packaging", "make_plastic", "pack_gasoline", "make_clothing", "board_wood", "goods_steel", "furn_wood", "weave_flax", "refine_gasoline", "extract_salt", "extract_limestone"} {
			f.Researched[rid] = true
		}
		e.World.Firms[f.ID] = f
		for _, b := range p.builds {
			if city := e.World.Cities[p.city]; city != nil && e.canSeedAIBuildAt(city, b.x, b.y) {
				if _, err := e.placeBuilding(f.ID, p.city, b.typ, b.recipe, b.x, b.y, true); err == nil {
					continue
				}
			}
			if x, y, ok := e.findAIBuildSite(p.city, b.typ, b.recipe); ok {
				_, _ = e.placeBuilding(f.ID, p.city, b.typ, b.recipe, x, y, true)
			}
		}
		e.seedAIInventory(f.ID, p.industry)
		e.openAIShop(f, p.city, p.shopItem, p.shopPrice)
	}
}

// Leave at least one accessible deposit of each generated resource when seeding competitors.
func (e *Engine) canSeedAIBuildAt(city *model.CityState, x, y int) bool {
	d := depositAt(city, x, y)
	if e.Catalog.Version < 2 || d == nil {
		return true
	}
	for _, other := range city.Deposits {
		if (other.X == x && other.Y == y) || other.Amount <= 0 || !depositMatchesNeed(other.ItemID, d.ItemID) {
			continue
		}
		if p := parcelAt(city, other.X, other.Y); p != nil && p.Building == "" {
			return true
		}
	}
	return false
}

// findAIBuildSite picks a free parcel for AI seeding (zone-aware, prefers matching deposits).
func (e *Engine) findAIBuildSite(cityID, buildingType, recipeID string) (int, int, bool) {
	city := e.World.Cities[cityID]
	if city == nil {
		return 0, 0, false
	}
	bdef, ok := e.Catalog.BuildingByID[buildingType]
	if !ok {
		return 0, 0, false
	}
	needRes := e.depositItemForBuild(buildingType, recipeID)
	var candidates []model.Parcel
	for _, p := range city.Parcels {
		if p.Building != "" || !e.canSeedAIBuildAt(city, p.X, p.Y) {
			continue
		}
		zoneOK := false
		switch bdef.Zone {
		case "industrial":
			zoneOK = p.Zone == "industrial" || p.Zone == "empty"
		case "rural":
			zoneOK = p.Zone == "rural" || p.Zone == "empty"
		default:
			zoneOK = p.Zone != "center"
		}
		if !zoneOK {
			continue
		}
		if needRes != "" {
			d := depositAt(city, p.X, p.Y)
			if d != nil && !depositMatchesNeed(d.ItemID, needRes) {
				continue
			}
		}
		candidates = append(candidates, p)
	}
	// Prefer deposit match (base id) when required.
	if needRes != "" {
		for _, p := range candidates {
			if d := depositAt(city, p.X, p.Y); d != nil && depositMatchesNeed(d.ItemID, needRes) {
				return p.X, p.Y, true
			}
		}
	}
	if len(candidates) == 0 {
		return 0, 0, false
	}
	p := candidates[len(candidates)/2]
	return p.X, p.Y, true
}

func (e *Engine) seedAIInventory(firmID, industry string) {
	v2 := e.Catalog != nil && e.Catalog.Version >= 2
	for _, b := range e.World.Buildings {
		if b.FirmID != firmID {
			continue
		}
		if v2 {
			switch industry {
			case "dairy":
				b.Inventory["milk"] += 40
				b.Inventory["milk_bottle"] += 30
			case "food":
				b.Inventory["wheat"] += 50
				b.Inventory["wheat_flour"] += 25
				b.Inventory["wheat_bread"] += 40
			case "furniture":
				b.Inventory["softwood"] += 40
				b.Inventory["lumber"] += 20
				b.Inventory["furniture_parts"] += 10
				b.Inventory["table"] += 8
			case "apparel":
				b.Inventory["flax"] += 30
				b.Inventory["yarn"] += 15
				b.Inventory["knitted_fabric"] += 10
				b.Inventory["tshirt"] += 20
			case "steel":
				b.Inventory["iron_ore"] += 50
				b.Inventory["coal"] += 40
				b.Inventory["pig_iron"] += 20
				b.Inventory["steel"] += 15
				b.Inventory["steel_pipe"] += 10
			case "oil":
				b.Inventory["crude_oil"] += 40
				b.Inventory["natural_gas"] += 20
				b.Inventory["gasoline_base"] += 15
				b.Inventory["ai92"] += 25
				b.Inventory["salt"] += 15
				b.Inventory["chemical_reagent"] += 15
				b.Inventory["detergent"] += 20
			}
			continue
		}
		switch industry {
		case "dairy":
			b.Inventory["milk"] += 30
			b.Inventory["bottled_milk"] += 40
		case "food":
			b.Inventory["wheat"] += 50
			b.Inventory["bread_wheat"] += 40
		case "steel":
			b.Inventory["iron_ore"] += 50
			b.Inventory["coal"] += 40
		case "oil":
			b.Inventory["oil"] += 40
			b.Inventory["fuel_can_gasoline"] += 35
		}
	}
}

func (e *Engine) openAIShop(f *model.Firm, cityID, item string, price float64) {
	if item == "" {
		if e.Catalog != nil && e.Catalog.Version >= 2 {
			item = "milk_bottle"
		} else {
			item = "bottled_milk"
		}
	}
	if _, ok := e.Catalog.ItemByID[item]; !ok {
		return
	}
	// Always snap into current corridor — old hardcoded AI prices break after rebalance.
	if c, ok := e.Catalog.Economy.PriceCorridors[item]; ok {
		if price < c.Min || price > c.Max || price <= 0 {
			price = (c.Min + c.Max) / 2
		}
	} else if price <= 0 {
		price = 100
	}
	shopType := e.defaultShopTypeForItem(item)
	city := e.World.Cities[cityID]
	for i := range city.Parcels {
		p := &city.Parcels[i]
		if p.Zone == "residential" && p.Slots > 0 && p.RentedBy == "" {
			id := uuid.NewString()
			e.World.Shops[id] = &model.ShopLease{
				ID: id, FirmID: f.ID, CityID: cityID, X: p.X, Y: p.Y,
				ShopType: shopType, ItemID: item, Price: price, Stock: 40, FixedRent: 500, RevenueShare: 0.08,
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
	if recipeID == "" {
		recipeID = e.defaultRecipeForBuilding(firm, typ)
	}
	if recipeID != "" {
		if err := e.RecipeFitsBuilding(typ, recipeID); err != nil {
			return "", err
		}
	}
	needZone := bdef.Zone
	needRes := e.depositItemForBuild(typ, recipeID)
	depositOK := false
	if extractBuildingResources[typ] || needRes != "" {
		if err := e.requireDepositForBuild(city, typ, recipeID, x, y); err != nil {
			if free && needRes != "" {
				if depositAt(city, x, y) == nil {
					base := resourceBaseID(needRes)
					renew := ruralDepositItems[base]
					city.Deposits = append(city.Deposits, model.ResourceDeposit{
						ID: uuid.NewString(), ItemID: base, X: x, Y: y,
						Amount: 200, MaxAmount: 200, Quality: 3.5, Renewable: renew,
					})
					depositOK = true
				}
			} else {
				return "", err
			}
		} else {
			depositOK = true
		}
	}
	// Extractors may sit on a matching deposit even if the parcel zone drifted
	// (e.g. oil on rural edge); convert the tile to the building zone on place.
	zoneOK := true
	if needZone == "industrial" && parcel.Zone != "industrial" && parcel.Zone != "empty" {
		zoneOK = depositOK && extractBuildingResources[typ]
		if !zoneOK {
			return "", fmt.Errorf("промка только в промзоне/пустых окраинах")
		}
	}
	if needZone == "rural" && parcel.Zone != "rural" && parcel.Zone != "empty" {
		zoneOK = depositOK && extractBuildingResources[typ]
		if !zoneOK {
			return "", fmt.Errorf("фермы только в сельской зоне")
		}
	}
	_ = zoneOK
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
	if parcel.Zone == "empty" || (extractBuildingResources[typ] && depositOK && parcel.Zone != needZone && needZone != "") {
		parcel.Zone = needZone
	}
	if !free {
		firm.Cash -= total
	}
	if roadTiles > 0 {
		e.World.Events = append(e.World.Events, fmt.Sprintf("%s проложил дорогу (%d кл., %.0f ₽) к (%d,%d) в %s", firm.Name, roadTiles, roadCost, x, y, cityID))
	}
	e.markDirty(cityID, x, y, "build")
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

func mapIndustryBranch(industry string) string {
	switch industry {
	case "dairy":
		return "dairy"
	case "food":
		return "food"
	case "furniture":
		return "timber"
	case "apparel":
		return "textile"
	case "steel":
		return "metal"
	case "oil":
		return "oil_chem"
	default:
		return "agro"
	}
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
	branch := mapIndustryBranch(industry)
	researched := map[string]bool{}
	nodes := map[string]bool{}
	// Starter unlock: tier 0 + free research only in primary branch (and shared packaging/logistics).
	for _, r := range e.Catalog.Recipes {
		if r.ResearchCost == 0 && r.TierUnlock == 0 && (r.Branch == branch || r.Branch == "packaging" || r.Branch == "retail" || r.Branch == "research" || r.Branch == "") {
			researched[r.ID] = true
		}
	}
	nodes[branch+"_t0"] = true
	nodes["packaging_t0"] = true
	nodes["research_t0"] = true
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
		Trucks: 0, PrimaryBranch: branch, ResearchNodes: nodes, ResearchPoints: 3,
	}
	e.World.Firms[id] = f
	e.World.Events = append(e.World.Events, fmt.Sprintf("Новая фирма: %s (%s)", name, industry))
	return f, nil
}
