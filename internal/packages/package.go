package packages

import (
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

// GetFile returns the platform-specific file for this package
func (p *Package) GetFile(platform string) string {
	if p.Metadata != nil {
		return p.Metadata.GetFile(platform)
	}

	// Return default based on type
	switch p.Type {
	case TypeAgent:
		return "AGENT.md"
	case TypeSkill:
		return "SKILL.md"
	case TypePrompt:
		return "prompt.txt"
	default:
		return ""
	}
}
