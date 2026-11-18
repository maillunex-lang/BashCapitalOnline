package store

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveAndLoadCatalogV2(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	dataDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "data"))
	dir := ResolveCatalogDir(dataDir)
	if filepath.Base(dir) != "catalog_v2" {
		t.Fatalf("expected catalog_v2, got %s", dir)
	}
	cat, err := LoadCatalog(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Version != 2 {
		t.Fatalf("version=%d", cat.Version)
	}
	if len(cat.Items) != 170 {
		t.Fatalf("items=%d", len(cat.Items))
	}
}
