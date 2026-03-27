package cli

import (
	"fmt"

	"github.com/apex-ai/ait/internal/adapters"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages",
	Long: `List all packages installed to AI tools.

Examples:
  # List all installed packages
  ait list

  # List packages for specific tool
  ait list --target opencode`,
	RunE: runList,
}

var (
	listTargets []string
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringSliceVarP(&listTargets, "target", "t", []string{}, "target tools to list (opencode, cursor, claude)")
}

func runList(cmd *cobra.Command, args []string) error {
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
			utils.PrintWarning(fmt.Sprintf("Skipping %s: %s", target, err.Error()))
			continue
		}

		packages, err := adapter.List()
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to list packages for %s: %s", target, err.Error()))
			continue
		}

		if len(packages) == 0 {
			continue
		}

		hasPackages = true
		utils.PrintInfo(fmt.Sprintf("\n%s:", target))
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
		utils.PrintInfo("Run 'ait install' to install packages from ait.yml")
	}

	return nil
}
