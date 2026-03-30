package cli

import (
	"fmt"
	"os"

	"github.com/apex-ai/ait/internal/adapters"
	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/sources"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [package-specs...]",
	Short: "Install packages from ait.yml or specific package specs",
	Long: `Install packages defined in ait.yml or install specific packages.

Examples:
  # Install all dependencies from ait.yml
  ait install

  # Install specific packages (GitHub shorthand - recommended)
  ait install org/repo/agents/code-reviewer@1.0.0
  ait install org/repo/skills/python@^2.0.0
  ait install gitlab.com/org/repo/agents/helper

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
	installCmd.Flags().BoolVarP(&installSave, "save", "s", false, "save installed packages to ait.yml")
	installCmd.Flags().BoolVarP(&installGlobal, "global", "g", false, "install globally to AI tools instead of project-local .ait/")
}

func runInstall(cmd *cobra.Command, args []string) error {
	var specsToInstall []string

	if len(args) > 0 {
		// Install specific packages from command line
		specsToInstall = args
	} else {
		// Install from ait.yml
		manifestPath := "ait.yml"
		if !config.ManifestExists(manifestPath) {
			return fmt.Errorf("no ait.yml found. Run 'ait init' first or provide package specs")
		}

		manifest, err := config.LoadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("failed to load ait.yml: %w", err)
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

	// Parse and install each package
	utils.PrintInfo(fmt.Sprintf("Installing %d package(s) to %d location(s)...", len(specsToInstall), len(targetAdapters)))

	// Load or create lock file
	lockPath := "ait.lock"
	var lockFile *config.LockFile
	if config.LockFileExists(lockPath) {
		var err error
		lockFile, err = config.LoadLockFile(lockPath)
		if err != nil {
			utils.PrintWarning(fmt.Sprintf("Failed to load existing lock file: %s", err.Error()))
			lockFile = config.NewLockFile()
		}
	} else {
		lockFile = config.NewLockFile()
	}

	installedPackages := []installResult{}

	for _, specStr := range specsToInstall {
		result, err := installPackage(specStr, targetAdapters)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to install %s: %s", specStr, err.Error()))
			continue
		}
		installedPackages = append(installedPackages, result)

		// Add to lock file
		installedToTools := []string{}
		for toolName := range targetAdapters {
			installedToTools = append(installedToTools, toolName)
		}

		lockFile.AddPackage(
			result.pkg.Name,
			result.spec.Version,
			string(result.pkg.Type),
			result.spec.Original,
			result.pkg.Version,
			installedToTools,
		)
	}

	if len(installedPackages) == 0 {
		return fmt.Errorf("no packages were installed successfully")
	}

	utils.PrintSuccess(fmt.Sprintf("Successfully installed %d package(s)", len(installedPackages)))

	// Write lock file
	if err := lockFile.Write(lockPath); err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to write lock file: %s", err.Error()))
	} else {
		utils.PrintInfo("Updated ait.lock")
	}

	return nil
}

// installPackage fetches and installs a single package to all target adapters
func installPackage(specStr string, targetAdapters map[string]adapters.Adapter) (installResult, error) {
	utils.PrintInfo(fmt.Sprintf("Installing %s...", specStr))

	// Parse package spec
	spec, err := sources.ParsePackageSpec(specStr)
	if err != nil {
		return installResult{}, fmt.Errorf("invalid package spec: %w", err)
	}

	// Get appropriate source
	source, err := sources.GetSource(*spec)
	if err != nil {
		return installResult{}, fmt.Errorf("failed to get source: %w", err)
	}

	// Fetch package
	pkg, err := source.Fetch(*spec)
	if err != nil {
		return installResult{}, fmt.Errorf("failed to fetch package: %w", err)
	}

	// Install to each target tool
	for toolName, adapter := range targetAdapters {
		if err := installToAdapter(pkg, adapter, toolName); err != nil {
			utils.PrintWarning(fmt.Sprintf("Failed to install to %s: %s", toolName, err.Error()))
			continue
		}
		utils.PrintSuccess(fmt.Sprintf("Installed %s to %s", pkg.Name, toolName))
	}

	return installResult{pkg: pkg, spec: spec}, nil
}

// installToAdapter installs a package using the appropriate adapter method based on package type
func installToAdapter(pkg *packages.Package, adapter adapters.Adapter, toolName string) error {
	switch pkg.Type {
	case packages.TypeAgent:
		return adapter.InstallAgent(pkg)

	case packages.TypeSkill:
		return adapter.InstallSkill(pkg)

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
		utils.PrintInfo(fmt.Sprintf("Found tools: %v", targets))
	}

	// Create adapters for target tools
	targetAdapters := make(map[string]adapters.Adapter)
	for _, target := range targets {
		adapter, err := adapters.GetAdapter(target)
		if err != nil {
			utils.PrintWarning(fmt.Sprintf("Skipping %s: %s", target, err.Error()))
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
	utils.PrintInfo("  • .github/copilot-instructions.md (GitHub Copilot auto-detects)")
	utils.PrintInfo("  • .opencode/ (proposed for OpenCode)")
	utils.PrintInfo("💡 Tip: Commit these files to git for team sharing!")
	utils.PrintInfo("💡 Tip: Use --global flag to install to AI tools globally")

	return map[string]adapters.Adapter{
		"project-root": projectRootAdapter,
	}, nil
}
