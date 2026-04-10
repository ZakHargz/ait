package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/sources"
	"github.com/apex-ai/ait/internal/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

var outdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "Check for outdated packages",
	Long: `Check for outdated packages by comparing installed versions with latest available versions.

This command reads ait.lock to get currently installed packages, checks the remote
repositories for the latest versions, and displays packages that have updates available.

Examples:
  # Check for outdated packages in current project
  ait outdated

  # Check outdated packages and show all packages (even up-to-date ones)
  ait outdated --all`,
	RunE: runOutdated,
}

var (
	outdatedShowAll bool
)

func init() {
	rootCmd.AddCommand(outdatedCmd)

	outdatedCmd.Flags().BoolVarP(&outdatedShowAll, "all", "a", false, "show all packages, not just outdated ones")
}

// PackageVersionInfo holds version information for a package
type PackageVersionInfo struct {
	Name           string
	CurrentVersion string
	LatestVersion  string
	IsOutdated     bool
	Error          string
	Type           string
	Source         string
}

func runOutdated(cmd *cobra.Command, args []string) error {
	// Check if we're in a project directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	manifestPath := filepath.Join(cwd, "ait.yml")
	lockfilePath := filepath.Join(cwd, "ait.lock")

	// Check if manifest exists
	if !config.ManifestExists(manifestPath) {
		utils.PrintWarning("No ait.yml found in current directory")
		utils.PrintInfo("Run 'ait init' to create a new project")
		return nil
	}

	// Check if lockfile exists
	if !config.LockFileExists(lockfilePath) {
		utils.PrintWarning("No ait.lock found in current directory")
		utils.PrintInfo("Run 'ait install' to install dependencies and create ait.lock")
		return nil
	}

	// Load lockfile
	lockfile, err := config.LoadLockFile(lockfilePath)
	if err != nil {
		return fmt.Errorf("failed to load ait.lock: %w", err)
	}

	// Check if there are any packages
	if len(lockfile.Packages) == 0 {
		utils.PrintInfo("No packages installed")
		utils.PrintInfo("Run 'ait install' to install dependencies")
		return nil
	}

	utils.PrintInfo("Checking for outdated packages...")
	fmt.Println()

	// Create git source for checking remote versions
	gitSource := sources.NewGitSource("")

	// Check each package
	var versionInfos []PackageVersionInfo
	for _, pkg := range lockfile.Packages {
		info := checkPackageVersion(pkg, gitSource)
		versionInfos = append(versionInfos, info)
	}

	// Display results
	return displayOutdatedPackages(versionInfos)
}

// checkPackageVersion checks if a package has a newer version available
func checkPackageVersion(pkg config.LockedPkg, gitSource *sources.GitSource) PackageVersionInfo {
	info := PackageVersionInfo{
		Name:           pkg.Name,
		CurrentVersion: pkg.Resolved,
		Type:           pkg.Type,
		Source:         pkg.Source,
		IsOutdated:     false,
	}

	// Parse the package source to get spec
	spec, err := sources.ParsePackageSpec(pkg.Source)
	if err != nil {
		info.Error = fmt.Sprintf("failed to parse source: %v", err)
		return info
	}

	// Skip local packages
	if spec.Type == "local" {
		info.LatestVersion = "local"
		return info
	}

	// Try to get the latest version from the remote repository
	latestVersion, err := getLatestVersion(spec, gitSource)
	if err != nil {
		info.Error = fmt.Sprintf("failed to check latest version: %v", err)
		return info
	}

	info.LatestVersion = latestVersion

	// Compare versions
	if pkg.Resolved != latestVersion {
		// Try to parse as semver to determine if it's actually newer
		currentSemver, currentErr := semver.NewVersion(strings.TrimPrefix(pkg.Resolved, "v"))
		latestSemver, latestErr := semver.NewVersion(strings.TrimPrefix(latestVersion, "v"))

		if currentErr == nil && latestErr == nil {
			// Both are valid semver - compare them
			if latestSemver.GreaterThan(currentSemver) {
				info.IsOutdated = true
			}
		} else {
			// Not semver, just mark as different
			info.IsOutdated = true
		}
	}

	return info
}

// getLatestVersion gets the latest version available for a package
func getLatestVersion(spec *sources.PackageSpec, gitSource *sources.GitSource) (string, error) {
	// Build a temporary spec with "latest-tag" version to get the newest tag
	tempSpec := *spec
	tempSpec.Version = "latest-tag"

	// Fetch the package to trigger repository update and version resolution
	// We don't need the actual package, just the version resolution
	repoURL := buildRepoURL(spec)
	repoPath := getRepoCachePath(spec, gitSource.CacheDir)

	// Ensure the repository is up-to-date
	repo, err := ensureRepo(gitSource, repoURL, repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to fetch repository: %w", err)
	}

	// Get all tags and find the latest semver tag
	tags, err := getTags(repo)
	if err != nil {
		return "", fmt.Errorf("failed to get tags: %w", err)
	}

	// Find the highest semver tag
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

	// If no semver tags found, return "latest"
	return "latest", nil
}

// Helper functions to access git source internals
// These mirror the private methods in git.go

func buildRepoURL(spec *sources.PackageSpec) string {
	switch spec.Type {
	case "github":
		return fmt.Sprintf("https://github.com/%s.git", spec.Repo)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s.git", spec.Repo)
	case "git":
		return spec.Repo
	default:
		return spec.Repo
	}
}

func getRepoCachePath(spec *sources.PackageSpec, cacheDir string) string {
	safeName := strings.ReplaceAll(spec.Repo, "/", "_")
	return filepath.Join(cacheDir, spec.Type, safeName)
}

func ensureRepo(gitSource *sources.GitSource, repoURL, repoPath string) (*git.Repository, error) {
	// Create a temporary spec to use the GitSource.Fetch method
	// This will handle cloning/updating the repository
	spec := &sources.PackageSpec{
		Type:    "github",
		Repo:    extractRepoFromURL(repoURL),
		Path:    "",
		Version: "latest-tag",
	}

	// Detect source type from URL
	if strings.Contains(repoURL, "gitlab.com") {
		spec.Type = "gitlab"
	} else if !strings.Contains(repoURL, "github.com") {
		spec.Type = "git"
	}

	// Use the internal method by creating a package fetch
	// This will ensure the repo exists and is up-to-date
	if utils.DirExists(repoPath) {
		// Repository exists, just open it
		repo, err := openRepo(repoPath)
		if err != nil {
			return nil, err
		}
		// Fetch updates
		err = fetchRepo(repo, gitSource)
		if err != nil {
			return nil, err
		}
		return repo, nil
	}

	// Clone the repository
	return cloneRepo(repoURL, repoPath, gitSource)
}

func extractRepoFromURL(repoURL string) string {
	// Remove .git suffix
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Extract org/repo from URL
	if strings.Contains(repoURL, "github.com/") {
		parts := strings.Split(repoURL, "github.com/")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	if strings.Contains(repoURL, "gitlab.com/") {
		parts := strings.Split(repoURL, "gitlab.com/")
		if len(parts) > 1 {
			return parts[1]
		}
	}

	return repoURL
}

// displayOutdatedPackages displays the results in a formatted table
func displayOutdatedPackages(infos []PackageVersionInfo) error {
	outdatedCount := 0
	upToDateCount := 0
	errorCount := 0

	// Filter based on --all flag
	displayInfos := make([]PackageVersionInfo, 0)
	for _, info := range infos {
		if info.Error != "" {
			errorCount++
			if outdatedShowAll {
				displayInfos = append(displayInfos, info)
			}
		} else if info.IsOutdated {
			outdatedCount++
			displayInfos = append(displayInfos, info)
		} else {
			upToDateCount++
			if outdatedShowAll {
				displayInfos = append(displayInfos, info)
			}
		}
	}

	// Display packages
	if len(displayInfos) > 0 {
		// Calculate column widths
		maxNameLen := 20
		maxCurrentLen := 15
		maxLatestLen := 15
		maxTypeLen := 10

		for _, info := range displayInfos {
			if len(info.Name) > maxNameLen {
				maxNameLen = len(info.Name)
			}
			if len(info.CurrentVersion) > maxCurrentLen {
				maxCurrentLen = len(info.CurrentVersion)
			}
			if len(info.LatestVersion) > maxLatestLen {
				maxLatestLen = len(info.LatestVersion)
			}
			if len(info.Type) > maxTypeLen {
				maxTypeLen = len(info.Type)
			}
		}

		// Print header
		headerFormat := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n",
			maxNameLen, maxCurrentLen, maxLatestLen, maxTypeLen)
		fmt.Printf(headerFormat, "PACKAGE", "CURRENT", "LATEST", "TYPE", "STATUS")
		fmt.Printf(headerFormat,
			strings.Repeat("-", maxNameLen),
			strings.Repeat("-", maxCurrentLen),
			strings.Repeat("-", maxLatestLen),
			strings.Repeat("-", maxTypeLen),
			strings.Repeat("-", 10))

		// Print packages
		rowFormat := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%s\n",
			maxNameLen, maxCurrentLen, maxLatestLen, maxTypeLen)

		for _, info := range displayInfos {
			status := ""
			if info.Error != "" {
				status = "error"
			} else if info.IsOutdated {
				status = "outdated"
			} else {
				status = "up-to-date"
			}

			fmt.Printf(rowFormat,
				info.Name,
				info.CurrentVersion,
				info.LatestVersion,
				info.Type,
				status)
		}
		fmt.Println()
	}

	// Print summary
	if outdatedCount > 0 {
		utils.PrintWarning("%d package(s) outdated", outdatedCount)
		utils.PrintInfo("Run 'ait update' to update packages")
	} else if len(infos) > 0 {
		utils.PrintSuccess("All packages are up-to-date!")
	}

	if errorCount > 0 {
		utils.PrintWarning("%d package(s) could not be checked", errorCount)
	}

	if upToDateCount > 0 && !outdatedShowAll {
		utils.PrintInfo("%d package(s) up-to-date (use --all to show)", upToDateCount)
	}

	return nil
}

// Wrappers for git operations - these will use go-git directly
// to avoid circular dependencies

func openRepo(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

func fetchRepo(repo *git.Repository, gitSource *sources.GitSource) error {
	err := repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Force:      true,
		Tags:       git.AllTags,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}

func cloneRepo(repoURL, repoPath string, gitSource *sources.GitSource) (*git.Repository, error) {
	if err := utils.EnsureDir(filepath.Dir(repoPath)); err != nil {
		return nil, err
	}

	return git.PlainClone(repoPath, false, &git.CloneOptions{
		URL:  repoURL,
		Tags: git.AllTags,
	})
}

func getTags(repo *git.Repository) ([]string, error) {
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
