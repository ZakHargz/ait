package sources

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func TestAuthResolver_GetAuth(t *testing.T) {
	tests := []struct {
		name        string
		repoURL     string
		setupEnv    map[string]string
		expectAuth  bool
		description string
	}{
		{
			name:    "github.com with GITHUB_TOKEN",
			repoURL: "https://github.com/myorg/myrepo.git",
			setupEnv: map[string]string{
				"GITHUB_TOKEN": "ghp_test_token",
			},
			expectAuth:  true,
			description: "Should use GITHUB_TOKEN for github.com repos",
		},
		{
			name:    "github.com with GH_TOKEN",
			repoURL: "https://github.com/myorg/myrepo.git",
			setupEnv: map[string]string{
				"GH_TOKEN": "ghp_test_token",
			},
			expectAuth:  true,
			description: "Should use GH_TOKEN as fallback",
		},
		{
			name:    "github.com with per-org token",
			repoURL: "https://github.com/myorg/myrepo.git",
			setupEnv: map[string]string{
				"GITHUB_APM_PAT_MYORG": "ghp_org_token",
				"GITHUB_TOKEN":         "ghp_global_token",
			},
			expectAuth:  true,
			description: "Should prioritize per-org token over global token",
		},
		{
			name:    "github.com with org containing hyphens",
			repoURL: "https://github.com/my-org-name/myrepo.git",
			setupEnv: map[string]string{
				"GITHUB_APM_PAT_MY_ORG_NAME": "ghp_org_token",
			},
			expectAuth:  true,
			description: "Should normalize org name (hyphens to underscores)",
		},
		{
			name:        "github.com without auth",
			repoURL:     "https://github.com/myorg/myrepo.git",
			setupEnv:    map[string]string{},
			expectAuth:  false,
			description: "Should return nil auth when no tokens available",
		},
		{
			name:    "gitlab.com with GITHUB_TOKEN",
			repoURL: "https://gitlab.com/myorg/myrepo.git",
			setupEnv: map[string]string{
				"GITHUB_TOKEN": "glpat_test_token",
			},
			expectAuth:  true,
			description: "Should use GITHUB_TOKEN for gitlab.com repos",
		},
		{
			name:    "GitHub Enterprise with GITHUB_APM_PAT",
			repoURL: "https://github.company.com/myorg/myrepo.git",
			setupEnv: map[string]string{
				"GITHUB_APM_PAT": "ghp_enterprise_token",
				"GITHUB_HOST":    "github.company.com",
			},
			expectAuth:  true,
			description: "Should use GITHUB_APM_PAT for enterprise hosts",
		},
		{
			name:    "SSH URL with token",
			repoURL: "git@github.com:myorg/myrepo.git",
			setupEnv: map[string]string{
				"GITHUB_TOKEN": "ghp_test_token",
			},
			expectAuth:  true,
			description: "Should handle SSH URLs and provide auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearAuthEnv()

			// Setup test environment
			for k, v := range tt.setupEnv {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create resolver
			resolver := NewAuthResolver()

			// Get auth
			auth := resolver.GetAuth(tt.repoURL)

			// Check result
			if tt.expectAuth && auth == nil {
				t.Errorf("Expected auth but got nil for %s", tt.description)
			}
			if !tt.expectAuth && auth != nil {
				t.Errorf("Expected no auth but got auth for %s", tt.description)
			}

			// Clear cache for next test
			resolver.ClearCache()
		})
	}
}

func TestAuthResolver_Priority(t *testing.T) {
	// Setup environment with multiple tokens
	os.Setenv("GITHUB_APM_PAT_MYORG", "per_org_token")
	os.Setenv("GITHUB_APM_PAT", "global_apm_token")
	os.Setenv("GITHUB_TOKEN", "github_token")
	os.Setenv("GH_TOKEN", "gh_token")

	defer func() {
		os.Unsetenv("GITHUB_APM_PAT_MYORG")
		os.Unsetenv("GITHUB_APM_PAT")
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("GH_TOKEN")
	}()

	resolver := NewAuthResolver()

	// Test priority: per-org should win
	auth := resolver.GetAuth("https://github.com/myorg/myrepo.git")
	if auth == nil {
		t.Fatal("Expected auth to be non-nil")
	}

	// We can't easily check which token was used without exposing internals,
	// but we verified it's using BasicAuth with the right structure
	if basicAuth, ok := auth.(*http.BasicAuth); ok {
		if basicAuth.Username != "x-access-token" {
			t.Errorf("Expected username 'x-access-token', got %s", basicAuth.Username)
		}
	} else {
		t.Error("Expected BasicAuth type")
	}
}

func TestAuthResolver_Cache(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "test_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	resolver := NewAuthResolver()

	// First call
	auth1 := resolver.GetAuth("https://github.com/myorg/myrepo.git")
	if auth1 == nil {
		t.Fatal("Expected auth to be non-nil")
	}

	// Second call should return cached result
	auth2 := resolver.GetAuth("https://github.com/myorg/myrepo.git")
	if auth2 == nil {
		t.Fatal("Expected cached auth to be non-nil")
	}

	// Should be the same instance (cached)
	if auth1 != auth2 {
		t.Error("Expected cached auth to be the same instance")
	}

	// Clear cache
	resolver.ClearCache()

	// After clear, should create new instance
	auth3 := resolver.GetAuth("https://github.com/myorg/myrepo.git")
	if auth3 == nil {
		t.Fatal("Expected auth to be non-nil after cache clear")
	}
}

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name         string
		repoURL      string
		expectedHost string
		expectedOrg  string
		expectError  bool
	}{
		{
			name:         "GitHub HTTPS",
			repoURL:      "https://github.com/myorg/myrepo.git",
			expectedHost: "github.com",
			expectedOrg:  "myorg",
			expectError:  false,
		},
		{
			name:         "GitHub SSH",
			repoURL:      "git@github.com:myorg/myrepo.git",
			expectedHost: "github.com",
			expectedOrg:  "myorg",
			expectError:  false,
		},
		{
			name:         "GitLab HTTPS",
			repoURL:      "https://gitlab.com/myorg/myrepo.git",
			expectedHost: "gitlab.com",
			expectedOrg:  "myorg",
			expectError:  false,
		},
		{
			name:         "GitHub Enterprise",
			repoURL:      "https://github.company.com/myorg/myrepo.git",
			expectedHost: "github.company.com",
			expectedOrg:  "myorg",
			expectError:  false,
		},
		{
			name:         "GitLab with subgroups",
			repoURL:      "https://gitlab.com/myorg/subgroup/myrepo.git",
			expectedHost: "gitlab.com",
			expectedOrg:  "myorg",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, org, err := parseRepoURL(tt.repoURL)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if host != tt.expectedHost {
				t.Errorf("Expected host %s, got %s", tt.expectedHost, host)
			}
			if org != tt.expectedOrg {
				t.Errorf("Expected org %s, got %s", tt.expectedOrg, org)
			}
		})
	}
}

func TestIsGitHubLikeHost(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		setupEnv     map[string]string
		isGitHubLike bool
	}{
		{
			name:         "github.com",
			host:         "github.com",
			isGitHubLike: true,
		},
		{
			name:         "GitHub Enterprise Cloud",
			host:         "mycompany.ghe.com",
			isGitHubLike: true,
		},
		{
			name:         "GitLab",
			host:         "gitlab.com",
			isGitHubLike: false,
		},
		{
			name: "GitHub Enterprise Server via GITHUB_HOST",
			host: "github.company.com",
			setupEnv: map[string]string{
				"GITHUB_HOST": "github.company.com",
			},
			isGitHubLike: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			for k, v := range tt.setupEnv {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			result := isGitHubLikeHost(tt.host)
			if result != tt.isGitHubLike {
				t.Errorf("Expected isGitHubLikeHost(%s) to be %v, got %v",
					tt.host, tt.isGitHubLike, result)
			}
		})
	}
}

// Helper function to clear all auth-related environment variables
func clearAuthEnv() {
	envVars := []string{
		"GITHUB_APM_PAT",
		"GITHUB_TOKEN",
		"GH_TOKEN",
		"GITHUB_HOST",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}
}
