package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apex-ai/ait/internal/packages"
)

// TestInstallPackageFile tests the common package installation logic
func TestInstallPackageFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	packageDir := filepath.Join(tmpDir, "package")

	// Create a test package directory with a test file
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		t.Fatalf("failed to create package directory: %v", err)
	}

	testContent := "# Test Agent\n\nThis is a test agent."
	testFile := filepath.Join(packageDir, "AGENT.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	testPkg := &packages.Package{
		Name:    "test-agent",
		Type:    packages.TypeAgent,
		Version: "1.0.0",
		Path:    packageDir,
	}

	tests := []struct {
		name             string
		config           PackageInstallConfig
		expectedDestPath string
	}{
		{
			name: "install with package subdirectory",
			config: PackageInstallConfig{
				TargetSubdir:     "agents",
				SourceFileName:   "AGENT.md",
				DestFileName:     "AGENT.md",
				UsePackageSubdir: true,
			},
			expectedDestPath: filepath.Join(configDir, "agents", "test-agent", "AGENT.md"),
		},
		{
			name: "install without package subdirectory",
			config: PackageInstallConfig{
				TargetSubdir:     "prompts",
				SourceFileName:   "AGENT.md",
				DestFileName:     "test-agent.txt",
				UsePackageSubdir: false,
			},
			expectedDestPath: filepath.Join(configDir, "prompts", "test-agent.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean config dir between tests
			os.RemoveAll(configDir)

			err := InstallPackageFile(testPkg, configDir, "test", tt.config)
			if err != nil {
				t.Errorf("InstallPackageFile() error = %v", err)
				return
			}

			// Verify file was created at expected location
			if _, err := os.Stat(tt.expectedDestPath); os.IsNotExist(err) {
				t.Errorf("expected file at %s was not created", tt.expectedDestPath)
				return
			}

			// Verify content matches
			content, err := os.ReadFile(tt.expectedDestPath)
			if err != nil {
				t.Errorf("failed to read installed file: %v", err)
				return
			}

			if string(content) != testContent {
				t.Errorf("installed file content = %q, want %q", string(content), testContent)
			}
		})
	}
}

// TestUninstallPackage tests the common package uninstallation logic
func TestUninstallPackage(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	tests := []struct {
		name       string
		pkg        *packages.Package
		setupFiles []string
		expectPath string
	}{
		{
			name: "uninstall agent",
			pkg: &packages.Package{
				Name: "test-agent",
				Type: packages.TypeAgent,
			},
			setupFiles: []string{
				filepath.Join(configDir, "agents", "test-agent", "AGENT.md"),
				filepath.Join(configDir, "agents", "test-agent", "README.md"),
			},
			expectPath: filepath.Join(configDir, "agents", "test-agent"),
		},
		{
			name: "uninstall skill",
			pkg: &packages.Package{
				Name: "test-skill",
				Type: packages.TypeSkill,
			},
			setupFiles: []string{
				filepath.Join(configDir, "skills", "test-skill", "SKILL.md"),
			},
			expectPath: filepath.Join(configDir, "skills", "test-skill"),
		},
		{
			name: "uninstall prompt",
			pkg: &packages.Package{
				Name: "test-prompt",
				Type: packages.TypePrompt,
			},
			setupFiles: []string{
				filepath.Join(configDir, "prompts", "test-prompt.txt"),
			},
			expectPath: filepath.Join(configDir, "prompts", "test-prompt.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean config dir
			os.RemoveAll(configDir)

			// Setup test files
			for _, file := range tt.setupFiles {
				dir := filepath.Dir(file)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			// Verify files exist before uninstall
			if _, err := os.Stat(tt.expectPath); os.IsNotExist(err) {
				t.Fatalf("test setup failed: expected path does not exist: %s", tt.expectPath)
			}

			// Uninstall
			err := UninstallPackage(tt.pkg, configDir, "agents", "skills", "prompts")
			if err != nil {
				t.Errorf("UninstallPackage() error = %v", err)
				return
			}

			// Verify files were removed
			if _, err := os.Stat(tt.expectPath); !os.IsNotExist(err) {
				t.Errorf("expected path still exists after uninstall: %s", tt.expectPath)
			}
		})
	}
}

// TestListPackagesInDir tests listing packages from a directory
func TestListPackagesInDir(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		setupDirs     []string
		pkgType       packages.PackageType
		expectedCount int
		expectedNames []string
	}{
		{
			name: "list multiple agents",
			setupDirs: []string{
				filepath.Join(tmpDir, "agent1"),
				filepath.Join(tmpDir, "agent2"),
				filepath.Join(tmpDir, "agent3"),
			},
			pkgType:       packages.TypeAgent,
			expectedCount: 3,
			expectedNames: []string{"agent1", "agent2", "agent3"},
		},
		{
			name:          "list empty directory",
			setupDirs:     []string{},
			pkgType:       packages.TypeSkill,
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			os.RemoveAll(testDir)

			// Setup directories
			for _, dir := range tt.setupDirs {
				fullPath := filepath.Join(testDir, filepath.Base(dir))
				if err := os.MkdirAll(fullPath, 0755); err != nil {
					t.Fatalf("failed to create test directory: %v", err)
				}
			}

			// List packages
			pkgs, err := ListPackagesInDir(testDir, tt.pkgType)
			if err != nil {
				t.Errorf("ListPackagesInDir() error = %v", err)
				return
			}

			if len(pkgs) != tt.expectedCount {
				t.Errorf("ListPackagesInDir() returned %d packages, want %d", len(pkgs), tt.expectedCount)
				return
			}

			// Verify package names
			for i, pkg := range pkgs {
				if i < len(tt.expectedNames) && pkg.Name != tt.expectedNames[i] {
					t.Errorf("package[%d].Name = %s, want %s", i, pkg.Name, tt.expectedNames[i])
				}
				if pkg.Type != tt.pkgType {
					t.Errorf("package[%d].Type = %s, want %s", i, pkg.Type, tt.pkgType)
				}
			}
		})
	}
}

// TestListPromptsInDir tests listing prompt packages from a directory
func TestListPromptsInDir(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		setupFiles    []string
		expectedCount int
		expectedNames []string
	}{
		{
			name: "list multiple prompts",
			setupFiles: []string{
				"prompt1.txt",
				"prompt2.txt",
				"prompt3.txt",
				"not-a-prompt.md", // Should be ignored
			},
			expectedCount: 3,
			expectedNames: []string{"prompt1", "prompt2", "prompt3"},
		},
		{
			name:          "list empty directory",
			setupFiles:    []string{},
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "ignore non-txt files",
			setupFiles: []string{
				"prompt.txt",
				"README.md",
				"config.json",
			},
			expectedCount: 1,
			expectedNames: []string{"prompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, tt.name)
			os.RemoveAll(testDir)
			os.MkdirAll(testDir, 0755)

			// Setup files
			for _, file := range tt.setupFiles {
				fullPath := filepath.Join(testDir, file)
				if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			// List prompts
			pkgs, err := ListPromptsInDir(testDir)
			if err != nil {
				t.Errorf("ListPromptsInDir() error = %v", err)
				return
			}

			if len(pkgs) != tt.expectedCount {
				t.Errorf("ListPromptsInDir() returned %d packages, want %d", len(pkgs), tt.expectedCount)
				return
			}

			// Verify package names (they should all be prompts)
			for _, pkg := range pkgs {
				if pkg.Type != packages.TypePrompt {
					t.Errorf("package.Type = %s, want %s", pkg.Type, packages.TypePrompt)
				}
			}
		})
	}
}

// TestListPackages tests the combined package listing
func TestListPackages(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	// Setup test structure
	testStructure := map[string][]string{
		"agents":  {"agent1", "agent2"},
		"skills":  {"skill1"},
		"prompts": {}, // Will add files, not directories
	}

	for subdir, items := range testStructure {
		dirPath := filepath.Join(configDir, subdir)
		os.MkdirAll(dirPath, 0755)

		if subdir == "prompts" {
			// Create prompt files
			os.WriteFile(filepath.Join(dirPath, "prompt1.txt"), []byte("test"), 0644)
			os.WriteFile(filepath.Join(dirPath, "prompt2.txt"), []byte("test"), 0644)
		} else {
			// Create subdirectories for agents/skills
			for _, item := range items {
				os.MkdirAll(filepath.Join(dirPath, item), 0755)
			}
		}
	}

	// List all packages
	pkgs, err := ListPackages(configDir, "agents", "skills", "prompts")
	if err != nil {
		t.Fatalf("ListPackages() error = %v", err)
	}

	// Expected: 2 agents + 1 skill + 2 prompts = 5 packages
	expectedCount := 5
	if len(pkgs) != expectedCount {
		t.Errorf("ListPackages() returned %d packages, want %d", len(pkgs), expectedCount)
	}

	// Count by type
	typeCounts := make(map[packages.PackageType]int)
	for _, pkg := range pkgs {
		typeCounts[pkg.Type]++
	}

	if typeCounts[packages.TypeAgent] != 2 {
		t.Errorf("found %d agents, want 2", typeCounts[packages.TypeAgent])
	}
	if typeCounts[packages.TypeSkill] != 1 {
		t.Errorf("found %d skills, want 1", typeCounts[packages.TypeSkill])
	}
	if typeCounts[packages.TypePrompt] != 2 {
		t.Errorf("found %d prompts, want 2", typeCounts[packages.TypePrompt])
	}
}

// TestValidateConfigDir tests the configuration directory validation
func TestValidateConfigDir(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name            string
		configDir       string
		createIfMissing bool
		setup           func(string)
		wantErr         bool
	}{
		{
			name:            "valid existing directory",
			configDir:       filepath.Join(tmpDir, "valid"),
			createIfMissing: false,
			setup: func(dir string) {
				os.MkdirAll(dir, 0755)
			},
			wantErr: false,
		},
		{
			name:            "create missing directory",
			configDir:       filepath.Join(tmpDir, "create-me"),
			createIfMissing: true,
			setup:           func(dir string) {},
			wantErr:         false,
		},
		{
			name:            "error on missing directory",
			configDir:       filepath.Join(tmpDir, "missing"),
			createIfMissing: false,
			setup:           func(dir string) {},
			wantErr:         true,
		},
		{
			name:            "error on non-writable directory",
			configDir:       filepath.Join(tmpDir, "readonly"),
			createIfMissing: false,
			setup: func(dir string) {
				os.MkdirAll(dir, 0555) // Read-only
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.RemoveAll(tt.configDir)

			// Setup
			tt.setup(tt.configDir)

			// Test
			err := ValidateConfigDir(tt.configDir, tt.createIfMissing)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Cleanup readonly directories for next test
			if tt.name == "error on non-writable directory" {
				os.Chmod(tt.configDir, 0755)
			}
		})
	}
}
