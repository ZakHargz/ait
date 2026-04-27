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
	TypeHybrid = config.PackageTypeHybrid
)

// Package represents an installed package
type Package struct {
	Name     string
	Version  string
	Type     PackageType
	Path     string // Local path to package files
	Metadata *config.PackageMetadata
	// ApmAgentFile is the absolute path to the agent file found in .apm/agents/
	// (e.g. .apm/agents/my-package.agent.md). Set when the package uses the APM layout.
	ApmAgentFile string
	// ApmSkillDir is the absolute path to the skill directory found in .apm/skills/<name>/.
	// When set, the entire directory (SKILL.md + bundled resources) should be copied on install.
	ApmSkillDir string
}

// GetFile returns the platform-specific file for this package.
// For APM-compatible packages that may not have the standard file name,
// it walks a list of candidates and returns the first one that exists on disk.
//
// The .apm/ layout takes priority: if ApmAgentFile or ApmSkillDir is set the
// caller should use those directly — GetFile returns "" in that case so the
// adapter can detect the situation and call InstallPackageDir instead.
func (p *Package) GetFile(platform string) string {
	if p.Metadata != nil {
		// If the metadata explicitly maps this platform, trust it.
		if file, ok := p.Metadata.Files[platform]; ok {
			return file
		}
	}

	// APM .apm/ layout: signal to the caller that it should use the dedicated
	// ApmAgentFile / ApmSkillDir paths instead of a single file copy.
	// Return "" so InstallPackageFile falls through to the SourceFileName, and
	// the adapter checks ApmAgentFile/ApmSkillDir before building the source path.
	if p.ApmAgentFile != "" || p.ApmSkillDir != "" {
		return ""
	}

	// Build candidate list based on normalised type, falling back through
	// common APM-compatible names so packages install cleanly.
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

// effectiveType returns the package type for file-resolution purposes.
// Hybrid packages are treated as agents when a single type is needed.
func (p *Package) effectiveType() PackageType {
	if p.Metadata != nil {
		return p.Metadata.NormaliseType()
	}
	return p.Type
}
