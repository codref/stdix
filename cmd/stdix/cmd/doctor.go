package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/stdix/stdix/internal/config"
	"github.com/stdix/stdix/internal/db"
)

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check the stdix setup for common problems",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			cwd, _ := os.Getwd()
			hasError := false

			// 1. .stdix.yaml
			if config.Exists(cwd) {
				fmt.Fprintln(out, "  ✓  .stdix.yaml found")
			} else {
				fmt.Fprintln(out, "  ✗  .stdix.yaml missing — run 'stdix init'")
				hasError = true
			}

			// 2. local registry.db
			dbPath, _ := resolveDBPath()
			if _, err := db.Read(dbPath); err == nil {
				fmt.Fprintf(out, "  ✓  registry.db readable (%s)\n", dbPath)
			} else {
				fmt.Fprintf(out, "  ✗  registry.db not found or unreadable at %s — run 'stdix sync'\n", dbPath)
				hasError = true
			}

			// 3. registry URL reachable (non-fatal)
			if cfg, err := loadConfig(cwd); err == nil && cfg.Registry.URL != "" {
				if reachable(cfg.Registry.URL) {
					fmt.Fprintf(out, "  ✓  registry URL reachable (%s)\n", cfg.Registry.URL)
				} else {
					fmt.Fprintf(out, "  ⚠  registry URL not reachable (%s) — offline?\n", cfg.Registry.URL)
				}
			}

			if hasError {
				return fmt.Errorf("doctor found problems — see above")
			}
			fmt.Fprintln(out, "\nAll checks passed.")
			return nil
		},
	}
}

func reachable(rawURL string) bool {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head(rawURL)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}
