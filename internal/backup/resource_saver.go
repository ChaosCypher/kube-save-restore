package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (bm *Manager) saveResource(resource interface{}, kind, filename string) error {
	wrapper := struct {
		Kind     string      `json:"kind"`
		Resource interface{} `json:"resource"`
	}{
		Kind:     kind,
		Resource: resource,
	}

	data, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling resource: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	bm.logger.Debugf("Saved resource to file: %s", filename)
	return nil
}
