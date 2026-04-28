package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	startMarker = "<!-- stdix:start -->"
	endMarker   = "<!-- stdix:end -->"
)

// Upsert inserts or replaces the stdix managed block in a file.
//   - If the file does not exist, it is created containing only the block.
//   - If the file exists but has no block, the block is appended.
//   - If the file has an existing block, only the block content is replaced.
//
// Content outside the markers is never modified.
func Upsert(path, content string) error {
	block := startMarker + "\n" + content + "\n" + endMarker

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", path, err)
		}
		return os.WriteFile(path, []byte(block+"\n"), 0o644)
	}
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	existing := string(data)
	startIdx := strings.Index(existing, startMarker)
	endIdx := strings.Index(existing, endMarker)

	if startIdx == -1 || endIdx == -1 || endIdx < startIdx {
		// No existing block: append.
		if !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		return os.WriteFile(path, []byte(existing+"\n"+block+"\n"), 0o644)
	}

	// Replace the existing block, preserving content outside it.
	before := existing[:startIdx]
	after := existing[endIdx+len(endMarker):]
	return os.WriteFile(path, []byte(before+block+after), 0o644)
}

// Extract returns the content inside the stdix managed block.
// Returns ("", false, nil) when no block is present.
func Extract(path string) (string, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false, err
	}
	content := string(data)
	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)
	if startIdx == -1 || endIdx == -1 || endIdx < startIdx {
		return "", false, nil
	}
	inner := content[startIdx+len(startMarker) : endIdx]
	return strings.TrimSpace(inner), true, nil
}
