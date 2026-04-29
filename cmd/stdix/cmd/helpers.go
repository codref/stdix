package cmd

import (
	"github.com/codref/stdix/internal/config"
)

// loadConfig loads .stdix.yaml from the preferred search paths: cwd first,
// then ~/.stdix (cwd wins).
func loadConfig(dir string) (*config.Config, error) {
	return config.LoadAuto(dir)
}
