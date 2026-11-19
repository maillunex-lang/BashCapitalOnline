package sim

import (
	"fmt"
	"math"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

// Logistics defaults: map coords are roughly km-scale across Bashkortostan.
// Ufa→Beloretsk ≈ 212 units → ~3 ticks at 80 units/tick (matches legacy routes.json).
const (
	defaultUnitsPerTick   = 80.0
	defaultCostPerUnitDist = 0.012 // ₽ per map-unit per cargo unit
	defaultLocalCost      = 0.8   // same-city last-mile
	defaultBaseCost       = 1.2
)

type routeQuote struct {
	Ticks    int
	CostUnit float64
	Distance float64
	Mode     string
}

func (e *Engine) cityDistance(from, to string) float64 {
	if from == to {
		return 0
	}
	a, okA := e.Catalog.CityByID[from]
	b, okB := e.Catalog.CityByID[to]
	if !okA || !okB {
		return 160 // unknown → ~2 ticks
	}
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Hypot(dx, dy)
}

func (e *Engine) logisticsParams() (unitsPerTick, costPerDist, localCost, baseCost float64) {
	unitsPerTick = defaultUnitsPerTick
	costPerDist = defaultCostPerUnitDist
	localCost = defaultLocalCost
	baseCost = defaultBaseCost
	eco := e.Catalog.Economy.Logistics
	if eco.UnitsPerTick > 0 {
		unitsPerTick = eco.UnitsPerTick
	}
	if eco.CostPerUnitDist > 0 {
		costPerDist = eco.CostPerUnitDist
	}
	if eco.LocalCostPerUnit > 0 {
		localCost = eco.LocalCostPerUnit
	}
	if eco.BaseCostPerUnit > 0 {
		baseCost = eco.BaseCostPerUnit
	}
	return
}

// quoteRoute — время и цена доставки по расстоянию. Всегда ≥ 1 тик.
func (e *Engine) quoteRoute(from, to string) routeQuote {
	unitsPerTick, costPerDist, localCost, baseCost := e.logisticsParams()
	dist := e.cityDistance(from, to)

	// Explicit route overrides ticks/cost when present, but never below 1 tick.
	for _, r := range e.Catalog.Routes {
		if (r.From == from && r.To == to) || (r.From == to && r.To == from) {
			ticks := r.Ticks
			if ticks < 1 {
				ticks = 1
			}
			// If route ticks look like a stub and coords exist, prefer distance ticks when larger.
			distTicks := 1
			if dist > 0 {
				distTicks = int(math.Ceil(dist / unitsPerTick))
				if distTicks < 1 {
					distTicks = 1
				}
			}
			if distTicks > ticks {
				ticks = distTicks
			}
			cost := r.CostPerUnit
			if cost <= 0 {
				cost = baseCost + dist*costPerDist
			}
			mode := r.Mode
			if mode == "" {
				mode = "road"
			}
			return routeQuote{Ticks: ticks, CostUnit: cost, Distance: dist, Mode: mode}
		}
	}

	if from == to || dist <= 0 {
		return routeQuote{Ticks: 1, CostUnit: localCost, Distance: 0, Mode: "local"}
	}
	ticks := int(math.Ceil(dist / unitsPerTick))
	if ticks < 1 {
		ticks = 1
	}
	cost := baseCost + dist*costPerDist
	return routeQuote{Ticks: ticks, CostUnit: cost, Distance: dist, Mode: "road"}
}

// routeInfo keeps legacy callers: ticks (>=1) and cost per unit.
func (e *Engine) routeInfo(from, to string) (int, float64) {
	q := e.quoteRoute(from, to)
	return q.Ticks, q.CostUnit
}

// RouteQuote / CityDistance — exported for API.
func (e *Engine) RouteQuote(from, to string) (int, float64) { return e.routeInfo(from, to) }
func (e *Engine) CityDistance(from, to string) float64      { return e.cityDistance(from, to) }

// FleetRouteQuote reports the tariff used when a firm has its own truck.
func (e *Engine) FleetRouteQuote(from, to string) (int, float64) {
	q := e.shipmentRoute(from, to, true)
	return q.Ticks, q.CostUnit
}

// TransferStock schedules a paid shipment between two buildings of the same firm.
func (e *Engine) TransferStock(firmID, fromID, toID, itemID string, qty int) (*model.Shipment, error) {
	firm, from, to := e.World.Firms[firmID], e.World.Buildings[fromID], e.World.Buildings[toID]
	if firm == nil || from == nil || to == nil || from.FirmID != firmID || to.FirmID != firmID || fromID == toID {
		return nil, fmt.Errorf("выберите два разных здания своей фирмы")
	}
	if _, ok := e.Catalog.ItemByID[itemID]; !ok || qty <= 0 {
		return nil, fmt.Errorf("выберите товар и положительное количество")
	}
	if from.Inventory[itemID] < qty {
		return nil, fmt.Errorf("на складе недостаточно товара для отправки")
	}
	ownFleet := firm.Trucks > 0
	cost := float64(qty) * e.shipmentRoute(from.CityID, to.CityID, ownFleet).CostUnit
	if firm.Cash < cost {
		return nil, fmt.Errorf("недостаточно средств на доставку")
	}
	shipment := e.createShipment(firmID, itemID, from.CityID, to.CityID, qty, from.Quality, ownFleet)
	shipment.ToBuildingID = to.ID
	from.Inventory[itemID] -= qty
	return shipment, nil
}

func (e *Engine) shipmentRoute(from, to string, ownFleet bool) routeQuote {
	q := e.quoteRoute(from, to)
	if ownFleet {
		q.CostUnit *= 0.55
		q.Ticks = max(1, q.Ticks-1)
	}
	return q
}

func (e *Engine) createShipment(firmID, itemID, from, to string, qty int, quality float64, ownFleet bool) *model.Shipment {
	if qty <= 0 {
		return nil
	}
	q := e.shipmentRoute(from, to, ownFleet)
	cost := float64(qty) * q.CostUnit
	firm := e.World.Firms[firmID]
	if firm != nil {
		if firm.Cash < cost {
			// still ship; charge what they have (debt-free MVP) — skip if broke
			if firm.Cash <= 0 {
				cost = 0
			} else {
				cost = math.Min(cost, firm.Cash)
			}
		}
		firm.Cash -= cost
	}
	id := uuid.NewString()
	s := &model.Shipment{
		ID: id, FirmID: firmID, ItemID: itemID, Qty: qty, Quality: quality,
		FromCity: from, ToCity: to, TicksLeft: q.Ticks, CostPaid: cost, OwnFleet: ownFleet,
		Distance: q.Distance, Mode: q.Mode,
	}
	e.World.Shipments[id] = s
	return s
}
