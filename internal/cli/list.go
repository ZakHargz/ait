package cli

import (
	"fmt"
	"os"

	"github.com/apex-ai/ait/internal/adapters"
	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages",
	Long: `List installed packages.

When run in a project with ait.yml, lists packages from project-local installations
(.cursorrules, .github/agents/, .opencode/). Otherwise, lists globally installed packages.

Examples:
  # List packages in current project (if ait.yml exists)
  ait list

  # List globally installed packages
  ait list --global

  # List packages for specific tool
  ait list --target opencode`,
	RunE: runList,
}

var (
	listTargets []string
	listGlobal  bool
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringSliceVarP(&listTargets, "target", "t", []string{}, "target tools to list (opencode, cursor, claude)")
	listCmd.Flags().BoolVarP(&listGlobal, "global", "g", false, "list globally installed packages (ignore ait.yml)")
}

func runList(cmd *cobra.Command, args []string) error {
	// Check if we're in a project with ait.yml (unless --global flag is set)
	manifestPath := "ait.yml"
	hasManifest := config.ManifestExists(manifestPath)

	if hasManifest && !listGlobal {
		// List project-local packages
		return listProjectPackages()
	}

	// List globally installed packages
	return listGlobalPackages()
}

// listProjectPackages lists packages installed in the current project
func listProjectPackages() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	utils.PrintInfo("Listing project-local packages from %s", cwd)

	// Use ProjectRootAdapter to list project-local installations
	projectAdapter := adapters.NewProjectRootAdapter(cwd)
	packages, err := projectAdapter.List()
	if err != nil {
		return fmt.Errorf("failed to list project packages: %w", err)
	}

	if len(packages) == 0 {
		utils.PrintInfo("No packages installed in this project")
		utils.PrintInfo("Run 'ait install' to install packages from ait.yml")
		return nil
	}

	// Group by type
	agentCount := 0
	skillCount := 0
	promptCount := 0

	for _, pkg := range packages {
		typeStr := string(pkg.Type)
		versionStr := ""
		if pkg.Version != "" {
			versionStr = fmt.Sprintf(" (%s)", pkg.Version)
		}
		fmt.Printf("  • %s%s [%s]\n", pkg.Name, versionStr, typeStr)

		// Count by type
		switch typeStr {
		case "agent":
			agentCount++
		case "skill":
			skillCount++
		case "prompt":
			promptCount++
		}
	}

	// Print summary
	fmt.Println()
	utils.PrintSuccess("Total: %d package(s) installed", len(packages))
	if agentCount > 0 {
		utils.PrintInfo("  • %d agent(s)", agentCount)
	}
	if skillCount > 0 {
		utils.PrintInfo("  • %d skill(s)", skillCount)
	}
	if promptCount > 0 {
		utils.PrintInfo("  • %d prompt(s)", promptCount)
	}

	return nil
}

// listGlobalPackages lists packages installed globally to AI tools
func listGlobalPackages() error {
	// Get targets to list
	targets := listTargets
	if len(targets) == 0 {
		// Detect all installed tools
		utils.PrintInfo("Detecting installed AI tools...")
		targets = adapters.DetectInstalledTools()
		if len(targets) == 0 {
			utils.PrintWarning("No AI tools detected")
			utils.PrintInfo("Install OpenCode, Cursor, or Claude Desktop first")
			return nil
		}
	}

	// List packages for each target
	hasPackages := false
	for _, target := range targets {
		adapter, err := adapters.GetAdapter(target)
		if err != nil {
			utils.PrintWarning("Skipping %s: %v", target, err)
			continue
		}

		packages, err := adapter.List()
		if err != nil {
			utils.PrintError("Failed to list packages for %s: %v", target, err)
			continue
		}

		if len(packages) == 0 {
			continue
		}

		hasPackages = true
		utils.PrintInfo("\n%s:", target)
		for _, pkg := range packages {
			typeStr := string(pkg.Type)
			versionStr := ""
			if pkg.Version != "" {
				versionStr = fmt.Sprintf(" (%s)", pkg.Version)
			}
			fmt.Printf("  • %s%s [%s]\n", pkg.Name, versionStr, typeStr)
		}
	}

	if !hasPackages {
		utils.PrintInfo("No packages installed")
		utils.PrintInfo("Run 'ait install --global' to install packages globally")
	}

	return nil
}
