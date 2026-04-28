package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/codref/stdix/internal/github"
	"gopkg.in/yaml.v3"
)

func pushCmd() *cobra.Command {
	var (
		message string
		repo    string
		branch  string
	)

	cmd := &cobra.Command{
		Use:   "push <yaml-file>",
		Short: "Push a standard YAML file to the registry repo via the GitHub API",
		Long: `Reads a standard YAML file, derives its registry path from the standard ID,
and creates or updates the file in the remote registry repository.

The commit is pushed directly to the target branch.

Authentication:
  Set STDIX_REGISTRY_TOKEN to a GitHub personal access token (or fine-grained
  token) with contents:write permission on the registry repository.

Registry repository:
  Set registry.repo in .stdix.yaml (e.g. "owner/stdix-registry") or use
  the --repo flag.

The registry path is derived from the standard ID:
  python.cli      → standards/python/cli.yaml
  shared.logging  → standards/shared/logging.yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			yamlFile := args[0]

			token := os.Getenv("STDIX_REGISTRY_TOKEN")
			if token == "" {
				return fmt.Errorf("STDIX_REGISTRY_TOKEN environment variable is not set\n" +
					"Create a GitHub token with contents:write on the registry repo and export it:\n" +
					"  export STDIX_REGISTRY_TOKEN=<your-token>")
			}

			content, err := os.ReadFile(yamlFile)
			if err != nil {
				return fmt.Errorf("reading %s: %w", yamlFile, err)
			}

			// Parse just enough to get id and language for path + message.
			var meta struct {
				ID       string `yaml:"id"`
				Language string `yaml:"language"`
			}
			if err := yaml.Unmarshal(content, &meta); err != nil {
				return fmt.Errorf("parsing YAML: %w", err)
			}
			if meta.ID == "" {
				return fmt.Errorf("standard YAML must have an 'id' field")
			}

			remotePath := idToPath(meta.ID)

			if message == "" {
				message = "add " + meta.ID + " standard"
			}

			// Resolve repo: flag > config > error
			if repo == "" {
				cwd, _ := os.Getwd()
				if cfg, err := loadConfig(cwd); err == nil {
					repo = cfg.Registry.Repo
				}
			}
			if repo == "" {
				return fmt.Errorf("registry repo not set\n" +
					"Set registry.repo in .stdix.yaml or use --repo owner/stdix-registry")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Pushing %s → %s:%s (%s)…\n", meta.ID, repo, remotePath, branch)

			htmlURL, err := github.PushFile(token, repo, branch, remotePath, message, content)
			if err != nil {
				return fmt.Errorf("pushing to GitHub: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Pushed: %s\n", htmlURL)
			fmt.Fprintln(cmd.OutOrStdout(), "The registry CI will rebuild registry.db. Run 'stdix sync' when the release is ready.")
			return nil
		},
	}

	cmd.Flags().StringVar(&message, "message", "", "commit message (default: \"add <id> standard\")")
	cmd.Flags().StringVar(&repo, "repo", "", "registry repo in owner/repo format (overrides registry.repo in .stdix.yaml)")
	cmd.Flags().StringVar(&branch, "branch", "main", "target branch")
	return cmd
}

// idToPath converts a standard ID to its registry file path.
// "python.cli"     → "standards/python/cli.yaml"
// "shared.logging" → "standards/shared/logging.yaml"
func idToPath(id string) string {
	parts := strings.SplitN(id, ".", 2)
	if len(parts) != 2 {
		return filepath.Join("standards", id+".yaml")
	}
	return filepath.Join("standards", parts[0], parts[1]+".yaml")
}
