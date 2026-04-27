package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex-ai/ait/internal/adapters"
	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/resolver"
	"github.com/apex-ai/ait/internal/sources"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [package-specs...]",
	Short: "Install packages from ait.yml or specific package specs",
	Long: `Install packages defined in ait.yml or install specific packages.

When installing specific packages from the command line, they are automatically
saved to ait.yml (creating it if it doesn't exist). This matches APM behavior.

Examples:
  # Install all dependencies from ait.yml
  ait install

  # Install specific packages (creates/updates ait.yml automatically)
  ait install org/repo/agents/code-reviewer@1.0.0
  ait install org/repo/skills/python@^2.0.0
  ait install gitlab.com/org/repo/agents/helper

  # Install without saving to ait.yml
  ait install org/repo/agents/code-reviewer --save=false

  # Install with explicit prefixes (legacy format)
  ait install github:org/repo/agents/code-reviewer@1.0.0
  ait install local:./path/to/package

Package spec formats:
  Shorthand (recommended):
    org/repo/path@version                    # Defaults to GitHub
    gitlab.com/org/repo/path@version         # Other hosts (use FQDN)
    ./path/to/package                        # Local path

  Legacy (with prefix):
    github:org/repo/path@version
    gitlab:org/repo/path@version
    git:https://git.example.com/repo@version
    local:./path/to/package@version

Version formats:
  - Exact: 1.0.0
  - Caret: ^1.0.0 (allows minor and patch updates)
  - Tilde: ~1.0.0 (allows patch updates only)
  - Range: >=1.0.0 <2.0.0
  - Branch: main, develop
  - Commit: abc123...`,
	RunE: runInstall,
}

var (
	installTargets []string
	installSave    bool
	installGlobal  bool
)

// installResult holds the result of installing a single package
type installResult struct {
	pkg  *packages.Package
	spec *sources.PackageSpec
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().StringSliceVarP(&installTargets, "target", "t", []string{}, "target tools to install to (opencode, cursor, claude, project)")
	installCmd.Flags().BoolVarP(&installSave, "save", "s", true, "save installed packages to ait.yml (default: true)")
	installCmd.Flags().BoolVarP(&installGlobal, "global", "g", false, "install globally to AI tools instead of project-local .ait/")
}

func runInstall(cmd *cobra.Command, args []string) error {
	var specsToInstall []string
	var installingFromCommandLine bool

	// manifestPath tracks which manifest file is active so later steps
	// (auto-save) write back to the same file.
	manifestPath := "ait.yml"

	if len(args) > 0 {
		// Install specific packages from command line
		specsToInstall = args
		installingFromCommandLine = true
	} else {
		// Install from manifest — support both ait.yml and apm.yml
		var err error
		manifestPath, err = config.FindManifest()
		if err != nil {
			return fmt.Errorf("%w\nRun 'ait init' to create one or provide package specs as arguments", err)
		}

		manifest, err := config.LoadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", manifestPath, err)
		}

		// Collect all dependencies (now a simple flat list)
		specsToInstall = append(specsToInstall, manifest.Dependencies...)

		if len(specsToInstall) == 0 {
			utils.PrintWarning("No dependencies found in ait.yml")
			utils.PrintInfo("Add dependencies to ait.yml or provide package specs as arguments")
			return nil
		}

		// Use targets from manifest if not specified via flag
		if len(installTargets) == 0 && len(manifest.Targets) > 0 {
			installTargets = manifest.Targets
		}
		installingFromCommandLine = false
	}

	// Determine installation mode: project-local or global
	var targetAdapters map[string]adapters.Adapter
	var err error

	if installGlobal {
		// Global installation to AI tools
		targetAdapters, err = getGlobalAdapters(installTargets)
		if err != nil {
			return err
		}
	} else {
		// Project-local installation to .ait/
		targetAdapters, err = getProjectLocalAdapters()
		if err != nil {
			return err
		}
	}

	if len(targetAdapters) == 0 {
		return fmt.Errorf("no valid target tools available")
	}

	// Resolve dependencies (including transitive dependencies)
	utils.PrintInfo("Resolving dependencies...")
	depResolver := resolver.NewResolver()
	resolvedPackages, err := depResolver.Resolve(specsToInstall)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	utils.PrintInfo("Installing %d package(s) (including dependencies) to %d location(s)...", len(resolvedPackages), len(targetAdapters))

	// Load or create lock file
	lockPath := "ait.lock"
	var lockFile *config.LockFile
	if config.LockFileExists(lockPath) {
		var err error
		lockFile, err = config.LoadLockFile(lockPath)
		if err != nil {
			utils.PrintWarning("Failed to load existing lock file: %v", err)
			lockFile = config.NewLockFile()
		}
	} else {
		lockFile = config.NewLockFile()
	}

	// Install all resolved packages in topological order (dependencies first)
	installedPackages := []installResult{}
	installedCount := 0

	for _, pkg := range resolvedPackages {
		utils.PrintInfo("Installing %s@%s...", pkg.Name, pkg.Version)

		// Install to each target tool
		for toolName, adapter := range targetAdapters {
			if err := installToAdapter(pkg, adapter, toolName); err != nil {
				utils.PrintWarning("Failed to install %s to %s: %v", pkg.Name, toolName, err)
				continue
			}
			utils.PrintSuccess("Installed %s to %s", pkg.Name, toolName)
		}

		installedCount++

		// Create result for tracking
		// Note: We don't have the original spec for transitive deps, so use constructed one
		spec := &sources.PackageSpec{
			Original: fmt.Sprintf("%s@%s", pkg.Name, pkg.Version),
			Version:  pkg.Version,
		}

		installedPackages = append(installedPackages, installResult{
			pkg:  pkg,
			spec: spec,
		})

		// Add to lock file
		installedToTools := []string{}
		for toolName := range targetAdapters {
			installedToTools = append(installedToTools, toolName)
		}

		lockFile.AddPackage(
			pkg.Name,
			pkg.Version,
			string(pkg.Type),
			spec.Original,
			pkg.Version,
			installedToTools,
		)
	}

	if installedCount == 0 {
		return fmt.Errorf("no packages were installed successfully")
	}

	utils.PrintSuccess("Successfully installed %d package(s)", installedCount)

	// Write lock file
	if err := lockFile.Write(lockPath); err != nil {
		utils.PrintWarning("Failed to write lock file: %v", err)
	} else {
		utils.PrintInfo("Updated ait.lock")
	}

	// Auto-save to manifest if installing from command line and save flag is true
	if installingFromCommandLine && installSave {
		if err := saveToManifest(installedPackages, manifestPath); err != nil {
			utils.PrintWarning("Failed to save to %s: %v", manifestPath, err)
		}
	}

	return nil
}

// installToAdapter installs a package using the appropriate adapter method based on package type.
// Hybrid packages install as both an agent AND a skill so that every file
// present in the package directory (AGENT.md and SKILL.md) lands in its
// correct destination (e.g. ~/.config/opencode/agents/ and
// ~/.config/opencode/skills/ for a global OpenCode install).
func installToAdapter(pkg *packages.Package, adapter adapters.Adapter, toolName string) error {
	switch pkg.Type {
	case packages.TypeAgent:
		return adapter.InstallAgent(pkg)

	case packages.TypeSkill:
		return adapter.InstallSkill(pkg)

	case packages.TypeHybrid:
		// Install the agent half first.
		if err := adapter.InstallAgent(pkg); err != nil {
			return fmt.Errorf("hybrid agent install failed: %w", err)
		}
		// Install the skill half — a missing SKILL.md is a soft warning, not a
		// hard error, so that packages which only ship AGENT.md still succeed.
		if err := adapter.InstallSkill(pkg); err != nil {
			utils.PrintWarning("[%s] skill install skipped for hybrid package %q: %v", toolName, pkg.Name, err)
		}
		return nil

	case packages.TypePrompt:
		return adapter.InstallPrompt(pkg)

	case packages.TypeMCP:
		return fmt.Errorf("MCP server installation not yet implemented")

	default:
		return fmt.Errorf("unknown package type: %s", pkg.Type)
	}
}

// getGlobalAdapters returns adapters for global AI tool installations
func getGlobalAdapters(targets []string) (map[string]adapters.Adapter, error) {
	// Detect available tools if no targets specified
	if len(targets) == 0 {
		utils.PrintInfo("Detecting installed AI tools...")
		detectedTools := adapters.DetectInstalledTools()
		if len(detectedTools) == 0 {
			return nil, fmt.Errorf("no AI tools detected. Install OpenCode, Cursor, or Claude Desktop first")
		}
		targets = detectedTools
		utils.PrintInfo("Found tools: %v", targets)
	}

	// Create adapters for target tools
	targetAdapters := make(map[string]adapters.Adapter)
	for _, target := range targets {
		adapter, err := adapters.GetAdapter(target)
		if err != nil {
			utils.PrintWarning("Skipping %s: %v", target, err)
			continue
		}
		targetAdapters[target] = adapter
	}

	return targetAdapters, nil
}

// getProjectLocalAdapters returns the project-root adapter for native tool detection
func getProjectLocalAdapters() (map[string]adapters.Adapter, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Use ProjectRootAdapter which creates files that tools auto-detect
	projectRootAdapter := adapters.NewProjectRootAdapter(cwd)

	utils.PrintInfo("Installing to project root using tool-native conventions:")
	utils.PrintInfo("  • .cursorrules (Cursor auto-detects)")
	utils.PrintInfo("  • .github/agents/*.agent.md (GitHub Copilot, VS Code, IntelliJ)")
	utils.PrintInfo("  • .opencode/ (proposed for OpenCode)")
	utils.PrintInfo("💡 Tip: Commit these files to git for team sharing!")
	utils.PrintInfo("💡 Tip: Use --global flag to install to AI tools globally")

	return map[string]adapters.Adapter{
		"project-root": projectRootAdapter,
	}, nil
}

// saveToManifest saves installed packages to the given manifest file (ait.yml or apm.yml),
// creating ait.yml if no manifest exists yet.
func saveToManifest(installedPackages []installResult, manifestPath string) error {
	// If no manifest path was determined (command-line install in a fresh dir),
	// default to ait.yml.
	if manifestPath == "" {
		manifestPath = "ait.yml"
	}

	var manifest *config.Manifest

	// Load existing manifest or create new one
	if config.ManifestExists(manifestPath) {
		var err error
		manifest, err = config.LoadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("failed to load existing %s: %w", manifestPath, err)
		}
	} else {
		// Creating a new manifest — always use ait.yml regardless of manifestPath
		manifestPath = "ait.yml"

		// Create new manifest with defaults
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Get directory name for project name
		projectName := filepath.Base(cwd)

		manifest = &config.Manifest{
			Name:         projectName,
			Version:      "1.0.0",
			Dependencies: []string{},
			Targets:      []string{},
		}

		utils.PrintInfo("Creating ait.yml...")
	}

	// Add new dependencies (avoid duplicates)
	existingDeps := make(map[string]bool)
	for _, dep := range manifest.Dependencies {
		existingDeps[dep] = true
	}

	newDepsAdded := 0
	for _, result := range installedPackages {
		// Use the original package spec string
		depSpec := result.spec.Original

		if !existingDeps[depSpec] {
			manifest.Dependencies = append(manifest.Dependencies, depSpec)
			newDepsAdded++
		}
	}

	// Write manifest
	if err := manifest.Write(manifestPath); err != nil {
		return fmt.Errorf("failed to write %s: %w", manifestPath, err)
	}

	if newDepsAdded > 0 {
		utils.PrintSuccess("Added %d package(s) to %s", newDepsAdded, manifestPath)
	} else {
		utils.PrintInfo("All packages already in %s", manifestPath)
	}

	return nil
}
