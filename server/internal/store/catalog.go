package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"bashcapital/server/internal/model"
)

// ResolveCatalogDir prefers data/catalog_v2 when present and complete; else dataDir root.
func ResolveCatalogDir(dataDir string) string {
	v2 := filepath.Join(dataDir, "catalog_v2")
	need := []string{"items.json", "recipes.json", "buildings.json", "shops.json", "cities.json", "economy.json"}
	ok := true
	for _, n := range need {
		if _, err := os.Stat(filepath.Join(v2, n)); err != nil {
			ok = false
			break
		}
	}
	if ok {
		return v2
	}
	return dataDir
}

func LoadCatalog(dataDir string) (*model.Catalog, error) {
	dir := ResolveCatalogDir(dataDir)
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
	if err := readJSON(filepath.Join(dir, "items.json"), &c.Items); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dir, "recipes.json"), &c.Recipes); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dir, "buildings.json"), &c.Buildings); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dir, "shops.json"), &c.Shops); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(dir, "cities.json"), &c.Cities); err != nil {
		return nil, err
	}
	routesPath := filepath.Join(dir, "routes.json")
	if _, err := os.Stat(routesPath); err != nil {
		routesPath = filepath.Join(dataDir, "routes.json")
	}
	if err := readJSON(routesPath, &c.Routes); err != nil {
		return nil, err
	}
	_ = readJSON(filepath.Join(dir, "economy.json"), &c.Economy)
	// Optional legacy sidecars from data root
	_ = readJSON(filepath.Join(dataDir, "ai_personas.json"), &c.AIPersonas)
	_ = readJSON(filepath.Join(dataDir, "rnd_tree.json"), &c.RnDTree)
	if c.Economy.Balance.Version >= 2 || c.Economy.PriceCorridors != nil {
		// economy.json may embed version
	}
	if metaVer := readMetaVersion(dir); metaVer > 0 {
		c.Version = metaVer
	} else if c.Economy.Balance.Version >= 2 {
		c.Version = c.Economy.Balance.Version
	} else {
		c.Version = 1
	}
	// Prefer explicit economy_catalog_version field if present in economy.json via raw peek
	if v := peekCatalogVersion(filepath.Join(dir, "economy.json")); v > 0 {
		c.Version = v
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

func readMetaVersion(dir string) int {
	var meta struct {
		EconomyCatalogVersion int `json:"economy_catalog_version"`
	}
	if err := readJSON(filepath.Join(dir, "catalog_meta.json"), &meta); err != nil {
		return 0
	}
	return meta.EconomyCatalogVersion
}

func peekCatalogVersion(path string) int {
	var raw map[string]json.RawMessage
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return 0
	}
	if v, ok := raw["economy_catalog_version"]; ok {
		var n int
		if json.Unmarshal(v, &n) == nil {
			return n
		}
	}
	return 0
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
