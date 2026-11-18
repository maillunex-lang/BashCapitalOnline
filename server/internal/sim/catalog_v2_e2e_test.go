package sim

import (
	"path/filepath"
	"runtime"
	"testing"

	"bashcapital/server/internal/model"
	"bashcapital/server/internal/store"
)

func v2DataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "data"))
}

func TestCatalogV2LoadsAndVersions(t *testing.T) {
	cat, err := store.LoadCatalog(v2DataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	if cat.Version < 2 {
		t.Fatalf("expected catalog v2, got %d (dir resolve failed?)", cat.Version)
	}
	if len(cat.Items) != 170 {
		t.Fatalf("items=%d", len(cat.Items))
	}
	if len(cat.Shops) != 14 {
		t.Fatalf("shops=%d", len(cat.Shops))
	}
}

func TestCatalogV2E2EChains(t *testing.T) {
	cat, err := store.LoadCatalog(v2DataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	if cat.Version < 2 {
		t.Skip("V2 catalog not active")
	}
	e := NewEngine(cat)
	// Clear AI noise for deterministic chains
	e.World.Buildings = map[string]*model.BuildingInst{}
	e.World.Shops = map[string]*model.ShopLease{}
	for id, f := range e.World.Firms {
		if f.IsAI {
			delete(e.World.Firms, id)
		}
	}
	for _, c := range e.World.Cities {
		for i := range c.Parcels {
			c.Parcels[i].Building = ""
			c.Parcels[i].RentedBy = ""
		}
	}

	firm, err := e.CreatePlayer("V2Цепи", "food")
	if err != nil {
		t.Fatal(err)
	}
	firm.Cash = 5_000_000
	for _, r := range cat.Recipes {
		firm.Researched[r.ID] = true
	}

	chains := []struct {
		name   string
		steps  []struct{ bld, recipe string }
		want   string
		city   string
		rawDep string
	}{
		{"bread", []struct{ bld, recipe string }{
			{"grain_farm", "extract_wheat"}, {"flour_mill", "food_01"}, {"bakery", "food_11"},
		}, "wheat_bread", "ufa", "wheat"},
		{"cheese", []struct{ bld, recipe string }{
			{"dairy_farm", "extract_milk"}, {"cheese_plant", "food_25"},
		}, "cheese", "ufa", "milk"},
		{"honey", []struct{ bld, recipe string }{
			{"apiary", "extract_honey"}, {"honey_packaging", "bash_01"},
		}, "bashkir_honey_pack", "ufa", "honey"},
		{"furniture", []struct{ bld, recipe string }{
			{"logging_camp", "extract_softwood"}, {"sawmill", "wood_01"}, {"wood_plant", "wood_03"}, {"furniture_plant", "wood_06"}, {"furniture_plant", "home_02"},
		}, "chair", "ufa", "softwood"},
		{"brick", []struct{ bld, recipe string }{
			{"quarry", "extract_clay"}, {"brickworks", "bld_03"}, {"building_packaging", "build_02"},
		}, "brick_pack", "beloretsk", "clay"},
		{"gasoline", []struct{ bld, recipe string }{
			{"oil_well", "extract_crude_oil"}, {"refinery", "oil_01"}, {"refinery", "fuel_01"},
		}, "ai92", "ufa", "crude_oil"},
		{"tires", []struct{ bld, recipe string }{
			{"oil_well", "extract_crude_oil"}, {"gas_well", "extract_natural_gas"},
			{"petrochemical_plant", "oil_05"}, {"petrochemical_plant", "oil_06"},
			{"mine", "extract_iron_ore"}, {"mine", "extract_coal"},
			{"blast_furnace", "met_01"}, {"steel_plant", "met_02"}, {"tire_plant", "auto_01"},
		}, "tire", "ufa", "crude_oil"},
		{"smartphone", []struct{ bld, recipe string }{
			{"electronics_assembly", "elec_01"},
		}, "smartphone", "ufa", ""},
	}

	for _, ch := range chains {
		t.Run(ch.name, func(t *testing.T) {
			city := e.World.Cities[ch.city]
			pickSite := func(bld string, needDep string) (int, int, bool) {
				bdef := cat.BuildingByID[bld]
				zoneOK := func(z string) bool {
					switch bdef.Zone {
					case "rural":
						return z == "rural" || z == "empty"
					case "industrial":
						return z == "industrial" || z == "empty"
					default:
						return z != "center"
					}
				}
				if needDep != "" {
					for _, d := range city.Deposits {
						if resourceBaseID(d.ItemID) != needDep {
							continue
						}
						for _, p := range city.Parcels {
							if p.X == d.X && p.Y == d.Y && p.Building == "" && zoneOK(p.Zone) {
								return p.X, p.Y, true
							}
						}
					}
				}
				for i := range city.Parcels {
					p := &city.Parcels[i]
					if p.Building != "" || !zoneOK(p.Zone) {
						continue
					}
					if needDep != "" {
						// clear conflicting deposit and plant required one
						out := city.Deposits[:0]
						for _, d := range city.Deposits {
							if d.X == p.X && d.Y == p.Y {
								continue
							}
							out = append(out, d)
						}
						city.Deposits = out
						city.Deposits = append(city.Deposits, model.ResourceDeposit{
							ID: "t-" + ch.name + "-" + needDep, ItemID: needDep, X: p.X, Y: p.Y,
							Amount: 500, MaxAmount: 500, Quality: 3, Renewable: ruralDepositItems[needDep],
						})
					}
					return p.X, p.Y, true
				}
				return 0, 0, false
			}
			cash0 := firm.Cash
			var lastID string
			for i, st := range ch.steps {
				need := ""
				if extractBuildingResources[st.bld] {
					// derive deposit from recipe output when possible
					if r, ok := cat.RecipeByID[st.recipe]; ok && len(r.Inputs) == 0 {
						need = r.Output.ItemID
					} else if i == 0 {
						need = ch.rawDep
					}
				}
				px, py, ok := pickSite(st.bld, need)
				if !ok {
					t.Fatalf("no site for %s", st.bld)
				}
				id, err := e.placeBuilding(firm.ID, ch.city, st.bld, st.recipe, px, py, true)
				if err != nil {
					t.Fatalf("build %s: %v", st.bld, err)
				}
				lastID = id
				b := e.World.Buildings[id]
				if r, ok := cat.RecipeByID[st.recipe]; ok {
					for _, in := range r.Inputs {
						b.Inventory[in.ItemID] += in.Qty * 20
					}
				}
			}
			for tick := 0; tick < 8; tick++ {
				e.Tick()
			}
			got := 0
			for _, b := range e.World.Buildings {
				if b.FirmID == firm.ID {
					got += b.Inventory[ch.want]
				}
			}
			if got <= 0 {
				t.Fatalf("chain %s produced 0 %s; last=%s cashΔ=%.0f log=%v", ch.name, ch.want, lastID, firm.Cash-cash0, e.World.LastTickLog)
			}
		})
	}
}

func TestCatalogV2UtilityWaterAndTireSmartphone(t *testing.T) {
	cat, err := store.LoadCatalog(v2DataDir(t))
	if err != nil {
		t.Fatal(err)
	}
	if cat.Version < 2 {
		t.Skip("V2 not active")
	}
	e := NewEngine(cat)
	firm, err := e.CreatePlayer("V2Util", "tech")
	if err != nil {
		t.Fatal(err)
	}
	firm.Cash = 2_000_000
	for _, r := range cat.Recipes {
		firm.Researched[r.ID] = true
	}
	city := e.World.Cities["ufa"]
	var x, y int
	for _, p := range city.Parcels {
		if p.Building == "" && (p.Zone == "industrial" || p.Zone == "empty") {
			x, y = p.X, p.Y
			break
		}
	}
	id, err := e.placeBuilding(firm.ID, "ufa", "water_plant", "drink_01", x, y, true)
	if err != nil {
		t.Fatal(err)
	}
	cashBefore := firm.Cash
	e.Tick()
	b := e.World.Buildings[id]
	if b.Inventory["bottled_water"] <= 0 {
		t.Fatalf("water plant should produce bottled_water")
	}
	if firm.Cash >= cashBefore {
		t.Fatalf("utility water should cost cash")
	}

	// Tire + smartphone recipe existence (full chain covered lightly by inventory inject)
	if _, ok := cat.RecipeByID["auto_01"]; !ok {
		t.Fatal("missing tire recipe")
	}
	if _, ok := cat.RecipeByID["elec_01"]; !ok {
		t.Fatal("missing smartphone recipe")
	}
}
