package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/codref/stdix/internal/db"
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all standards in the local registry.db",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath, err := resolveDBPath()
			if err != nil {
				return err
			}
			d, err := db.Read(dbPath)
			if err != nil {
				return fmt.Errorf("loading registry.db: %w (run 'stdix sync' first)", err)
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tLANGUAGE\tVERSION")
			fmt.Fprintln(w, "──\t─────\t────────\t───────")
			for _, s := range d.Standards {
				lang := s.Language
				if lang == "" {
					lang = "any"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.ID, s.Title, lang, s.Version)
			}
			w.Flush()
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d standards  (registry version: %s)\n",
				len(d.Standards), d.Metadata.Version)
			return nil
		},
	}
}
