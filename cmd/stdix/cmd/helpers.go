package cmd

import (
	"github.com/codref/stdix/internal/config"
)

// loadConfig loads .stdix.yaml from dir, returning a default config on error.
func loadConfig(dir string) (*config.Config, error) {
	return config.Load(dir)
}
