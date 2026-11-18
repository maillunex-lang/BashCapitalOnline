package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bashcapital/server/internal/model"
)

// MigrateWorldCopyToV2 rewrites inventories/buildings/shops on a world copy using migration maps.
// Never mutates the live DB path — caller must pass a snapshot file.
func MigrateWorldCopyToV2(world *model.World, itemMap map[string]map[string]any, buildingMap map[string]string) (compensated float64, report []string) {
	mapItem := func(old string) (string, bool) {
		m, ok := itemMap[old]
		if !ok {
			return "", false
		}
		if m["action"] == "map_to_new" {
			if nid, ok := m["new_id"].(string); ok && nid != "" {
				return nid, true
			}
		}
		return "", false
	}
	mid := 50.0
	for _, b := range world.Buildings {
		if typ, ok := buildingMap[b.Type]; ok && typ != "" {
			report = append(report, fmt.Sprintf("building %s type %s→%s", b.ID, b.Type, typ))
			b.Type = typ
		}
		next := map[string]int{}
		for item, qty := range b.Inventory {
			if qty <= 0 {
				continue
			}
			if nid, ok := mapItem(item); ok {
				next[nid] += qty
			} else {
				compensated += float64(qty) * mid * 0.85
				report = append(report, fmt.Sprintf("compensate %d×%s", qty, item))
			}
		}
		b.Inventory = next
	}
	for _, s := range world.Shops {
		if nid, ok := mapItem(s.ItemID); ok {
			s.ItemID = nid
		} else {
			compensated += float64(s.Stock) * mid * 0.85
			s.Stock = 0
			report = append(report, fmt.Sprintf("shop %s item compensated", s.ID))
		}
	}
	for _, f := range world.Firms {
		f.Cash += compensated / float64(max(1, len(world.Firms)))
	}
	world.CatalogVersion = 2
	return compensated, report
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// SnapshotDB copies a sqlite file for safe migration experiments.
func SnapshotDB(src, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	name := fmt.Sprintf("pre_v2_%s.db", time.Now().Format("20060102_150405"))
	dest := filepath.Join(destDir, name)
	in, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(dest, in, 0644); err != nil {
		return "", err
	}
	return dest, nil
}

// LoadItemMigrationMap reads catalog_v2/migration_items.json.
func LoadItemMigrationMap(dataDir string) (map[string]map[string]any, error) {
	path := filepath.Join(ResolveCatalogDir(dataDir), "migration_items.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
