package adapters

import (
	"path/filepath"

	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/utils"
)

// ClaudeAdapter implements the Adapter interface for Claude Desktop
type ClaudeAdapter struct {
	BaseAdapter
}

// NewClaudeAdapter creates a new Claude Desktop adapter
func NewClaudeAdapter() *ClaudeAdapter {
	home := utils.HomeDir()
	configDir := filepath.Join(home, ".claude")

	return &ClaudeAdapter{
		BaseAdapter: NewBaseAdapter("claude", configDir),
	}
}

// Detect checks if Claude Desktop is installed
func (a *ClaudeAdapter) Detect() bool {
	// Check if config directory exists
	if utils.DirExists(a.configDir) {
		return true
	}

	// Also check for Claude app on macOS
	if utils.DirExists("/Applications/Claude.app") {
		return true
	}

	return false
}

// InstallAgent installs an agent package for Claude Desktop
func (a *ClaudeAdapter) InstallAgent(pkg *packages.Package) error {
	// Claude uses same format as OpenCode: agents/<name>/AGENT.md
	return InstallPackageFile(pkg, a.configDir, "claude", PackageInstallConfig{
		TargetSubdir:     "agents",
		SourceFileName:   "AGENT.md",
		DestFileName:     "AGENT.md",
		UsePackageSubdir: true,
	})
}

// InstallSkill installs a skill package for Claude Desktop
func (a *ClaudeAdapter) InstallSkill(pkg *packages.Package) error {
	// When the package uses the APM .apm/skills/ layout, copy the entire skill
	// directory (SKILL.md + bundled resources) to match APM's install behaviour.
	if pkg.ApmSkillDir != "" {
		return InstallSkillDir(pkg, a.configDir, "skills")
	}
	// Claude uses same format as OpenCode: skills/<name>/SKILL.md
	return InstallPackageFile(pkg, a.configDir, "claude", PackageInstallConfig{
		TargetSubdir:     "skills",
		SourceFileName:   "SKILL.md",
		DestFileName:     "SKILL.md",
		UsePackageSubdir: true,
	})
}

// InstallPrompt installs a prompt package for Claude Desktop
func (a *ClaudeAdapter) InstallPrompt(pkg *packages.Package) error {
	// Claude uses same format as OpenCode: prompts/<name>.txt
	return InstallPackageFile(pkg, a.configDir, "claude", PackageInstallConfig{
		TargetSubdir:     "prompts",
		SourceFileName:   "prompt.txt",
		DestFileName:     pkg.Name + ".txt",
		UsePackageSubdir: false,
	})
}

// Uninstall removes a package from Claude Desktop
func (a *ClaudeAdapter) Uninstall(pkg *packages.Package) error {
	return UninstallPackage(pkg, a.configDir, "agents", "skills", "prompts")
}

// List returns all installed packages for Claude Desktop
func (a *ClaudeAdapter) List() ([]*packages.Package, error) {
	return ListPackages(a.configDir, "agents", "skills", "prompts")
}

// Validate checks if Claude Desktop installation is healthy
func (a *ClaudeAdapter) Validate() error {
	return ValidateConfigDir(a.configDir, true)
}
