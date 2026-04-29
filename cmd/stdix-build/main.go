package main

import (
	"fmt"
	"os"
	"time"

	"github.com/codref/stdix/internal/db"
	"github.com/codref/stdix/internal/registry"
	"github.com/codref/stdix/internal/search"
	"github.com/codref/stdix/internal/version"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:     "stdix-build",
		Version: version.Version,
		Short:   "Build and validate the stdix standards registry database",
		Long: `stdix-build is the CI-side tool for managing registry.db.

Run 'stdix-build validate' on every pull request.
Run 'stdix-build build' on merge to produce the registry.db artifact.`,
	}
	root.AddCommand(validateCmd(), buildCmd())
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func validateCmd() *cobra.Command {
	var registryPath string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate all standards in a registry without writing output",
		Long: `Validate parses every YAML file under <registry>/standards/, checks required
fields, semver format, and duplicate IDs. Designed for CI pull-request checks.

Exits non-zero when any validation error is found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			standards, err := registry.Load(registryPath)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Validated %d standards in %s — OK\n", len(standards), registryPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&registryPath, "registry", ".", "path to the registry root directory")
	return cmd
}

func buildCmd() *cobra.Command {
	var registryPath, outPath, version string
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build registry.db from a standards registry",
		Long: `Build validates all standards, indexes them for BM25 search, and writes
the result as a single JSON artifact (registry.db).

Exits non-zero when any validation error is found — safe to run in CI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			standards, err := registry.Load(registryPath)
			if err != nil {
				return err
			}

			indexed := make([]db.IndexedStandard, len(standards))
			for i, s := range standards {
				indexed[i] = db.IndexedStandard{
					Standard: s,
					Terms:    search.IndexTerms(s),
				}
			}

			d := &db.DB{
				Metadata: db.Metadata{
					Version: version,
					BuiltAt: time.Now().UTC(),
				},
				Standards: indexed,
			}
			if err := db.Write(outPath, d); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"Built registry.db: %d standards, version %s\n", len(standards), version)
			return nil
		},
	}
	cmd.Flags().StringVar(&registryPath, "registry", ".", "path to the registry root directory")
	cmd.Flags().StringVarP(&outPath, "out", "o", "registry.db", "output path for registry.db")
	cmd.Flags().StringVar(&version, "version", "0.0.1", "registry version to embed in metadata")
	return cmd
}
