package adapters

import (
	"fmt"
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
	targetDir := filepath.Join(a.configDir, "agents", pkg.Name)

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Get source file
	sourceFile := pkg.GetFile("opencode")
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

// InstallSkill installs a skill package for OpenCode
func (a *OpenCodeAdapter) InstallSkill(pkg *packages.Package) error {
	// OpenCode expects: ~/.config/opencode/skills/<name>/SKILL.md
	targetDir := filepath.Join(a.configDir, "skills", pkg.Name)

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Get source file
	sourceFile := pkg.GetFile("opencode")
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

// InstallPrompt installs a prompt package for OpenCode
func (a *OpenCodeAdapter) InstallPrompt(pkg *packages.Package) error {
	// OpenCode expects: ~/.config/opencode/prompts/<name>.txt
	targetDir := filepath.Join(a.configDir, "prompts")

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Get source file
	sourceFile := pkg.GetFile("opencode")
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

// Uninstall removes a package from OpenCode
func (a *OpenCodeAdapter) Uninstall(pkg *packages.Package) error {
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

// List returns all installed packages for OpenCode
func (a *OpenCodeAdapter) List() ([]*packages.Package, error) {
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

// Validate checks if OpenCode installation is healthy
func (a *OpenCodeAdapter) Validate() error {
	// Check if config directory exists and is writable
	if !utils.DirExists(a.configDir) {
		return fmt.Errorf("opencode config directory does not exist: %s", a.configDir)
	}

	if err := utils.CheckDirWritable(a.configDir); err != nil {
		return fmt.Errorf("opencode config directory not writable: %w", err)
	}

	return nil
}

// Helper functions

func listPackagesInDir(dir string, pkgType packages.PackageType) ([]*packages.Package, error) {
	if !utils.DirExists(dir) {
		return []*packages.Package{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	pkgs := []*packages.Package{}
	for _, entry := range entries {
		if entry.IsDir() {
			pkgs = append(pkgs, &packages.Package{
				Name: entry.Name(),
				Type: pkgType,
				Path: filepath.Join(dir, entry.Name()),
			})
		}
	}

	return pkgs, nil
}

func listPromptsInDir(dir string) ([]*packages.Package, error) {
	if !utils.DirExists(dir) {
		return []*packages.Package{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	pkgs := []*packages.Package{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".txt" {
			name := entry.Name()[:len(entry.Name())-4] // Remove .txt extension
			pkgs = append(pkgs, &packages.Package{
				Name: name,
				Type: packages.TypePrompt,
				Path: filepath.Join(dir, entry.Name()),
			})
		}
	}

	return pkgs, nil
}
