package db_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stdix/stdix/internal/db"
	"github.com/stdix/stdix/internal/registry"
)

func testdataPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(file), "..", "..", "testdata")
	return filepath.Join(append([]string{base}, parts...)...)
}

func TestRoundTrip(t *testing.T) {
	original := &db.DB{
		Metadata: db.Metadata{
			Version: "1.0.0",
			BuiltAt: time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC),
		},
		Standards: []db.IndexedStandard{
			{
				Standard: registry.Standard{
					ID:      "python.cli",
					Title:   "Python CLI Standard",
					Version: "1.0.0",
					Rules:   []string{"Use Typer"},
				},
				Terms: []string{"python", "cli", "typer"},
			},
		},
	}

	tmp := filepath.Join(t.TempDir(), "registry.db")
	if err := db.Write(tmp, original); err != nil {
		t.Fatalf("Write: %v", err)
	}

	loaded, err := db.Read(tmp)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if loaded.Metadata.Version != original.Metadata.Version {
		t.Errorf("version mismatch: got %q, want %q", loaded.Metadata.Version, original.Metadata.Version)
	}
	if len(loaded.Standards) != 1 {
		t.Errorf("expected 1 standard, got %d", len(loaded.Standards))
	}
	if loaded.Standards[0].ID != "python.cli" {
		t.Errorf("expected id python.cli, got %q", loaded.Standards[0].ID)
	}
}

func TestRead_CorruptFile(t *testing.T) {
	_, err := db.Read(testdataPath("fixtures", "corrupt.db"))
	if err == nil {
		t.Fatal("expected error for corrupt db, got nil")
	}
}

func TestRead_EmptyFile(t *testing.T) {
	_, err := db.Read(testdataPath("fixtures", "empty.db"))
	if err == nil {
		t.Fatal("expected error for empty db, got nil")
	}
}

func TestRead_MissingMetadata(t *testing.T) {
	_, err := db.Read(testdataPath("fixtures", "missing-metadata.db"))
	if err == nil {
		t.Fatal("expected error for missing metadata, got nil")
	}
}

func TestRead_Missing(t *testing.T) {
	_, err := db.Read("/nonexistent/path/registry.db")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestWrite_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	d := &db.DB{
		Metadata:  db.Metadata{Version: "0.1.0", BuiltAt: time.Now()},
		Standards: nil,
	}
	path := filepath.Join(dir, "out.db")
	if err := db.Write(path, d); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}
