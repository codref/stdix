package generator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codref/stdix/internal/generator"
)

const testContent = "## My Standard\n\n- Rule one.\n- Rule two."

func TestUpsert_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")

	if err := generator.Upsert(path, testContent); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "<!-- stdix:start -->") {
		t.Error("expected start marker in new file")
	}
	if !strings.Contains(string(data), "<!-- stdix:end -->") {
		t.Error("expected end marker in new file")
	}
	if !strings.Contains(string(data), testContent) {
		t.Error("expected content in new file")
	}
}

func TestUpsert_AppendsToExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	existing := "# My Project\n\nSome user content.\n"
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := generator.Upsert(path, testContent); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	data, _ := os.ReadFile(path)
	s := string(data)
	if !strings.Contains(s, "# My Project") {
		t.Error("user content was removed")
	}
	if !strings.Contains(s, "<!-- stdix:start -->") {
		t.Error("start marker missing")
	}
}

func TestUpsert_ReplacesExistingBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	initial := "# Project\n\n<!-- stdix:start -->\nold content\n<!-- stdix:end -->\n\nUser footer.\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := generator.Upsert(path, testContent); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	data, _ := os.ReadFile(path)
	s := string(data)
	if strings.Contains(s, "old content") {
		t.Error("old block content should have been replaced")
	}
	if !strings.Contains(s, testContent) {
		t.Error("new content missing")
	}
	if !strings.Contains(s, "# Project") {
		t.Error("user prefix was removed")
	}
	if !strings.Contains(s, "User footer.") {
		t.Error("user footer was removed")
	}
}

func TestUpsert_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")

	for i := 0; i < 3; i++ {
		if err := generator.Upsert(path, testContent); err != nil {
			t.Fatalf("Upsert iteration %d: %v", i, err)
		}
	}
	data, _ := os.ReadFile(path)
	s := string(data)
	if count := strings.Count(s, "<!-- stdix:start -->"); count != 1 {
		t.Errorf("expected exactly 1 start marker after 3 upserts, got %d", count)
	}
}

func TestUpsert_CreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".github", "copilot-instructions.md")

	if err := generator.Upsert(path, testContent); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestExtract_ReturnsBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	if err := generator.Upsert(path, testContent); err != nil {
		t.Fatal(err)
	}
	extracted, ok, err := generator.Extract(path)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if !ok {
		t.Fatal("expected block to be present")
	}
	if extracted != testContent {
		t.Errorf("extracted content mismatch:\ngot:  %q\nwant: %q", extracted, testContent)
	}
}

func TestExtract_NoBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(path, []byte("no markers here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, ok, err := generator.Extract(path)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if ok {
		t.Error("expected ok=false for file with no block")
	}
}
