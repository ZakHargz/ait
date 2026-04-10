package adapters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apex-ai/ait/internal/packages"
)

func TestOpenCodeAdapter_InstallAgent(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "opencode")
	packageDir := filepath.Join(tmpDir, "package")

	// Setup test package
	os.MkdirAll(packageDir, 0755)
	agentContent := "# Test Agent\nThis is a test agent."
	os.WriteFile(filepath.Join(packageDir, "AGENT.md"), []byte(agentContent), 0644)

	pkg := &packages.Package{
		Name:    "test-agent",
		Type:    packages.TypeAgent,
		Version: "1.0.0",
		Path:    packageDir,
	}

	adapter := &OpenCodeAdapter{
		BaseAdapter: NewBaseAdapter("opencode", configDir),
	}

	// Test installation
	err := adapter.InstallAgent(pkg)
	if err != nil {
		t.Fatalf("InstallAgent() error = %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(configDir, "agents", "test-agent", "AGENT.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s was not created", expectedPath)
	}

	// Verify content
	content, _ := os.ReadFile(expectedPath)
	if string(content) != agentContent {
		t.Errorf("installed file content = %q, want %q", string(content), agentContent)
	}
}

func TestOpenCodeAdapter_InstallSkill(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "opencode")
	packageDir := filepath.Join(tmpDir, "package")

	// Setup test package
	os.MkdirAll(packageDir, 0755)
	skillContent := "# Test Skill\nThis is a test skill."
	os.WriteFile(filepath.Join(packageDir, "SKILL.md"), []byte(skillContent), 0644)

	pkg := &packages.Package{
		Name:    "test-skill",
		Type:    packages.TypeSkill,
		Version: "1.0.0",
		Path:    packageDir,
	}

	adapter := &OpenCodeAdapter{
		BaseAdapter: NewBaseAdapter("opencode", configDir),
	}

	// Test installation
	err := adapter.InstallSkill(pkg)
	if err != nil {
		t.Fatalf("InstallSkill() error = %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(configDir, "skills", "test-skill", "SKILL.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s was not created", expectedPath)
	}
}

func TestOpenCodeAdapter_InstallPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "opencode")
	packageDir := filepath.Join(tmpDir, "package")

	// Setup test package
	os.MkdirAll(packageDir, 0755)
	promptContent := "This is a test prompt."
	os.WriteFile(filepath.Join(packageDir, "prompt.txt"), []byte(promptContent), 0644)

	pkg := &packages.Package{
		Name:    "test-prompt",
		Type:    packages.TypePrompt,
		Version: "1.0.0",
		Path:    packageDir,
	}

	adapter := &OpenCodeAdapter{
		BaseAdapter: NewBaseAdapter("opencode", configDir),
	}

	// Test installation
	err := adapter.InstallPrompt(pkg)
	if err != nil {
		t.Fatalf("InstallPrompt() error = %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(configDir, "prompts", "test-prompt.txt")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at %s was not created", expectedPath)
	}
}

func TestOpenCodeAdapter_Uninstall(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "opencode")

	// Setup installed agent
	agentPath := filepath.Join(configDir, "agents", "test-agent")
	os.MkdirAll(agentPath, 0755)
	os.WriteFile(filepath.Join(agentPath, "AGENT.md"), []byte("test"), 0644)

	pkg := &packages.Package{
		Name: "test-agent",
		Type: packages.TypeAgent,
	}

	adapter := &OpenCodeAdapter{
		BaseAdapter: NewBaseAdapter("opencode", configDir),
	}

	// Test uninstallation
	err := adapter.Uninstall(pkg)
	if err != nil {
		t.Fatalf("Uninstall() error = %v", err)
	}

	// Verify directory was removed
	if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
		t.Errorf("expected agent directory to be removed")
	}
}

func TestOpenCodeAdapter_List(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "opencode")

	// Setup test structure
	os.MkdirAll(filepath.Join(configDir, "agents", "agent1"), 0755)
	os.MkdirAll(filepath.Join(configDir, "agents", "agent2"), 0755)
	os.MkdirAll(filepath.Join(configDir, "skills", "skill1"), 0755)
	os.MkdirAll(filepath.Join(configDir, "prompts"), 0755) // Create prompts directory
	os.WriteFile(filepath.Join(configDir, "prompts", "prompt1.txt"), []byte("test"), 0644)

	adapter := &OpenCodeAdapter{
		BaseAdapter: NewBaseAdapter("opencode", configDir),
	}

	// Test listing
	pkgs, err := adapter.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should find 2 agents + 1 skill + 1 prompt = 4 packages
	if len(pkgs) != 4 {
		t.Errorf("List() returned %d packages, want 4", len(pkgs))
	}
}

func TestOpenCodeAdapter_Validate(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		configDir string
		setup     func()
		wantErr   bool
	}{
		{
			name:      "valid existing directory",
			configDir: filepath.Join(tmpDir, "valid"),
			setup: func() {
				os.MkdirAll(filepath.Join(tmpDir, "valid"), 0755)
			},
			wantErr: false,
		},
		{
			name:      "missing directory",
			configDir: filepath.Join(tmpDir, "missing"),
			setup:     func() {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			adapter := &OpenCodeAdapter{
				BaseAdapter: NewBaseAdapter("opencode", tt.configDir),
			}

			err := adapter.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
