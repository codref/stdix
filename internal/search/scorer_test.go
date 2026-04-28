package search_test

import (
	"testing"

	"github.com/codref/stdix/internal/db"
	"github.com/codref/stdix/internal/registry"
	"github.com/codref/stdix/internal/search"
)

// buildDB creates a minimal DB from hand-crafted standards.
func buildDB(standards []registry.Standard) []db.IndexedStandard {
	indexed := make([]db.IndexedStandard, len(standards))
	for i, s := range standards {
		indexed[i] = db.IndexedStandard{Standard: s, Terms: search.IndexTerms(s)}
	}
	return indexed
}

var testStandards = []registry.Standard{
	{
		ID:       "python.cli",
		Title:    "Python CLI Standard",
		Version:  "1.0.0",
		Language: "python",
		Tags:     []string{"cli", "terminal", "command-line"},
		AppliesWhen: []string{
			"building a CLI application",
			"creating terminal commands",
		},
		Rules: []string{
			"Use Typer for CLI command definitions.",
			"Use Rich for terminal output.",
		},
	},
	{
		ID:       "go.cli",
		Title:    "Go CLI Standard",
		Version:  "1.0.0",
		Language: "go",
		Tags:     []string{"cli", "cobra", "terminal"},
		AppliesWhen: []string{
			"building a CLI application in Go",
			"creating Go command-line tools",
		},
		Rules: []string{
			"Use Cobra for CLI command definitions.",
			"Use os.Exit only in main.",
		},
	},
	{
		ID:       "shared.logging",
		Title:    "Shared Logging Standard",
		Version:  "1.0.0",
		Language: "",
		Tags:     []string{"logging", "observability"},
		AppliesWhen: []string{
			"adding structured logging",
		},
		Rules: []string{"Use structured log output."},
	},
}

func TestScore_PythonCLIRanksFirst(t *testing.T) {
	indexed := buildDB(testStandards)
	results := search.Score("build a Python CLI app", "", indexed)
	if len(results) == 0 {
		t.Fatal("no results")
	}
	if results[0].Standard.ID != "python.cli" {
		t.Errorf("expected python.cli first, got %q (score %.2f)", results[0].Standard.ID, results[0].Score)
	}
}

func TestScore_LanguageBonusApplied(t *testing.T) {
	indexed := buildDB(testStandards)

	withLang := search.Score("build a CLI app", "python", indexed)
	withoutLang := search.Score("build a CLI app", "", indexed)

	var pythonWithBonus, pythonWithout float64
	for _, r := range withLang {
		if r.Standard.ID == "python.cli" {
			pythonWithBonus = r.Score
		}
	}
	for _, r := range withoutLang {
		if r.Standard.ID == "python.cli" {
			pythonWithout = r.Score
		}
	}
	if pythonWithBonus <= pythonWithout {
		t.Errorf("expected language bonus to increase score: with=%.2f without=%.2f", pythonWithBonus, pythonWithout)
	}
}

func TestScore_LanguageBonusAmount(t *testing.T) {
	indexed := buildDB(testStandards)
	withLang := search.Score("build a CLI", "python", indexed)
	withoutLang := search.Score("build a CLI", "", indexed)

	find := func(results []search.Result, id string) float64 {
		for _, r := range results {
			if r.Standard.ID == id {
				return r.Score
			}
		}
		return 0
	}
	diff := find(withLang, "python.cli") - find(withoutLang, "python.cli")
	if diff != search.LangBonus {
		t.Errorf("expected diff == LangBonus (%.1f), got %.2f", search.LangBonus, diff)
	}
}

func TestScore_MatchedWhyPopulated(t *testing.T) {
	indexed := buildDB(testStandards)
	results := search.Score("CLI application", "", indexed)
	for _, r := range results {
		if r.Standard.ID == "python.cli" && r.MatchedWhy == "" {
			t.Error("expected MatchedWhy to be populated for python.cli")
		}
	}
}

func TestScore_EmptyQuery(t *testing.T) {
	indexed := buildDB(testStandards)
	results := search.Score("", "", indexed)
	// Should return all standards with score 0; no panic.
	if len(results) != len(indexed) {
		t.Errorf("expected %d results, got %d", len(indexed), len(results))
	}
}

func TestScore_NoStandards(t *testing.T) {
	results := search.Score("build a CLI", "python", nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// Ensure IndexTerms is exported and usable (compilation check).
func TestIndexTerms_NonEmpty(t *testing.T) {
	s := testStandards[0]
	terms := search.IndexTerms(s)
	if len(terms) == 0 {
		t.Error("expected non-empty terms for python.cli")
	}
}
