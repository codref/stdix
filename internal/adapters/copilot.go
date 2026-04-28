package adapters

import "github.com/stdix/stdix/internal/db"

// CopilotAdapter generates content for .github/copilot-instructions.md.
type CopilotAdapter struct{}

func (CopilotAdapter) TargetPath() string { return ".github/copilot-instructions.md" }

func (CopilotAdapter) Render(s db.IndexedStandard) string {
	return renderMarkdown(s)
}
