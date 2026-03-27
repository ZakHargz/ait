package adapters

import (
	"fmt"
	"os"
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
	targetDir := filepath.Join(a.configDir, "agents", pkg.Name)

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Get source file
	sourceFile := pkg.GetFile("claude")
	if sourceFile == "" {
		sourceFile = "AGENT.md"
	}

	source := filepath.Join(pkg.Path, sourceFile)
	dest := filepath.Join(targetDir, "AGENT.md")

	// Copy file
	if err := utils.CopyFile(source, dest); err != nil {
		return fmt.Errorf("failed to install agent: %w", err)
	}

	return nil
}

// InstallSkill installs a skill package for Claude Desktop
func (a *ClaudeAdapter) InstallSkill(pkg *packages.Package) error {
	// Claude uses same format as OpenCode: skills/<name>/SKILL.md
	targetDir := filepath.Join(a.configDir, "skills", pkg.Name)

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Get source file
	sourceFile := pkg.GetFile("claude")
	if sourceFile == "" {
		sourceFile = "SKILL.md"
	}

	source := filepath.Join(pkg.Path, sourceFile)
	dest := filepath.Join(targetDir, "SKILL.md")

	// Copy file
	if err := utils.CopyFile(source, dest); err != nil {
		return fmt.Errorf("failed to install skill: %w", err)
	}

	return nil
}

// InstallPrompt installs a prompt package for Claude Desktop
func (a *ClaudeAdapter) InstallPrompt(pkg *packages.Package) error {
	// Claude uses same format as OpenCode: prompts/<name>.txt
	targetDir := filepath.Join(a.configDir, "prompts")

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Get source file
	sourceFile := pkg.GetFile("claude")
	if sourceFile == "" {
		sourceFile = "prompt.txt"
	}

	source := filepath.Join(pkg.Path, sourceFile)
	dest := filepath.Join(targetDir, pkg.Name+".txt")

	// Copy file
	if err := utils.CopyFile(source, dest); err != nil {
		return fmt.Errorf("failed to install prompt: %w", err)
	}

	return nil
}

// Uninstall removes a package from Claude Desktop
func (a *ClaudeAdapter) Uninstall(pkg *packages.Package) error {
	var targetPath string

	switch pkg.Type {
	case packages.TypeAgent:
		targetPath = filepath.Join(a.configDir, "agents", pkg.Name)
	case packages.TypeSkill:
		targetPath = filepath.Join(a.configDir, "skills", pkg.Name)
	case packages.TypePrompt:
		targetPath = filepath.Join(a.configDir, "prompts", pkg.Name+".txt")
	default:
		return fmt.Errorf("unsupported package type: %s", pkg.Type)
	}

	return os.RemoveAll(targetPath)
}

// List returns all installed packages for Claude Desktop
func (a *ClaudeAdapter) List() ([]*packages.Package, error) {
	installed := []*packages.Package{}

	// List agents
	agentsDir := filepath.Join(a.configDir, "agents")
	if agents, err := listPackagesInDir(agentsDir, packages.TypeAgent); err == nil {
		installed = append(installed, agents...)
	}

	// List skills
	skillsDir := filepath.Join(a.configDir, "skills")
	if skills, err := listPackagesInDir(skillsDir, packages.TypeSkill); err == nil {
		installed = append(installed, skills...)
	}

	// List prompts
	promptsDir := filepath.Join(a.configDir, "prompts")
	if prompts, err := listPromptsInDir(promptsDir); err == nil {
		installed = append(installed, prompts...)
	}

	return installed, nil
}

// Validate checks if Claude Desktop installation is healthy
func (a *ClaudeAdapter) Validate() error {
	// Check if config directory exists and is writable
	if !utils.DirExists(a.configDir) {
		// Try to create it
		if err := utils.EnsureDir(a.configDir); err != nil {
			return fmt.Errorf("claude config directory does not exist and cannot be created: %s", a.configDir)
		}
	}

	if err := utils.CheckDirWritable(a.configDir); err != nil {
		return fmt.Errorf("claude config directory not writable: %w", err)
	}

	return nil
}
