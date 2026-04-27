package packages

import (
	"os"
	"path/filepath"

	"github.com/apex-ai/ait/internal/config"
)

// PackageType represents the type of package
type PackageType = config.PackageType

const (
	TypeAgent  = config.PackageTypeAgent
	TypeSkill  = config.PackageTypeSkill
	TypePrompt = config.PackageTypePrompt
	TypeMCP    = config.PackageTypeMCP
)

// Package represents an installed package
type Package struct {
	Name     string
	Version  string
	Type     PackageType
	Path     string // Local path to package files
	Metadata *config.PackageMetadata
}

// GetFile returns the platform-specific file for this package.
// For APM-compatible packages that may not have the standard file name,
// it walks a list of candidates and returns the first one that exists on disk.
func (p *Package) GetFile(platform string) string {
	if p.Metadata != nil {
		// If the metadata explicitly maps this platform, trust it.
		if file, ok := p.Metadata.Files[platform]; ok {
			return file
		}
	}

	// Build candidate list based on normalised type, falling back through
	// common APM-compatible names so hybrid packages install cleanly.
	var candidates []string
	switch p.effectiveType() {
	case TypeAgent:
		candidates = []string{"AGENT.md", "README.md", "INSTRUCTIONS.md"}
	case TypeSkill:
		candidates = []string{"SKILL.md", "README.md"}
	case TypePrompt:
		candidates = []string{"prompt.txt", "PROMPT.md"}
	default:
		return ""
	}

	// Return the first candidate that actually exists on disk.
	for _, name := range candidates {
		if p.Path != "" {
			if _, err := os.Stat(filepath.Join(p.Path, name)); err == nil {
				return name
			}
		}
	}

	// Nothing found on disk — return the canonical default so callers get a
	// meaningful error rather than an empty string.
	return candidates[0]
}

// effectiveType returns the package type, normalising APM hybrid → agent.
func (p *Package) effectiveType() PackageType {
	if p.Metadata != nil {
		return p.Metadata.NormaliseType()
	}
	return p.Type
}
