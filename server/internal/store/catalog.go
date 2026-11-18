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
		CityByID:     map[string]model.CityDef{},
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
	if err := readJSON(filepath.Join(dataDir, "cities.json"), &c.Cities); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dataDir, "routes.json"), &c.Routes); err != nil {
		return nil, err
	}
	for _, it := range c.Items {
		c.ItemByID[it.ID] = it
	}
	for _, r := range c.Recipes {
		c.RecipeByID[r.ID] = r
	}
	for _, b := range c.Buildings {
		c.BuildingByID[b.ID] = b
	}
	for _, city := range c.Cities {
		c.CityByID[city.ID] = city
	}
	return c, nil
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
