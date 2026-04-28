package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed scaffold/new-standard.prompt.md
var newStandardPrompt []byte

func scaffoldCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scaffold",
		Short: "Install the stdix Copilot agent prompt into this project",
		Long: `Writes .github/prompts/new-standard.prompt.md to the current directory.

Once installed, use the '/New stdix Standard' Copilot agent to generate and
push new standards to your registry directly from this project.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			dest := filepath.Join(cwd, ".github", "prompts", "new-standard.prompt.md")

			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}

			if _, err := os.Stat(dest); err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "%s already exists — skipping.\n", dest)
				return nil
			}

			if err := os.WriteFile(dest, newStandardPrompt, 0o644); err != nil {
				return fmt.Errorf("writing prompt: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", dest)
			fmt.Fprintln(cmd.OutOrStdout(), "Use the '/New stdix Standard' Copilot agent to add standards to your registry.")
			return nil
		},
	}
}
