package adapters

import "github.com/codref/stdix/internal/db"

// Adapter generates agent-specific instruction content for a standard.
type Adapter interface {
	// TargetPath returns the output file path relative to the project root.
	TargetPath() string
	// Render produces the block content to be inserted between the stdix markers.
	Render(s db.IndexedStandard) string
}
