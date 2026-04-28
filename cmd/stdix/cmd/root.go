package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stdix",
	Short: "Connect your project to engineering standards",
	Long: `stdix matches your project against a curated standards registry and
materialises the relevant standards into AI-agent instruction files
(AGENTS.md, CLAUDE.md, .github/copilot-instructions.md, .cursor/rules/).`,
}

// Execute is the entry point called from main.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(
		initCmd(),
		syncCmd(),
		matchCmd(),
		listCmd(),
		doctorCmd(),
		applyCmd(),
		deployCmd(),
		pushCmd(),
	)
}
