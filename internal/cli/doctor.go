package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apex-ai/ait/internal/adapters"
	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and configuration",
	Long: `Run diagnostics to check if AIT is properly configured.

Checks performed:
  • Git installation
  • AI tool detection (OpenCode, Cursor, Claude)
  • Configuration directory accessibility
  • Manifest and lock file validation
  • Authentication setup

Examples:
  # Run all health checks
  ait doctor

  # Run checks and show verbose output
  ait doctor --verbose`,
	RunE: runDoctor,
}

type healthCheck struct {
	name     string
	status   string // "pass", "warn", "fail"
	message  string
	critical bool
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	utils.PrintInfo("Running AIT health checks...\n")

	checks := []healthCheck{}
	warnings := 0
	errors := 0

	// Check 1: Git installation
	gitCheck := checkGit()
	checks = append(checks, gitCheck)
	if gitCheck.status == "fail" {
		errors++
	}

	// Check 2: AI Tools detection
	toolChecks := checkAITools()
	checks = append(checks, toolChecks...)
	for _, check := range toolChecks {
		if check.status == "warn" {
			warnings++
		}
	}

	// Check 3: Configuration directories
	configChecks := checkConfigDirs()
	checks = append(checks, configChecks...)
	for _, check := range configChecks {
		if check.status == "fail" {
			errors++
		} else if check.status == "warn" {
			warnings++
		}
	}

	// Check 4: Manifest validation (if exists)
	manifestCheck := checkManifest()
	checks = append(checks, manifestCheck)
	if manifestCheck.status == "warn" {
		warnings++
	} else if manifestCheck.status == "fail" {
		errors++
	}

	// Check 5: Lock file validation (if exists)
	lockCheck := checkLockFile()
	checks = append(checks, lockCheck)
	if lockCheck.status == "warn" {
		warnings++
	} else if lockCheck.status == "fail" {
		errors++
	}

	// Check 6: Authentication setup
	authCheck := checkAuthentication()
	checks = append(checks, authCheck)
	if authCheck.status == "warn" {
		warnings++
	}

	// Print results
	fmt.Println()
	for _, check := range checks {
		printCheck(check)
	}

	// Summary
	fmt.Println()
	if errors > 0 {
		utils.PrintError("%d error(s), %d warning(s)", errors, warnings)
		return fmt.Errorf("health check failed with %d error(s)", errors)
	} else if warnings > 0 {
		utils.PrintWarning("%d warning(s)", warnings)
		utils.PrintInfo("Your AIT installation has warnings but should work")
	} else {
		utils.PrintSuccess("All checks passed! AIT is ready to use")
	}

	return nil
}

func checkGit() healthCheck {
	cmd := exec.Command("git", "--version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return healthCheck{
			name:     "Git installation",
			status:   "fail",
			message:  "Git is not installed or not in PATH",
			critical: true,
		}
	}

	return healthCheck{
		name:    "Git installation",
		status:  "pass",
		message: string(output),
	}
}

func checkAITools() []healthCheck {
	checks := []healthCheck{}

	detectedTools := adapters.DetectInstalledTools()

	if len(detectedTools) == 0 {
		checks = append(checks, healthCheck{
			name:    "AI tools detection",
			status:  "warn",
			message: "No AI tools detected (OpenCode, Cursor, or Claude Desktop)",
		})
	} else {
		checks = append(checks, healthCheck{
			name:    "AI tools detection",
			status:  "pass",
			message: fmt.Sprintf("Found: %v", detectedTools),
		})
	}

	// Check individual tools
	tools := map[string]string{
		"opencode": "OpenCode",
		"cursor":   "Cursor",
		"claude":   "Claude Desktop",
	}

	for toolID, toolName := range tools {
		adapter, err := adapters.GetAdapter(toolID)
		if err != nil {
			checks = append(checks, healthCheck{
				name:    toolName,
				status:  "warn",
				message: "Not detected",
			})
			continue
		}

		if adapter.Detect() {
			configDir, _ := adapter.GetConfigDir()
			checks = append(checks, healthCheck{
				name:    toolName,
				status:  "pass",
				message: fmt.Sprintf("Detected at %s", configDir),
			})
		} else {
			checks = append(checks, healthCheck{
				name:    toolName,
				status:  "warn",
				message: "Not detected",
			})
		}
	}

	return checks
}

func checkConfigDirs() []healthCheck {
	checks := []healthCheck{}

	// Check if we're in a project directory
	cwd, err := os.Getwd()
	if err != nil {
		checks = append(checks, healthCheck{
			name:    "Current directory",
			status:  "fail",
			message: "Cannot determine current directory",
		})
		return checks
	}

	checks = append(checks, healthCheck{
		name:    "Current directory",
		status:  "pass",
		message: cwd,
	})

	// Check write permissions
	testFile := filepath.Join(cwd, ".ait-test")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		checks = append(checks, healthCheck{
			name:    "Write permissions",
			status:  "fail",
			message: "Current directory is not writable",
		})
	} else {
		os.Remove(testFile)
		checks = append(checks, healthCheck{
			name:    "Write permissions",
			status:  "pass",
			message: "Current directory is writable",
		})
	}

	return checks
}

func checkManifest() healthCheck {
	manifestPath := "ait.yml"

	if !config.ManifestExists(manifestPath) {
		return healthCheck{
			name:    "ait.yml",
			status:  "warn",
			message: "Not found (run 'ait init' to create)",
		}
	}

	manifest, err := config.LoadManifest(manifestPath)
	if err != nil {
		return healthCheck{
			name:    "ait.yml",
			status:  "fail",
			message: fmt.Sprintf("Invalid: %v", err),
		}
	}

	depCount := len(manifest.Dependencies)
	return healthCheck{
		name:    "ait.yml",
		status:  "pass",
		message: fmt.Sprintf("Valid (%d dependencies)", depCount),
	}
}

func checkLockFile() healthCheck {
	lockPath := "ait.lock"

	if !config.LockFileExists(lockPath) {
		return healthCheck{
			name:    "ait.lock",
			status:  "warn",
			message: "Not found (will be created on first install)",
		}
	}

	lockFile, err := config.LoadLockFile(lockPath)
	if err != nil {
		return healthCheck{
			name:    "ait.lock",
			status:  "fail",
			message: fmt.Sprintf("Invalid: %v", err),
		}
	}

	pkgCount := len(lockFile.Packages)
	return healthCheck{
		name:    "ait.lock",
		status:  "pass",
		message: fmt.Sprintf("Valid (%d packages locked)", pkgCount),
	}
}

func checkAuthentication() healthCheck {
	// Check for GitHub tokens
	tokens := []string{
		"GITHUB_TOKEN",
		"GH_TOKEN",
		"GITHUB_APM_PAT",
	}

	hasToken := false
	for _, token := range tokens {
		if os.Getenv(token) != "" {
			hasToken = true
			break
		}
	}

	if !hasToken {
		return healthCheck{
			name:    "GitHub authentication",
			status:  "warn",
			message: "No GitHub token found (private repos won't work)",
		}
	}

	return healthCheck{
		name:    "GitHub authentication",
		status:  "pass",
		message: "GitHub token configured",
	}
}

func printCheck(check healthCheck) {
	var icon string
	var colorFunc func(...interface{}) string

	switch check.status {
	case "pass":
		icon = "✓"
		colorFunc = func(args ...interface{}) string {
			return fmt.Sprint(args...)
		}
	case "warn":
		icon = "⚠"
		colorFunc = func(args ...interface{}) string {
			return fmt.Sprint(args...)
		}
	case "fail":
		icon = "✗"
		colorFunc = func(args ...interface{}) string {
			return fmt.Sprint(args...)
		}
	}

	fmt.Printf("%s %s: %s\n", icon, check.name, colorFunc(check.message))
}
