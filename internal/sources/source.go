package sources

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex-ai/ait/internal/packages"
)

// Source represents a package source (git repository, local path, etc.)
type Source interface {
	// Fetch downloads the package from the source to the cache directory
	Fetch(spec PackageSpec) (*packages.Package, error)

	// GetCachePath returns the local cache path for a package
	GetCachePath(spec PackageSpec) string

	// IsCached checks if a package version is already in the cache
	IsCached(spec PackageSpec) bool
}

// PackageSpec represents a parsed package specification
// Examples:
//
//	github:org/repo/path/to/package@1.0.0
//	github:org/repo/agents/code-reviewer@^1.2.0
//	gitlab:org/repo/skills/python@~2.0.0
type PackageSpec struct {
	// Source type (github, gitlab, git, local)
	Type string

	// Repository path (e.g., "org/repo")
	Repo string

	// Path within repository (e.g., "agents/code-reviewer")
	Path string

	// Version constraint (e.g., "1.0.0", "^1.2.0", "~2.0.0", "main")
	Version string

	// Original spec string
	Original string
}

// ParsePackageSpec parses a package specification string
// Supported formats:
//
//	github:org/repo/path/to/package@version
//	gitlab:org/repo/path/to/package@version
//	git:https://git.example.com/repo/path/to/package@version
//	local:./path/to/package
func ParsePackageSpec(spec string) (*PackageSpec, error) {
	// Split by colon to get source type
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid package spec format (expected 'type:path@version'): %s", spec)
	}

	sourceType := parts[0]
	remainder := parts[1]

	// Split by @ to get version
	pathParts := strings.SplitN(remainder, "@", 2)
	if len(pathParts) != 2 {
		return nil, fmt.Errorf("invalid package spec format (missing @version): %s", spec)
	}

	pathStr := pathParts[0]
	version := pathParts[1]

	// For local paths, no need to split further
	if sourceType == "local" {
		return &PackageSpec{
			Type:     sourceType,
			Path:     pathStr,
			Version:  version,
			Original: spec,
		}, nil
	}

	// For git sources, split path into repo and package path
	// Format: org/repo/path/to/package
	pathComponents := strings.SplitN(pathStr, "/", 3)
	if len(pathComponents) < 2 {
		return nil, fmt.Errorf("invalid package spec format (expected at least org/repo): %s", spec)
	}

	repo := filepath.Join(pathComponents[0], pathComponents[1])
	pkgPath := ""
	if len(pathComponents) > 2 {
		pkgPath = pathComponents[2]
	}

	return &PackageSpec{
		Type:     sourceType,
		Repo:     repo,
		Path:     pkgPath,
		Version:  version,
		Original: spec,
	}, nil
}

// String returns the string representation of the package spec
func (ps *PackageSpec) String() string {
	return ps.Original
}

// GetPackageName returns the package name from the spec
// This is typically the last component of the path
func (ps *PackageSpec) GetPackageName() string {
	if ps.Path == "" {
		// Use repo name if no path
		parts := strings.Split(ps.Repo, "/")
		return parts[len(parts)-1]
	}

	// Get last component of path
	parts := strings.Split(ps.Path, "/")
	return parts[len(parts)-1]
}

// GetFullPath returns the full path (repo + package path)
func (ps *PackageSpec) GetFullPath() string {
	if ps.Path == "" {
		return ps.Repo
	}
	return filepath.Join(ps.Repo, ps.Path)
}
