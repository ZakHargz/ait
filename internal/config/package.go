package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PackageType represents the type of package
type PackageType string

const (
	PackageTypeAgent  PackageType = "agent"
	PackageTypeSkill  PackageType = "skill"
	PackageTypePrompt PackageType = "prompt"
	PackageTypeMCP    PackageType = "mcp"
)

// PackageMetadata represents the package.yml metadata file
type PackageMetadata struct {
	Name          string            `yaml:"name"`
	Version       string            `yaml:"version"`
	Type          PackageType       `yaml:"type"`
	Description   string            `yaml:"description"`
	Author        Author            `yaml:"author,omitempty"`
	License       string            `yaml:"license,omitempty"`
	Dependencies  Dependencies      `yaml:"dependencies,omitempty"`
	Compatibility []string          `yaml:"compatibility"`
	Files         map[string]string `yaml:"files"`
	Tags          []string          `yaml:"tags,omitempty"`
	Keywords      []string          `yaml:"keywords,omitempty"`
	Requires      Requirements      `yaml:"requires,omitempty"`
	Repository    Repository        `yaml:"repository,omitempty"`
}

// Author represents package author information
type Author struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email,omitempty"`
	URL   string `yaml:"url,omitempty"`
}

// Requirements represents minimum version requirements
type Requirements struct {
	AIT string `yaml:"ait,omitempty"`
}

// Repository represents the package repository information
type Repository struct {
	Type      string `yaml:"type"`
	URL       string `yaml:"url"`
	Directory string `yaml:"directory,omitempty"`
}

// LoadPackageMetadata loads package.yml from the specified path
func LoadPackageMetadata(path string) (*PackageMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read package metadata: %w", err)
	}

	var metadata PackageMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse package metadata: %w", err)
	}

	return &metadata, nil
}

// Write saves the package metadata to the specified path
func (p *PackageMetadata) Write(path string) error {
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal package metadata: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write package metadata: %w", err)
	}

	return nil
}

// GetFile returns the platform-specific file path
func (p *PackageMetadata) GetFile(platform string) string {
	if file, ok := p.Files[platform]; ok {
		return file
	}

	// Return default based on type
	switch p.Type {
	case PackageTypeAgent:
		return "AGENT.md"
	case PackageTypeSkill:
		return "SKILL.md"
	case PackageTypePrompt:
		return "prompt.txt"
	default:
		return ""
	}
}
