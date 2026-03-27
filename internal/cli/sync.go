package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex-ai/ait/internal/adapters"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync project-local .ait/ packages to AI tools",
	Long: `Sync packages from project-local .ait/ directory to your AI tools (OpenCode, Cursor, Claude, etc).

This command bridges project-local packages with your global AI tools:
- Reads packages from .ait/ directory
- Copies them to your AI tools' global directories
- Each tool can now access the project's AI agents/skills/prompts

This is useful when:
- Working on a team project with shared AI tools in .ait/
- You want your AI tools to see the project's custom agents
- After cloning a repo with .ait/ directory

Examples:
  # Sync to all detected AI tools
  ait sync

  # Sync to specific tools
  ait sync --target opencode --target cursor

  # Sync and show what would be copied (dry-run)
  ait sync --dry-run

Tip: Add this to your project setup or use a git hook to auto-sync!`,
	RunE: runSync,
}

var (
	syncTargets []string
	syncDryRun  bool
	syncForce   bool
)

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringSliceVarP(&syncTargets, "target", "t", []string{}, "target tools to sync to (opencode, cursor, claude)")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "show what would be synced without copying")
	syncCmd.Flags().BoolVarP(&syncForce, "force", "f", false, "force overwrite existing packages in tools")
}

func runSync(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if .ait directory exists
	aitDir := filepath.Join(cwd, ".ait")
	if _, err := os.Stat(aitDir); os.IsNotExist(err) {
		return fmt.Errorf(".ait directory not found. Run 'ait install' first to create project-local packages")
	}

	// List packages in .ait/
	projectAdapter := adapters.NewProjectAdapter(cwd)
	projectPackages, err := projectAdapter.List()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	if len(projectPackages) == 0 {
		utils.PrintWarning("No packages found in .ait/ directory")
		return nil
	}

	utils.PrintInfo(fmt.Sprintf("Found %d package(s) in .ait/", len(projectPackages)))

	// Get global AI tool adapters
	targetAdapters, err := getGlobalAdapters(syncTargets)
	if err != nil {
		return err
	}

	if len(targetAdapters) == 0 {
		return fmt.Errorf("no AI tools available to sync to")
	}

	if syncDryRun {
		utils.PrintInfo("\n--- Dry Run: Packages that would be synced ---")
		for _, pkg := range projectPackages {
			utils.PrintInfo(fmt.Sprintf("  • %s [%s]", pkg.Name, pkg.Type))
			for toolName := range targetAdapters {
				utils.PrintInfo(fmt.Sprintf("    → %s", toolName))
			}
		}
		utils.PrintInfo("--- End of dry-run ---")
		utils.PrintInfo("Run without --dry-run to sync packages")
		return nil
	}

	// Sync each package to each tool
	utils.PrintInfo(fmt.Sprintf("Syncing to %d tool(s)...", len(targetAdapters)))

	syncedCount := 0
	for _, pkg := range projectPackages {
		utils.PrintInfo(fmt.Sprintf("Syncing %s...", pkg.Name))

		for toolName, adapter := range targetAdapters {
			// Check if package already exists in tool
			existingPkgs, _ := adapter.List()
			exists := false
			for _, existing := range existingPkgs {
				if existing.Name == pkg.Name && existing.Type == pkg.Type {
					exists = true
					break
				}
			}

			if exists && !syncForce {
				utils.PrintWarning(fmt.Sprintf("  • %s already has %s (use --force to overwrite)", toolName, pkg.Name))
				continue
			}

			// Install to tool
			if err := installToAdapter(pkg, adapter, toolName); err != nil {
				utils.PrintWarning(fmt.Sprintf("  • Failed to sync to %s: %s", toolName, err.Error()))
				continue
			}

			utils.PrintSuccess(fmt.Sprintf("  • Synced to %s", toolName))
		}
		syncedCount++
	}

	if syncedCount > 0 {
		utils.PrintSuccess(fmt.Sprintf("\nSuccessfully synced %d package(s)", syncedCount))
		utils.PrintInfo("Your AI tools can now access these packages!")
	}

	return nil
}
