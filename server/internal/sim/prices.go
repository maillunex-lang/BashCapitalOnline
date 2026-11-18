package sim

import (
	"fmt"

	"bashcapital/server/internal/model"
)

func (e *Engine) ClampPrice(itemID string, price float64) (float64, error) {
	c, ok := e.Catalog.Economy.PriceCorridors[itemID]
	if !ok {
		if price <= 0 {
			return 0, fmt.Errorf("цена должна быть > 0")
		}
		if price > 100000 {
			return 0, fmt.Errorf("цена вне разумного коридора")
		}
		return price, nil
	}
	if price < c.Min || price > c.Max {
		return 0, fmt.Errorf("цена %s должна быть в коридоре %.0f…%.0f ₽", itemID, c.Min, c.Max)
	}
	return price, nil
}

// SnapPriceToCorridor clamps price into the item corridor (mid if out of range or <=0).
// Used for retail lease setup so a stale client default (e.g. 55₽) still rents successfully.
func (e *Engine) SnapPriceToCorridor(itemID string, price float64) float64 {
	c, ok := e.Catalog.Economy.PriceCorridors[itemID]
	if !ok {
		if price <= 0 {
			return 50
		}
		return price
	}
	if price < c.Min || price > c.Max {
		return (c.Min + c.Max) / 2
	}
	return price
}

func (e *Engine) initCityMarket() {
	e.World.CityMarket = map[string]model.CityMarketStock{}
	for _, o := range e.Catalog.Economy.CityMarket {
		e.World.CityMarket[o.ItemID] = model.CityMarketStock{
			ItemID: o.ItemID, Price: o.Price, Quality: o.Quality, Stock: o.StockPerTick * 3,
		}
	}
}

func (e *Engine) refreshCityMarket() {
	for _, o := range e.Catalog.Economy.CityMarket {
		cur := e.World.CityMarket[o.ItemID]
		cur.ItemID = o.ItemID
		cur.Price = o.Price
		cur.Quality = o.Quality
		cur.Stock = min(o.StockPerTick*5, cur.Stock+o.StockPerTick)
		e.World.CityMarket[o.ItemID] = cur
	}
}

func (e *Engine) BuyFromCity(firmID, itemID string, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("qty")
	}
	firm := e.World.Firms[firmID]
	offer, ok := e.World.CityMarket[itemID]
	if firm == nil || !ok {
		return fmt.Errorf("нет предложения города")
	}
	if offer.Stock < qty {
		return fmt.Errorf("у города мало товара (осталось %d)", offer.Stock)
	}
	pay := float64(qty) * offer.Price
	if firm.Cash < pay {
		return fmt.Errorf("недостаточно средств")
	}
	var target *model.BuildingInst
	for _, b := range e.World.Buildings {
		if b.FirmID == firmID {
			target = b
			break
		}
	}
	if target == nil {
		return fmt.Errorf("нужен хотя бы один склад/здание для приёмки")
	}
	firm.Cash -= pay
	offer.Stock -= qty
	e.World.CityMarket[itemID] = offer
	if target.Inventory == nil {
		target.Inventory = map[string]int{}
	}
	target.Inventory[itemID] += qty
	target.Quality = clamp(target.Quality*0.7+offer.Quality*0.3, 1, 5)
	e.World.Events = append(e.World.Events, fmt.Sprintf("%s купил у города %d %s (качество %.1f)", firm.Name, qty, itemID, offer.Quality))
	return nil
}
