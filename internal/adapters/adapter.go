package adapters

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/utils"
)

// Adapter is the interface that all platform adapters must implement
type Adapter interface {
	// Name returns the adapter identifier
	Name() string

	// Detect checks if this tool is installed on the system
	Detect() bool

	// GetConfigDir returns the tool's configuration directory
	GetConfigDir() (string, error)

	// InstallAgent installs an agent package
	InstallAgent(pkg *packages.Package) error

	// InstallSkill installs a skill package
	InstallSkill(pkg *packages.Package) error

	// InstallPrompt installs a prompt package
	InstallPrompt(pkg *packages.Package) error

	// Uninstall removes a package
	Uninstall(pkg *packages.Package) error

	// List returns all installed packages for this adapter
	List() ([]*packages.Package, error)

	// Validate checks if the adapter installation is healthy
	Validate() error
}

// BaseAdapter provides common functionality for adapters
type BaseAdapter struct {
	name      string
	configDir string
}

// NewBaseAdapter creates a new base adapter
func NewBaseAdapter(name, configDir string) BaseAdapter {
	return BaseAdapter{
		name:      name,
		configDir: configDir,
	}
}

// Name returns the adapter name
func (b *BaseAdapter) Name() string {
	return b.name
}

// GetConfigDir returns the configuration directory
func (b *BaseAdapter) GetConfigDir() (string, error) {
	return b.configDir, nil
}

// PackageInstallConfig holds configuration for package installation
type PackageInstallConfig struct {
	// TargetSubdir is the subdirectory within configDir (e.g., "agents", "skills", "prompts")
	TargetSubdir string
	// SourceFileName is the default source file name if not specified in package
	SourceFileName string
	// DestFileName is the destination file name (may include pkg.Name)
	DestFileName string
	// UsePackageSubdir indicates whether to create a subdirectory per package
	UsePackageSubdir bool
}

// InstallPackageFile is a common helper for installing packages
// It handles directory creation, file copying, and source file resolution.
// When pkg.ApmAgentFile is set (for agent installs), it is used as the source directly.
func InstallPackageFile(pkg *packages.Package, configDir, adapterName string, cfg PackageInstallConfig) error {
	var targetDir string

	if cfg.UsePackageSubdir {
		// Create per-package subdirectory (e.g., agents/package-name/)
		targetDir = filepath.Join(configDir, cfg.TargetSubdir, pkg.Name)
	} else {
		// Use shared directory (e.g., prompts/)
		targetDir = filepath.Join(configDir, cfg.TargetSubdir)
	}

	// Create directory
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", cfg.TargetSubdir, err)
	}

	// Determine source path.
	// Priority: ApmAgentFile (for agent installs from .apm/ layout) > GetFile() > SourceFileName fallback.
	var source string
	if pkg.ApmAgentFile != "" && cfg.SourceFileName == "AGENT.md" {
		// APM .apm/agents/ layout — use the discovered .agent.md file directly.
		source = pkg.ApmAgentFile
	} else {
		sourceFile := pkg.GetFile(adapterName)
		if sourceFile == "" {
			sourceFile = cfg.SourceFileName
		}
		source = filepath.Join(pkg.Path, sourceFile)
	}

	dest := filepath.Join(targetDir, cfg.DestFileName)

	// Copy file
	if err := utils.CopyFile(source, dest); err != nil {
		return fmt.Errorf("failed to install %s: %w", cfg.TargetSubdir, err)
	}

	return nil
}

// InstallSkillDir installs a skill from the APM .apm/skills/<name>/ directory layout.
// The entire source directory (SKILL.md + bundled resources) is copied to
// targetDir/<pkg.Name>/, mirroring what APM does on install.
func InstallSkillDir(pkg *packages.Package, configDir, skillSubdir string) error {
	targetDir := filepath.Join(configDir, skillSubdir, pkg.Name)

	// Remove any pre-existing install so stale files are not left behind.
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to remove existing skill directory: %w", err)
	}

	if err := utils.CopyDir(pkg.ApmSkillDir, targetDir); err != nil {
		return fmt.Errorf("failed to install skill directory: %w", err)
	}

	return nil
}

// UninstallPackage is a common helper for uninstalling packages
func UninstallPackage(pkg *packages.Package, configDir string, agentSubdir, skillSubdir, promptSubdir string) error {
	switch pkg.Type {
	case packages.TypeAgent:
		return os.RemoveAll(filepath.Join(configDir, agentSubdir, pkg.Name))
	case packages.TypeSkill:
		return os.RemoveAll(filepath.Join(configDir, skillSubdir, pkg.Name))
	case packages.TypeHybrid:
		// Remove both the agent and skill directories installed by a hybrid package.
		agentErr := os.RemoveAll(filepath.Join(configDir, agentSubdir, pkg.Name))
		skillErr := os.RemoveAll(filepath.Join(configDir, skillSubdir, pkg.Name))
		if agentErr != nil {
			return agentErr
		}
		return skillErr
	case packages.TypePrompt:
		return os.RemoveAll(filepath.Join(configDir, promptSubdir, pkg.Name+".txt"))
	default:
		return fmt.Errorf("unsupported package type: %s", pkg.Type)
	}
}

// ListPackages is a common helper for listing all installed packages
func ListPackages(configDir string, agentSubdir, skillSubdir, promptSubdir string) ([]*packages.Package, error) {
	installed := []*packages.Package{}

	// List agents
	agentsDir := filepath.Join(configDir, agentSubdir)
	if agents, err := ListPackagesInDir(agentsDir, packages.TypeAgent); err == nil {
		installed = append(installed, agents...)
	}

	// List skills
	skillsDir := filepath.Join(configDir, skillSubdir)
	if skills, err := ListPackagesInDir(skillsDir, packages.TypeSkill); err == nil {
		installed = append(installed, skills...)
	}

	// List prompts
	promptsDir := filepath.Join(configDir, promptSubdir)
	if prompts, err := ListPromptsInDir(promptsDir); err == nil {
		installed = append(installed, prompts...)
	}

	return installed, nil
}

// ListPackagesInDir lists packages in a directory (for agents and skills)
func ListPackagesInDir(dir string, pkgType packages.PackageType) ([]*packages.Package, error) {
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

// ListPromptsInDir lists prompt packages in a directory
func ListPromptsInDir(dir string) ([]*packages.Package, error) {
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

// ValidateConfigDir is a common helper for validating adapter configuration
func ValidateConfigDir(configDir string, createIfMissing bool) error {
	// Check if config directory exists
	if !utils.DirExists(configDir) {
		if createIfMissing {
			// Try to create it
			if err := utils.EnsureDir(configDir); err != nil {
				return fmt.Errorf("config directory does not exist and cannot be created: %s", configDir)
			}
		} else {
			return fmt.Errorf("config directory does not exist: %s", configDir)
		}
	}

	// Check if writable
	if err := utils.CheckDirWritable(configDir); err != nil {
		return fmt.Errorf("config directory not writable: %w", err)
	}

	return nil
}
