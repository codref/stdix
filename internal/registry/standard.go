package registry

// Standard represents a single engineering standard parsed from YAML.
type Standard struct {
	ID          string   `yaml:"id"           json:"id"`
	Title       string   `yaml:"title"        json:"title"`
	Version     string   `yaml:"version"      json:"version"`
	Language    string   `yaml:"language,omitempty" json:"language,omitempty"`
	Tags        []string `yaml:"tags,omitempty"    json:"tags,omitempty"`
	AppliesWhen []string `yaml:"applies_when,omitempty" json:"applies_when,omitempty"`
	Rules       []string `yaml:"rules"        json:"rules"`
	Outputs     Outputs  `yaml:"outputs,omitempty" json:"outputs,omitempty"`
}

// Outputs controls which agent files are generated for a standard.
type Outputs struct {
	Agents  bool `yaml:"agents"  json:"agents"`
	Claude  bool `yaml:"claude"  json:"claude"`
	Copilot bool `yaml:"copilot" json:"copilot"`
	Cursor  bool `yaml:"cursor"  json:"cursor"`
}
