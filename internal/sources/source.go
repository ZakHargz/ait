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
//	org/repo/path/to/package@1.0.0              (APM shorthand, defaults to GitHub)
//	gitlab.com/org/repo/skills/python@~2.0.0    (FQDN host format)
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

	// IsVirtualPackage indicates if this is a single file package (.agent.md, .skill.md, .prompt.md)
	IsVirtualPackage bool
}

// ParsePackageSpec parses a package specification string
// Supported formats:
//
//	github:org/repo/path/to/package@version
//	github:org/repo/path/to/package           (version defaults to "latest")
//	gitlab:org/repo/path/to/package@version
//	git:https://git.example.com/repo/path/to/package@version
//	local:./path/to/package@version
//	local:./path/to/package
//	org/repo/path/to/package@version          (APM shorthand, defaults to GitHub)
//	gitlab.com/org/repo/path/to/package       (FQDN format for non-GitHub hosts)
//	org/repo/agents/reviewer.agent.md@1.0.0   (Virtual package - single file)
func ParsePackageSpec(spec string) (*PackageSpec, error) {
	originalSpec := spec
	var sourceType string
	var remainder string

	// Check if spec has explicit source type prefix (contains : but not ://)
	colonIndex := strings.Index(spec, ":")
	hasExplicitType := colonIndex > 0 && !strings.HasPrefix(spec[colonIndex:], "://")

	if hasExplicitType {
		// Format: type:path[@version]
		parts := strings.SplitN(spec, ":", 2)
		sourceType = parts[0]
		remainder = parts[1]
	} else {
		// APM shorthand format: org/repo/path[@version] or host.com/org/repo/path[@version]
		// Check if it starts with a known host FQDN
		if strings.HasPrefix(spec, "gitlab.com/") {
			sourceType = "gitlab"
			remainder = strings.TrimPrefix(spec, "gitlab.com/")
		} else if strings.HasPrefix(spec, "bitbucket.org/") {
			sourceType = "git"
			remainder = spec // Keep full URL for git source
		} else if strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "/") || strings.HasPrefix(spec, "~/") {
			// Local path without prefix
			sourceType = "local"
			remainder = spec
		} else {
			// Default to GitHub for shorthand format
			sourceType = "github"
			remainder = spec
		}
	}

	// Split by @ to get version (optional)
	pathParts := strings.SplitN(remainder, "@", 2)
	pathStr := pathParts[0]

	// Default to "latest" if no version specified
	version := "latest"
	if len(pathParts) == 2 {
		version = pathParts[1]
	}

	// Check if this is a virtual package (single file)
	isVirtual := isVirtualPackage(pathStr)

	// For local paths, no need to split further
	if sourceType == "local" {
		return &PackageSpec{
			Type:             sourceType,
			Path:             pathStr,
			Version:          version,
			Original:         originalSpec,
			IsVirtualPackage: isVirtual,
		}, nil
	}

	// For git sources, split path into repo and package path
	// Format: org/repo/path/to/package
	pathComponents := strings.SplitN(pathStr, "/", 3)
	if len(pathComponents) < 2 {
		return nil, fmt.Errorf("invalid package spec format (expected at least org/repo): %s", originalSpec)
	}

	repo := filepath.Join(pathComponents[0], pathComponents[1])
	pkgPath := ""
	if len(pathComponents) > 2 {
		pkgPath = pathComponents[2]
	}

	return &PackageSpec{
		Type:             sourceType,
		Repo:             repo,
		Path:             pkgPath,
		Version:          version,
		Original:         originalSpec,
		IsVirtualPackage: isVirtual,
	}, nil
}

// isVirtualPackage checks if the path represents a virtual package (single file)
// Virtual packages end with specific extensions like .agent.md, .skill.md, .prompt.md
func isVirtualPackage(path string) bool {
	virtualExtensions := []string{
		".agent.md",
		".skill.md",
		".prompt.md",
		".instructions.md",
		".chatmode.md",
	}

	for _, ext := range virtualExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
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
