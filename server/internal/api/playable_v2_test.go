package api

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"bashcapital/server/internal/model"
	"bashcapital/server/internal/sim"
	"bashcapital/server/internal/store"
)

type economyPlaytest struct {
	t       *testing.T
	catalog *model.Catalog
	path    string
	db      *store.DB
	handler http.Handler
	token   string
	firm    string
}

func newEconomyPlaytest(t *testing.T, industry string) *economyPlaytest {
	t.Helper()
	catalog, err := store.LoadCatalog("../../../data")
	if err != nil || catalog.Version != 2 {
		t.Fatalf("load V2 catalog: %v", err)
	}
	p := &economyPlaytest{t: t, catalog: catalog, path: filepath.Join(t.TempDir(), "world.db")}
	p.start()
	t.Cleanup(func() {
		if p.db != nil {
			p.db.Close()
		}
	})
	var auth map[string]string
	p.post("/api/auth/register", map[string]any{"username": "playtester", "password": "test-password"}, &auth)
	p.token = auth["session_token"]
	var firm model.Firm
	p.post("/api/register", map[string]any{"name": "Игровая фирма", "industry": industry}, &firm)
	p.firm = firm.ID
	return p
}

func (p *economyPlaytest) start() {
	p.t.Helper()
	db, err := store.OpenDB(p.path)
	if err != nil {
		p.t.Fatal(err)
	}
	p.db = db
	e := sim.NewEngine(p.catalog)
	if db.HasWorld() {
		e.World, err = db.LoadWorld()
		if err != nil {
			p.t.Fatal(err)
		}
	}
	e.Store = db
	p.handler = New(e, 0, true, "test-admin", db).Handler()
}

func (p *economyPlaytest) request(method, url string, body map[string]any, code int, out any) string {
	p.t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		p.t.Fatal(err)
	}
	req := httptest.NewRequest(method, url, bytes.NewReader(data))
	if p.token != "" {
		req.Header.Set("X-Session-Token", p.token)
	}
	res := httptest.NewRecorder()
	p.handler.ServeHTTP(res, req)
	if res.Code != code {
		p.t.Fatalf("%s %s: got %d want %d: %s", method, url, res.Code, code, res.Body.String())
	}
	if out != nil {
		if err := json.Unmarshal(res.Body.Bytes(), out); err != nil {
			p.t.Fatalf("decode %s: %v", url, err)
		}
	}
	return res.Body.String()
}

func (p *economyPlaytest) post(url string, body map[string]any, out any) {
	p.t.Helper()
	p.request(http.MethodPost, url, body, 200, out)
}

func (p *economyPlaytest) world() *model.World {
	p.t.Helper()
	var world model.World
	p.request(http.MethodGet, "/api/world", nil, 200, &world)
	return &world
}

func (p *economyPlaytest) tick() *model.World {
	p.t.Helper()
	p.post("/api/tick", nil, nil)
	return p.world()
}

func (p *economyPlaytest) build(cityID, typ, recipe, resource string) string {
	p.t.Helper()
	city := p.world().Cities[cityID]
	var site *model.Parcel
	for i := range city.Parcels {
		parcel := &city.Parcels[i]
		if parcel.Building != "" {
			continue
		}
		if resource != "" {
			for _, deposit := range city.Deposits {
				if deposit.X == parcel.X && deposit.Y == parcel.Y && deposit.ItemID == resource && deposit.Amount > 0 {
					site = parcel
					break
				}
			}
		} else if parcel.Zone == "industrial" || parcel.Zone == "empty" {
			site = parcel
		}
		if site != nil && site.HasRoad {
			break
		}
	}
	if site == nil {
		p.t.Fatalf("no open site for %s (%s) in %s", typ, resource, cityID)
	}
	var b model.BuildingInst
	p.post("/api/action/build", map[string]any{
		"city_id": cityID, "building_type": typ, "recipe_id": recipe, "x": site.X, "y": site.Y,
	}, &b)
	return b.ID
}

func (p *economyPlaytest) rent(cityID, typ, item string) string {
	p.t.Helper()
	city := p.world().Cities[cityID]
	var site *model.Parcel
	for i := range city.Parcels {
		parcel := &city.Parcels[i]
		if parcel.Slots > 0 && parcel.RentedBy == "" && parcel.HasRoad && (site == nil || parcel.Slots > site.Slots) {
			site = parcel
		}
	}
	if site == nil {
		p.t.Fatal("no available retail slot")
	}
	var shop model.ShopLease
	maxPrice := p.catalog.Economy.PriceCorridors[item].Max
	p.post("/api/action/rent_shop", map[string]any{
		"city_id": cityID, "shop_type": typ, "item_id": item, "x": site.X, "y": site.Y, "price": maxPrice,
	}, &shop)
	if shop.Price != maxPrice {
		p.t.Fatalf("shop price %.2f, expected %.2f", shop.Price, maxPrice)
	}
	return shop.ID
}

func (p *economyPlaytest) restart() {
	p.t.Helper()
	if err := p.db.Close(); err != nil {
		p.t.Fatal(err)
	}
	p.db = nil
	p.start()
	p.token = ""
	var auth map[string]string
	p.post("/api/auth/login", map[string]any{"username": "playtester", "password": "test-password"}, &auth)
	if auth["firm_id"] != p.firm || auth["session_token"] == "" {
		p.t.Fatalf("login did not restore the firm: %v", auth)
	}
	p.token = auth["session_token"]
}

func TestPlayableV2Bread(t *testing.T) {
	p := newEconomyPlaytest(t, "food")
	initialCash := p.world().Firms[p.firm].Cash
	p.request(http.MethodPost, "/api/action/buy_city_market", map[string]any{"item_id": "wheat", "qty": 1}, 400, nil)
	if p.world().Firms[p.firm].Cash != initialCash {
		t.Fatal("purchase without a building debited cash")
	}
	farm := p.build("ufa", "grain_farm", "extract_wheat", "wheat")
	world := p.tick()
	production := world.Buildings[farm].Production
	if production == nil || production.Status != "produced" || production.ItemID != "wheat" ||
		production.Qty != world.Buildings[farm].Inventory["wheat"] || production.OutputTick != world.Tick {
		t.Fatalf("farm reported the wrong production state: %+v", production)
	}
	if world.Buildings[farm].Inventory["wheat"] < 4 || world.Firms[p.firm].Cash >= initialCash {
		t.Fatal("farm should produce wheat and incur expenses")
	}
	mill := p.build("ufa", "flour_mill", "food_01", "")
	world = p.tick()
	if world.Buildings[mill].Inventory["wheat_flour"] == 0 {
		t.Fatal("mill did not turn the firm's wheat into flour")
	}
	bakery := p.build("ufa", "bakery", "food_11", "")
	warehouse := p.build("sterlitamak", "warehouse", "", "")
	shop := p.rent("sterlitamak", "grocery", "wheat_bread")
	p.request(http.MethodPost, "/api/action/stock_shop", map[string]any{"shop_id": shop, "qty": 1}, 400, nil)

	world = p.tick()
	sendBread := func(world *model.World) *model.Shipment {
		t.Helper()
		qty := world.Buildings[bakery].Inventory["wheat_bread"]
		if qty == 0 {
			return nil
		}
		var shipment model.Shipment
		p.post("/api/action/transfer_stock", map[string]any{
			"from_building_id": bakery, "to_building_id": warehouse, "item_id": "wheat_bread", "qty": qty,
		}, &shipment)
		return &shipment
	}

	bad := map[string]any{
		"from_building_id": bakery, "to_building_id": warehouse, "item_id": "wheat_bread", "qty": 1000,
	}
	p.request(http.MethodPost, "/api/action/transfer_stock", bad, 400, nil)
	unchanged := p.world()
	if unchanged.Firms[p.firm].Cash != world.Firms[p.firm].Cash ||
		unchanged.Buildings[bakery].Inventory["wheat_bread"] != world.Buildings[bakery].Inventory["wheat_bread"] {
		t.Fatal("rejected shipment altered cash or inventory")
	}
	first := sendBread(world)
	if first == nil || first.FromBuildingID != bakery || first.ToBuildingID != warehouse ||
		first.DispatchedTick != world.Tick || first.TicksLeft < 1 || first.CostPaid <= 0 {
		t.Fatalf("unexpected shipment: %+v", first)
	}
	var publicQuote struct {
		Ticks int     `json:"ticks"`
		Cost  float64 `json:"cost_per_unit"`
		Fleet struct {
			Ticks int     `json:"ticks"`
			Cost  float64 `json:"cost_per_unit"`
		} `json:"fleet"`
	}
	p.request(http.MethodGet, "/api/route?from=ufa&to=sterlitamak", nil, http.StatusOK, &publicQuote)
	if first.TicksLeft != publicQuote.Ticks || math.Abs(first.CostPaid-float64(first.Qty)*publicQuote.Cost) > 0.001 {
		t.Fatal("dispatch differs from the public route quote")
	}
	if publicQuote.Fleet.Ticks != max(1, publicQuote.Ticks-1) || math.Abs(publicQuote.Fleet.Cost-publicQuote.Cost*0.55) > 0.001 {
		t.Fatalf("unexpected own-fleet quote: %+v", publicQuote.Fleet)
	}
	cashAfterDispatch := p.world().Firms[p.firm].Cash
	if math.Abs(world.Firms[p.firm].Cash-cashAfterDispatch-first.CostPaid) > 0.001 {
		t.Fatal("shipping charge was not deducted")
	}
	p.restart()
	world = p.world()
	if world.Shipments[first.ID] == nil || world.Shipments[first.ID].TicksLeft != first.TicksLeft ||
		world.Shops[shop] == nil || world.Buildings[bakery].Inventory["wheat_bread"] != 0 ||
		world.Firms[p.firm].Cash != cashAfterDispatch {
		t.Fatal("in-transit shipment, shop, inventory or cash did not survive restart")
	}

	sold := false
	var warmCash float64
	for i := 0; i < 90; i++ {
		world = p.tick()
		if i+1 < first.TicksLeft && world.Shops[shop].Stock != 0 {
			t.Fatal("bread reached the shop before the route elapsed")
		}
		if i+1 == first.TicksLeft {
			receipt := *first
			receipt.TicksLeft = 0
			want := model.ShipmentReceipt{Shipment: receipt, DeliveredTick: world.Tick}
			if world.Shipments[first.ID] != nil || len(world.ShipmentHistory[p.firm]) != 1 || world.ShipmentHistory[p.firm][0] != want {
				t.Fatal("arrived shipment was not replaced with its receipt")
			}
			p.restart()
			world = p.world()
			if world.Shipments[first.ID] != nil || len(world.ShipmentHistory[p.firm]) != 1 || world.ShipmentHistory[p.firm][0] != want {
				t.Fatal("delivery receipt did not survive restart")
			}
		}
		for _, line := range world.LastTickLog {
			if strings.Contains(line, "Розница Игровая фирма") && strings.Contains(line, "sold") && strings.Contains(line, "wheat_bread") {
				sold = true
			}
		}
		sendBread(world)
		if i == first.TicksLeft+3 {
			warmCash = p.world().Firms[p.firm].Cash
		}
		if warmCash > 0 && p.world().Firms[p.firm].Cash-warmCash >= 5000 {
			break
		}
	}
	profit := p.world().Firms[p.firm].Cash - warmCash
	if !sold || warmCash <= 0 || profit < 5000 {
		t.Fatalf("bread did not finance expansion: sold=%v net cash change %.2f, log %v", sold, profit, world.LastTickLog)
	}
	for _, parcel := range p.world().Cities["sterlitamak"].Parcels {
		if parcel.OwnerID == "" {
			p.post("/api/action/buy_parcel", map[string]any{"city_id": "sterlitamak", "x": parcel.X, "y": parcel.Y}, nil)
			if math.Abs(p.world().Firms[p.firm].Cash-(warmCash+profit-5000)) > 0.001 {
				t.Fatal("expansion did not debit the price of the parcel")
			}
			t.Logf("bread net operating cash +%.2f ₽, then reinvested 5000 ₽; first shipment %d ticks", profit, first.TicksLeft)
			return
		}
	}
	t.Fatal("no parcel to reinvest in")
}

func TestPlayableV2Fuel(t *testing.T) {
	p := newEconomyPlaytest(t, "oil")
	refinery := p.build("sterlitamak", "refinery", "oil_01", "")
	p.request(http.MethodPost, "/api/action/rent_shop", map[string]any{
		"city_id": "sterlitamak", "shop_type": "grocery", "item_id": "ai92", "price": 150,
	}, http.StatusBadRequest, nil)
	before := p.world()
	p.post("/api/action/buy_city_market", map[string]any{"item_id": "crude_oil", "qty": 5}, nil)
	world := p.world()
	pay := 5 * before.CityMarket["crude_oil"].Price
	if world.Buildings[refinery].Inventory["crude_oil"] != 5 ||
		math.Abs(before.Firms[p.firm].Cash-world.Firms[p.firm].Cash-pay) > 0.001 {
		t.Fatal("city purchase did not debit cash and credit refinery inventory")
	}
	cash := world.Firms[p.firm].Cash
	world = p.tick()
	r := p.catalog.RecipeByID["oil_01"]
	balance := p.catalog.Economy.Balance
	cost := float64(r.Energy)*balance.EnergyCostFactory + r.WaterCost*balance.WaterCostPerUnit +
		r.UtilityCost*balance.UtilityCostScale + float64(r.Labor.Worker)*balance.WagePerTick["worker"] +
		float64(r.Labor.Engineer)*balance.WagePerTick["engineer"]
	if world.Cities["sterlitamak"].LaborUsed.Worker < r.Labor.Worker ||
		world.Buildings[refinery].Inventory["crude_oil"] != 0 || world.Buildings[refinery].Inventory["gasoline_base"] != 3 ||
		math.Abs(cash-world.Firms[p.firm].Cash-cost) > 0.001 {
		t.Fatalf("refining inputs/output and wages/utilities: charged %.2f, expected %.2f", cash-world.Firms[p.firm].Cash, cost)
	}

	world = p.tick()
	if report := world.Buildings[refinery].Production; report.Status != "inputs" ||
		report.Qty != 3 || report.OutputTick >= world.Tick {
		t.Fatalf("missing inputs did not preserve the previous output: %+v", report)
	}

	p.post("/api/action/set_recipe", map[string]any{"building_id": refinery, "recipe_id": "fuel_01"}, nil)
	world = p.tick()
	if report := world.Buildings[refinery].Production; report.Status != "retooling" ||
		report.ItemID != "gasoline_base" || report.OutputTick >= world.Tick {
		t.Fatalf("retooling did not preserve the previous output: %+v", report)
	}
	world = p.tick()
	if world.Buildings[refinery].Inventory["ai92"] != 3 || world.Buildings[refinery].Inventory["gasoline_base"] != 0 {
		t.Fatalf("refinery inventory: %v", world.Buildings[refinery].Inventory)
	}
	shop := p.rent("sterlitamak", "fuel_station", "ai92")
	p.post("/api/action/stock_shop", map[string]any{"shop_id": shop, "qty": 3}, nil)
	world = p.world()
	if world.Shops[shop].Stock != 3 || world.Buildings[refinery].Inventory["ai92"] != 0 {
		t.Fatal("station did not receive the three produced units")
	}
	world = p.tick()
	if world.Shops[shop].Stock >= 3 || world.Firms[p.firm].Cash <= cash-cost {
		t.Fatalf("station did not sell its fuel: %v", world.LastTickLog)
	}
}
