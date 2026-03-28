package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new ait.yml manifest",
	Long: `Initialize a new ait.yml manifest file in the current directory.

This command will interactively prompt for project information and create
a new ait.yml file with sensible defaults.`,
	RunE: runInit,
}

var (
	initName     string
	initVersion  string
	initDefaults bool
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initName, "name", "", "project name")
	initCmd.Flags().StringVar(&initVersion, "version", "1.0.0", "project version")
	initCmd.Flags().BoolVar(&initDefaults, "defaults", false, "use default values without prompting")
}

func runInit(cmd *cobra.Command, args []string) error {
	manifestPath := "ait.yml"

	// Check if manifest already exists
	if config.ManifestExists(manifestPath) {
		utils.PrintWarning("ait.yml already exists in this directory")

		if !initDefaults {
			if !promptYesNo("Overwrite existing ait.yml?", false) {
				utils.PrintInfo("Initialization cancelled")
				return nil
			}
		}
	}

	var manifest *config.Manifest

	if initDefaults {
		manifest = createDefaultManifest()
	} else {
		manifest = promptForManifest()
	}

	// Write manifest
	if err := manifest.Write(manifestPath); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	utils.PrintSuccess("Created ait.yml")
	utils.PrintInfo("Edit ait.yml to add dependencies, then run 'ait install'")

	return nil
}

func createDefaultManifest() *config.Manifest {
	name := initName
	if name == "" {
		// Use current directory name
		cwd, _ := os.Getwd()
		name = strings.ToLower(strings.ReplaceAll(filepath.Base(cwd), " ", "-"))
	}

	return &config.Manifest{
		Name:    name,
		Version: initVersion,
		Dependencies: config.Dependencies{
			Agents:  []string{},
			Skills:  []string{},
			Prompts: []string{},
		},
		Targets: []string{},
	}
}

func promptForManifest() *config.Manifest {
	reader := bufio.NewReader(os.Stdin)

	// Project name
	name := initName
	if name == "" {
		cwd, _ := os.Getwd()
		defaultName := strings.ToLower(strings.ReplaceAll(filepath.Base(cwd), " ", "-"))
		name = promptString(reader, "Project name", defaultName)
	}

	// Project version
	version := promptString(reader, "Version", initVersion)

	// Description
	description := promptString(reader, "Description", "")

	// Targets
	utils.PrintInfo("Detecting installed AI tools...")
	// TODO: Implement tool detection
	targets := []string{}

	manifest := &config.Manifest{
		Name:        name,
		Version:     version,
		Description: description,
		Dependencies: config.Dependencies{
			Agents:  []string{},
			Skills:  []string{},
			Prompts: []string{},
		},
		Targets: targets,
	}

	return manifest
}

func promptString(reader *bufio.Reader, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s (%s): ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)

	if text == "" && defaultValue != "" {
		return defaultValue
	}

	return text
}

func promptYesNo(prompt string, defaultValue bool) bool {
	reader := bufio.NewReader(os.Stdin)

	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s [%s]: ", prompt, defaultStr)

	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))

	if text == "" {
		return defaultValue
	}

	return text == "y" || text == "yes"
}
