package adapters

import "fmt"

var availableAdapters = []Adapter{
	NewOpenCodeAdapter(),
	NewCursorAdapter(),
	NewClaudeAdapter(),
}

// DetectInstalledTools returns the names of all detected AI tools on the system
func DetectInstalledTools() []string {
	detected := []string{}

	for _, adapter := range availableAdapters {
		if adapter.Detect() {
			detected = append(detected, adapter.Name())
		}
	}

	return detected
}

// GetAdapter returns an adapter by name
func GetAdapter(name string) (Adapter, error) {
	for _, adapter := range availableAdapters {
		if adapter.Name() == name {
			return adapter, nil
		}
	}
	return nil, fmt.Errorf("adapter not found: %s", name)
}

// GetAdapterNames returns the names of all available adapters
func GetAdapterNames() []string {
	names := make([]string, len(availableAdapters))
	for i, adapter := range availableAdapters {
		names[i] = adapter.Name()
	}
	return names
}
