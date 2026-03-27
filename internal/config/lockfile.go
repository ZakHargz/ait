package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LockFile represents the ait.lock file
type LockFile struct {
	Version   string               `yaml:"version"`
	Generated time.Time            `yaml:"generated"`
	Packages  map[string]LockedPkg `yaml:"packages"`
}

// LockedPkg represents a locked package with resolved version
type LockedPkg struct {
	Name      string   `yaml:"name"`
	Version   string   `yaml:"version"`
	Type      string   `yaml:"type"`
	Source    string   `yaml:"source"`
	Resolved  string   `yaml:"resolved"`            // Resolved version (exact tag/commit)
	Integrity string   `yaml:"integrity,omitempty"` // Future: checksum
	Installed []string `yaml:"installed"`           // List of tools it was installed to
}

// NewLockFile creates a new lock file
func NewLockFile() *LockFile {
	return &LockFile{
		Version:   "1.0",
		Generated: time.Now(),
		Packages:  make(map[string]LockedPkg),
	}
}

// LoadLockFile loads ait.lock from the specified path
func LoadLockFile(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	var lockFile LockFile
	if err := yaml.Unmarshal(data, &lockFile); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	return &lockFile, nil
}

// Write saves the lock file to the specified path
func (l *LockFile) Write(path string) error {
	l.Generated = time.Now()

	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	return nil
}

// AddPackage adds or updates a package in the lock file
func (l *LockFile) AddPackage(name, version, pkgType, source, resolved string, installedTo []string) {
	l.Packages[name] = LockedPkg{
		Name:      name,
		Version:   version,
		Type:      pkgType,
		Source:    source,
		Resolved:  resolved,
		Installed: installedTo,
	}
}

// GetPackage retrieves a locked package by name
func (l *LockFile) GetPackage(name string) (LockedPkg, bool) {
	pkg, ok := l.Packages[name]
	return pkg, ok
}

// LockFileExists checks if a lock file exists at the given path
func LockFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
