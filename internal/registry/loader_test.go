package registry_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/codref/stdix/internal/registry"
)

func testdataPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(file), "..", "..", "testdata")
	return filepath.Join(append([]string{base}, parts...)...)
}

func TestLoad_ValidRegistry(t *testing.T) {
	standards, err := registry.Load(testdataPath("stdix-registry"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(standards) != 6 {
		t.Errorf("expected 6 standards, got %d", len(standards))
	}
}

func TestLoad_MissingStandardsDir(t *testing.T) {
	_, err := registry.Load(testdataPath("fixtures", "nonexistent"))
	if err == nil {
		t.Fatal("expected error for missing directory, got nil")
	}
}

func TestValidate_MissingID(t *testing.T) {
	s := registry.Standard{Title: "X", Version: "1.0.0", Rules: []string{"Do something"}}
	err := registry.Validate(&s)
	if err == nil {
		t.Fatal("expected validation error for missing id")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("error should mention 'id', got: %s", err)
	}
}

func TestValidate_MissingRules(t *testing.T) {
	s := registry.Standard{ID: "x.y", Title: "X", Version: "1.0.0"}
	err := registry.Validate(&s)
	if err == nil {
		t.Fatal("expected validation error for missing rules")
	}
	if !strings.Contains(err.Error(), "rules") {
		t.Errorf("error should mention 'rules', got: %s", err)
	}
}

func TestValidate_MissingTitle(t *testing.T) {
	s := registry.Standard{ID: "x.y", Version: "1.0.0", Rules: []string{"Do something"}}
	err := registry.Validate(&s)
	if err == nil {
		t.Fatal("expected validation error for missing title")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention 'title', got: %s", err)
	}
}

func TestValidate_MissingVersion(t *testing.T) {
	s := registry.Standard{ID: "x.y", Title: "X", Rules: []string{"Do something"}}
	err := registry.Validate(&s)
	if err == nil {
		t.Fatal("expected validation error for missing version")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Errorf("error should mention 'version', got: %s", err)
	}
}

func TestValidate_BadSemver(t *testing.T) {
	s := registry.Standard{ID: "x.y", Title: "X", Version: "v1", Rules: []string{"Do something"}}
	err := registry.Validate(&s)
	if err == nil {
		t.Fatal("expected validation error for bad semver")
	}
	if !strings.Contains(err.Error(), "semver") {
		t.Errorf("error should mention 'semver', got: %s", err)
	}
}

func TestLoad_DuplicateID(t *testing.T) {
	_, err := registry.Load(testdataPath("fixtures", "invalid-duplicate"))
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention 'duplicate', got: %s", err)
	}
}

func TestLoad_InvalidFileInRegistry(t *testing.T) {
	_, err := registry.Load(testdataPath("fixtures", "invalid-missing-rules"))
	if err == nil {
		t.Fatal("expected error for invalid standard, got nil")
	}
	if !strings.Contains(err.Error(), "rules") {
		t.Errorf("error should mention 'rules', got: %s", err)
	}
}
