package adapters

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/utils"
)

// OpenCodeAdapter implements the Adapter interface for OpenCode
type OpenCodeAdapter struct {
	BaseAdapter
}

// NewOpenCodeAdapter creates a new OpenCode adapter
func NewOpenCodeAdapter() *OpenCodeAdapter {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "opencode")

	return &OpenCodeAdapter{
		BaseAdapter: NewBaseAdapter("opencode", configDir),
	}
}

// Detect checks if OpenCode is installed
func (a *OpenCodeAdapter) Detect() bool {
	// Check if opencode binary exists
	if _, err := exec.LookPath("opencode"); err == nil {
		return true
	}

	// Check if config directory exists
	if utils.DirExists(a.configDir) {
		return true
	}

	return false
}

// InstallAgent installs an agent package for OpenCode
func (a *OpenCodeAdapter) InstallAgent(pkg *packages.Package) error {
	// OpenCode expects: ~/.config/opencode/agents/<name>/AGENT.md
	return InstallPackageFile(pkg, a.configDir, "opencode", PackageInstallConfig{
		TargetSubdir:     "agents",
		SourceFileName:   "AGENT.md",
		DestFileName:     "AGENT.md",
		UsePackageSubdir: true,
	})
}

// InstallSkill installs a skill package for OpenCode
func (a *OpenCodeAdapter) InstallSkill(pkg *packages.Package) error {
	// When the package uses the APM .apm/skills/ layout, copy the entire skill
	// directory (SKILL.md + bundled resources) to match APM's install behaviour.
	if pkg.ApmSkillDir != "" {
		return InstallSkillDir(pkg, a.configDir, "skills")
	}
	// OpenCode expects: ~/.config/opencode/skills/<name>/SKILL.md
	return InstallPackageFile(pkg, a.configDir, "opencode", PackageInstallConfig{
		TargetSubdir:     "skills",
		SourceFileName:   "SKILL.md",
		DestFileName:     "SKILL.md",
		UsePackageSubdir: true,
	})
}

// InstallPrompt installs a prompt package for OpenCode
func (a *OpenCodeAdapter) InstallPrompt(pkg *packages.Package) error {
	// OpenCode expects: ~/.config/opencode/prompts/<name>.txt
	return InstallPackageFile(pkg, a.configDir, "opencode", PackageInstallConfig{
		TargetSubdir:     "prompts",
		SourceFileName:   "prompt.txt",
		DestFileName:     pkg.Name + ".txt",
		UsePackageSubdir: false,
	})
}

// Uninstall removes a package from OpenCode
func (a *OpenCodeAdapter) Uninstall(pkg *packages.Package) error {
	return UninstallPackage(pkg, a.configDir, "agents", "skills", "prompts")
}

// List returns all installed packages for OpenCode
func (a *OpenCodeAdapter) List() ([]*packages.Package, error) {
	return ListPackages(a.configDir, "agents", "skills", "prompts")
}

// Validate checks if OpenCode installation is healthy
func (a *OpenCodeAdapter) Validate() error {
	return ValidateConfigDir(a.configDir, false)
}
