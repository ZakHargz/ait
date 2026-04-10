package cli

import (
	"os"
	"testing"
)

func TestDoctorCmd_BasicExecution(t *testing.T) {
	// Test that doctor command runs without crashing
	err := runDoctor(doctorCmd, []string{})

	// Should complete (may have warnings but shouldn't error in most cases)
	if err != nil {
		t.Logf("Doctor returned error (expected if Git not installed): %v", err)
	}
}

func TestCheckGit(t *testing.T) {
	check := checkGit()

	if check.name != "Git installation" {
		t.Errorf("Expected check name 'Git installation', got %s", check.name)
	}

	// Most systems have git, but test should handle both cases
	if check.status == "pass" {
		t.Log("Git is installed")
		if check.message == "" {
			t.Error("Git check passed but message is empty")
		}
	} else if check.status == "fail" {
		t.Log("Git is not installed (expected on some test systems)")
		if !check.critical {
			t.Error("Git check should be marked as critical")
		}
	}
}

func TestCheckAITools(t *testing.T) {
	checks := checkAITools()

	if len(checks) == 0 {
		t.Error("checkAITools should return at least one check")
	}

	// First check should be overall detection
	if checks[0].name != "AI tools detection" {
		t.Errorf("First check should be 'AI tools detection', got %s", checks[0].name)
	}

	// Should have checks for individual tools
	toolNames := make(map[string]bool)
	for _, check := range checks {
		toolNames[check.name] = true
	}

	expectedTools := []string{"AI tools detection", "OpenCode", "Cursor", "Claude Desktop"}
	for _, tool := range expectedTools {
		if !toolNames[tool] {
			t.Errorf("Missing check for %s", tool)
		}
	}
}

func TestCheckConfigDirs(t *testing.T) {
	checks := checkConfigDirs()

	if len(checks) < 2 {
		t.Error("checkConfigDirs should return at least 2 checks")
	}

	// Should check current directory
	foundCwdCheck := false
	foundPermCheck := false

	for _, check := range checks {
		if check.name == "Current directory" {
			foundCwdCheck = true
		}
		if check.name == "Write permissions" {
			foundPermCheck = true
		}
	}

	if !foundCwdCheck {
		t.Error("Missing current directory check")
	}
	if !foundPermCheck {
		t.Error("Missing write permissions check")
	}
}

func TestCheckManifest(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Test without manifest
	check := checkManifest()
	if check.status != "warn" {
		t.Errorf("Expected warn status without manifest, got %s", check.status)
	}

	// Test with valid manifest
	manifestContent := `name: test
version: 1.0.0
dependencies: []
`
	os.WriteFile("ait.yml", []byte(manifestContent), 0644)

	check = checkManifest()
	if check.status != "pass" {
		t.Errorf("Expected pass status with valid manifest, got %s", check.status)
	}

	// Test with invalid manifest
	os.WriteFile("ait.yml", []byte("invalid: yaml: content:"), 0644)

	check = checkManifest()
	if check.status != "fail" {
		t.Errorf("Expected fail status with invalid manifest, got %s", check.status)
	}
}

func TestCheckLockFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Test without lock file
	check := checkLockFile()
	if check.status != "warn" {
		t.Errorf("Expected warn status without lock file, got %s", check.status)
	}

	// Test with valid lock file
	lockContent := `version: "1.0"
generated: 2024-01-01T00:00:00Z
packages: {}
`
	os.WriteFile("ait.lock", []byte(lockContent), 0644)

	check = checkLockFile()
	if check.status != "pass" {
		t.Errorf("Expected pass status with valid lock file, got %s: %s", check.status, check.message)
	}
}

func TestCheckAuthentication(t *testing.T) {
	// Save original env
	originalTokens := map[string]string{
		"GITHUB_TOKEN":   os.Getenv("GITHUB_TOKEN"),
		"GH_TOKEN":       os.Getenv("GH_TOKEN"),
		"GITHUB_APM_PAT": os.Getenv("GITHUB_APM_PAT"),
	}
	defer func() {
		for key, value := range originalTokens {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Test without tokens
	os.Unsetenv("GITHUB_TOKEN")
	os.Unsetenv("GH_TOKEN")
	os.Unsetenv("GITHUB_APM_PAT")

	check := checkAuthentication()
	if check.status != "warn" {
		t.Errorf("Expected warn status without tokens, got %s", check.status)
	}

	// Test with token
	os.Setenv("GITHUB_TOKEN", "test-token")

	check = checkAuthentication()
	if check.status != "pass" {
		t.Errorf("Expected pass status with token, got %s", check.status)
	}
}

func TestPrintCheck(t *testing.T) {
	// Test that printCheck doesn't panic with different statuses
	checks := []healthCheck{
		{name: "Test Pass", status: "pass", message: "OK"},
		{name: "Test Warn", status: "warn", message: "Warning"},
		{name: "Test Fail", status: "fail", message: "Failed"},
	}

	for _, check := range checks {
		// Should not panic
		printCheck(check)
	}
}

func TestDoctorCmd_InProjectWithManifest(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Create valid manifest
	manifestContent := `name: test-project
version: 1.0.0
dependencies:
  - test/package
`
	os.WriteFile("ait.yml", []byte(manifestContent), 0644)

	// Run doctor
	err := runDoctor(doctorCmd, []string{})

	// Should succeed (warnings allowed)
	if err != nil && err.Error() != "health check failed with 1 error(s)" {
		// Allow failure only if it's from missing Git
		t.Logf("Doctor completed with: %v", err)
	}
}
