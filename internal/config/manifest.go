package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Manifest represents the ait.yml project manifest
type Manifest struct {
	Name         string         `yaml:"name"`
	Version      string         `yaml:"version"`
	Description  string         `yaml:"description,omitempty"`
	Dependencies []string       `yaml:"dependencies"`
	Targets      []string       `yaml:"targets,omitempty"`
	Overrides    map[string]any `yaml:"overrides,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshaling to support both
// the new flat list format and legacy nested formats
func (m *Manifest) UnmarshalYAML(node *yaml.Node) error {
	// Create a temporary struct to decode into
	type rawManifest struct {
		Name         string         `yaml:"name"`
		Version      string         `yaml:"version"`
		Description  string         `yaml:"description,omitempty"`
		Dependencies yaml.Node      `yaml:"dependencies"`
		Targets      []string       `yaml:"targets,omitempty"`
		Overrides    map[string]any `yaml:"overrides,omitempty"`
	}

	var raw rawManifest
	if err := node.Decode(&raw); err != nil {
		return err
	}

	// Copy simple fields
	m.Name = raw.Name
	m.Version = raw.Version
	m.Description = raw.Description
	m.Targets = raw.Targets
	m.Overrides = raw.Overrides

	// Parse dependencies - support both flat list and nested formats
	m.Dependencies = parseDependencies(&raw.Dependencies)

	return nil
}

// parseDependencies handles both flat list and legacy nested formats
func parseDependencies(node *yaml.Node) []string {
	deps := make([]string, 0)

	// Check if it's a sequence (flat list - new format)
	if node.Kind == yaml.SequenceNode {
		var list []string
		if err := node.Decode(&list); err == nil {
			return list
		}
	}

	// Check if it's a mapping (nested format - legacy or APM-style)
	if node.Kind == yaml.MappingNode {
		var nested map[string][]string
		if err := node.Decode(&nested); err == nil {
			// Support new APM-style nested format
			if apm, ok := nested["apm"]; ok {
				deps = append(deps, apm...)
			}
			// Support legacy format
			if agents, ok := nested["agents"]; ok {
				deps = append(deps, agents...)
			}
			if skills, ok := nested["skills"]; ok {
				deps = append(deps, skills...)
			}
			if prompts, ok := nested["prompts"]; ok {
				deps = append(deps, prompts...)
			}
			// Always include MCP
			if mcp, ok := nested["mcp"]; ok {
				deps = append(deps, mcp...)
			}
		}
	}

	return deps
}

// LoadManifest loads ait.yml from the specified path
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// Write saves the manifest to the specified path
func (m *Manifest) Write(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// ToYAML converts the manifest to YAML string
func (m *Manifest) ToYAML() (string, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest: %w", err)
	}
	return string(data), nil
}

// Exists checks if a manifest file exists at the given path
func ManifestExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
