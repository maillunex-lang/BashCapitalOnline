package sim

import (
	"testing"

	"bashcapital/server/internal/model"
)

func TestDistanceDeliveryMinOneTick(t *testing.T) {
	cat := &model.Catalog{
		CityByID: map[string]model.CityDef{
			"ufa":         {ID: "ufa", X: 420, Y: 310},
			"beloretsk":   {ID: "beloretsk", X: 620, Y: 380},
			"sterlitamak": {ID: "sterlitamak", X: 480, Y: 420},
		},
		Routes: []model.RouteDef{},
		Economy: model.EconomyConfig{Logistics: model.LogisticsConfig{
			UnitsPerTick: 80, CostPerUnitDist: 0.012, LocalCostPerUnit: 0.8, BaseCostPerUnit: 1.2,
		}},
	}
	e := &Engine{Catalog: cat, World: &model.World{Firms: map[string]*model.Firm{
		"f1": {ID: "f1", Cash: 1e6},
	}, Shipments: map[string]*model.Shipment{}}}

	localTicks, _ := e.routeInfo("ufa", "ufa")
	if localTicks != 1 {
		t.Fatalf("same-city ticks=%d want 1", localTicks)
	}
	farTicks, cost := e.routeInfo("ufa", "beloretsk")
	if farTicks < 2 {
		t.Fatalf("ufa-beloretsk ticks=%d want >=2", farTicks)
	}
	if cost <= 0 {
		t.Fatalf("cost should be >0, got %v", cost)
	}
	nearTicks, _ := e.routeInfo("ufa", "sterlitamak")
	if nearTicks < 1 || nearTicks >= farTicks {
		t.Fatalf("sterlitamak should be shorter than beloretsk: near=%d far=%d", nearTicks, farTicks)
	}
	ship := e.createShipment("f1", "wheat", "ufa", "beloretsk", 10, 3, false)
	if ship == nil || ship.TicksLeft < 1 {
		t.Fatalf("shipment ticks invalid: %+v", ship)
	}
	if ship.Distance <= 0 {
		t.Fatalf("expected distance > 0")
	}
}

func TestShipmentReceiptAfterDestinationReturns(t *testing.T) {
	receiver := &model.BuildingInst{ID: "receiver", FirmID: "f1", CityID: "ufa"}
	world := &model.World{
		Tick:      40,
		Buildings: map[string]*model.BuildingInst{},
		Shipments: map[string]*model.Shipment{
			"delivery": {ID: "delivery", FirmID: "f1", ToCity: "ufa", ToBuildingID: receiver.ID, ItemID: "wheat", Qty: 3, TicksLeft: 1},
		},
	}
	e := &Engine{World: world}
	var log []string
	e.advanceShipments(&log)
	if shipment := world.Shipments["delivery"]; shipment == nil || !shipment.WaitingForDestination ||
		shipment.TicksLeft != 1 || len(world.ShipmentHistory["f1"]) != 0 {
		t.Fatal("unavailable destination must leave cargo pending without a receipt")
	}

	world.Buildings[receiver.ID] = receiver
	world.Tick++
	e.advanceShipments(&log)
	e.advanceShipments(&log)
	history := world.ShipmentHistory["f1"]
	if len(history) != 1 || history[0].WaitingForDestination || history[0].DeliveredTick != 41 ||
		world.Shipments["delivery"] != nil || receiver.Inventory["wheat"] != 3 {
		t.Fatalf("expected one delivery to the restored destination, got %+v", history)
	}

	for i := 0; i < shipmentHistoryLimit; i++ {
		world.Tick++
		e.recordShipmentReceipt(&model.Shipment{ID: "recent", FirmID: "f1"})
	}
	e.recordShipmentReceipt(&model.Shipment{ID: "other", FirmID: "f2"})
	history = world.ShipmentHistory["f1"]
	if len(history) != 20 || history[0].ID != "recent" || history[0].DeliveredTick != 42 ||
		history[len(history)-1].DeliveredTick != world.Tick || len(world.ShipmentHistory["f2"]) != 1 {
		t.Fatalf("expected separate histories capped at the last 20 receipts, got %+v", world.ShipmentHistory)
	}
}
