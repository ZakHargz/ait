package cli

import (
	"fmt"

	"github.com/apex-ai/ait/internal/adapters"
	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/sources"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [package-names...]",
	Short: "Update installed packages to latest versions",
	Long: `Update packages to their latest compatible versions.

Examples:
  # Update all packages from ait.yml
  ait update

  # Update specific packages
  ait update code-reviewer python

  # Update to specific version
  ait install github:org/repo/agents/code-reviewer@2.0.0`,
	RunE: runUpdate,
}

var (
	updateTargets []string
)

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringSliceVarP(&updateTargets, "target", "t", []string{}, "target tools to update (opencode, cursor, claude)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	manifestPath := "ait.yml"
	if !config.ManifestExists(manifestPath) {
		return fmt.Errorf("no ait.yml found. Run 'ait init' first")
	}

	manifest, err := config.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load ait.yml: %w", err)
	}

	// Get all dependencies (now a simple flat list)
	allDeps := manifest.Dependencies

	if len(allDeps) == 0 {
		utils.PrintWarning("No dependencies found in ait.yml")
		return nil
	}

	// Filter by package names if specified
	var specsToUpdate []string
	if len(args) > 0 {
		// Find specs matching the package names
		for _, dep := range allDeps {
			spec, err := sources.ParsePackageSpec(dep)
			if err != nil {
				continue
			}
			for _, name := range args {
				if spec.GetPackageName() == name {
					specsToUpdate = append(specsToUpdate, dep)
					break
				}
			}
		}
		if len(specsToUpdate) == 0 {
			return fmt.Errorf("no matching packages found for: %v", args)
		}
	} else {
		specsToUpdate = allDeps
	}

	// Use targets from manifest if not specified via flag
	if len(updateTargets) == 0 && len(manifest.Targets) > 0 {
		updateTargets = manifest.Targets
	}

	// Detect available tools if no targets specified
	if len(updateTargets) == 0 {
		utils.PrintInfo("Detecting installed AI tools...")
		detectedTools := adapters.DetectInstalledTools()
		if len(detectedTools) == 0 {
			return fmt.Errorf("no AI tools detected")
		}
		updateTargets = detectedTools
	}

	// Create adapters for target tools
	targetAdapters := make(map[string]adapters.Adapter)
	for _, target := range updateTargets {
		adapter, err := adapters.GetAdapter(target)
		if err != nil {
			utils.PrintWarning(fmt.Sprintf("Skipping %s: %s", target, err.Error()))
			continue
		}
		targetAdapters[target] = adapter
	}

	if len(targetAdapters) == 0 {
		return fmt.Errorf("no valid target tools available")
	}

	utils.PrintInfo(fmt.Sprintf("Updating %d package(s) to %d tool(s)...", len(specsToUpdate), len(targetAdapters)))

	// Load lock file to check current versions
	lockPath := "ait.lock"
	var lockFile *config.LockFile
	if config.LockFileExists(lockPath) {
		lockFile, err = config.LoadLockFile(lockPath)
		if err != nil {
			utils.PrintWarning("Could not load lock file, will do fresh install")
			lockFile = config.NewLockFile()
		}
	} else {
		lockFile = config.NewLockFile()
	}

	updatedCount := 0

	for _, specStr := range specsToUpdate {
		// Parse spec
		spec, err := sources.ParsePackageSpec(specStr)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Invalid spec %s: %s", specStr, err.Error()))
			continue
		}

		pkgName := spec.GetPackageName()

		// Check current version
		currentPkg, exists := lockFile.GetPackage(pkgName)
		if exists {
			utils.PrintInfo(fmt.Sprintf("Updating %s from %s...", pkgName, currentPkg.Resolved))
		} else {
			utils.PrintInfo(fmt.Sprintf("Installing %s (not currently installed)...", pkgName))
		}

		// Get source
		source, err := sources.GetSource(*spec)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to get source for %s: %s", pkgName, err.Error()))
			continue
		}

		// Fetch latest version matching constraint
		pkg, err := source.Fetch(*spec)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to fetch %s: %s", pkgName, err.Error()))
			continue
		}

		// Check if version changed
		if exists && pkg.Version == currentPkg.Resolved {
			utils.PrintInfo(fmt.Sprintf("✓ %s is already up to date (%s)", pkgName, pkg.Version))
			continue
		}

		// Install to each target tool
		installSuccess := false
		for toolName, adapter := range targetAdapters {
			if err := installToAdapter(pkg, adapter, toolName); err != nil {
				utils.PrintWarning(fmt.Sprintf("Failed to install to %s: %s", toolName, err.Error()))
				continue
			}
			utils.PrintSuccess(fmt.Sprintf("✓ Updated %s to %s in %s", pkgName, pkg.Version, toolName))
			installSuccess = true
		}

		if installSuccess {
			// Update lock file
			installedToTools := []string{}
			for toolName := range targetAdapters {
				installedToTools = append(installedToTools, toolName)
			}

			lockFile.AddPackage(
				pkg.Name,
				spec.Version,
				string(pkg.Type),
				spec.Original,
				pkg.Version,
				installedToTools,
			)
			updatedCount++
		}
	}

	if updatedCount > 0 {
		utils.PrintSuccess(fmt.Sprintf("Successfully updated %d package(s)", updatedCount))

		// Write lock file
		if err := lockFile.Write(lockPath); err != nil {
			utils.PrintWarning(fmt.Sprintf("Failed to write lock file: %s", err.Error()))
		} else {
			utils.PrintInfo("Updated ait.lock")
		}
	} else {
		utils.PrintInfo("All packages are up to date")
	}

	return nil
}
