package config

import (
	"os"
	"testing"
)

func TestManifestUnmarshal_FlatList(t *testing.T) {
	yaml := `name: test-project
version: 1.0.0
dependencies:
  - org/repo/agents/code-reviewer@1.0.0
  - org/repo/skills/python
  - org/repo/prompts/debug
`
	tmpFile, err := os.CreateTemp("", "manifest-flat-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yaml)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	manifest, err := LoadManifest(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if manifest.Name != "test-project" {
		t.Errorf("Name: expected 'test-project', got '%s'", manifest.Name)
	}

	if len(manifest.Dependencies) != 3 {
		t.Errorf("Dependencies: expected 3, got %d", len(manifest.Dependencies))
	}

	expected := []string{
		"org/repo/agents/code-reviewer@1.0.0",
		"org/repo/skills/python",
		"org/repo/prompts/debug",
	}

	for i, dep := range manifest.Dependencies {
		if dep != expected[i] {
			t.Errorf("Dependency %d: expected '%s', got '%s'", i, expected[i], dep)
		}
	}
}

func TestManifestUnmarshal_LegacyFormat(t *testing.T) {
	yaml := `name: test-legacy
version: 1.0.0
dependencies:
  agents:
    - github:org/repo/agents/code-reviewer@1.0.0
  skills:
    - github:org/repo/skills/python
  prompts:
    - github:org/repo/prompts/debug
`
	tmpFile, err := os.CreateTemp("", "manifest-legacy-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yaml)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	manifest, err := LoadManifest(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if manifest.Name != "test-legacy" {
		t.Errorf("Name: expected 'test-legacy', got '%s'", manifest.Name)
	}

	if len(manifest.Dependencies) != 3 {
		t.Errorf("Dependencies: expected 3, got %d", len(manifest.Dependencies))
	}

	// Legacy format should be merged into flat list
	expectedDeps := map[string]bool{
		"github:org/repo/agents/code-reviewer@1.0.0": true,
		"github:org/repo/skills/python":              true,
		"github:org/repo/prompts/debug":              true,
	}

	for _, dep := range manifest.Dependencies {
		if !expectedDeps[dep] {
			t.Errorf("Unexpected dependency: %s", dep)
		}
	}
}

func TestManifestUnmarshal_APMStyleNested(t *testing.T) {
	yaml := `name: test-apm
version: 1.0.0
dependencies:
  apm:
    - org/repo/agents/code-reviewer@1.0.0
    - org/repo/skills/python
  mcp:
    - org/repo/servers/custom-mcp
`
	tmpFile, err := os.CreateTemp("", "manifest-apm-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yaml)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	manifest, err := LoadManifest(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if manifest.Name != "test-apm" {
		t.Errorf("Name: expected 'test-apm', got '%s'", manifest.Name)
	}

	if len(manifest.Dependencies) != 3 {
		t.Errorf("Dependencies: expected 3 (2 apm + 1 mcp), got %d", len(manifest.Dependencies))
	}

	// Should include both APM and MCP dependencies
	expectedDeps := map[string]bool{
		"org/repo/agents/code-reviewer@1.0.0": true,
		"org/repo/skills/python":              true,
		"org/repo/servers/custom-mcp":         true,
	}

	for _, dep := range manifest.Dependencies {
		if !expectedDeps[dep] {
			t.Errorf("Unexpected dependency: %s", dep)
		}
	}
}

func TestManifestWrite_FlatFormat(t *testing.T) {
	manifest := &Manifest{
		Name:    "test-write",
		Version: "1.0.0",
		Dependencies: []string{
			"org/repo/agents/code-reviewer",
			"org/repo/skills/python",
		},
		Targets: []string{"opencode"},
	}

	tmpFile, err := os.CreateTemp("", "manifest-write-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	if err := manifest.Write(tmpFile.Name()); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Read it back
	loaded, err := LoadManifest(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load written manifest: %v", err)
	}

	if loaded.Name != "test-write" {
		t.Errorf("Name: expected 'test-write', got '%s'", loaded.Name)
	}

	if len(loaded.Dependencies) != 2 {
		t.Errorf("Dependencies: expected 2, got %d", len(loaded.Dependencies))
	}
}
