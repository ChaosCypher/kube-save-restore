package diff

import (
	"encoding/json"
	"fmt"
	"os"
)

// ResourceMetadata holds metadata for a Kubernetes resource.
type ResourceMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
}

// ResourceWrapper wraps a Kubernetes resource with its metadata.
type ResourceWrapper struct {
	Metadata ResourceMetadata     `json:"metadata"`
	Resource map[string]interface{} `json:"resource"`
}

// LoadResourceFromFile loads a ResourceWrapper from a JSON file.
func LoadResourceFromFile(filename string) (*ResourceWrapper, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filename, err)
	}

	var wrapper ResourceWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON from file %s: %v", filename, err)
	}

	return &wrapper, nil
}