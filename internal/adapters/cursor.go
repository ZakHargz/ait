package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/utils"
)

// CursorAdapter implements the Adapter interface for Cursor
type CursorAdapter struct {
	BaseAdapter
}

// NewCursorAdapter creates a new Cursor adapter
func NewCursorAdapter() *CursorAdapter {
	home := utils.HomeDir()

	// Cursor config location varies by OS
	var configDir string
	if _, err := os.Stat(filepath.Join(home, "Library", "Application Support", "Cursor")); err == nil {
		// macOS
		configDir = filepath.Join(home, "Library", "Application Support", "Cursor", "User")
	} else if _, err := os.Stat(filepath.Join(home, ".config", "Cursor")); err == nil {
		// Linux
		configDir = filepath.Join(home, ".config", "Cursor", "User")
	} else {
		// Windows (AppData)
		configDir = filepath.Join(home, "AppData", "Roaming", "Cursor", "User")
	}

	return &CursorAdapter{
		BaseAdapter: NewBaseAdapter("cursor", configDir),
	}
}

// Detect checks if Cursor is installed
func (a *CursorAdapter) Detect() bool {
	// Check if config directory exists
	if utils.DirExists(a.configDir) {
		return true
	}
	return false
}

// InstallAgent installs an agent package for Cursor
// Cursor doesn't have native agent support, so we'll create custom directories
func (a *CursorAdapter) InstallAgent(pkg *packages.Package) error {
	// Create a custom directory for agents
	// Cursor/User/ait-agents/<name>/.cursorrules
	targetDir := filepath.Join(a.configDir, "ait-agents", pkg.Name)

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create agent directory: %w", err)
	}

	// Get source AGENT.md file
	sourceFile := pkg.GetFile("cursor")
	if sourceFile == "" {
		sourceFile = "AGENT.md"
	}
	source := filepath.Join(pkg.Path, sourceFile)

	// Read agent content
	content, err := utils.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read agent file: %w", err)
	}

	// Convert AGENT.md to .cursorrules format
	// Extract the main content (skip YAML frontmatter)
	cursorRules := convertToCursorRules(string(content), pkg.Name)

	// Write .cursorrules file
	dest := filepath.Join(targetDir, ".cursorrules")
	if err := os.WriteFile(dest, []byte(cursorRules), 0644); err != nil {
		return fmt.Errorf("failed to write .cursorrules: %w", err)
	}

	// Also write README with instructions
	readme := fmt.Sprintf(`# %s Agent for Cursor

This agent was installed by AIT (AI Toolkit Package Manager).

## Usage

To use this agent in Cursor:
1. Open your project in Cursor
2. Copy .cursorrules to your project root
3. Or set this directory as your Cursor rules location

## Original Package

- Name: %s
- Version: %s
- Type: %s

## Files

- .cursorrules - Agent rules for Cursor
`, pkg.Name, pkg.Name, pkg.Version, pkg.Type)

	readmePath := filepath.Join(targetDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		// Non-fatal, just warn
		utils.PrintWarning(fmt.Sprintf("Could not write README: %s", err.Error()))
	}

	return nil
}

// InstallSkill installs a skill package for Cursor
func (a *CursorAdapter) InstallSkill(pkg *packages.Package) error {
	// Similar to agents, create custom directory
	targetDir := filepath.Join(a.configDir, "ait-skills", pkg.Name)

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Get source file
	sourceFile := pkg.GetFile("cursor")
	if sourceFile == "" {
		sourceFile = "SKILL.md"
	}

	source := filepath.Join(pkg.Path, sourceFile)
	dest := filepath.Join(targetDir, "skill.md")

	// Copy file
	if err := utils.CopyFile(source, dest); err != nil {
		return fmt.Errorf("failed to install skill: %w", err)
	}

	return nil
}

// InstallPrompt installs a prompt package for Cursor
func (a *CursorAdapter) InstallPrompt(pkg *packages.Package) error {
	// Create prompts directory
	targetDir := filepath.Join(a.configDir, "ait-prompts")

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Get source file
	sourceFile := pkg.GetFile("cursor")
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

// Uninstall removes a package from Cursor
func (a *CursorAdapter) Uninstall(pkg *packages.Package) error {
	var targetPath string

	switch pkg.Type {
	case packages.TypeAgent:
		targetPath = filepath.Join(a.configDir, "ait-agents", pkg.Name)
	case packages.TypeSkill:
		targetPath = filepath.Join(a.configDir, "ait-skills", pkg.Name)
	case packages.TypePrompt:
		targetPath = filepath.Join(a.configDir, "ait-prompts", pkg.Name+".txt")
	default:
		return fmt.Errorf("unsupported package type: %s", pkg.Type)
	}

	return os.RemoveAll(targetPath)
}

// List returns all installed packages for Cursor
func (a *CursorAdapter) List() ([]*packages.Package, error) {
	installed := []*packages.Package{}

	// List agents
	agentsDir := filepath.Join(a.configDir, "ait-agents")
	if agents, err := listPackagesInDir(agentsDir, packages.TypeAgent); err == nil {
		installed = append(installed, agents...)
	}

	// List skills
	skillsDir := filepath.Join(a.configDir, "ait-skills")
	if skills, err := listPackagesInDir(skillsDir, packages.TypeSkill); err == nil {
		installed = append(installed, skills...)
	}

	// List prompts
	promptsDir := filepath.Join(a.configDir, "ait-prompts")
	if prompts, err := listPromptsInDir(promptsDir); err == nil {
		installed = append(installed, prompts...)
	}

	return installed, nil
}

// Validate checks if Cursor installation is healthy
func (a *CursorAdapter) Validate() error {
	// Check if config directory exists
	if !utils.DirExists(a.configDir) {
		return fmt.Errorf("cursor config directory does not exist: %s", a.configDir)
	}

	return nil
}

// convertToCursorRules converts AGENT.md content to .cursorrules format
func convertToCursorRules(content string, agentName string) string {
	// Split content into lines
	lines := strings.Split(content, "\n")

	// Skip YAML frontmatter if present
	inFrontmatter := false
	startIdx := 0
	for i, line := range lines {
		if i == 0 && strings.TrimSpace(line) == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.TrimSpace(line) == "---" {
			startIdx = i + 1
			break
		}
	}

	// Get content after frontmatter
	var contentLines []string
	if startIdx > 0 {
		contentLines = lines[startIdx:]
	} else {
		contentLines = lines
	}

	// Join back together
	cleanContent := strings.Join(contentLines, "\n")
	cleanContent = strings.TrimSpace(cleanContent)

	// Add Cursor-specific header
	result := fmt.Sprintf(`# Cursor Rules: %s

%s

---
Generated by AIT (AI Toolkit Package Manager)
`, agentName, cleanContent)

	return result
}
