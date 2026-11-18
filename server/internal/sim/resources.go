package sim

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"bashcapital/server/internal/model"
)

// extractBuildingResources — какие здания требуют залежь какого сырья.
var extractBuildingResources = map[string]string{
	"mine":         "iron_ore",
	"logging_camp": "wood",
	"quarry":       "", // зависит от рецепта
	"farm":         "", // milk / flax по рецепту
	"orchard":      "cherry",
	"chem_plant":   "chemicals",
	"oil_well":     "oil",
}

var recipeOutputResource = map[string]string{
	"mine_iron":         "iron_ore",
	"extract_wood":      "wood",
	"extract_sand":      "sand",
	"extract_quartz":    "quartz",
	"farm_milk":         "milk",
	"extract_flax":      "flax",
	"farm_cherry":       "cherry",
	"extract_chemicals": "chemicals",
	"extract_oil":       "oil",
}

func resourceForRecipe(buildingType, recipeID string) string {
	if item, ok := recipeOutputResource[recipeID]; ok {
		return item
	}
	if item, ok := extractBuildingResources[buildingType]; ok && item != "" {
		return item
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
			n := 2
			if res == "iron_ore" || res == "oil" {
				n = 3
			}
			for i := 0; i < n; i++ {
				e.spawnDeposit(city, res, true)
			}
		}
		// гарантия песка/дерева рядом с промкой если город металлургический
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
		// ресурсы на окраине / rural / industrial / empty
		if p.Zone == "center" {
			continue
		}
		if p.Zone == "residential" && itemID != "milk" && itemID != "flax" && itemID != "cherry" {
			continue
		}
		candidates = append(candidates, p)
	}
	if len(candidates) == 0 {
		// расширим город слегка поиском любых empty
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
	case "iron_ore", "oil":
		base = 150 + rand.Intn(200)
	case "wood", "flax", "cherry", "milk":
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
	need := resourceForRecipe(buildingType, recipeID)
	if need == "" && extractBuildingResources[buildingType] == "" && buildingType != "quarry" && buildingType != "farm" {
		// не добыча
		if _, isExtract := extractBuildingResources[buildingType]; !isExtract {
			return nil
		}
	}
	if need == "" {
		// quarry/farm без рецепта — любая подходящая залежь на клетке
		d := depositAt(city, x, y)
		if d == nil {
			return fmt.Errorf("нужна залежь ресурса на этой клетке")
		}
		return nil
	}
	d := depositAt(city, x, y)
	if d == nil || d.ItemID != need {
		return fmt.Errorf("нужна залежь «%s» на клетке (сгенерированы на окраине)", need)
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
		// возобновление в другом месте
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

// occasional discovery: rare new deposits on monthly event
func (e *Engine) discoverResources(log *[]string) {
	for _, city := range e.World.Cities {
		def := e.Catalog.CityByID[city.ID]
		if len(def.Resources) == 0 || rand.Float64() > 0.45 {
			continue
		}
		res := def.Resources[rand.Intn(len(def.Resources))]
		if d := e.spawnDeposit(city, res, false); d != nil {
			*log = append(*log, fmt.Sprintf("Геологи нашли %s у %s", res, city.ID))
		}
	}
}
