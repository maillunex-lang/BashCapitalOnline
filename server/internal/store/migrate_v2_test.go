package store

import (
	"path/filepath"
	"runtime"
	"testing"

	"bashcapital/server/internal/model"
)

func TestMigrateWorldCopyToV2MapsAndCompensates(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	dataDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "data"))
	itemMap, err := LoadItemMigrationMap(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	world := &model.World{
		Firms: map[string]*model.Firm{
			"f1": {ID: "f1", Cash: 1000},
		},
		Buildings: map[string]*model.BuildingInst{
			"b1": {
				ID: "b1", Type: "old_farm",
				Inventory: map[string]int{
					"wheat":          10,
					"unknown_legacy": 5,
				},
			},
		},
		Shops: map[string]*model.ShopLease{
			"s1": {ID: "s1", ItemID: "unknown_shop_sku", Stock: 3},
		},
	}
	// Ensure wheat maps if present; otherwise invent map for test determinism.
	if _, ok := itemMap["wheat"]; !ok {
		itemMap["wheat"] = map[string]any{"action": "map_to_new", "new_id": "wheat"}
	}
	buildingMap := map[string]string{"old_farm": "grain_farm"}
	comp, report := MigrateWorldCopyToV2(world, itemMap, buildingMap)
	if world.CatalogVersion != 2 {
		t.Fatalf("catalog version %d", world.CatalogVersion)
	}
	if world.Buildings["b1"].Type != "grain_farm" {
		t.Fatalf("building type %s", world.Buildings["b1"].Type)
	}
	if world.Buildings["b1"].Inventory["wheat"] != 10 {
		t.Fatalf("wheat qty=%v", world.Buildings["b1"].Inventory)
	}
	if _, still := world.Buildings["b1"].Inventory["unknown_legacy"]; still {
		t.Fatal("unknown_legacy should be compensated away")
	}
	if comp <= 0 {
		t.Fatalf("expected compensation >0, got %v report=%v", comp, report)
	}
	if world.Shops["s1"].Stock != 0 {
		t.Fatalf("shop stock should clear on unmapped item")
	}
	if world.Firms["f1"].Cash <= 1000 {
		t.Fatalf("firm should receive compensation cash, cash=%v", world.Firms["f1"].Cash)
	}
}
