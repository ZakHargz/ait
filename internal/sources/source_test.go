package sources

import (
	"testing"
)

func TestParsePackageSpec_LegacyFormat(t *testing.T) {
	tests := []struct {
		name            string
		spec            string
		expectedType    string
		expectedRepo    string
		expectedPath    string
		expectedVer     string
		expectedVirtual bool
		shouldError     bool
	}{
		{
			name:            "github with path and version",
			spec:            "github:org/repo/agents/code-reviewer@1.0.0",
			expectedType:    "github",
			expectedRepo:    "org/repo",
			expectedPath:    "agents/code-reviewer",
			expectedVer:     "1.0.0",
			expectedVirtual: false,
		},
		{
			name:            "gitlab without version",
			spec:            "gitlab:org/repo/skills/python",
			expectedType:    "gitlab",
			expectedRepo:    "org/repo",
			expectedPath:    "skills/python",
			expectedVer:     "latest",
			expectedVirtual: false,
		},
		{
			name:            "local path",
			spec:            "local:./my-package",
			expectedType:    "local",
			expectedRepo:    "",
			expectedPath:    "./my-package",
			expectedVer:     "latest",
			expectedVirtual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := ParsePackageSpec(tt.spec)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if spec.Type != tt.expectedType {
				t.Errorf("Type: expected %q, got %q", tt.expectedType, spec.Type)
			}
			if spec.Repo != tt.expectedRepo {
				t.Errorf("Repo: expected %q, got %q", tt.expectedRepo, spec.Repo)
			}
			if spec.Path != tt.expectedPath {
				t.Errorf("Path: expected %q, got %q", tt.expectedPath, spec.Path)
			}
			if spec.Version != tt.expectedVer {
				t.Errorf("Version: expected %q, got %q", tt.expectedVer, spec.Version)
			}
			if spec.IsVirtualPackage != tt.expectedVirtual {
				t.Errorf("IsVirtualPackage: expected %v, got %v", tt.expectedVirtual, spec.IsVirtualPackage)
			}
		})
	}
}

func TestParsePackageSpec_APMShorthand(t *testing.T) {
	tests := []struct {
		name            string
		spec            string
		expectedType    string
		expectedRepo    string
		expectedPath    string
		expectedVer     string
		expectedVirtual bool
	}{
		{
			name:            "github shorthand with path and version",
			spec:            "org/repo/agents/code-reviewer@1.0.0",
			expectedType:    "github",
			expectedRepo:    "org/repo",
			expectedPath:    "agents/code-reviewer",
			expectedVer:     "1.0.0",
			expectedVirtual: false,
		},
		{
			name:            "github shorthand without version",
			spec:            "awesome/packages/skills/python",
			expectedType:    "github",
			expectedRepo:    "awesome/packages",
			expectedPath:    "skills/python",
			expectedVer:     "latest",
			expectedVirtual: false,
		},
		{
			name:            "gitlab FQDN format",
			spec:            "gitlab.com/myorg/myrepo/agents/helper",
			expectedType:    "gitlab",
			expectedRepo:    "myorg/myrepo",
			expectedPath:    "agents/helper",
			expectedVer:     "latest",
			expectedVirtual: false,
		},
		{
			name:            "github shorthand with virtual package",
			spec:            "org/repo/agents/reviewer.agent.md@1.0.0",
			expectedType:    "github",
			expectedRepo:    "org/repo",
			expectedPath:    "agents/reviewer.agent.md",
			expectedVer:     "1.0.0",
			expectedVirtual: true,
		},
		{
			name:            "skill virtual package",
			spec:            "org/repo/skills/python-expert.skill.md",
			expectedType:    "github",
			expectedRepo:    "org/repo",
			expectedPath:    "skills/python-expert.skill.md",
			expectedVer:     "latest",
			expectedVirtual: true,
		},
		{
			name:            "prompt virtual package",
			spec:            "org/repo/prompts/debug.prompt.md@2.1.0",
			expectedType:    "github",
			expectedRepo:    "org/repo",
			expectedPath:    "prompts/debug.prompt.md",
			expectedVer:     "2.1.0",
			expectedVirtual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := ParsePackageSpec(tt.spec)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if spec.Type != tt.expectedType {
				t.Errorf("Type: expected %q, got %q", tt.expectedType, spec.Type)
			}
			if spec.Repo != tt.expectedRepo {
				t.Errorf("Repo: expected %q, got %q", tt.expectedRepo, spec.Repo)
			}
			if spec.Path != tt.expectedPath {
				t.Errorf("Path: expected %q, got %q", tt.expectedPath, spec.Path)
			}
			if spec.Version != tt.expectedVer {
				t.Errorf("Version: expected %q, got %q", tt.expectedVer, spec.Version)
			}
			if spec.IsVirtualPackage != tt.expectedVirtual {
				t.Errorf("IsVirtualPackage: expected %v, got %v", tt.expectedVirtual, spec.IsVirtualPackage)
			}
		})
	}
}

func TestIsVirtualPackage(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"agents/reviewer.agent.md", true},
		{"skills/python.skill.md", true},
		{"prompts/debug.prompt.md", true},
		{"instructions/guide.instructions.md", true},
		{"chatmodes/dev.chatmode.md", true},
		{"agents/code-reviewer", false},
		{"skills/python", false},
		{"README.md", false},
		{"package.yml", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isVirtualPackage(tt.path)
			if result != tt.expected {
				t.Errorf("isVirtualPackage(%q): expected %v, got %v", tt.path, tt.expected, result)
			}
		})
	}
}
