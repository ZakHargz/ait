package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PackageType represents the type of package
type PackageType string

const (
	PackageTypeAgent  PackageType = "agent"
	PackageTypeSkill  PackageType = "skill"
	PackageTypePrompt PackageType = "prompt"
	PackageTypeMCP    PackageType = "mcp"
	// PackageTypeHybrid is an APM-compatible type for packages that combine
	// instructions, skills and prompts. AIT treats it as an agent.
	PackageTypeHybrid PackageType = "hybrid"
)

// PackageMetadata represents the package.yml metadata file
type PackageMetadata struct {
	Name          string            `yaml:"name"`
	Version       string            `yaml:"version"`
	Type          PackageType       `yaml:"type"`
	Description   string            `yaml:"description"`
	Author        Author            `yaml:"author,omitempty"`
	License       string            `yaml:"license,omitempty"`
	Dependencies  []string          `yaml:"dependencies,omitempty"`
	Compatibility []string          `yaml:"compatibility"`
	Files         map[string]string `yaml:"files"`
	Tags          []string          `yaml:"tags,omitempty"`
	Keywords      []string          `yaml:"keywords,omitempty"`
	Requires      Requirements      `yaml:"requires,omitempty"`
	Repository    Repository        `yaml:"repository,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshaling to support both
// the new flat list format and legacy nested formats for dependencies
func (p *PackageMetadata) UnmarshalYAML(node *yaml.Node) error {
	// Create a temporary struct to decode into
	type rawPackageMetadata struct {
		Name          string            `yaml:"name"`
		Version       string            `yaml:"version"`
		Type          PackageType       `yaml:"type"`
		Description   string            `yaml:"description"`
		Author        Author            `yaml:"author,omitempty"`
		License       string            `yaml:"license,omitempty"`
		Dependencies  yaml.Node         `yaml:"dependencies,omitempty"`
		Compatibility []string          `yaml:"compatibility"`
		Files         map[string]string `yaml:"files"`
		Tags          []string          `yaml:"tags,omitempty"`
		Keywords      []string          `yaml:"keywords,omitempty"`
		Requires      Requirements      `yaml:"requires,omitempty"`
		Repository    Repository        `yaml:"repository,omitempty"`
	}

	var raw rawPackageMetadata
	if err := node.Decode(&raw); err != nil {
		return err
	}

	// Copy all fields
	p.Name = raw.Name
	p.Version = raw.Version
	p.Type = raw.Type
	p.Description = raw.Description
	p.Author = raw.Author
	p.License = raw.License
	p.Compatibility = raw.Compatibility
	p.Files = raw.Files
	p.Tags = raw.Tags
	p.Keywords = raw.Keywords
	p.Requires = raw.Requires
	p.Repository = raw.Repository

	// Parse dependencies - support both flat list and nested formats
	p.Dependencies = parseDependencies(&raw.Dependencies)

	return nil
}

// Author represents package author information.
// Supports both struct form { name: ..., email: ... } and plain string form "Name <email>".
type Author struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email,omitempty"`
	URL   string `yaml:"url,omitempty"`
}

// UnmarshalYAML allows Author to be specified as either a mapping or a plain string.
func (a *Author) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		// Plain string: "Author Name" or "Author Name <email>"
		a.Name = node.Value
		return nil
	case yaml.MappingNode:
		type authorRaw Author
		var raw authorRaw
		if err := node.Decode(&raw); err != nil {
			return err
		}
		*a = Author(raw)
		return nil
	default:
		return fmt.Errorf("cannot unmarshal author field")
	}
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

// packageMetadataCandidates is the ordered list of metadata filenames AIT looks for
// inside a package directory. package.yml is the native format; apm.yml is the APM fallback.
var packageMetadataCandidates = []string{"package.yml", "apm.yml"}

// FindPackageMetadata searches pkgDir for a supported metadata file and loads it.
// It tries package.yml first, then apm.yml.
func FindPackageMetadata(pkgDir string) (*PackageMetadata, error) {
	for _, name := range packageMetadataCandidates {
		path := pkgDir + "/" + name
		if _, err := os.Stat(path); err == nil {
			return LoadPackageMetadata(path)
		}
	}
	return nil, fmt.Errorf("no package metadata found in %s (looked for: package.yml, apm.yml)", pkgDir)
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

// FindApmPrimitives scans <pkgDir>/.apm/ for APM-layout agent and skill primitives.
// It returns:
//   - agentFile: absolute path to the first *.agent.md found in .apm/agents/
//     (prefers <pkgName>.agent.md if it exists, otherwise the first match)
//   - skillDir: absolute path to the skill directory in .apm/skills/<pkgName>/
//     (or the first available skill subdirectory if <pkgName> doesn't match)
//
// Either return value may be empty string if not found.
// This mirrors the APM spec layout: https://microsoft.github.io/apm/introduction/anatomy-of-an-apm-package/
func FindApmPrimitives(pkgDir, pkgName string) (agentFile, skillDir string) {
	apmDir := filepath.Join(pkgDir, ".apm")
	if _, err := os.Stat(apmDir); err != nil {
		return "", ""
	}

	// --- Agent: look in .apm/agents/ for *.agent.md ---
	agentsDir := filepath.Join(apmDir, "agents")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		var firstMatch string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".agent.md") {
				full := filepath.Join(agentsDir, entry.Name())
				// Prefer the file named <pkgName>.agent.md
				if entry.Name() == pkgName+".agent.md" {
					agentFile = full
					break
				}
				if firstMatch == "" {
					firstMatch = full
				}
			}
		}
		if agentFile == "" {
			agentFile = firstMatch
		}
	}

	// --- Skill: look in .apm/skills/<pkgName>/ ---
	skillsDir := filepath.Join(apmDir, "skills")
	namedSkill := filepath.Join(skillsDir, pkgName)
	if info, err := os.Stat(namedSkill); err == nil && info.IsDir() {
		skillDir = namedSkill
	} else {
		// Fall back to the first available skill subdirectory
		if entries, err := os.ReadDir(skillsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					skillDir = filepath.Join(skillsDir, entry.Name())
					break
				}
			}
		}
	}

	return agentFile, skillDir
}

// NormaliseType maps APM-compatible types to the closest AIT equivalent.
// hybrid → agent (instructions + prompts bundle, installed like an agent).
func (p *PackageMetadata) NormaliseType() PackageType {
	switch p.Type {
	case PackageTypeHybrid:
		return PackageTypeAgent
	default:
		return p.Type
	}
}

// GetFile returns the platform-specific file path
func (p *PackageMetadata) GetFile(platform string) string {
	if file, ok := p.Files[platform]; ok {
		return file
	}

	// Return default based on normalised type
	switch p.NormaliseType() {
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
