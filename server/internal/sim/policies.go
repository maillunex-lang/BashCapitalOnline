package sim

import (
	"fmt"
	"strings"

	"bashcapital/server/internal/model"
)

func (e *Engine) ShopAllowsItem(shopType, itemID string) error {
	def, ok := e.Catalog.ShopByID[shopType]
	if !ok {
		return fmt.Errorf("неизвестный тип магазина «%s»", shopType)
	}
	item, ok := e.Catalog.ItemByID[itemID]
	if !ok {
		return fmt.Errorf("неизвестный товар")
	}
	if len(def.AllowedItems) > 0 {
		for _, id := range def.AllowedItems {
			if id == itemID {
				return nil
			}
		}
		return fmt.Errorf("%s не продаёт «%s» — выберите товар из ассортимента", def.Name, item.Name)
	}
	tierOK := len(def.AllowedTiers) == 0
	for _, t := range def.AllowedTiers {
		if t == item.Tier {
			tierOK = true
			break
		}
	}
	if !tierOK {
		return fmt.Errorf("%s не торгует товарами уровня «%s»", def.Name, item.Tier)
	}
	catOK := false
	for _, c := range def.AllowedCategories {
		if c == item.Category {
			catOK = true
			break
		}
	}
	if !catOK {
		return fmt.Errorf("%s не продаёт категорию «%s» (как в Capitalism Lab: специализация витрины)", def.Name, item.Category)
	}
	return nil
}

func (e *Engine) RecipeFitsBuilding(buildingType, recipeID string) error {
	r, ok := e.Catalog.RecipeByID[recipeID]
	if !ok {
		return fmt.Errorf("рецепт не найден")
	}
	if r.BuildingType != buildingType {
		bname := buildingType
		if bd, ok := e.Catalog.BuildingByID[buildingType]; ok {
			bname = bd.Name
		}
		return fmt.Errorf("%s не может выполнять «%s» — только рецепты своего типа здания", bname, r.Name)
	}
	return nil
}

func (e *Engine) firmAllowsExternalSale(firmID, itemID string) bool {
	// If any building of the firm that holds the item has SellExternal=false and owns stock, block.
	// Policy: firm-level effective = false if ALL stock-bearing buildings forbid, OR if firm has explicit
	// sell_external false on the building we take from. Simpler: if firm has any building with
	// SellExternal=false for this item's production chain, check takeInventory sources.
	for _, b := range e.World.Buildings {
		if b.FirmID != firmID {
			continue
		}
		if b.Inventory[itemID] > 0 && !b.SellExternal {
			return false
		}
	}
	// If firm has global preference via buildings with stock elsewhere allowing — still allow
	// only when some stock exists on SellExternal=true buildings.
	hasStock := false
	hasOpen := false
	for _, b := range e.World.Buildings {
		if b.FirmID != firmID || b.Inventory[itemID] <= 0 {
			continue
		}
		hasStock = true
		if b.SellExternal {
			hasOpen = true
		}
	}
	if !hasStock {
		return true
	}
	return hasOpen
}

func (e *Engine) takeInventoryRespectingSales(firmID, cityID, itemID string, qty int, externalBuyer bool) int {
	if externalBuyer && !e.firmAllowsExternalSale(firmID, itemID) {
		return 0
	}
	left := qty
	for _, b := range e.World.Buildings {
		if b.FirmID != firmID || b.CityID != cityID {
			continue
		}
		if externalBuyer && !b.SellExternal {
			continue
		}
		have := b.Inventory[itemID]
		take := min(have, left)
		b.Inventory[itemID] -= take
		left -= take
		if left == 0 {
			break
		}
	}
	if left > 0 {
		for _, b := range e.World.Buildings {
			if b.FirmID != firmID {
				continue
			}
			if externalBuyer && !b.SellExternal {
				continue
			}
			have := b.Inventory[itemID]
			take := min(have, left)
			b.Inventory[itemID] -= take
			left -= take
			if left == 0 {
				break
			}
		}
	}
	return qty - left
}

func (e *Engine) inputSourceFor(b *model.BuildingInst, itemID string) model.InputSource {
	if b.InputSources != nil {
		if s, ok := b.InputSources[itemID]; ok && s.Mode != "" {
			return s
		}
	}
	if b.DefaultInput.Mode != "" {
		return b.DefaultInput
	}
	return model.InputSource{Mode: "own"}
}

// procureInputs tries to fill missing recipe inputs according to building buy policy.
func (e *Engine) procureInputs(b *model.BuildingInst, recipe model.RecipeDef, log *[]string) {
	if !b.AutoProcure {
		return
	}
	firm := e.World.Firms[b.FirmID]
	if firm == nil {
		return
	}
	for _, in := range recipe.Inputs {
		have := b.Inventory[in.ItemID]
		need := in.Qty - have
		if need <= 0 {
			continue
		}
		src := e.inputSourceFor(b, in.ItemID)
		switch src.Mode {
		case "own":
			moved := e.pullFromOtherBuildings(b, in.ItemID, need)
			b.Inventory[in.ItemID] += moved
		case "city":
			_ = e.buyCityIntoBuilding(b, in.ItemID, need, log)
		case "market":
			e.buyMarketIntoBuilding(b, in.ItemID, need, "", log)
		case "firm":
			e.buyMarketIntoBuilding(b, in.ItemID, need, src.FirmID, log)
		default:
			moved := e.pullFromOtherBuildings(b, in.ItemID, need)
			b.Inventory[in.ItemID] += moved
		}
	}
}

func (e *Engine) buyCityIntoBuilding(b *model.BuildingInst, itemID string, qty int, log *[]string) int {
	firm := e.World.Firms[b.FirmID]
	offer, ok := e.World.CityMarket[itemID]
	if firm == nil || !ok || qty <= 0 {
		return 0
	}
	take := min(qty, offer.Stock)
	if take <= 0 {
		return 0
	}
	pay := float64(take) * offer.Price
	if firm.Cash < pay {
		take = int(firm.Cash / offer.Price)
		if take <= 0 {
			return 0
		}
		pay = float64(take) * offer.Price
	}
	firm.Cash -= pay
	offer.Stock -= take
	e.World.CityMarket[itemID] = offer
	b.Inventory[itemID] += take
	b.Quality = clamp((b.Quality*0.85)+(offer.Quality*0.15), 1, 5)
	if log != nil {
		*log = append(*log, fmt.Sprintf("%s закупил у города %d %s", firm.Name, take, itemID))
	}
	return take
}

func (e *Engine) buyMarketIntoBuilding(b *model.BuildingInst, itemID string, qty int, onlyFirm string, log *[]string) int {
	firm := e.World.Firms[b.FirmID]
	if firm == nil || qty <= 0 {
		return 0
	}
	bought := 0
	for _, o := range e.World.Orders {
		if bought >= qty {
			break
		}
		if o.Side != "sell" || o.ItemID != itemID || o.CityID != b.CityID {
			continue
		}
		if o.FirmID == b.FirmID {
			continue
		}
		if onlyFirm != "" && o.FirmID != onlyFirm {
			continue
		}
		remain := o.Qty - o.Filled
		if remain <= 0 {
			continue
		}
		if !e.firmAllowsExternalSale(o.FirmID, itemID) {
			continue
		}
		want := min(qty-bought, remain)
		taken := e.takeInventoryRespectingSales(o.FirmID, o.CityID, itemID, want, true)
		if taken <= 0 {
			continue
		}
		pay := float64(taken) * o.Price
		if firm.Cash < pay {
			e.putInventory(o.FirmID, o.CityID, itemID, taken)
			continue
		}
		firm.Cash -= pay
		if seller := e.World.Firms[o.FirmID]; seller != nil {
			seller.Cash += pay
		}
		b.Inventory[itemID] += taken
		o.Filled += taken
		bought += taken
		if log != nil {
			*log = append(*log, fmt.Sprintf("%s купил на бирже %d %s у %s", firm.Name, taken, itemID, o.FirmID[:8]))
		}
	}
	return bought
}

func (e *Engine) pullFromOtherBuildings(target *model.BuildingInst, itemID string, qty int) int {
	left := qty
	for _, b := range e.World.Buildings {
		if b.ID == target.ID || b.FirmID != target.FirmID || b.CityID != target.CityID {
			continue
		}
		have := b.Inventory[itemID]
		take := min(have, left)
		b.Inventory[itemID] -= take
		left -= take
		if left == 0 {
			break
		}
	}
	return qty - left
}

func normalizeInputMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "own", "self", "warehouse":
		return "own"
	case "city", "town":
		return "city"
	case "market", "exchange", "players":
		return "market"
	case "firm", "player", "supplier":
		return "firm"
	default:
		return "own"
	}
}
