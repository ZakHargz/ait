package adapters

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/utils"
)

// ProjectAdapter installs packages to the project's .ait/ directory
// This allows project-specific AI tools that are committed to version control
type ProjectAdapter struct {
	BaseAdapter
	projectRoot string
}

// NewProjectAdapter creates a new project-local adapter
func NewProjectAdapter(projectRoot string) *ProjectAdapter {
	return &ProjectAdapter{
		BaseAdapter: NewBaseAdapter("project", filepath.Join(projectRoot, ".ait")),
		projectRoot: projectRoot,
	}
}

// Detect checks if we're in a project with ait.yml
func (p *ProjectAdapter) Detect() bool {
	aitYmlPath := filepath.Join(p.projectRoot, "ait.yml")
	_, err := os.Stat(aitYmlPath)
	return err == nil
}

// GetConfigDir returns the .ait directory path
func (p *ProjectAdapter) GetConfigDir() (string, error) {
	return filepath.Join(p.projectRoot, ".ait"), nil
}

// InstallAgent installs an agent to .ait/agents/
func (p *ProjectAdapter) InstallAgent(pkg *packages.Package) error {
	agentsDir := filepath.Join(p.projectRoot, ".ait", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	targetDir := filepath.Join(agentsDir, pkg.Name)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	// Copy AGENT.md
	agentFile := filepath.Join(pkg.Path, "AGENT.md")
	if _, err := os.Stat(agentFile); err != nil {
		return fmt.Errorf("AGENT.md not found: %w", err)
	}

	targetFile := filepath.Join(targetDir, "AGENT.md")
	if err := utils.CopyFile(agentFile, targetFile); err != nil {
		return fmt.Errorf("failed to copy AGENT.md: %w", err)
	}

	// Copy README if exists
	readmeFile := filepath.Join(pkg.Path, "README.md")
	if _, err := os.Stat(readmeFile); err == nil {
		targetReadme := filepath.Join(targetDir, "README.md")
		utils.CopyFile(readmeFile, targetReadme) // Ignore error, README is optional
	}

	return nil
}

// InstallSkill installs a skill to .ait/skills/
func (p *ProjectAdapter) InstallSkill(pkg *packages.Package) error {
	skillsDir := filepath.Join(p.projectRoot, ".ait", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	targetDir := filepath.Join(skillsDir, pkg.Name)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	// Copy SKILL.md
	skillFile := filepath.Join(pkg.Path, "SKILL.md")
	if _, err := os.Stat(skillFile); err != nil {
		return fmt.Errorf("SKILL.md not found: %w", err)
	}

	targetFile := filepath.Join(targetDir, "SKILL.md")
	if err := utils.CopyFile(skillFile, targetFile); err != nil {
		return fmt.Errorf("failed to copy SKILL.md: %w", err)
	}

	// Copy README if exists
	readmeFile := filepath.Join(pkg.Path, "README.md")
	if _, err := os.Stat(readmeFile); err == nil {
		targetReadme := filepath.Join(targetDir, "README.md")
		utils.CopyFile(readmeFile, targetReadme)
	}

	return nil
}

// InstallPrompt installs a prompt to .ait/prompts/
func (p *ProjectAdapter) InstallPrompt(pkg *packages.Package) error {
	promptsDir := filepath.Join(p.projectRoot, ".ait", "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		return fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Find prompt file (could be .txt, .md, or other)
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
	targetFile := filepath.Join(promptsDir, pkg.Name+filepath.Ext(promptFile))

	if err := utils.CopyFile(srcFile, targetFile); err != nil {
		return fmt.Errorf("failed to copy prompt file: %w", err)
	}

	return nil
}

// Uninstall removes a package from .ait/
func (p *ProjectAdapter) Uninstall(pkg *packages.Package) error {
	var targetPath string

	switch pkg.Type {
	case packages.TypeAgent:
		targetPath = filepath.Join(p.projectRoot, ".ait", "agents", pkg.Name)
	case packages.TypeSkill:
		targetPath = filepath.Join(p.projectRoot, ".ait", "skills", pkg.Name)
	case packages.TypePrompt:
		// Find the prompt file
		promptsDir := filepath.Join(p.projectRoot, ".ait", "prompts")
		files, _ := os.ReadDir(promptsDir)
		for _, file := range files {
			if !file.IsDir() && filepath.Base(file.Name()) == pkg.Name+filepath.Ext(file.Name()) {
				targetPath = filepath.Join(promptsDir, file.Name())
				break
			}
		}
		if targetPath == "" {
			return fmt.Errorf("prompt file not found for %s", pkg.Name)
		}
	default:
		return fmt.Errorf("unknown package type: %s", pkg.Type)
	}

	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("failed to remove package: %w", err)
	}

	return nil
}

// List returns all installed packages in .ait/
func (p *ProjectAdapter) List() ([]*packages.Package, error) {
	var pkgs []*packages.Package

	aitDir := filepath.Join(p.projectRoot, ".ait")
	if _, err := os.Stat(aitDir); os.IsNotExist(err) {
		return pkgs, nil
	}

	// List agents
	agentsDir := filepath.Join(aitDir, "agents")
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
	skillsDir := filepath.Join(aitDir, "skills")
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
	promptsDir := filepath.Join(aitDir, "prompts")
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

// Validate checks if the .ait directory structure is valid
func (p *ProjectAdapter) Validate() error {
	aitDir := filepath.Join(p.projectRoot, ".ait")
	if _, err := os.Stat(aitDir); os.IsNotExist(err) {
		return fmt.Errorf(".ait directory does not exist")
	}
	return nil
}
