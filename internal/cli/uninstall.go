package cli

import (
	"fmt"

	"github.com/apex-ai/ait/internal/adapters"
	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <package-name> [package-names...]",
	Short: "Uninstall packages",
	Long: `Uninstall one or more packages from AI tools.

Examples:
  # Uninstall a package from all tools where it's installed
  ait uninstall code-reviewer

  # Uninstall multiple packages
  ait uninstall code-reviewer python-linter

  # Uninstall from specific tools only
  ait uninstall code-reviewer --target opencode --target cursor

The command will:
1. Remove packages from the specified tools (or all tools if no target specified)
2. Update the lock file to reflect the changes`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUninstall,
}

var (
	uninstallTargets []string
	uninstallGlobal  bool
)

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().StringSliceVarP(&uninstallTargets, "target", "t", []string{}, "target tools to uninstall from (opencode, cursor, claude, project)")
	uninstallCmd.Flags().BoolVarP(&uninstallGlobal, "global", "g", false, "uninstall globally from AI tools instead of project-local .ait/")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	packageNames := args

	// Load lock file to find installed packages
	lockPath := "ait.lock"
	if !config.LockFileExists(lockPath) {
		return fmt.Errorf("no ait.lock found. No packages installed")
	}

	lockFile, err := config.LoadLockFile(lockPath)
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	// Determine which tools to uninstall from
	var targetAdapters map[string]adapters.Adapter

	if uninstallGlobal {
		// Global uninstall from AI tools
		targetAdapters, err = getGlobalAdaptersForUninstall(uninstallTargets)
		if err != nil {
			return err
		}
	} else {
		// Project-local uninstall from .ait/
		targetAdapters, err = getProjectLocalAdapters()
		if err != nil {
			return err
		}
	}

	if len(targetAdapters) == 0 {
		return fmt.Errorf("no valid target tools available")
	}

	// Track uninstall results
	uninstalledCount := 0
	packagesToRemoveFromLock := []string{}

	// Uninstall each package
	for _, pkgName := range packageNames {
		utils.PrintInfo("Uninstalling %s...", pkgName)

		// Check if package exists in lock file
		lockedPkg, ok := lockFile.GetPackage(pkgName)
		if !ok {
			utils.PrintWarning("Package %s not found in lock file, attempting to uninstall anyway", pkgName)
		}

		// Create a minimal package struct for uninstall
		pkg := &packages.Package{
			Name: pkgName,
		}

		// If we found it in lock file, add more details
		if ok {
			pkg.Type = packages.PackageType(lockedPkg.Type)
			pkg.Version = lockedPkg.Resolved
		}

		// Attempt to uninstall from each target tool
		uninstalledFromAny := false
		for toolName, adapter := range targetAdapters {
			// Skip if not installed to this tool (if we have lock file info)
			if ok && !isInstalledToTool(lockedPkg.Installed, toolName) {
				utils.PrintInfo("Package %s not installed to %s, skipping", pkgName, toolName)
				continue
			}

			// Try to uninstall
			if err := adapter.Uninstall(pkg); err != nil {
				utils.PrintWarning("Failed to uninstall from %s: %s", toolName, err.Error())
				continue
			}

			utils.PrintSuccess("Uninstalled %s from %s", pkgName, toolName)
			uninstalledFromAny = true
		}

		if uninstalledFromAny {
			uninstalledCount++
			// Get list of tool names from adapters
			toolNames := make([]string, 0, len(targetAdapters))
			for toolName := range targetAdapters {
				toolNames = append(toolNames, toolName)
			}
			// If uninstalled from all tools (or specific targets), remove from lock
			if shouldRemoveFromLock(lockedPkg.Installed, toolNames, len(uninstallTargets) == 0) {
				packagesToRemoveFromLock = append(packagesToRemoveFromLock, pkgName)
			} else {
				// Update lock file to remove the tools we uninstalled from
				updateLockedPackageTools(&lockedPkg, toolNames)
				lockFile.Packages[pkgName] = lockedPkg
			}
		} else {
			utils.PrintError("Failed to uninstall %s from any tool", pkgName)
		}
	}

	// Remove packages from lock file
	for _, pkgName := range packagesToRemoveFromLock {
		delete(lockFile.Packages, pkgName)
	}

	// Write updated lock file
	if uninstalledCount > 0 {
		if err := lockFile.Write(lockPath); err != nil {
			utils.PrintWarning("Failed to update lock file: %s", err.Error())
		} else {
			utils.PrintInfo("Updated ait.lock")
		}

		utils.PrintSuccess("Successfully uninstalled %d package(s)", uninstalledCount)
	} else {
		return fmt.Errorf("no packages were uninstalled")
	}

	return nil
}

// isInstalledToTool checks if a package is installed to the given tool
func isInstalledToTool(installedTools []string, tool string) bool {
	for _, t := range installedTools {
		if t == tool {
			return true
		}
	}
	return false
}

// shouldRemoveFromLock determines if a package should be completely removed from lock file
func shouldRemoveFromLock(installedTools []string, uninstallTargets []string, uninstallFromAll bool) bool {
	if uninstallFromAll {
		// If uninstalling from all detected tools, remove from lock
		return true
	}

	// Check if we're removing from all tools where it's installed
	remainingTools := []string{}
	for _, tool := range installedTools {
		isTarget := false
		for _, target := range uninstallTargets {
			if tool == target {
				isTarget = true
				break
			}
		}
		if !isTarget {
			remainingTools = append(remainingTools, tool)
		}
	}

	return len(remainingTools) == 0
}

// updateLockedPackageTools updates the installed tools list for a locked package
func updateLockedPackageTools(pkg *config.LockedPkg, toolsToRemove []string) {
	newInstalled := []string{}
	for _, tool := range pkg.Installed {
		shouldKeep := true
		for _, remove := range toolsToRemove {
			if tool == remove {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			newInstalled = append(newInstalled, tool)
		}
	}
	pkg.Installed = newInstalled
}

// getGlobalAdaptersForUninstall returns adapters for global AI tool uninstalls
func getGlobalAdaptersForUninstall(targets []string) (map[string]adapters.Adapter, error) {
	var targetTools []string
	if len(targets) > 0 {
		targetTools = targets
	} else {
		// If no targets specified, detect installed tools
		utils.PrintInfo("Detecting installed AI tools...")
		detectedTools := adapters.DetectInstalledTools()
		if len(detectedTools) == 0 {
			return nil, fmt.Errorf("no AI tools detected")
		}
		targetTools = detectedTools
		utils.PrintInfo("Found tools: %v", targetTools)
	}

	// Create adapters for target tools
	targetAdapters := make(map[string]adapters.Adapter)
	for _, target := range targetTools {
		adapter, err := adapters.GetAdapter(target)
		if err != nil {
			utils.PrintWarning("Skipping %s: %s", target, err.Error())
			continue
		}
		targetAdapters[target] = adapter
	}

	return targetAdapters, nil
}
