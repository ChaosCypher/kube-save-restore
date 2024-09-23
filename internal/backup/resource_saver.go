package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// saveResource saves a Kubernetes resource to a JSON file.
func (bm *Manager) saveResource(resource interface{}, kind, filename string) error {
	// Create a wrapper struct to include the resource kind
	wrapper := struct {
		Kind     string      `json:"kind"`
		Resource interface{} `json:"resource"`
	}{
		Kind:     kind,
		Resource: resource,
	}

	// Marshal the wrapper struct to JSON with indentation
	data, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling resource: %v", err)
	}

	// Create the directory structure if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Write the JSON data to the file
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	bm.logger.Debugf("Saved resource to file: %s", filename)
	return nil
}
