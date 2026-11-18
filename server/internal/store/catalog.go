package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"bashcapital/server/internal/model"
)

func LoadCatalog(dataDir string) (*model.Catalog, error) {
	c := &model.Catalog{
		ItemByID:     map[string]model.ItemDef{},
		RecipeByID:   map[string]model.RecipeDef{},
		BuildingByID: map[string]model.BuildingDef{},
		ShopByID:     map[string]model.ShopDef{},
		CityByID:     map[string]model.CityDef{},
		PersonaByID:  map[string]model.AIPersona{},
		Economy: model.EconomyConfig{
			PriceCorridors: map[string]model.PriceCorridor{},
			Chat:           model.ChatConfig{MaxMessages: 120, AIMessagesPerTickCap: 2, MinTicksBetweenAIMsg: 2},
		},
	}
	if err := readJSON(filepath.Join(dataDir, "items.json"), &c.Items); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dataDir, "recipes.json"), &c.Recipes); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dataDir, "buildings.json"), &c.Buildings); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dataDir, "shops.json"), &c.Shops); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dataDir, "cities.json"), &c.Cities); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dataDir, "routes.json"), &c.Routes); err != nil {
		return nil, err
	}
	_ = readJSON(filepath.Join(dataDir, "economy.json"), &c.Economy)
	_ = readJSON(filepath.Join(dataDir, "ai_personas.json"), &c.AIPersonas)
	for _, it := range c.Items {
		c.ItemByID[it.ID] = it
	}
	for _, r := range c.Recipes {
		c.RecipeByID[r.ID] = r
	}
	for _, b := range c.Buildings {
		c.BuildingByID[b.ID] = b
	}
	for _, s := range c.Shops {
		c.ShopByID[s.ID] = s
	}
	for _, city := range c.Cities {
		c.CityByID[city.ID] = city
	}
	for _, p := range c.AIPersonas {
		c.PersonaByID[p.ID] = p
	}
	return c, nil
}

func SaveJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func DataPath(dataDir, name string) string {
	return filepath.Join(dataDir, name)
}

func readJSON(path string, dst any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}
