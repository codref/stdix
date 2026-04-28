package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func syncCmd() *cobra.Command {
	var registryURL string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Download registry.db to the local cache",
		Long: `Downloads registry.db from the configured URL and stores it at
~/.cache/stdix/registry.db. Verifies SHA-256 if a checksum is set in .stdix.yaml.

Supports http://, https://, and file:// URLs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if registryURL == "" {
				// Try to load from config.
				cwd, _ := os.Getwd()
				cfg, err := loadConfig(cwd)
				if err != nil {
					return fmt.Errorf("no --registry-url provided and failed to load config: %w", err)
				}
				registryURL = cfg.Registry.URL
				if registryURL == "" {
					return fmt.Errorf("registry URL not set; use --registry-url or set registry.url in .stdix.yaml")
				}
			}

			destDir := defaultCacheDir()
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return fmt.Errorf("creating cache directory: %w", err)
			}
			dest := filepath.Join(destDir, "registry.db")

			if err := downloadDB(registryURL, dest); err != nil {
				return err
			}

			// Verify checksum if configured.
			cwd, _ := os.Getwd()
			if cfg, err := loadConfig(cwd); err == nil && cfg.Registry.Checksum != "" {
				if err := verifySHA256(dest, cfg.Registry.Checksum); err != nil {
					os.Remove(dest)
					return fmt.Errorf("checksum verification failed: %w", err)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Synced registry.db → %s\n", dest)
			return nil
		},
	}
	cmd.Flags().StringVar(&registryURL, "registry-url", "", "URL of registry.db (http/https/file)")
	return cmd
}

func downloadDB(rawURL, dest string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}

	switch u.Scheme {
	case "file":
		data, err := os.ReadFile(u.Path)
		if err != nil {
			return fmt.Errorf("reading local registry.db from %s: %w", u.Path, err)
		}
		return os.WriteFile(dest, data, 0o644)

	case "http", "https":
		resp, err := http.Get(rawURL) //nolint:gosec // URL is user-supplied and validated above
		if err != nil {
			return fmt.Errorf("downloading registry.db: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned %s", resp.Status)
		}
		f, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("creating %s: %w", dest, err)
		}
		defer f.Close()
		if _, err := io.Copy(f, resp.Body); err != nil {
			return fmt.Errorf("writing registry.db: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported URL scheme %q (supported: file, http, https)", u.Scheme)
	}
}

func verifySHA256(path, expected string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	sum := fmt.Sprintf("%x", sha256.Sum256(data))
	if sum != expected {
		return fmt.Errorf("got %s, expected %s", sum, expected)
	}
	return nil
}

func defaultCacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "stdix")
}
