package sources

import (
	"fmt"
)

// GetSource returns the appropriate Source implementation for a package spec
func GetSource(spec PackageSpec) (Source, error) {
	switch spec.Type {
	case "github", "gitlab", "git":
		return NewGitSource(""), nil
	case "local":
		return NewLocalSource(), nil
	default:
		return nil, fmt.Errorf("unsupported source type: %s", spec.Type)
	}
}
