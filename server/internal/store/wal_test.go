package store

import (
	"database/sql"
	"path/filepath"
	"testing"

	"bashcapital/server/internal/model"
)

func TestOpenDBEnablesDurableWALForExistingSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "world #1.db")
	legacy, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = legacy.Close() })
	old := &DB{sql: legacy}
	if err := old.migrate(); err != nil {
		t.Fatal(err)
	}
	if err := old.SaveWorld(&model.World{Tick: 12}); err != nil {
		t.Fatal(err)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	writer, err := OpenDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	var mode string
	var sync int
	if err := writer.sql.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil || mode != "wal" {
		t.Fatalf("journal mode %q: %v", mode, err)
	}
	if err := writer.sql.QueryRow("PRAGMA synchronous").Scan(&sync); err != nil || sync != 2 {
		t.Fatalf("expected FULL synchronous mode, got %d: %v", sync, err)
	}
	world, err := writer.LoadWorld()
	if err != nil || world.Tick != 12 {
		t.Fatalf("legacy world not restored: %v", err)
	}

	reader, err := OpenDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	tx, err := reader.sql.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	var original, stillVisible []byte
	readSnapshot := `SELECT payload FROM world_save WHERE id=1`
	if err := tx.QueryRow(readSnapshot).Scan(&original); err != nil {
		t.Fatal(err)
	}
	world.Tick++
	if err := writer.SaveWorld(world); err != nil {
		t.Fatalf("save while another connection holds a read transaction: %v", err)
	}
	if err := tx.QueryRow(readSnapshot).Scan(&stillVisible); err != nil || string(stillVisible) != string(original) {
		t.Fatalf("read transaction did not retain its snapshot: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	latest, err := reader.LoadWorld()
	if err != nil || latest.Tick != 13 {
		t.Fatalf("new read did not see the committed tick: %v", err)
	}
}
