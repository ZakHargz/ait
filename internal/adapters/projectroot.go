package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/utils"
)

// ProjectRootAdapter installs packages using tool-native project-level conventions
// This adapter creates files that AI tools automatically detect (e.g., .cursorrules, .github/copilot-instructions.md)
type ProjectRootAdapter struct {
	BaseAdapter
	projectRoot string
}

// NewProjectRootAdapter creates a new project-root adapter
func NewProjectRootAdapter(projectRoot string) *ProjectRootAdapter {
	return &ProjectRootAdapter{
		BaseAdapter: NewBaseAdapter("project-root", projectRoot),
		projectRoot: projectRoot,
	}
}

// Detect checks if we're in a project with ait.yml
func (a *ProjectRootAdapter) Detect() bool {
	aitYmlPath := filepath.Join(a.projectRoot, "ait.yml")
	_, err := os.Stat(aitYmlPath)
	return err == nil
}

// GetConfigDir returns the project root
func (a *ProjectRootAdapter) GetConfigDir() (string, error) {
	return a.projectRoot, nil
}

// InstallAgent installs an agent using tool-native conventions
// Creates:
// - .cursorrules (for Cursor)
// - .github/copilot-instructions.md (for GitHub Copilot)
// - .opencode/agents/<name>/ (for OpenCode, if they support it)
func (a *ProjectRootAdapter) InstallAgent(pkg *packages.Package) error {
	// Read AGENT.md content
	agentFile := filepath.Join(pkg.Path, "AGENT.md")
	content, err := os.ReadFile(agentFile)
	if err != nil {
		return fmt.Errorf("AGENT.md not found: %w", err)
	}

	agentContent := string(content)

	// 1. Install for Cursor (.cursorrules at project root)
	if err := a.installCursorRules(pkg, agentContent); err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to create .cursorrules: %s", err.Error()))
	}

	// 2. Install for GitHub Copilot (.github/copilot-instructions.md)
	if err := a.installGitHubCopilot(pkg, agentContent); err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to create .github/copilot-instructions.md: %s", err.Error()))
	}

	// 3. Install for OpenCode (proposed .opencode/agents/ convention)
	if err := a.installOpenCodeAgent(pkg, agentContent); err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to create .opencode/agents/: %s", err.Error()))
	}

	utils.PrintSuccess("Created project-level agent files:")
	utils.PrintInfo("  • .cursorrules (Cursor auto-detects)")
	utils.PrintInfo("  • .github/copilot-instructions.md (GitHub Copilot auto-detects)")
	utils.PrintInfo("  • .opencode/agents/ (OpenCode may support in future)")

	return nil
}

// installCursorRules creates .cursorrules file for Cursor
func (a *ProjectRootAdapter) installCursorRules(pkg *packages.Package, agentContent string) error {
	cursorRulesPath := filepath.Join(a.projectRoot, ".cursorrules")

	// Convert AGENT.md to Cursor format
	cursorContent := convertAgentToCursorRules(pkg.Name, agentContent)

	// Check if .cursorrules already exists
	if _, err := os.Stat(cursorRulesPath); err == nil {
		// File exists - append or merge
		utils.PrintWarning(".cursorrules already exists, appending agent")
		existing, _ := os.ReadFile(cursorRulesPath)
		cursorContent = string(existing) + "\n\n" + cursorContent
	}

	if err := os.WriteFile(cursorRulesPath, []byte(cursorContent), 0644); err != nil {
		return fmt.Errorf("failed to write .cursorrules: %w", err)
	}

	return nil
}

// installGitHubCopilot creates .github/copilot-instructions.md for GitHub Copilot
func (a *ProjectRootAdapter) installGitHubCopilot(pkg *packages.Package, agentContent string) error {
	githubDir := filepath.Join(a.projectRoot, ".github")
	if err := utils.EnsureDir(githubDir); err != nil {
		return fmt.Errorf("failed to create .github directory: %w", err)
	}

	copilotPath := filepath.Join(githubDir, "copilot-instructions.md")

	// Convert AGENT.md to GitHub Copilot format
	copilotContent := convertToGitHubCopilot(pkg.Name, agentContent)

	// Check if file already exists
	if _, err := os.Stat(copilotPath); err == nil {
		utils.PrintWarning(".github/copilot-instructions.md already exists, appending agent")
		existing, _ := os.ReadFile(copilotPath)
		copilotContent = string(existing) + "\n\n---\n\n" + copilotContent
	}

	if err := os.WriteFile(copilotPath, []byte(copilotContent), 0644); err != nil {
		return fmt.Errorf("failed to write copilot-instructions.md: %w", err)
	}

	return nil
}

// installOpenCodeAgent creates .opencode/agents/ structure (proposed convention)
func (a *ProjectRootAdapter) installOpenCodeAgent(pkg *packages.Package, agentContent string) error {
	opencodeAgentsDir := filepath.Join(a.projectRoot, ".opencode", "agents", pkg.Name)
	if err := utils.EnsureDir(opencodeAgentsDir); err != nil {
		return fmt.Errorf("failed to create .opencode/agents directory: %w", err)
	}

	agentPath := filepath.Join(opencodeAgentsDir, "AGENT.md")
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		return fmt.Errorf("failed to write AGENT.md: %w", err)
	}

	return nil
}

// InstallSkill installs a skill (not all tools support skills)
func (a *ProjectRootAdapter) InstallSkill(pkg *packages.Package) error {
	// For now, only OpenCode supports skills
	opencodeSkillsDir := filepath.Join(a.projectRoot, ".opencode", "skills", pkg.Name)
	if err := utils.EnsureDir(opencodeSkillsDir); err != nil {
		return fmt.Errorf("failed to create .opencode/skills directory: %w", err)
	}

	skillFile := filepath.Join(pkg.Path, "SKILL.md")
	targetFile := filepath.Join(opencodeSkillsDir, "SKILL.md")

	if err := utils.CopyFile(skillFile, targetFile); err != nil {
		return fmt.Errorf("failed to copy SKILL.md: %w", err)
	}

	utils.PrintInfo("Created .opencode/skills/ (OpenCode may support in future)")
	return nil
}

// InstallPrompt installs a prompt
func (a *ProjectRootAdapter) InstallPrompt(pkg *packages.Package) error {
	// Prompts go to .opencode/prompts/ (proposed)
	opencodePromptsDir := filepath.Join(a.projectRoot, ".opencode", "prompts")
	if err := utils.EnsureDir(opencodePromptsDir); err != nil {
		return fmt.Errorf("failed to create .opencode/prompts directory: %w", err)
	}

	// Find prompt file
	files, err := os.ReadDir(pkg.Path)
	if err != nil {
		return fmt.Errorf("failed to read package directory: %w", err)
	}

	promptFile := ""
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if name == "prompt.txt" || name == "prompt.md" || name == pkg.Name+".txt" {
			promptFile = name
			break
		}
	}

	if promptFile == "" {
		return fmt.Errorf("no prompt file found in package")
	}

	srcFile := filepath.Join(pkg.Path, promptFile)
	targetFile := filepath.Join(opencodePromptsDir, pkg.Name+filepath.Ext(promptFile))

	if err := utils.CopyFile(srcFile, targetFile); err != nil {
		return fmt.Errorf("failed to copy prompt file: %w", err)
	}

	return nil
}

// Uninstall removes project-root files
func (a *ProjectRootAdapter) Uninstall(pkg *packages.Package) error {
	// This is complex as we may need to edit .cursorrules, copilot-instructions.md
	// For now, just remove .opencode entries
	var errors []string

	// Remove from .opencode/
	opencodeAgentPath := filepath.Join(a.projectRoot, ".opencode", "agents", pkg.Name)
	if err := os.RemoveAll(opencodeAgentPath); err != nil {
		errors = append(errors, fmt.Sprintf(".opencode/agents: %s", err.Error()))
	}

	opencodeSkillPath := filepath.Join(a.projectRoot, ".opencode", "skills", pkg.Name)
	if err := os.RemoveAll(opencodeSkillPath); err != nil {
		errors = append(errors, fmt.Sprintf(".opencode/skills: %s", err.Error()))
	}

	// TODO: Remove from .cursorrules and .github/copilot-instructions.md (requires parsing)
	utils.PrintWarning("Note: .cursorrules and .github/copilot-instructions.md may still contain agent - manual cleanup required")

	if len(errors) > 0 {
		return fmt.Errorf("uninstall errors: %s", strings.Join(errors, ", "))
	}

	return nil
}

// List returns all installed packages (from .opencode/ directory)
func (a *ProjectRootAdapter) List() ([]*packages.Package, error) {
	var pkgs []*packages.Package

	opencodeDir := filepath.Join(a.projectRoot, ".opencode")
	if _, err := os.Stat(opencodeDir); os.IsNotExist(err) {
		return pkgs, nil
	}

	// List agents
	agentsDir := filepath.Join(opencodeDir, "agents")
	if dirs, err := os.ReadDir(agentsDir); err == nil {
		for _, dir := range dirs {
			if dir.IsDir() {
				pkgs = append(pkgs, &packages.Package{
					Name: dir.Name(),
					Type: packages.TypeAgent,
					Path: filepath.Join(agentsDir, dir.Name()),
				})
			}
		}
	}

	// List skills
	skillsDir := filepath.Join(opencodeDir, "skills")
	if dirs, err := os.ReadDir(skillsDir); err == nil {
		for _, dir := range dirs {
			if dir.IsDir() {
				pkgs = append(pkgs, &packages.Package{
					Name: dir.Name(),
					Type: packages.TypeSkill,
					Path: filepath.Join(skillsDir, dir.Name()),
				})
			}
		}
	}

	// List prompts
	promptsDir := filepath.Join(opencodeDir, "prompts")
	if files, err := os.ReadDir(promptsDir); err == nil {
		for _, file := range files {
			if !file.IsDir() {
				name := file.Name()
				ext := filepath.Ext(name)
				pkgName := name[:len(name)-len(ext)]
				pkgs = append(pkgs, &packages.Package{
					Name: pkgName,
					Type: packages.TypePrompt,
					Path: filepath.Join(promptsDir, name),
				})
			}
		}
	}

	return pkgs, nil
}

// Validate checks if the project root is valid
func (a *ProjectRootAdapter) Validate() error {
	if _, err := os.Stat(a.projectRoot); os.IsNotExist(err) {
		return fmt.Errorf("project root does not exist: %s", a.projectRoot)
	}
	return nil
}

// convertAgentToCursorRules converts AGENT.md content to Cursor .cursorrules format
func convertAgentToCursorRules(name string, content string) string {
	// Strip YAML frontmatter if present
	content = stripYAMLFrontmatter(content)

	// Add Cursor-specific header
	header := fmt.Sprintf("# Cursor Rules: %s\n\n", name)

	footer := "\n\n---\nGenerated by AIT (AI Toolkit Package Manager)"

	return header + content + footer
}

// convertToGitHubCopilot converts AGENT.md content to GitHub Copilot format
func convertToGitHubCopilot(name string, content string) string {
	// Strip YAML frontmatter if present
	content = stripYAMLFrontmatter(content)

	// Add GitHub Copilot-specific header
	header := fmt.Sprintf("# GitHub Copilot Instructions: %s\n\n", name)

	footer := "\n\n---\n*Generated by AIT (AI Toolkit Package Manager)*"

	return header + content + footer
}

// stripYAMLFrontmatter removes YAML frontmatter from markdown content
func stripYAMLFrontmatter(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		// Find closing ---
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				// Return everything after the frontmatter
				return strings.Join(lines[i+1:], "\n")
			}
		}
	}
	return content
}
