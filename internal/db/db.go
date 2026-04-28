package db

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/codref/stdix/internal/registry"
)

// Metadata holds registry-level information embedded in every registry.db.
type Metadata struct {
	Version string    `json:"version"`
	BuiltAt time.Time `json:"built_at"`
}

// IndexedStandard extends a Standard with pre-tokenized terms for BM25 scoring.
type IndexedStandard struct {
	registry.Standard
	Terms []string `json:"terms"`
}

// DB is the full in-memory representation of a registry.db file.
type DB struct {
	Metadata  Metadata          `json:"metadata"`
	Standards []IndexedStandard `json:"standards"`
}

// Write serialises db to path as indented JSON.
func Write(path string, d *DB) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling registry.db: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// Read deserialises a registry.db file from path.
// Returns an error if the file is missing, malformed, or lacks required metadata.
func Read(path string) (*DB, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var d DB
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if d.Metadata.Version == "" {
		return nil, fmt.Errorf("%s: missing metadata.version", path)
	}
	if d.Metadata.BuiltAt.IsZero() {
		return nil, fmt.Errorf("%s: missing metadata.built_at", path)
	}
	return &d, nil
}
