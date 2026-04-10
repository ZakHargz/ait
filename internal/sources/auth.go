package sources

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// AuthResolver handles authentication for git operations
type AuthResolver struct {
	// Cache for resolved auth per (host, org) pair
	cache map[string]transport.AuthMethod
}

// NewAuthResolver creates a new AuthResolver
func NewAuthResolver() *AuthResolver {
	return &AuthResolver{
		cache: make(map[string]transport.AuthMethod),
	}
}

// GetAuth resolves authentication for a git repository URL
// Follows APM's resolution chain:
// 1. Per-org env var (GITHUB_APM_PAT_{ORG})
// 2. Global env vars (GITHUB_APM_PAT -> GITHUB_TOKEN -> GH_TOKEN)
// 3. Git credential helper (fallback)
func (ar *AuthResolver) GetAuth(repoURL string) transport.AuthMethod {
	// Parse the URL to extract host and org
	host, org, err := parseRepoURL(repoURL)
	if err != nil {
		// If we can't parse, try global token
		return ar.getGlobalAuth()
	}

	// Check cache
	cacheKey := fmt.Sprintf("%s:%s", host, org)
	if auth, ok := ar.cache[cacheKey]; ok {
		return auth
	}

	// Try resolution chain
	var auth transport.AuthMethod

	// 1. Try per-org token (only for GitHub-like hosts)
	if isGitHubLikeHost(host) && org != "" {
		auth = ar.getPerOrgAuth(org)
	}

	// 2. Try global tokens
	if auth == nil {
		auth = ar.getGlobalAuth()
	}

	// Cache the result (even if nil)
	ar.cache[cacheKey] = auth

	return auth
}

// getPerOrgAuth attempts to get a per-org token
// Converts org name to env var name: contoso-microsoft -> GITHUB_APM_PAT_CONTOSO_MICROSOFT
func (ar *AuthResolver) getPerOrgAuth(org string) transport.AuthMethod {
	// Normalize org name: uppercase, replace hyphens with underscores
	normalizedOrg := strings.ToUpper(strings.ReplaceAll(org, "-", "_"))
	envVar := fmt.Sprintf("GITHUB_APM_PAT_%s", normalizedOrg)

	token := os.Getenv(envVar)
	if token != "" {
		return &http.BasicAuth{
			Username: "x-access-token", // GitHub standard for PAT
			Password: token,
		}
	}

	return nil
}

// getGlobalAuth attempts to get a global token
// Priority: GITHUB_APM_PAT -> GITHUB_TOKEN -> GH_TOKEN
func (ar *AuthResolver) getGlobalAuth() transport.AuthMethod {
	// Try in priority order
	envVars := []string{"GITHUB_APM_PAT", "GITHUB_TOKEN", "GH_TOKEN"}

	for _, envVar := range envVars {
		token := os.Getenv(envVar)
		if token != "" {
			return &http.BasicAuth{
				Username: "x-access-token", // GitHub standard for PAT
				Password: token,
			}
		}
	}

	return nil
}

// parseRepoURL extracts host and org from a git repository URL
// Supports:
//   - https://github.com/org/repo.git -> github.com, org
//   - https://gitlab.com/org/repo.git -> gitlab.com, org
//   - git@github.com:org/repo.git -> github.com, org
//   - https://github.company.com/org/repo.git -> github.company.com, org
func parseRepoURL(repoURL string) (host, org string, err error) {
	// Handle SSH URLs (git@host:org/repo.git)
	if strings.HasPrefix(repoURL, "git@") {
		parts := strings.SplitN(strings.TrimPrefix(repoURL, "git@"), ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH URL format")
		}
		host = parts[0]

		// Extract org from path (org/repo.git)
		pathParts := strings.Split(parts[1], "/")
		if len(pathParts) > 0 {
			org = pathParts[0]
		}
		return host, org, nil
	}

	// Handle HTTPS URLs
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", err
	}

	host = u.Host

	// Extract org from path (/org/repo.git)
	path := strings.TrimPrefix(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	pathParts := strings.Split(path, "/")

	if len(pathParts) > 0 {
		org = pathParts[0]
	}

	return host, org, nil
}

// isGitHubLikeHost checks if a host is GitHub or GitHub Enterprise
// GitHub-like hosts support per-org tokens
func isGitHubLikeHost(host string) bool {
	// github.com
	if host == "github.com" {
		return true
	}

	// GitHub Enterprise Cloud (*.ghe.com)
	if strings.HasSuffix(host, ".ghe.com") {
		return true
	}

	// Check for GITHUB_HOST env var (GitHub Enterprise Server)
	ghHost := os.Getenv("GITHUB_HOST")
	if ghHost != "" && host == ghHost {
		return true
	}

	return false
}

// ClearCache clears the authentication cache
// Useful for testing or when tokens change
func (ar *AuthResolver) ClearCache() {
	ar.cache = make(map[string]transport.AuthMethod)
}
