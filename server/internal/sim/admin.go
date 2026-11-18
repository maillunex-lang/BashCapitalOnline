package sim

import (
	"fmt"

	"bashcapital/server/internal/model"
	"bashcapital/server/internal/store"
)

func (e *Engine) SetPaused(paused bool) {
	e.Paused = paused
	if paused {
		e.World.Events = append(e.World.Events, "Админ поставил мир на паузу")
	} else {
		e.World.Events = append(e.World.Events, "Админ снял паузу")
	}
}

func (e *Engine) AdminStatus() map[string]any {
	enabled := 0
	for _, c := range e.Catalog.Cities {
		if c.Enabled {
			enabled++
		}
	}
	return map[string]any{
		"tick":           e.World.Tick,
		"month":          e.World.Month,
		"year":           e.World.Year,
		"paused":         e.Paused,
		"firms":          len(e.World.Firms),
		"buildings":      len(e.World.Buildings),
		"shops":          len(e.World.Shops),
		"cities_live":    len(e.World.Cities),
		"cities_enabled": enabled,
		"map_revision":   e.World.MapRevision,
		"dirty_tiles":    len(e.World.DirtyTiles),
	}
}

func (e *Engine) UpdateEconomy(corridors map[string]model.PriceCorridor, market []model.CityMarketOffer) error {
	if corridors != nil {
		if e.Catalog.Economy.PriceCorridors == nil {
			e.Catalog.Economy.PriceCorridors = map[string]model.PriceCorridor{}
		}
		for k, v := range corridors {
			if v.Min < 0 || v.Max < v.Min {
				return fmt.Errorf("коридор %s: min/max некорректны", k)
			}
			e.Catalog.Economy.PriceCorridors[k] = v
		}
	}
	if market != nil {
		e.Catalog.Economy.CityMarket = market
		e.initCityMarket()
	}
	if e.DataDir != "" {
		if err := store.SaveJSON(store.DataPath(e.DataDir, "economy.json"), e.Catalog.Economy); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) SetCityEnabled(cityID string, enabled bool) error {
	idx := -1
	for i := range e.Catalog.Cities {
		if e.Catalog.Cities[i].ID == cityID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("город не найден")
	}
	def := e.Catalog.Cities[idx]
	def.Enabled = enabled
	e.Catalog.Cities[idx] = def
	e.Catalog.CityByID[cityID] = def
	if enabled {
		if _, ok := e.World.Cities[cityID]; !ok {
			e.seedOneCity(def)
		}
	}
	if e.DataDir != "" {
		if err := store.SaveJSON(store.DataPath(e.DataDir, "cities.json"), e.Catalog.Cities); err != nil {
			return err
		}
	}
	e.World.Events = append(e.World.Events, fmt.Sprintf("Админ: город %s enabled=%v", cityID, enabled))
	return nil
}

func (e *Engine) seedOneCity(def model.CityDef) {
	cs := &model.CityState{
		ID:         def.ID,
		Population: def.StartPopulation,
		Radius:     def.StartRadius,
		LaborPool:  def.Labor,
		LaborUsed:  model.Labor{},
		Parcels:    generateParcels(def.StartRadius),
		DemandSeg: map[string]float64{
			"food":    float64(def.StartPopulation) * 0.08,
			"home":    float64(def.StartPopulation) * 0.02,
			"apparel": float64(def.StartPopulation) * 0.015,
			"fuel":    float64(def.StartPopulation) * 0.01,
			"tech":    float64(def.StartPopulation) * 0.008,
		},
		HousingOcc: 0.75,
	}
	for i := range cs.Parcels {
		p := &cs.Parcels[i]
		dist := absf(float64(p.X)) + absf(float64(p.Y))
		if dist == 0 {
			p.Zone = "center"
			p.Traffic = 1.0
			p.Slots = 4
		} else if dist <= 2 {
			p.Zone = "residential"
			p.Slots = 1
			p.Traffic = 0.7
		} else if dist <= float64(def.StartRadius)-1 {
			if p.X > 0 {
				p.Zone = "industrial"
				p.Traffic = 0.3
			} else {
				p.Zone = "rural"
				p.Traffic = 0.2
			}
		} else {
			p.Zone = "empty"
		}
	}
	seedCityRoads(cs)
	e.World.Cities[def.ID] = cs
	for _, res := range def.Resources {
		if it, ok := e.Catalog.ItemByID[res]; ok && it.Tier != "raw" {
			continue
		}
		e.spawnDeposit(cs, res, true)
		e.spawnDeposit(cs, res, true)
	}
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func (e *Engine) UpdateRecipeResearch(recipeID string, researchCost int) error {
	r, ok := e.Catalog.RecipeByID[recipeID]
	if !ok {
		return fmt.Errorf("рецепт не найден")
	}
	if researchCost < 0 {
		return fmt.Errorf("research_cost >= 0")
	}
	r.ResearchCost = researchCost
	e.Catalog.RecipeByID[recipeID] = r
	for i := range e.Catalog.Recipes {
		if e.Catalog.Recipes[i].ID == recipeID {
			e.Catalog.Recipes[i].ResearchCost = researchCost
			break
		}
	}
	if e.DataDir != "" {
		if err := store.SaveJSON(store.DataPath(e.DataDir, "recipes.json"), e.Catalog.Recipes); err != nil {
			return err
		}
	}
	return nil
}
