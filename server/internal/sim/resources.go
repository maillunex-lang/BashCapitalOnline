package sim

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

// extractBuildingResources — здания добычи/сбора (каталог V1+V2).
var extractBuildingResources = map[string]bool{
	"mine": true, "quarry": true, "oil_well": true, "gas_well": true, "peat_field": true,
	"grain_farm": true, "crop_farm": true, "oilseed_farm": true, "fiber_farm": true, "vegetable_farm": true,
	"dairy_farm": true, "livestock_farm": true, "poultry_farm": true, "sheep_farm": true, "horse_farm": true,
	"logging_camp": true, "forest_gathering": true, "apiary": true, "orchard": true,
	"farm": true, "greenhouse": true, "chem_plant": true,
}

var ruralDepositItems = map[string]bool{
	"wheat": true, "rye": true, "barley": true, "oats": true, "buckwheat": true,
	"potato": true, "sugar_beet": true, "sunflower": true, "flax": true, "vegetables": true,
	"milk": true, "beef": true, "pork": true, "poultry": true, "eggs": true, "wool": true, "raw_hide": true,
	"softwood": true, "hardwood": true, "wood_chips": true, "forest_berries": true, "mushrooms": true,
	"honey": true, "apples": true, "cherries": true, "horse_meat": true, "mare_milk": true,
	"cotton": true, "cattle": true, "cherry": true, "apple": true, "wood": true,
}

func resourceBaseID(id string) string {
	cur := id
	for {
		trimmed := false
		for _, s := range []string{"_export", "_prem", "_eco", "_std", "_ufa", "_sterli", "_beloretsk"} {
			if strings.HasSuffix(cur, s) {
				cur = strings.TrimSuffix(cur, s)
				trimmed = true
				break
			}
		}
		if !trimmed {
			return cur
		}
	}
}

func depositMatchesNeed(depositItemID, needItemID string) bool {
	if needItemID == "" {
		return true
	}
	if depositItemID == needItemID {
		return true
	}
	return resourceBaseID(depositItemID) == resourceBaseID(needItemID)
}

func (e *Engine) depositItemForBuild(buildingType, recipeID string) string {
	if recipeID != "" {
		if r, ok := e.Catalog.RecipeByID[recipeID]; ok && isExtractRecipe(r) {
			return r.Output.ItemID
		}
	}
	switch buildingType {
	case "orchard":
		return "apples"
	case "logging_camp":
		return "softwood"
	case "mine":
		return "iron_ore"
	case "oil_well":
		return "crude_oil"
	case "gas_well":
		return "natural_gas"
	case "peat_field":
		return "peat"
	case "dairy_farm":
		return "milk"
	case "livestock_farm":
		return "beef"
	case "apiary":
		return "honey"
	case "grain_farm", "farm":
		return "wheat"
	}
	return ""
}

func isExtractRecipe(recipe model.RecipeDef) bool {
	if recipe.ExtractMode == "depleting" || recipe.ExtractMode == "renewable" {
		return true
	}
	return len(recipe.Inputs) == 0 && recipe.Output.ItemID != "" && recipe.WaterCost == 0
}

func isUtilityExtract(recipe model.RecipeDef) bool {
	return len(recipe.Inputs) == 0 && recipe.WaterCost > 0 && recipe.ExtractMode != "depleting" && recipe.ExtractMode != "renewable"
}

func (e *Engine) seedResourceDeposits() {
	for _, city := range e.World.Cities {
		def := e.Catalog.CityByID[city.ID]
		city.Deposits = nil
		for _, res := range def.Resources {
			if it, ok := e.Catalog.ItemByID[res]; ok && it.Tier != "raw" {
				continue
			}
			n := 2
			switch res {
			case "iron_ore", "crude_oil", "coal", "wheat", "milk", "oil", "gas":
				n = 3
			}
			for i := 0; i < n; i++ {
				e.spawnDeposit(city, res, true)
			}
		}
		if len(city.Deposits) == 0 {
			e.spawnDeposit(city, "softwood", true)
			e.spawnDeposit(city, "sand", true)
		}
	}
}

func (e *Engine) spawnDeposit(city *model.CityState, itemID string, initial bool) *model.ResourceDeposit {
	renewableHint := ruralDepositItems[itemID]
	switch itemID {
	case "iron_ore", "crude_oil", "coal", "copper_ore", "natural_gas", "oil", "gas", "bauxite":
		renewableHint = false
	case "softwood", "hardwood", "wood", "flax", "wheat", "potato", "milk", "honey", "apples", "cherries":
		renewableHint = true
	case "sand", "quartz", "stone", "clay", "limestone", "peat", "salt":
		renewableHint = itemID == "peat"
	}
	pick := func(preferZones map[string]bool) []model.Parcel {
		out := []model.Parcel{}
		for _, p := range city.Parcels {
			if p.Building != "" || depositAt(city, p.X, p.Y) != nil {
				continue
			}
			if p.Zone == "center" {
				continue
			}
			if p.Zone == "residential" && !renewableHint {
				continue
			}
			if preferZones != nil && !preferZones[p.Zone] {
				continue
			}
			out = append(out, p)
		}
		return out
	}
	var candidates []model.Parcel
	if renewableHint {
		candidates = pick(map[string]bool{"rural": true, "empty": true})
	} else {
		candidates = pick(map[string]bool{"industrial": true, "empty": true})
	}
	if len(candidates) == 0 {
		candidates = pick(nil)
	}
	if len(candidates) == 0 {
		for _, p := range city.Parcels {
			if p.Building == "" && depositAt(city, p.X, p.Y) == nil {
				candidates = append(candidates, p)
			}
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	p := candidates[rand.Intn(len(candidates))]
	base := 80 + rand.Intn(120)
	renewable := renewableHint
	switch itemID {
	case "iron_ore", "crude_oil", "coal", "copper_ore", "natural_gas", "oil", "gas", "bauxite":
		base = 150 + rand.Intn(200)
		renewable = false
	case "softwood", "hardwood", "wood", "flax", "wheat", "potato", "milk", "honey", "apples", "cherries":
		base = 100 + rand.Intn(80)
		renewable = true
	case "sand", "quartz", "stone", "clay", "limestone", "peat", "salt":
		base = 100 + rand.Intn(100)
		renewable = itemID == "peat"
	}
	d := model.ResourceDeposit{
		ID: uuid.NewString(), ItemID: itemID, X: p.X, Y: p.Y,
		Amount: base, MaxAmount: base, Quality: 2.5 + rand.Float64()*2.0, Renewable: renewable,
	}
	city.Deposits = append(city.Deposits, d)
	if !initial {
		e.World.Events = append(e.World.Events, fmt.Sprintf("Новая залежь %s в %s (%d,%d), запас %d", itemID, city.ID, d.X, d.Y, d.Amount))
	}
	return &city.Deposits[len(city.Deposits)-1]
}

func depositAt(city *model.CityState, x, y int) *model.ResourceDeposit {
	for i := range city.Deposits {
		if city.Deposits[i].X == x && city.Deposits[i].Y == y && city.Deposits[i].Amount > 0 {
			return &city.Deposits[i]
		}
	}
	return nil
}

func depositAtAny(city *model.CityState, x, y int) *model.ResourceDeposit {
	for i := range city.Deposits {
		if city.Deposits[i].X == x && city.Deposits[i].Y == y {
			return &city.Deposits[i]
		}
	}
	return nil
}

// ensureOpenExtractDeposits spawns one free deposit per city resource when AI
// seeding occupied every matching tile — players must still be able to extract.
func (e *Engine) ensureOpenExtractDeposits() {
	for _, city := range e.World.Cities {
		def := e.Catalog.CityByID[city.ID]
		for _, res := range def.Resources {
			if it, ok := e.Catalog.ItemByID[res]; ok && it.Tier != "raw" {
				continue
			}
			hasFree := false
			for _, d := range city.Deposits {
				if resourceBaseID(d.ItemID) != resourceBaseID(res) {
					continue
				}
				occ := false
				for _, p := range city.Parcels {
					if p.X == d.X && p.Y == d.Y && p.Building != "" {
						occ = true
						break
					}
				}
				if !occ {
					hasFree = true
					break
				}
			}
			if !hasFree {
				e.spawnDeposit(city, res, true)
			}
		}
	}
}

func (e *Engine) requireDepositForBuild(city *model.CityState, buildingType, recipeID string, x, y int) error {
	if !extractBuildingResources[buildingType] {
		return nil
	}
	if recipeID != "" {
		if r, ok := e.Catalog.RecipeByID[recipeID]; ok && isUtilityExtract(r) {
			return nil
		}
	}
	need := e.depositItemForBuild(buildingType, recipeID)
	if need == "" {
		d := depositAt(city, x, y)
		if d == nil {
			return fmt.Errorf("нужна залежь/участок ресурса на этой клетке")
		}
		return nil
	}
	d := depositAt(city, x, y)
	if d == nil || !depositMatchesNeed(d.ItemID, need) {
		name := resourceBaseID(need)
		if it, ok := e.Catalog.ItemByID[name]; ok {
			name = it.Name
		} else if it, ok := e.Catalog.ItemByID[need]; ok {
			name = it.Name
		}
		return fmt.Errorf("нужна залежь «%s» на клетке (сгенерированы на окраине)", name)
	}
	return nil
}

func (e *Engine) consumeDeposit(city *model.CityState, x, y, qty int, itemID string, log *[]string) (float64, bool) {
	d := depositAt(city, x, y)
	if d == nil || !depositMatchesNeed(d.ItemID, itemID) {
		return 0, false
	}
	if d.Amount < qty {
		qty = d.Amount
	}
	if qty <= 0 {
		return d.Quality, false
	}
	d.Amount -= qty
	q := d.Quality
	spawnID := resourceBaseID(d.ItemID)
	if d.Renewable {
		regen := d.MaxAmount / 8
		if regen < 1 {
			regen = 1
		}
		d.Amount += regen
		if d.Amount > d.MaxAmount {
			d.Amount = d.MaxAmount
		}
		return q, true
	}
	if d.Amount <= 0 {
		d.Amount = 0
		*log = append(*log, fmt.Sprintf("Залежь %s в %s (%d,%d) исчерпана", spawnID, city.ID, x, y))
		e.removeEmptyDeposits(city)
		if nd := e.spawnDeposit(city, spawnID, false); nd != nil {
			*log = append(*log, fmt.Sprintf("Ресурс %s появился в другом месте (%d,%d)", spawnID, nd.X, nd.Y))
		}
	}
	return q, true
}

func (e *Engine) removeEmptyDeposits(city *model.CityState) {
	out := city.Deposits[:0]
	for _, d := range city.Deposits {
		if d.Amount > 0 {
			out = append(out, d)
		}
	}
	city.Deposits = out
}

func (e *Engine) discoverResources(log *[]string) {
	for _, city := range e.World.Cities {
		def := e.Catalog.CityByID[city.ID]
		if len(def.Resources) == 0 || rand.Float64() > 0.45 {
			continue
		}
		res := def.Resources[rand.Intn(len(def.Resources))]
		if it, ok := e.Catalog.ItemByID[res]; ok && it.Tier != "raw" {
			continue
		}
		if d := e.spawnDeposit(city, res, false); d != nil {
			*log = append(*log, fmt.Sprintf("Геологи нашли %s у %s", res, city.ID))
		}
	}
}

func (e *Engine) opsCost(recipe model.RecipeDef, extract bool) float64 {
	energyCost := float64(recipe.Energy) * e.energyRate(extract)
	waterUnit := e.Catalog.Economy.Balance.WaterCostPerUnit
	if waterUnit <= 0 {
		waterUnit = 2.5
	}
	scale := e.Catalog.Economy.Balance.UtilityCostScale
	if scale <= 0 {
		scale = 1
	}
	return energyCost + recipe.WaterCost*waterUnit + recipe.UtilityCost*scale
}
