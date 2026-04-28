package adapters

import (
	"fmt"
	"strings"

	"github.com/codref/stdix/internal/db"
)

// CursorAdapter generates content for .cursor/rules/stdix.mdc (MDC format).
type CursorAdapter struct{}

func (CursorAdapter) TargetPath() string { return ".cursor/rules/stdix.mdc" }

func (CursorAdapter) Render(s db.IndexedStandard) string {
	globs := "**/*"
	if s.Language != "" {
		switch strings.ToLower(s.Language) {
		case "python":
			globs = "**/*.py"
		case "go":
			globs = "**/*.go"
		case "node", "typescript", "javascript":
			globs = "**/*.{ts,js}"
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "---\n")
	fmt.Fprintf(&b, "description: %s\n", s.Title)
	fmt.Fprintf(&b, "globs: \"%s\"\n", globs)
	fmt.Fprintf(&b, "alwaysApply: true\n")
	fmt.Fprintf(&b, "---\n\n")
	fmt.Fprintf(&b, "## %s\n\n", s.Title)
	for _, rule := range s.Rules {
		fmt.Fprintf(&b, "- %s\n", rule)
	}
	return strings.TrimRight(b.String(), "\n")
}
