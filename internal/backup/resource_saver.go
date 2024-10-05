package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// saveResource saves a Kubernetes resource to a JSON file.
func (bm *Manager) saveResource(resource interface{}, kind, filename string) error {
	bm.logger.Debugf("Saving %s resource to %s", kind, filename)

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
		bm.logger.Errorf("Error marshaling %s resource: %v", kind, err)
		return fmt.Errorf("error marshaling resource: %v", err)
	}

	// Create the directory structure if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		bm.logger.Errorf("Error creating directory for %s: %v", filename, err)
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Write the JSON data to the file
	if err := os.WriteFile(filename, data, 0600); err != nil {
		bm.logger.Errorf("Error writing %s resource to file %s: %v", kind, filename, err)
		return fmt.Errorf("error writing file: %v", err)
	}

	bm.logger.Debugf("Successfully saved %s resource to file: %s", kind, filename)
	return nil
}
