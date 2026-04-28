package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load walks registryRoot/standards/**/*.yaml, parses each file, validates it,
// and returns the full list of standards. Returns an error on the first
// parse or validation failure, including duplicate ID detection.
func Load(registryRoot string) ([]Standard, error) {
	standardsDir := filepath.Join(registryRoot, "standards")
	if _, err := os.Stat(standardsDir); err != nil {
		return nil, fmt.Errorf("standards directory not found at %s", standardsDir)
	}

	var standards []Standard
	seen := map[string]string{} // id → file path

	err := filepath.WalkDir(standardsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		var s Standard
		if err := yaml.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}
		if err := Validate(&s); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if prev, ok := seen[s.ID]; ok {
			return fmt.Errorf("duplicate standard ID %q: found in %s and %s", s.ID, prev, path)
		}
		seen[s.ID] = path
		standards = append(standards, s)
		return nil
	})
	return standards, err
}
