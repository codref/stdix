package adapters

import (
	"fmt"
	"strings"

	"github.com/stdix/stdix/internal/db"
)

// AgentsAdapter generates content for AGENTS.md.
type AgentsAdapter struct{}

func (AgentsAdapter) TargetPath() string { return "AGENTS.md" }

func (AgentsAdapter) Render(s db.IndexedStandard) string {
	return renderMarkdown(s)
}

// renderMarkdown produces the shared Markdown block used by AGENTS.md and CLAUDE.md.
func renderMarkdown(s db.IndexedStandard) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Applied stdix standards\n\n")
	fmt.Fprintf(&b, "### %s (%s v%s)\n\n", s.Title, s.ID, s.Version)
	for _, rule := range s.Rules {
		fmt.Fprintf(&b, "- %s\n", rule)
	}
	return strings.TrimRight(b.String(), "\n")
}
