package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const Filename = ".stdix.yaml"

type Registry struct {
	Source   string `yaml:"source"`
	URL      string `yaml:"url,omitempty"`
	DB       string `yaml:"db,omitempty"`
	Checksum string `yaml:"checksum,omitempty"`
	Repo     string `yaml:"repo,omitempty"` // GitHub repo for stdix push, e.g. "owner/stdix-registry"
}

type Project struct {
	Language   string   `yaml:"language,omitempty"`
	Frameworks []string `yaml:"frameworks,omitempty"`
}

type Outputs struct {
	Agents  bool `yaml:"agents"`
	Claude  bool `yaml:"claude"`
	Copilot bool `yaml:"copilot"`
	Cursor  bool `yaml:"cursor"`
}

type Config struct {
	Registry  Registry `yaml:"registry"`
	Project   Project  `yaml:"project,omitempty"`
	Standards []string `yaml:"standards,omitempty"`
	Outputs   Outputs  `yaml:"outputs"`
}

// Default returns the configuration written by stdix init.
func Default() *Config {
	return &Config{
		Registry: Registry{
			Source: "local",
		},
		Outputs: Outputs{
			Agents:  true,
			Claude:  true,
			Copilot: true,
			Cursor:  true,
		},
	}
}

// Load reads and parses .stdix.yaml from dir.
func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, Filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

// Save writes cfg as YAML to dir/.stdix.yaml.
func Save(dir string, cfg *Config) error {
	path := filepath.Join(dir, Filename)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// Exists reports whether .stdix.yaml is present in dir.
func Exists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, Filename))
	return err == nil
}

// HomeConfigDir returns the directory where the user-level config lives: ~/.stdix.
func HomeConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".stdix")
}

// SearchPaths returns the ordered list of directories to search for .stdix.yaml,
// highest priority first. The current working directory wins over the home config.
func SearchPaths(cwd string) []string {
	return []string{cwd, HomeConfigDir()}
}

// LoadAuto searches for .stdix.yaml in the preferred path order — cwd first,
// then ~/.stdix — and returns the first config found. cwd always wins.
func LoadAuto(cwd string) (*Config, error) {
	for _, dir := range SearchPaths(cwd) {
		cfg, err := Load(dir)
		if err == nil {
			return cfg, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("%s not found in cwd or %s", Filename, HomeConfigDir())
}

// ExistsAny reports whether .stdix.yaml is present in any of the search paths.
func ExistsAny(cwd string) bool {
	for _, dir := range SearchPaths(cwd) {
		if Exists(dir) {
			return true
		}
	}
	return false
}

// DBPath returns the path to the local registry.db file.
func DBPath(cfg *Config) string {
	if cfg.Registry.Source == "local" && cfg.Registry.DB != "" {
		return cfg.Registry.DB
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "stdix", "registry.db")
}
