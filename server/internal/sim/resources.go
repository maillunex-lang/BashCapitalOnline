package sim

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

// extractBuildingResources — здания добычи/сбора (могут требовать залежь).
var extractBuildingResources = map[string]bool{
	"farm":           true,
	"livestock_farm": true,
	"orchard":        true,
	"logging_camp":   true,
	"quarry":         true,
	"mine":           true,
	"oil_well":       true,
	"gas_well":       true,
	"chem_plant":     true, // базовая химия как «залежь»
}

var ruralDepositItems = map[string]bool{
	"wheat": true, "potato": true, "flax": true, "cotton": true,
	"milk": true, "cattle": true, "cherry": true, "apple": true,
}

func resourceForRecipe(buildingType, recipeID string) string {
	if recipeID != "" {
		// Prefer recipe output for extract recipes (Capitalism Lab style: many raws per building).
		return ""
	}
	_ = buildingType
	return ""
}

func (e *Engine) depositItemForBuild(buildingType, recipeID string) string {
	if recipeID != "" {
		if r, ok := e.Catalog.RecipeByID[recipeID]; ok && isExtractRecipe(r) {
			return r.Output.ItemID
		}
	}
	// fallback defaults by building
	switch buildingType {
	case "orchard":
		return "cherry"
	case "logging_camp":
		return "wood"
	case "mine":
		return "iron_ore"
	case "oil_well":
		return "oil"
	case "gas_well":
		return "gas"
	case "chem_plant":
		return "chemicals"
	case "livestock_farm":
		return "milk"
	}
	return ""
}

func isExtractRecipe(recipe model.RecipeDef) bool {
	return len(recipe.Inputs) == 0 && recipe.Output.ItemID != ""
}

func (e *Engine) seedResourceDeposits() {
	for _, city := range e.World.Cities {
		def := e.Catalog.CityByID[city.ID]
		city.Deposits = nil
		for _, res := range def.Resources {
			// skip non-raw catalog items accidentally listed
			if it, ok := e.Catalog.ItemByID[res]; ok && it.Tier != "raw" {
				continue
			}
			n := 2
			switch res {
			case "iron_ore", "oil", "coal", "wheat", "milk":
				n = 3
			}
			for i := 0; i < n; i++ {
				e.spawnDeposit(city, res, true)
			}
		}
		if len(city.Deposits) == 0 {
			e.spawnDeposit(city, "wood", true)
			e.spawnDeposit(city, "sand", true)
		}
	}
}

func (e *Engine) spawnDeposit(city *model.CityState, itemID string, initial bool) *model.ResourceDeposit {
	candidates := []model.Parcel{}
	for _, p := range city.Parcels {
		if p.Building != "" {
			continue
		}
		if depositAt(city, p.X, p.Y) != nil {
			continue
		}
		if p.Zone == "center" {
			continue
		}
		if p.Zone == "residential" && !ruralDepositItems[itemID] {
			continue
		}
		candidates = append(candidates, p)
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
	switch itemID {
	case "iron_ore", "oil", "coal", "copper_ore", "gas":
		base = 150 + rand.Intn(200)
	case "wood", "flax", "cherry", "milk", "wheat", "potato", "apple", "cattle", "cotton":
		base = 60 + rand.Intn(80)
	case "sand", "quartz":
		base = 100 + rand.Intn(100)
	}
	d := model.ResourceDeposit{
		ID:        uuid.NewString(),
		ItemID:    itemID,
		X:         p.X,
		Y:         p.Y,
		Amount:    base,
		MaxAmount: base,
		Quality:   2.5 + rand.Float64()*2.0,
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

func (e *Engine) requireDepositForBuild(city *model.CityState, buildingType, recipeID string, x, y int) error {
	if !extractBuildingResources[buildingType] {
		return nil
	}
	need := e.depositItemForBuild(buildingType, recipeID)
	if need == "" {
		d := depositAt(city, x, y)
		if d == nil {
			return fmt.Errorf("нужна залежь ресурса на этой клетке")
		}
		return nil
	}
	d := depositAt(city, x, y)
	if d == nil || d.ItemID != need {
		name := need
		if it, ok := e.Catalog.ItemByID[need]; ok {
			name = it.Name
		}
		return fmt.Errorf("нужна залежь «%s» на клетке (сгенерированы на окраине)", name)
	}
	return nil
}

func (e *Engine) consumeDeposit(city *model.CityState, x, y, qty int, itemID string, log *[]string) (float64, bool) {
	d := depositAt(city, x, y)
	if d == nil || d.ItemID != itemID {
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
	if d.Amount <= 0 {
		d.Amount = 0
		*log = append(*log, fmt.Sprintf("Залежь %s в %s (%d,%d) исчерпана", itemID, city.ID, x, y))
		e.removeEmptyDeposits(city)
		if nd := e.spawnDeposit(city, itemID, false); nd != nil {
			*log = append(*log, fmt.Sprintf("Ресурс %s появился в другом месте (%d,%d)", itemID, nd.X, nd.Y))
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
