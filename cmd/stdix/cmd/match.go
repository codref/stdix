package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/stdix/stdix/internal/db"
	"github.com/stdix/stdix/internal/search"
)

func matchCmd() *cobra.Command {
	var limit int
	var lang string

	cmd := &cobra.Command{
		Use:   "match <query>",
		Short: "Find standards matching a task description",
		Long: `Loads the local registry.db and ranks standards against your query
using BM25 keyword scoring. Use --lang to apply a language bonus.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			dbPath, err := resolveDBPath()
			if err != nil {
				return err
			}
			d, err := db.Read(dbPath)
			if err != nil {
				return fmt.Errorf("loading registry.db: %w (run 'stdix sync' first)", err)
			}

			results := search.Score(query, lang, d.Standards)

			if limit > 0 && len(results) > limit {
				results = results[:limit]
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tSCORE\tMATCHED WHY")
			fmt.Fprintln(w, "──\t─────\t─────\t───────────")
			for _, r := range results {
				if r.Score == 0 {
					continue
				}
				why := r.MatchedWhy
				if why == "" {
					why = "—"
				}
				fmt.Fprintf(w, "%s\t%s\t%.2f\t%s\n",
					r.Standard.ID, r.Standard.Title, r.Score, why)
			}
			w.Flush()
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "maximum number of results")
	cmd.Flags().StringVar(&lang, "lang", "", "filter bonus for language (e.g. python, go)")
	return cmd
}

// resolveDBPath finds the registry.db to use, preferring the .stdix.yaml config.
func resolveDBPath() (string, error) {
	cwd, _ := os.Getwd()
	if cfg, err := loadConfig(cwd); err == nil {
		p := cfg.Registry.DB
		if p != "" && cfg.Registry.Source == "local" {
			return p, nil
		}
	}
	home, _ := os.UserHomeDir()
	p := fmt.Sprintf("%s/.cache/stdix/registry.db", home)
	return p, nil
}
