package sim

import (
	"sync"
	"testing"

	"bashcapital/server/internal/model"
)

type memStore struct {
	mu   sync.Mutex
	data *model.World
	n    int
}

func (m *memStore) SaveWorld(w *model.World) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.n++
	// shallow-copy firm map keys so we can assert presence after "restart"
	cp := *w
	firms := map[string]*model.Firm{}
	for id, f := range w.Firms {
		fc := *f
		firms[id] = &fc
	}
	cp.Firms = firms
	m.data = &cp
	return nil
}

func (m *memStore) HasWorld() bool { return m.data != nil }

func testEngine() *Engine {
	cat := testCatalog()
	e := &Engine{Catalog: cat, World: &model.World{
		Tick: 0, Month: 1, Year: 2026,
		Cities: map[string]*model.CityState{}, Firms: map[string]*model.Firm{},
		Buildings: map[string]*model.BuildingInst{}, Shops: map[string]*model.ShopLease{},
		Orders: map[string]*model.MarketOrder{}, Contracts: map[string]*model.Contract{},
		Shipments: map[string]*model.Shipment{}, Events: []string{}, Chat: []model.ChatMessage{},
	}}
	e.seedCities()
	e.initCityMarket()
	return e
}

func testCatalog() *model.Catalog {
	return &model.Catalog{
		Items: []model.ItemDef{
			{ID: "bread", Name: "Хлеб", Tier: "finished", Category: "food"},
			{ID: "flour", Name: "Мука", Tier: "intermediate", Category: "food"},
		},
		ItemByID: map[string]model.ItemDef{
			"bread": {ID: "bread", Name: "Хлеб", Tier: "finished", Category: "food"},
			"flour": {ID: "flour", Name: "Мука", Tier: "intermediate", Category: "food"},
		},
		Recipes: []model.RecipeDef{
			{ID: "bake_bread", Name: "Выпечка", BuildingType: "bakery", Output: model.RecipeIO{ItemID: "bread", Qty: 1}},
			{ID: "mill_flour", Name: "Помол", BuildingType: "mill", Output: model.RecipeIO{ItemID: "flour", Qty: 1}},
		},
		RecipeByID: map[string]model.RecipeDef{
			"bake_bread": {ID: "bake_bread", Name: "Выпечка", BuildingType: "bakery", Output: model.RecipeIO{ItemID: "bread", Qty: 1}},
			"mill_flour":  {ID: "mill_flour", Name: "Помол", BuildingType: "mill", Output: model.RecipeIO{ItemID: "flour", Qty: 1}},
		},
		Buildings: []model.BuildingDef{{ID: "bakery", Name: "Пекарня"}},
		BuildingByID: map[string]model.BuildingDef{
			"bakery": {ID: "bakery", Name: "Пекарня"},
			"mill":   {ID: "mill", Name: "Мельница"},
		},
		Shops: []model.ShopDef{{ID: "grocery", Name: "Продукты", AllowedCategories: []string{"food"}, AllowedTiers: []string{"finished"}, RentMult: 1}},
		ShopByID: map[string]model.ShopDef{
			"grocery": {ID: "grocery", Name: "Продукты", AllowedCategories: []string{"food"}, AllowedTiers: []string{"finished"}, RentMult: 1},
		},
		Cities: []model.CityDef{{ID: "ufa", Name: "Уфа", Enabled: true, X: 100, Y: 100, StartPopulation: 1000, StartRadius: 3}},
		CityByID: map[string]model.CityDef{
			"ufa": {ID: "ufa", Name: "Уфа", Enabled: true, X: 100, Y: 100, StartPopulation: 1000, StartRadius: 3},
		},
		Economy: model.EconomyConfig{
			PriceCorridors: map[string]model.PriceCorridor{
				"bread": {Min: 99, Max: 180},
			},
			Logistics: model.LogisticsConfig{UnitsPerTick: 80, CostPerUnitDist: 0.01, LocalCostPerUnit: 0.5, BaseCostPerUnit: 1},
		},
	}
}

func TestPersistSavesFirmBeforeTick(t *testing.T) {
	e := testEngine()
	st := &memStore{}
	e.Store = st
	firm, err := e.CreatePlayer("ТестФирма", "food")
	if err != nil {
		t.Fatal(err)
	}
	e.Persist()
	if st.n != 1 {
		t.Fatalf("expected 1 save, got %d", st.n)
	}
	if st.data == nil || st.data.Firms[firm.ID] == nil {
		t.Fatal("firm missing from persisted world")
	}
	if st.data.Firms[firm.ID].Name != "ТестФирма" {
		t.Fatalf("wrong name: %s", st.data.Firms[firm.ID].Name)
	}
}

func TestSnapPriceToCorridor(t *testing.T) {
	e := testEngine()
	if got := e.SnapPriceToCorridor("bread", 55); got < 99 || got > 180 {
		t.Fatalf("55 should snap into corridor, got %v", got)
	}
	if got := e.SnapPriceToCorridor("bread", 120); got != 120 {
		t.Fatalf("in-range price must stay, got %v", got)
	}
	mid := e.SnapPriceToCorridor("bread", 0)
	want := (99.0 + 180.0) / 2
	if mid != want {
		t.Fatalf("zero snaps to mid: got %v want %v", mid, want)
	}
}

func TestRentShopSnapsLowPrice(t *testing.T) {
	e := testEngine()
	firm, err := e.CreatePlayer("Арендатор", "food")
	if err != nil {
		t.Fatal(err)
	}
	city := e.World.Cities["ufa"]
	var x, y int
	found := false
	for i := range city.Parcels {
		p := &city.Parcels[i]
		if p.Slots > 0 && p.RentedBy == "" {
			x, y = p.X, p.Y
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no free commercial parcel")
	}
	shop, err := e.RentShop(firm.ID, "ufa", x, y, "grocery", "bread", 55)
	if err != nil {
		t.Fatalf("rent with 55₽ should snap, got err: %v", err)
	}
	if shop.Price < 99 || shop.Price > 180 {
		t.Fatalf("shop price out of corridor: %v", shop.Price)
	}
	if !shop.AutoRestock {
		t.Fatal("new shop should AutoRestock")
	}
}

func TestPlayerShopAutoRestockFromFactory(t *testing.T) {
	e := testEngine()
	firm, err := e.CreatePlayer("Пекарь", "food")
	if err != nil {
		t.Fatal(err)
	}
	b := &model.BuildingInst{
		ID: "b1", FirmID: firm.ID, CityID: "ufa", Type: "bakery", RecipeID: "bake_bread",
		Inventory: map[string]int{"bread": 40}, Quality: 3, SellExternal: false,
	}
	e.World.Buildings[b.ID] = b
	shop := &model.ShopLease{
		ID: "s1", FirmID: firm.ID, CityID: "ufa", ShopType: "grocery",
		ItemID: "bread", Price: 120, Stock: 0, AutoRestock: true, StockSource: "own",
	}
	e.World.Shops[shop.ID] = shop

	log := []string{}
	e.restockAllAutoShops(&log)
	if shop.Stock <= 0 {
		t.Fatalf("player shop should auto-restock from factory, stock=%d inv=%d log=%v", shop.Stock, b.Inventory["bread"], log)
	}
	if b.Inventory["bread"]+shop.Stock != 40 {
		t.Fatalf("inventory not conserved: factory=%d shop=%d", b.Inventory["bread"], shop.Stock)
	}
}

func TestConfigureBuildingAtomicOnError(t *testing.T) {
	e := testEngine()
	firm, err := e.CreatePlayer("Конфиг", "food")
	if err != nil {
		t.Fatal(err)
	}
	b := &model.BuildingInst{
		ID: "b2", FirmID: firm.ID, CityID: "ufa", Type: "bakery", RecipeID: "bake_bread",
		Inventory: map[string]int{}, Quality: 3,
		AutoProcure: false, SellExternal: false,
		DefaultInput: model.InputSource{Mode: "own"},
		InputSources: map[string]model.InputSource{},
	}
	e.World.Buildings[b.ID] = b

	_, err = e.ConfigureBuilding(firm.ID, b.ID, "mill_flour", true, "firm", "", true, nil)
	if err == nil {
		t.Fatal("expected error for wrong recipe + missing firm_id")
	}
	if b.RecipeID != "bake_bread" {
		t.Fatalf("recipe should not change on error, got %s", b.RecipeID)
	}
	if b.AutoProcure || b.SellExternal {
		t.Fatal("flags should not partially apply on error")
	}
	if b.DefaultInput.Mode != "own" {
		t.Fatalf("DefaultInput mutated: %+v", b.DefaultInput)
	}

	_, err = e.ConfigureBuilding(firm.ID, b.ID, "", true, "firm", "", true, nil)
	if err == nil {
		t.Fatal("expected firm_id error")
	}
	if b.AutoProcure || b.SellExternal {
		t.Fatal("partial apply after firm_id error")
	}

	out, err := e.ConfigureBuilding(firm.ID, b.ID, "bake_bread", true, "own", "", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !out.AutoProcure || !out.SellExternal {
		t.Fatal("valid configure should apply flags")
	}
}

func TestClampPriceStillRejectsOutOfRange(t *testing.T) {
	e := testEngine()
	_, err := e.ClampPrice("bread", 55)
	if err == nil {
		t.Fatal("ClampPrice should still reject out-of-range (market orders)")
	}
}
