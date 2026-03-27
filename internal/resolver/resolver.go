package resolver

import (
	"fmt"

	"github.com/apex-ai/ait/internal/config"
	"github.com/apex-ai/ait/internal/packages"
	"github.com/apex-ai/ait/internal/sources"
)

// Resolver handles dependency resolution
type Resolver struct {
	resolved map[string]*packages.Package
	visited  map[string]bool
}

// NewResolver creates a new dependency resolver
func NewResolver() *Resolver {
	return &Resolver{
		resolved: make(map[string]*packages.Package),
		visited:  make(map[string]bool),
	}
}

// Resolve resolves all dependencies recursively
func (r *Resolver) Resolve(specs []string) ([]*packages.Package, error) {
	result := []*packages.Package{}

	for _, spec := range specs {
		pkgs, err := r.resolveSpec(spec)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", spec, err)
		}
		result = append(result, pkgs...)
	}

	return result, nil
}

// resolveSpec resolves a single package spec and its dependencies
func (r *Resolver) resolveSpec(specStr string) ([]*packages.Package, error) {
	// Check if already resolved
	if r.visited[specStr] {
		// Check for cycles
		if pkg, ok := r.resolved[specStr]; ok {
			return []*packages.Package{pkg}, nil
		}
		return nil, fmt.Errorf("circular dependency detected: %s", specStr)
	}

	// Mark as visiting
	r.visited[specStr] = true

	// Parse spec
	spec, err := sources.ParsePackageSpec(specStr)
	if err != nil {
		return nil, fmt.Errorf("invalid package spec: %w", err)
	}

	// Get source
	source, err := sources.GetSource(*spec)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	// Fetch package
	pkg, err := source.Fetch(*spec)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package: %w", err)
	}

	// Mark as resolved
	r.resolved[specStr] = pkg

	// Collect results starting with dependencies
	result := []*packages.Package{}

	// Resolve dependencies recursively
	if pkg.Metadata != nil && pkg.Metadata.Dependencies != nil {
		deps := collectDependencies(pkg.Metadata.Dependencies)
		for _, depSpec := range deps {
			depPkgs, err := r.resolveSpec(depSpec)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve dependency %s of %s: %w", depSpec, pkg.Name, err)
			}
			result = append(result, depPkgs...)
		}
	}

	// Add this package after its dependencies
	result = append(result, pkg)

	return result, nil
}

// collectDependencies collects all dependency specs from a Dependencies struct
func collectDependencies(deps *config.Dependencies) []string {
	if deps == nil {
		return []string{}
	}

	allDeps := []string{}
	allDeps = append(allDeps, deps.Agents...)
	allDeps = append(allDeps, deps.Skills...)
	allDeps = append(allDeps, deps.Prompts...)
	allDeps = append(allDeps, deps.MCP...)

	return allDeps
}

// GetResolutionOrder returns packages in topological order (dependencies first)
func (r *Resolver) GetResolutionOrder() []*packages.Package {
	result := []*packages.Package{}
	for _, pkg := range r.resolved {
		result = append(result, pkg)
	}
	return result
}
