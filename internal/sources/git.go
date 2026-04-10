package sources

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitSource handles fetching packages from Git repositories
type GitSource struct {
	// CacheDir is the root directory for cached repositories
	CacheDir string
	// AuthResolver handles authentication for git operations
	authResolver *AuthResolver
}

// NewGitSource creates a new GitSource with the specified cache directory
func NewGitSource(cacheDir string) *GitSource {
	if cacheDir == "" {
		// Default to ~/.ait/cache
		home := utils.HomeDir()
		cacheDir = filepath.Join(home, ".ait", "cache")
	}

	return &GitSource{
		CacheDir:     cacheDir,
		authResolver: NewAuthResolver(),
	}
}

// Fetch downloads the package from the git repository
func (gs *GitSource) Fetch(spec PackageSpec) (*packages.Package, error) {
	// Build repository URL based on source type
	repoURL := gs.buildRepoURL(spec)

	// Get cache path for this repository
	repoPath := gs.getRepoCachePath(spec)

	// Clone or update repository
	repo, err := gs.ensureRepo(repoURL, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	// Resolve version to a specific tag/commit
	ref, err := gs.resolveVersion(repo, spec.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve version %s: %w", spec.Version, err)
	}

	// Checkout the specific version
	if err := gs.checkout(repo, ref); err != nil {
		return nil, fmt.Errorf("failed to checkout version: %w", err)
	}

	// Get package path within repository
	pkgPath := filepath.Join(repoPath, spec.Path)

	// Read package metadata
	metadataPath := filepath.Join(pkgPath, "package.yml")
	if !utils.FileExists(metadataPath) {
		return nil, fmt.Errorf("package.yml not found at %s", metadataPath)
	}

	metadata, err := config.LoadPackageMetadata(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load package metadata: %w", err)
	}

	// Create package struct
	pkg := &packages.Package{
		Name:     metadata.Name,
		Version:  metadata.Version,
		Type:     metadata.Type,
		Path:     pkgPath,
		Metadata: metadata,
	}

	return pkg, nil
}

// GetCachePath returns the local cache path for a package
func (gs *GitSource) GetCachePath(spec PackageSpec) string {
	return filepath.Join(gs.getRepoCachePath(spec), spec.Path)
}

// IsCached checks if a package version is already in the cache
func (gs *GitSource) IsCached(spec PackageSpec) bool {
	repoPath := gs.getRepoCachePath(spec)
	if !utils.DirExists(repoPath) {
		return false
	}

	// Check if package directory exists
	pkgPath := filepath.Join(repoPath, spec.Path)
	return utils.DirExists(pkgPath)
}

// buildRepoURL constructs the git repository URL from the spec
func (gs *GitSource) buildRepoURL(spec PackageSpec) string {
	switch spec.Type {
	case "github":
		return fmt.Sprintf("https://github.com/%s.git", spec.Repo)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s.git", spec.Repo)
	case "git":
		// For generic git sources, the Repo field should contain the full URL
		return spec.Repo
	default:
		return spec.Repo
	}
}

// getRepoCachePath returns the local cache directory for a repository
func (gs *GitSource) getRepoCachePath(spec PackageSpec) string {
	// Create a safe directory name from the repo
	// Replace slashes with underscores: org/repo -> org_repo
	safeName := strings.ReplaceAll(spec.Repo, "/", "_")
	return filepath.Join(gs.CacheDir, spec.Type, safeName)
}

// ensureRepo clones the repository if it doesn't exist, or fetches updates if it does
func (gs *GitSource) ensureRepo(repoURL, repoPath string) (*git.Repository, error) {
	// Get authentication for this repository
	auth := gs.authResolver.GetAuth(repoURL)

	// Check if repository already exists
	if utils.DirExists(repoPath) {
		// Open existing repository
		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open repository: %w", err)
		}

		// Fetch latest changes
		utils.PrintInfo("Updating repository cache...")
		err = repo.Fetch(&git.FetchOptions{
			RemoteName: "origin",
			Auth:       auth,
			Force:      true,
			Tags:       git.AllTags,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return nil, fmt.Errorf("failed to fetch updates: %w", err)
		}

		return repo, nil
	}

	// Clone repository
	utils.PrintInfo("Cloning repository from %s...", repoURL)

	// Ensure cache directory exists
	if err := utils.EnsureDir(filepath.Dir(repoPath)); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	repo, err := git.PlainClone(repoPath, false, &git.CloneOptions{
		URL:      repoURL,
		Auth:     auth,
		Progress: nil, // Could add progress bar here
		Tags:     git.AllTags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return repo, nil
}

// resolveVersion resolves a version constraint to a specific git reference
// Supports:
//   - "latest" -> HEAD of default branch (master or main)
//   - "latest-tag" -> highest semver tag
//   - Exact versions: "1.0.0" -> tag "v1.0.0"
//   - Semver constraints: "^1.0.0", "~1.2.0" -> matching tags
//   - Branch names: "main", "develop"
//   - Commit hashes: "abc123..."
func (gs *GitSource) resolveVersion(repo *git.Repository, versionSpec string) (string, error) {
	// Handle "latest" - use HEAD of default branch (master or main)
	if versionSpec == "latest" {
		branches, err := repo.Branches()
		if err == nil {
			var foundBranch string
			branches.ForEach(func(ref *plumbing.Reference) error {
				branchName := ref.Name().Short()
				// Prefer main over master if both exist
				if branchName == "main" {
					foundBranch = "main"
					return nil
				}
				if branchName == "master" && foundBranch == "" {
					foundBranch = "master"
				}
				return nil
			})
			if foundBranch != "" {
				return foundBranch, nil
			}
		}
		// Fallback to master
		return "master", nil
	}

	// Handle "latest-tag" - find the highest semver tag
	if versionSpec == "latest-tag" {
		tags, err := gs.getTags(repo)
		if err != nil {
			return "", fmt.Errorf("failed to list tags: %w", err)
		}

		var latestVersion *semver.Version
		var latestTag string

		for _, tag := range tags {
			// Try to parse tag as semver (strip 'v' prefix if present)
			tagName := strings.TrimPrefix(tag, "v")
			v, err := semver.NewVersion(tagName)
			if err != nil {
				continue // Skip non-semver tags
			}

			if latestVersion == nil || v.GreaterThan(latestVersion) {
				latestVersion = v
				latestTag = tag
			}
		}

		if latestVersion != nil {
			return latestTag, nil
		}

		// If no semver tags found, fallback to "latest" behavior
		return gs.resolveVersion(repo, "latest")
	}

	// If it looks like a commit hash, use it directly
	if len(versionSpec) == 40 || len(versionSpec) == 7 {
		// Verify it exists
		_, err := repo.CommitObject(plumbing.NewHash(versionSpec))
		if err == nil {
			return versionSpec, nil
		}
	}

	// Try to parse as semantic version constraint
	if strings.HasPrefix(versionSpec, "^") || strings.HasPrefix(versionSpec, "~") ||
		strings.HasPrefix(versionSpec, ">=") || strings.HasPrefix(versionSpec, "<=") ||
		strings.Contains(versionSpec, ".") {

		constraint, err := semver.NewConstraint(versionSpec)
		if err == nil {
			// Find matching tags
			tags, err := gs.getTags(repo)
			if err != nil {
				return "", fmt.Errorf("failed to list tags: %w", err)
			}

			// Find the highest version that matches constraint
			var matchedVersion *semver.Version
			for _, tag := range tags {
				// Try to parse tag as semver (strip 'v' prefix if present)
				tagName := strings.TrimPrefix(tag, "v")
				v, err := semver.NewVersion(tagName)
				if err != nil {
					continue // Skip non-semver tags
				}

				if constraint.Check(v) {
					if matchedVersion == nil || v.GreaterThan(matchedVersion) {
						matchedVersion = v
					}
				}
			}

			if matchedVersion != nil {
				// Return the tag name (with 'v' prefix if original had it)
				for _, tag := range tags {
					if strings.TrimPrefix(tag, "v") == matchedVersion.String() {
						return tag, nil
					}
				}
				return "v" + matchedVersion.String(), nil
			}
		}
	}

	// Try as exact tag name (with or without 'v' prefix)
	tags, err := gs.getTags(repo)
	if err != nil {
		return "", fmt.Errorf("failed to list tags: %w", err)
	}

	// Try exact match
	for _, tag := range tags {
		if tag == versionSpec || tag == "v"+versionSpec {
			return tag, nil
		}
	}

	// Try as branch name
	branches, err := repo.Branches()
	if err != nil {
		return "", fmt.Errorf("failed to list branches: %w", err)
	}

	var branchFound bool
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == versionSpec {
			branchFound = true
			return nil
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to iterate branches: %w", err)
	}

	if branchFound {
		return versionSpec, nil
	}

	return "", fmt.Errorf("could not resolve version: %s", versionSpec)
}

// getTags returns all tags in the repository
func (gs *GitSource) getTags(repo *git.Repository) ([]string, error) {
	tags, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	var tagNames []string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagNames = append(tagNames, ref.Name().Short())
		return nil
	})

	return tagNames, err
}

// checkout checks out a specific version (tag, branch, or commit)
func (gs *GitSource) checkout(repo *git.Repository, ref string) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Try to resolve as tag first
	tagRef, err := repo.Tag(ref)
	if err == nil {
		return w.Checkout(&git.CheckoutOptions{
			Branch: tagRef.Name(),
		})
	}

	// Try as remote branch (origin/ref) first to avoid detached HEAD
	remoteBranchRef := plumbing.NewRemoteReferenceName("origin", ref)
	remoteRef, err := repo.Reference(remoteBranchRef, true)
	if err == nil {
		// Create/update local branch to track remote
		localBranchRef := plumbing.NewBranchReferenceName(ref)

		// Check if local branch exists
		_, err := repo.Reference(localBranchRef, false)
		if err != nil {
			// Create new local branch tracking remote
			newRef := plumbing.NewHashReference(localBranchRef, remoteRef.Hash())
			err = repo.Storer.SetReference(newRef)
			if err != nil {
				return fmt.Errorf("failed to create local branch: %w", err)
			}
		}

		// Checkout the local branch (this will track the remote)
		err = w.Checkout(&git.CheckoutOptions{
			Branch: localBranchRef,
		})
		if err == nil {
			// Update to remote HEAD
			return w.Reset(&git.ResetOptions{
				Commit: remoteRef.Hash(),
				Mode:   git.HardReset,
			})
		}
	}

	// Try as local branch
	branchRef := plumbing.NewBranchReferenceName(ref)
	_, err = repo.Reference(branchRef, false)
	if err == nil {
		return w.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
		})
	}

	// Try as commit hash
	hash := plumbing.NewHash(ref)
	return w.Checkout(&git.CheckoutOptions{
		Hash: hash,
	})
}

// LocalSource handles packages from local filesystem
type LocalSource struct{}

// NewLocalSource creates a new LocalSource
func NewLocalSource() *LocalSource {
	return &LocalSource{}
}

// Fetch reads a package from the local filesystem
func (ls *LocalSource) Fetch(spec PackageSpec) (*packages.Package, error) {
	pkgPath := spec.Path

	// Expand ~ to home directory
	if strings.HasPrefix(pkgPath, "~") {
		pkgPath = filepath.Join(utils.HomeDir(), pkgPath[1:])
	}

	// Make absolute if relative
	if !filepath.IsAbs(pkgPath) {
		var err error
		pkgPath, err = filepath.Abs(pkgPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve path: %w", err)
		}
	}

	// Check if path exists
	if !utils.DirExists(pkgPath) {
		return nil, fmt.Errorf("package path does not exist: %s", pkgPath)
	}

	// Read package metadata
	metadataPath := filepath.Join(pkgPath, "package.yml")
	if !utils.FileExists(metadataPath) {
		return nil, fmt.Errorf("package.yml not found at %s", metadataPath)
	}

	metadata, err := config.LoadPackageMetadata(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load package metadata: %w", err)
	}

	// Create package struct
	pkg := &packages.Package{
		Name:     metadata.Name,
		Version:  metadata.Version,
		Type:     metadata.Type,
		Path:     pkgPath,
		Metadata: metadata,
	}

	return pkg, nil
}

// GetCachePath returns the path (local sources don't use cache)
func (ls *LocalSource) GetCachePath(spec PackageSpec) string {
	return spec.Path
}

// IsCached returns true for local sources (always available)
func (ls *LocalSource) IsCached(spec PackageSpec) bool {
	return utils.DirExists(spec.Path)
}
