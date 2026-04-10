package cli

import (
	"os"
	"testing"

	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/sources"
)

func TestOutdatedCmd_NoManifest(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Run outdated command without ait.yml
	err := runOutdated(outdatedCmd, []string{})

	// Should complete without error (just print warnings)
	if err != nil {
		t.Errorf("Expected no error when no manifest exists, got: %v", err)
	}
}

func TestOutdatedCmd_NoLockfile(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Create ait.yml but no ait.lock
	manifest := &config.Manifest{
		Name:         "test-project",
		Version:      "1.0.0",
		Dependencies: []string{"github:test/repo/package@1.0.0"},
	}
	err := manifest.Write("ait.yml")
	if err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Run outdated command
	err = runOutdated(outdatedCmd, []string{})

	// Should complete without error (just print warnings)
	if err != nil {
		t.Errorf("Expected no error when no lockfile exists, got: %v", err)
	}
}

func TestOutdatedCmd_NoPackages(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Create ait.yml
	manifest := &config.Manifest{
		Name:         "test-project",
		Version:      "1.0.0",
		Dependencies: []string{},
	}
	err := manifest.Write("ait.yml")
	if err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Create empty ait.lock
	lockfile := config.NewLockFile()
	err = lockfile.Write("ait.lock")
	if err != nil {
		t.Fatalf("Failed to create lockfile: %v", err)
	}

	// Run outdated command
	err = runOutdated(outdatedCmd, []string{})

	// Should complete without error
	if err != nil {
		t.Errorf("Expected no error with no packages, got: %v", err)
	}
}

func TestCheckPackageVersion_LocalPackage(t *testing.T) {
	pkg := config.LockedPkg{
		Name:     "test-package",
		Version:  "1.0.0",
		Type:     "agent",
		Source:   "local:./packages/test",
		Resolved: "1.0.0",
	}

	gitSource := sources.NewGitSource("")

	info := checkPackageVersion(pkg, gitSource)

	if info.Name != "test-package" {
		t.Errorf("Expected name 'test-package', got %s", info.Name)
	}

	if info.LatestVersion != "local" {
		t.Errorf("Expected latest version 'local' for local package, got %s", info.LatestVersion)
	}

	if info.IsOutdated {
		t.Error("Local package should not be marked as outdated")
	}

	if info.Error != "" {
		t.Errorf("Expected no error for local package, got %s", info.Error)
	}
}

func TestDisplayOutdatedPackages_Empty(t *testing.T) {
	infos := []PackageVersionInfo{}

	err := displayOutdatedPackages(infos)

	if err != nil {
		t.Errorf("Expected no error for empty list, got: %v", err)
	}
}

func TestDisplayOutdatedPackages_AllUpToDate(t *testing.T) {
	infos := []PackageVersionInfo{
		{
			Name:           "package1",
			CurrentVersion: "1.0.0",
			LatestVersion:  "1.0.0",
			IsOutdated:     false,
			Type:           "agent",
		},
		{
			Name:           "package2",
			CurrentVersion: "2.0.0",
			LatestVersion:  "2.0.0",
			IsOutdated:     false,
			Type:           "skill",
		},
	}

	err := displayOutdatedPackages(infos)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestDisplayOutdatedPackages_WithOutdated(t *testing.T) {
	infos := []PackageVersionInfo{
		{
			Name:           "package1",
			CurrentVersion: "1.0.0",
			LatestVersion:  "2.0.0",
			IsOutdated:     true,
			Type:           "agent",
		},
		{
			Name:           "package2",
			CurrentVersion: "2.0.0",
			LatestVersion:  "2.0.0",
			IsOutdated:     false,
			Type:           "skill",
		},
	}

	err := displayOutdatedPackages(infos)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestDisplayOutdatedPackages_WithErrors(t *testing.T) {
	infos := []PackageVersionInfo{
		{
			Name:           "package1",
			CurrentVersion: "1.0.0",
			LatestVersion:  "",
			IsOutdated:     false,
			Error:          "failed to fetch",
			Type:           "agent",
		},
	}

	err := displayOutdatedPackages(infos)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestDisplayOutdatedPackages_ShowAll(t *testing.T) {
	// Set show all flag
	originalShowAll := outdatedShowAll
	outdatedShowAll = true
	defer func() { outdatedShowAll = originalShowAll }()

	infos := []PackageVersionInfo{
		{
			Name:           "package1",
			CurrentVersion: "1.0.0",
			LatestVersion:  "2.0.0",
			IsOutdated:     true,
			Type:           "agent",
		},
		{
			Name:           "package2",
			CurrentVersion: "2.0.0",
			LatestVersion:  "2.0.0",
			IsOutdated:     false,
			Type:           "skill",
		},
	}

	err := displayOutdatedPackages(infos)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
