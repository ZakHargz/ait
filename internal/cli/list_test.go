package cli

import (
	"os"
	"testing"
)

// TestListCmd_NoManifest tests list command without ait.yml
func TestListCmd_NoManifest(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Without ait.yml, should list globally
	err := runList(listCmd, []string{})

	// May succeed or fail depending on tools installed
	// Should not panic
	t.Logf("List without manifest result: %v", err)
}

// TestListCmd_WithManifest tests list command with ait.yml
func TestListCmd_WithManifest(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Create ait.yml
	manifestContent := `name: test-project
version: 1.0.0
dependencies: []
`
	err := os.WriteFile("ait.yml", []byte(manifestContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test ait.yml: %v", err)
	}

	// With ait.yml, should list project-local packages
	err = runList(listCmd, []string{})

	// Should succeed even with no packages
	if err != nil {
		t.Errorf("List with manifest failed: %v", err)
	}
}

// TestListCmd_GlobalFlag tests --global flag behavior
func TestListCmd_GlobalFlag(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Create ait.yml
	manifestContent := `name: test-project
version: 1.0.0
dependencies: []
`
	err := os.WriteFile("ait.yml", []byte(manifestContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test ait.yml: %v", err)
	}

	// Set global flag
	listGlobal = true
	defer func() { listGlobal = false }()

	// Should list globally even with ait.yml
	err = runList(listCmd, []string{})

	t.Logf("List with --global flag result: %v", err)
}

// TestListProjectPackages tests project-local package listing
func TestListProjectPackages(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Create project structure with mock agents
	os.MkdirAll(".github/agents", 0755)
	os.WriteFile(".github/agents/test-agent.agent.md", []byte("# Test Agent"), 0644)

	err := listProjectPackages()

	// Should succeed and find 1 package
	if err != nil {
		t.Errorf("listProjectPackages() failed: %v", err)
	}
}

// TestListGlobalPackages tests global package listing
func TestListGlobalPackages(t *testing.T) {
	// This will try to detect tools, may or may not find any
	err := listGlobalPackages()

	// Should not panic, may succeed or fail gracefully
	t.Logf("listGlobalPackages() result: %v", err)
}

// TestListCmd_Flags tests list command flags
func TestListCmd_Flags(t *testing.T) {
	// Verify flags are registered
	if listCmd.Flags().Lookup("global") == nil {
		t.Error("--global flag not registered")
	}

	if listCmd.Flags().Lookup("target") == nil {
		t.Error("--target flag not registered")
	}

	// Verify default values
	globalFlag := listCmd.Flags().Lookup("global")
	if globalFlag.DefValue != "false" {
		t.Errorf("--global flag default should be false, got %s", globalFlag.DefValue)
	}
}
