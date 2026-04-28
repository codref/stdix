package adapters

import "github.com/stdix/stdix/internal/db"

// ClaudeAdapter generates content for CLAUDE.md.
type ClaudeAdapter struct{}

func (ClaudeAdapter) TargetPath() string { return "CLAUDE.md" }

func (ClaudeAdapter) Render(s db.IndexedStandard) string {
	return renderMarkdown(s)
}
