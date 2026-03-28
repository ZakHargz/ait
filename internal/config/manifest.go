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
	Dependencies Dependencies   `yaml:"dependencies"`
	Targets      []string       `yaml:"targets,omitempty"`
	Overrides    map[string]any `yaml:"overrides,omitempty"`
}

// Dependencies represents all types of dependencies
type Dependencies struct {
	Agents  []string `yaml:"agents,omitempty"`
	Skills  []string `yaml:"skills,omitempty"`
	Prompts []string `yaml:"prompts,omitempty"`
	MCP     []string `yaml:"mcp,omitempty"`
}

// All returns all dependencies as a flat list
func (d *Dependencies) All() []string {
	all := make([]string, 0)
	all = append(all, d.Agents...)
	all = append(all, d.Skills...)
	all = append(all, d.Prompts...)
	all = append(all, d.MCP...)
	return all
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
