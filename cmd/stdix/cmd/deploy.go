package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stdix/stdix/internal/db"
	"github.com/stdix/stdix/internal/generator"
)

func deployCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Re-apply all standards listed in .stdix.yaml to agent files",
		Long: `Reads the standards list from .stdix.yaml, looks each one up in registry.db,
and writes (or refreshes) the managed blocks in all enabled agent files.

Useful after 'stdix sync' to regenerate agent files from an updated registry,
or to restore agent files that were deleted or corrupted.

Files modified depend on the outputs flags in .stdix.yaml:
  agents  → AGENTS.md
  claude  → CLAUDE.md
  copilot → .github/copilot-instructions.md
  cursor  → .cursor/rules/stdix.mdc`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			cfg, err := loadConfig(cwd)
			if err != nil {
				return fmt.Errorf("loading .stdix.yaml: %w (run 'stdix init' first)", err)
			}

			if len(cfg.Standards) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No standards in .stdix.yaml — nothing to deploy.")
				fmt.Fprintln(cmd.OutOrStdout(), "Run 'stdix apply <id>' to add standards first.")
				return nil
			}

			dbPath, err := resolveDBPath()
			if err != nil {
				return err
			}
			d, err := db.Read(dbPath)
			if err != nil {
				return fmt.Errorf("loading registry.db: %w (run 'stdix sync' first)", err)
			}

			// Index standards by ID for fast lookup.
			byID := make(map[string]*db.IndexedStandard, len(d.Standards))
			for i := range d.Standards {
				byID[d.Standards[i].ID] = &d.Standards[i]
			}

			activeAdapters := activeAdapters(cfg)
			if len(activeAdapters) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No outputs enabled in .stdix.yaml — nothing to write.")
				return nil
			}

			deployed := 0
			for _, id := range cfg.Standards {
				standard, ok := byID[id]
				if !ok {
					fmt.Fprintf(cmd.OutOrStdout(), "  warning: standard %q not found in registry.db — skipped\n", id)
					continue
				}
				for _, ad := range activeAdapters {
					target := filepath.Join(cwd, ad.TargetPath())
					content := ad.Render(*standard)
					if err := generator.Upsert(target, content); err != nil {
						return fmt.Errorf("writing %s: %w", ad.TargetPath(), err)
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  applied %s (%s)\n", standard.Title, id)
				deployed++
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deployed %d standard(s) to %d file(s).\n", deployed, len(activeAdapters))
			return nil
		},
	}
}
