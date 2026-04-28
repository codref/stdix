package config

import (
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

// DBPath returns the path to the local registry.db file.
func DBPath(cfg *Config) string {
	if cfg.Registry.Source == "local" && cfg.Registry.DB != "" {
		return cfg.Registry.DB
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "stdix", "registry.db")
}
