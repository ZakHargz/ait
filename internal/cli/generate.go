package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex-ai/ait/internal/adapters"
	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate ait.yml from installed packages in .ait/ directory",
	Long: `Generate an ait.yml manifest from packages already installed in the .ait/ directory.

This is useful when:
- You want to create a manifest from an existing .ait/ directory
- You've manually copied packages into .ait/ and want to track them
- You're setting up a new project based on existing packages

Example:
  # Generate ait.yml from .ait/ directory
  ait generate

  # Preview without writing
  ait generate --dry-run`,
	RunE: runGenerate,
}

var (
	generateDryRun bool
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().BoolVar(&generateDryRun, "dry-run", false, "show what would be generated without writing")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if .ait directory exists
	aitDir := filepath.Join(cwd, ".ait")
	if _, err := os.Stat(aitDir); os.IsNotExist(err) {
		return fmt.Errorf(".ait directory not found. No packages to generate from")
	}

	// Create project adapter to list packages
	projectAdapter := adapters.NewProjectAdapter(cwd)
	packages, err := projectAdapter.List()
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	if len(packages) == 0 {
		utils.PrintWarning("No packages found in .ait/ directory")
		return nil
	}

	utils.PrintInfo(fmt.Sprintf("Found %d package(s) in .ait/", len(packages)))

	// Get project name from directory
	projectName := filepath.Base(cwd)

	// Build manifest
	manifest := &config.Manifest{
		Name:         projectName,
		Version:      "1.0.0",
		Targets:      []string{"project"}, // Default to project-local
		Dependencies: []string{},
	}

	// Add all packages to flat dependency list
	for _, pkg := range packages {
		// Create a local reference since packages are already in .ait/
		spec := fmt.Sprintf("local:.ait/%s/%s@1.0.0", getTypeDir(string(pkg.Type)), pkg.Name)
		manifest.Dependencies = append(manifest.Dependencies, spec)
		utils.PrintInfo(fmt.Sprintf("  • %s [%s]", pkg.Name, pkg.Type))
	}

	if generateDryRun {
		utils.PrintInfo("\n--- Generated ait.yml (dry-run) ---")
		manifestYaml, err := manifest.ToYAML()
		if err != nil {
			return fmt.Errorf("failed to generate YAML: %w", err)
		}
		fmt.Println(manifestYaml)
		utils.PrintInfo("--- End of dry-run ---")
		utils.PrintInfo("Run without --dry-run to write ait.yml")
		return nil
	}

	// Check if ait.yml already exists
	manifestPath := "ait.yml"
	if config.ManifestExists(manifestPath) {
		utils.PrintWarning("ait.yml already exists")
		utils.PrintInfo("Overwriting with generated manifest...")
	}

	// Write manifest
	if err := manifest.Write(manifestPath); err != nil {
		return fmt.Errorf("failed to write ait.yml: %w", err)
	}

	utils.PrintSuccess("Generated ait.yml")
	utils.PrintInfo("Review and edit ait.yml to customize your configuration")

	return nil
}

// getTypeDir returns the directory name for a package type
func getTypeDir(pkgType string) string {
	switch pkgType {
	case "agent":
		return "agents"
	case "skill":
		return "skills"
	case "prompt":
		return "prompts"
	case "mcp":
		return "mcp"
	default:
		return pkgType + "s"
	}
}
