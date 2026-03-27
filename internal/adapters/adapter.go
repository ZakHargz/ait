package adapters

import (
	"github.com/apex-ai/ait/internal/packages"
)

// Adapter is the interface that all platform adapters must implement
type Adapter interface {
	// Name returns the adapter identifier
	Name() string

	// Detect checks if this tool is installed on the system
	Detect() bool

	// GetConfigDir returns the tool's configuration directory
	GetConfigDir() (string, error)

	// InstallAgent installs an agent package
	InstallAgent(pkg *packages.Package) error

	// InstallSkill installs a skill package
	InstallSkill(pkg *packages.Package) error

	// InstallPrompt installs a prompt package
	InstallPrompt(pkg *packages.Package) error

	// Uninstall removes a package
	Uninstall(pkg *packages.Package) error

	// List returns all installed packages for this adapter
	List() ([]*packages.Package, error)

	// Validate checks if the adapter installation is healthy
	Validate() error
}

// BaseAdapter provides common functionality for adapters
type BaseAdapter struct {
	name      string
	configDir string
}

// NewBaseAdapter creates a new base adapter
func NewBaseAdapter(name, configDir string) BaseAdapter {
	return BaseAdapter{
		name:      name,
		configDir: configDir,
	}
}

// Name returns the adapter name
func (b *BaseAdapter) Name() string {
	return b.name
}

// GetConfigDir returns the configuration directory
func (b *BaseAdapter) GetConfigDir() (string, error) {
	return b.configDir, nil
}
