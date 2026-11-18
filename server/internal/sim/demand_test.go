package sim

import (
	"strings"
	"testing"

	"bashcapital/server/internal/model"
)

func demandTestEngine() *Engine {
	cat := &model.Catalog{
		Items: []model.ItemDef{{ID: "bread", Name: "Хлеб", Tier: "finished", Category: "food"}},
		ItemByID: map[string]model.ItemDef{
			"bread": {ID: "bread", Name: "Хлеб", Tier: "finished", Category: "food"},
		},
		Cities: []model.CityDef{{ID: "ufa", Name: "Уфа", Enabled: true, X: 100, Y: 100, StartPopulation: 10000, StartRadius: 3}},
		CityByID: map[string]model.CityDef{
			"ufa": {ID: "ufa", Name: "Уфа", Enabled: true, X: 100, Y: 100, StartPopulation: 10000, StartRadius: 3},
		},
		Economy: model.EconomyConfig{
			PriceCorridors: map[string]model.PriceCorridor{"bread": {Min: 90, Max: 150}},
			Demand: model.DemandConfig{
				ClassShares: struct {
					Budget  float64 `json:"budget"`
					Mid     float64 `json:"mid"`
					Premium float64 `json:"premium"`
				}{Budget: 0.45, Mid: 0.4, Premium: 0.15},
				QualityWeight: 0.4, BrandWeight: 0.15, PriceWeight: 0.25, DistanceWeight: 0.2,
				BrandDecay: 0.002, BrandFloor: 0.03, BrandGainOnSale: 0.006,
				TooExpensiveAt: 1.35,
				PopDemandPerDay: map[string]float64{"food": 0.12},
			},
		},
	}
	e := &Engine{Catalog: cat, World: &model.World{
		Tick: 1, Month: 1, Year: 2026,
		Cities: map[string]*model.CityState{},
		Firms: map[string]*model.Firm{}, Buildings: map[string]*model.BuildingInst{},
		Shops: map[string]*model.ShopLease{}, Orders: map[string]*model.MarketOrder{},
		Contracts: map[string]*model.Contract{}, Shipments: map[string]*model.Shipment{},
		Events: []string{},
	}}
	e.seedCities()
	city := e.World.Cities["ufa"]
	// Tight demand so shops compete (quality order matters).
	city.Population = 2500
	city.DemandSeg = map[string]float64{"food": 18, "home": 4}
	return e
}

func TestHigherQualitySellsMore(t *testing.T) {
	e := demandTestEngine()
	hi := &model.Firm{ID: "hi", Name: "HiQ", Cash: 1e6, BrandAwareness: map[string]float64{"food": 0.2}}
	lo := &model.Firm{ID: "lo", Name: "LoQ", Cash: 1e6, BrandAwareness: map[string]float64{"food": 0.2}}
	e.World.Firms[hi.ID] = hi
	e.World.Firms[lo.ID] = lo
	e.World.Buildings["bhi"] = &model.BuildingInst{
		ID: "bhi", FirmID: hi.ID, CityID: "ufa", Type: "bakery", Quality: 4.8, Inventory: map[string]int{},
	}
	e.World.Buildings["blo"] = &model.BuildingInst{
		ID: "blo", FirmID: lo.ID, CityID: "ufa", Type: "bakery", Quality: 2.0, Inventory: map[string]int{},
	}
	price := 120.0
	e.World.Shops["shi"] = &model.ShopLease{
		ID: "shi", FirmID: hi.ID, CityID: "ufa", X: 0, Y: 0, ShopType: "grocery",
		ItemID: "bread", Price: price, Stock: 40, Radius: 4, RevenueShare: 0.08, Quality: 4.8,
	}
	e.World.Shops["slo"] = &model.ShopLease{
		ID: "slo", FirmID: lo.ID, CityID: "ufa", X: 1, Y: 0, ShopType: "grocery",
		ItemID: "bread", Price: price, Stock: 40, Radius: 4, RevenueShare: 0.08, Quality: 2.0,
	}

	log := []string{}
	e.runRetail(&log)
	soldHi := 40 - e.World.Shops["shi"].Stock
	soldLo := 40 - e.World.Shops["slo"].Stock
	if soldHi <= soldLo {
		t.Fatalf("higher quality should sell more: hi=%d lo=%d log=%v", soldHi, soldLo, log)
	}
	if soldHi == 0 {
		t.Fatalf("expected some sales, log=%v", log)
	}
}

func TestTooExpensiveBlocksSales(t *testing.T) {
	e := demandTestEngine()
	f := &model.Firm{ID: "f1", Name: "Dear", Cash: 1e6, BrandAwareness: map[string]float64{"food": 0.5}}
	e.World.Firms[f.ID] = f
	e.World.Shops["s1"] = &model.ShopLease{
		ID: "s1", FirmID: f.ID, CityID: "ufa", X: 0, Y: 0, ShopType: "grocery",
		ItemID: "bread", Price: 400, Stock: 30, Radius: 4, Quality: 5,
	}
	log := []string{}
	e.runRetail(&log)
	if e.World.Shops["s1"].Stock != 30 {
		t.Fatalf("overpriced shop should not sell, stock=%d log=%v", e.World.Shops["s1"].Stock, log)
	}
	found := false
	for _, line := range log {
		if contains(line, "too_expensive") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected too_expensive log, got %v", log)
	}
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
