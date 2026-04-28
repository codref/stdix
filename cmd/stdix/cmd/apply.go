package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/codref/stdix/internal/adapters"
	"github.com/codref/stdix/internal/config"
	"github.com/codref/stdix/internal/db"
	"github.com/codref/stdix/internal/generator"
)

func applyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply <standard-id>",
		Short: "Apply a standard to the current project",
		Long: `Looks up the standard in the local registry.db, renders its rules into
agent instruction files, and records the standard ID in .stdix.yaml.

Files modified depend on the outputs flags in .stdix.yaml:
  agents  → AGENTS.md
  claude  → CLAUDE.md
  copilot → .github/copilot-instructions.md
  cursor  → .cursor/rules/stdix.mdc`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			dbPath, err := resolveDBPath()
			if err != nil {
				return err
			}
			d, err := db.Read(dbPath)
			if err != nil {
				return fmt.Errorf("loading registry.db: %w (run 'stdix sync' first)", err)
			}

			var standard *db.IndexedStandard
			for i := range d.Standards {
				if d.Standards[i].ID == id {
					standard = &d.Standards[i]
					break
				}
			}
			if standard == nil {
				return fmt.Errorf("standard %q not found in registry.db", id)
			}

			cfg, err := loadConfig(cwd)
			if err != nil {
				return fmt.Errorf("loading .stdix.yaml: %w (run 'stdix init' first)", err)
			}

			activeAdapters := activeAdapters(cfg)
			for _, ad := range activeAdapters {
				target := filepath.Join(cwd, ad.TargetPath())
				content := ad.Render(*standard)
				if err := generator.Upsert(target, content); err != nil {
					return fmt.Errorf("writing %s: %w", ad.TargetPath(), err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  updated %s\n", ad.TargetPath())
			}

			// Record standard ID in .stdix.yaml if not already present.
			if !containsString(cfg.Standards, id) {
				cfg.Standards = append(cfg.Standards, id)
				if err := config.Save(cwd, cfg); err != nil {
					return fmt.Errorf("updating .stdix.yaml: %w", err)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Applied %s (%s)\n", standard.Title, id)
			return nil
		},
	}
}

// activeAdapters returns the adapters enabled by cfg.Outputs.
func activeAdapters(cfg *config.Config) []adapters.Adapter {
	var result []adapters.Adapter
	if cfg.Outputs.Agents {
		result = append(result, adapters.AgentsAdapter{})
	}
	if cfg.Outputs.Claude {
		result = append(result, adapters.ClaudeAdapter{})
	}
	if cfg.Outputs.Copilot {
		result = append(result, adapters.CopilotAdapter{})
	}
	if cfg.Outputs.Cursor {
		result = append(result, adapters.CursorAdapter{})
	}
	return result
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
