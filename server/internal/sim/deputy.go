package sim

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

var deputyTaskCatalog = []model.DeputyTask{
	{Kind: "procure_inputs", Label: "Искать поставщиков входов для заводов"},
	{Kind: "restock", Label: "Пополнять магазины со склада"},
	{Kind: "price_retail", Label: "Выставлять розничные цены"},
	{Kind: "sell_excess", Label: "Сбывать излишки на бирже"},
	{Kind: "research", Label: "Исследовать доступные рецепты"},
	{Kind: "auto_contracts", Label: "Закрывать дыры контрактами на месяц"},
}

func (e *Engine) ActivateSubscription(firmID string, months int) error {
	firm := e.World.Firms[firmID]
	if firm == nil || firm.IsAI {
		return fmt.Errorf("фирма не найдена")
	}
	if months <= 0 {
		months = 1
	}
	// Mock donate: in production replace with payment webhook (YooKassa etc.)
	cost := 490.0 * float64(months) // decorative price in soft currency for MVP mock
	if firm.Cash < cost {
		// allow activating even with low cash in mock via "donate flag" — still charge if possible
		cost = 0
	}
	firm.Cash -= cost
	extend := months * 24 // 24 hourly ticks per game month
	if firm.SubscriptionActive && firm.SubscriptionUntil > e.World.Tick {
		firm.SubscriptionUntil += extend
	} else {
		firm.SubscriptionUntil = e.World.Tick + extend
	}
	firm.SubscriptionActive = true
	e.World.Events = append(e.World.Events, fmt.Sprintf("%s оформил подписку до тика %d", firm.Name, firm.SubscriptionUntil))
	return nil
}

func (e *Engine) HireDeputy(firmID, name string) error {
	firm := e.World.Firms[firmID]
	if firm == nil {
		return fmt.Errorf("фирма не найдена")
	}
	if !e.subscriptionOK(firm) {
		return fmt.Errorf("нужна активная подписка")
	}
	if strings.TrimSpace(name) == "" {
		name = "Заместитель"
	}
	if firm.Deputy == nil {
		firm.Deputy = &model.Deputy{Hired: true, Name: name, Tasks: []model.DeputyTask{}}
	} else {
		firm.Deputy.Hired = true
		firm.Deputy.Name = name
	}
	return nil
}

func (e *Engine) SetDeputyTasks(firmID string, kinds []string) error {
	firm := e.World.Firms[firmID]
	if firm == nil {
		return fmt.Errorf("фирма не найдена")
	}
	if !e.subscriptionOK(firm) || firm.Deputy == nil || !firm.Deputy.Hired {
		return fmt.Errorf("заместитель недоступен без подписки")
	}
	var tasks []model.DeputyTask
	for _, kind := range kinds {
		label := kind
		for _, c := range deputyTaskCatalog {
			if c.Kind == kind {
				label = c.Label
				break
			}
		}
		tasks = append(tasks, model.DeputyTask{
			ID: uuid.NewString(), Kind: kind, Label: label, Active: true, Params: map[string]string{},
		})
	}
	firm.Deputy.Tasks = tasks
	return nil
}

func (e *Engine) DeputyCatalog() []model.DeputyTask {
	return deputyTaskCatalog
}

func (e *Engine) subscriptionOK(f *model.Firm) bool {
	if f == nil {
		return false
	}
	if f.SubscriptionUntil > 0 && e.World.Tick > f.SubscriptionUntil {
		f.SubscriptionActive = false
		if f.Deputy != nil {
			f.Deputy.Hired = false
		}
		return false
	}
	return f.SubscriptionActive
}

func (e *Engine) runDeputies(log *[]string) {
	for _, firm := range e.World.Firms {
		if firm.IsAI || !e.subscriptionOK(firm) || firm.Deputy == nil || !firm.Deputy.Hired {
			continue
		}
		for i := range firm.Deputy.Tasks {
			t := &firm.Deputy.Tasks[i]
			if !t.Active {
				continue
			}
			note := e.runDeputyTask(firm, t)
			t.LastNote = note
			if note != "" {
				*log = append(*log, fmt.Sprintf("Заместитель «%s»: %s", firm.Deputy.Name, note))
			}
		}
	}
}

func (e *Engine) runDeputyTask(firm *model.Firm, t *model.DeputyTask) string {
	switch t.Kind {
	case "restock":
		n := 0
		for _, shop := range e.World.Shops {
			if shop.FirmID != firm.ID || shop.Stock >= 12 {
				continue
			}
			need := 12 - shop.Stock
			taken := e.takeInventory(firm.ID, shop.CityID, shop.ItemID, need)
			shop.Stock += taken
			n += taken
		}
		if n == 0 {
			return "магазины укомплектованы или нет товара"
		}
		return fmt.Sprintf("пополнил магазины на %d ед.", n)
	case "price_retail":
		for _, shop := range e.World.Shops {
			if shop.FirmID != firm.ID {
				continue
			}
			item := e.Catalog.ItemByID[shop.ItemID]
			base := 40.0
			switch item.Tier {
			case "intermediate":
				base = 55
			case "finished":
				base = 75
			}
			switch item.Category {
			case "home":
				base = 180
			case "apparel":
				base = 120
			case "tech":
				base = 250
			case "fuel":
				base = 90
			case "food":
				base = 55
			}
			if corr, ok := e.Catalog.Economy.PriceCorridors[shop.ItemID]; ok {
				mid := (corr.Min + corr.Max) / 2
				base = mid
			}
			shop.Price = base * (0.9 + firm.BrandAwareness[item.Category]*0.4)
			if corr, ok := e.Catalog.Economy.PriceCorridors[shop.ItemID]; ok {
				if shop.Price < corr.Min {
					shop.Price = corr.Min
				}
				if shop.Price > corr.Max {
					shop.Price = corr.Max
				}
			}
		}
		return "обновил розничные цены"
	case "sell_excess":
		sold := 0
		for _, b := range e.World.Buildings {
			if b.FirmID != firm.ID || !b.SellExternal {
				continue
			}
			for item, qty := range b.Inventory {
				idef := e.Catalog.ItemByID[item]
				if idef.Tier == "raw" || qty < 20 {
					continue
				}
				sellQty := qty / 4
				b.Inventory[item] -= sellQty
				id := uuid.NewString()
				price := 25.0
				if idef.Tier == "finished" {
					price = 75
				} else if idef.Tier == "intermediate" {
					price = 40
				}
				e.World.Orders[id] = &model.MarketOrder{
					ID: id, FirmID: firm.ID, CityID: b.CityID, ItemID: item, Side: "sell",
					Qty: sellQty, Price: price, Quality: b.Quality,
				}
				sold += sellQty
			}
		}
		if sold == 0 {
			return "излишков нет"
		}
		return fmt.Sprintf("выставил на биржу %d ед.", sold)
	case "procure_inputs":
		placed := 0
		for _, b := range e.World.Buildings {
			if b.FirmID != firm.ID {
				continue
			}
			r, ok := e.Catalog.RecipeByID[b.RecipeID]
			if !ok {
				continue
			}
			for _, in := range r.Inputs {
				have := b.Inventory[in.ItemID]
				if have >= in.Qty*3 {
					continue
				}
				need := in.Qty*5 - have
				id := uuid.NewString()
				e.World.Orders[id] = &model.MarketOrder{
					ID: id, FirmID: firm.ID, CityID: b.CityID, ItemID: in.ItemID, Side: "buy",
					Qty: need, Price: 55, Quality: 3,
				}
				placed++
			}
		}
		if placed == 0 {
			return "входы в норме"
		}
		return fmt.Sprintf("разместил %d заявок на закупку входов", placed)
	case "research":
		for _, r := range e.Catalog.Recipes {
			if r.ResearchCost <= 0 || firm.Researched[r.ID] {
				continue
			}
			if firm.Cash < float64(r.ResearchCost) {
				continue
			}
			firm.Cash -= float64(r.ResearchCost)
			firm.Researched[r.ID] = true
			return fmt.Sprintf("исследовал «%s»", r.Name)
		}
		return "новых исследований нет"
	case "auto_contracts":
		// find sell orders from AI and contract if we need that item
		needed := map[string]bool{}
		homeCity := ""
		for _, b := range e.World.Buildings {
			if b.FirmID != firm.ID {
				continue
			}
			homeCity = b.CityID
			r, ok := e.Catalog.RecipeByID[b.RecipeID]
			if !ok {
				continue
			}
			for _, in := range r.Inputs {
				if b.Inventory[in.ItemID] < in.Qty*2 {
					needed[in.ItemID] = true
				}
			}
		}
		made := 0
		for itemID := range needed {
			for _, o := range e.World.Orders {
				if o.Side != "sell" || o.ItemID != itemID || o.FirmID == firm.ID {
					continue
				}
				qty := min(10, o.Qty-o.Filled)
				if qty <= 0 {
					continue
				}
				_, err := e.CreateContract(o.FirmID, firm.ID, itemID, o.CityID, homeCity, qty, o.Price, 4, firm.Trucks > 0)
				if err == nil {
					made++
					break
				}
			}
		}
		if made == 0 {
			return "подходящих контрактов не нашёл"
		}
		return fmt.Sprintf("заключил %d контракт(ов)", made)
	default:
		return ""
	}
}
