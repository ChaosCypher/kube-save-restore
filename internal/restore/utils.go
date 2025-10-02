package restore

import (
	"fmt"
	"os"
	"path/filepath"
)

// getResourceFiles walks through the restoreDir and collects all .json files.
// It returns a slice of file paths and an error if any occurs during the walk.
func getResourceFiles(restoreDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(restoreDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// adjustResourceStructure adjusts the structure of the rawResource map.
// It ensures the resource has the correct "kind" and "apiVersion" fields.
// It returns the adjusted resource, its kind, and an error if type assertions fail.
func adjustResourceStructure(rawResource map[string]interface{}) (map[string]interface{}, string, error) {
	var resource map[string]interface{}
	var kind string

	if rawKind, ok := rawResource["kind"].(string); ok {
		resourceMap, ok := rawResource["resource"].(map[string]interface{})
		if !ok {
			return nil, "", fmt.Errorf("expected resource field to be map[string]interface{}, got %T", rawResource["resource"])
		}
		resource = resourceMap
		resource["kind"] = rawKind
		resource["apiVersion"] = "v1"
		kind = rawKind
	} else {
		resource = rawResource
		kindStr, ok := resource["kind"].(string)
		if !ok {
			return nil, "", fmt.Errorf("expected kind field to be string, got %T", resource["kind"])
		}
		kind = kindStr
	}

	metadata, ok := resource["metadata"].(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("expected metadata field to be map[string]interface{}, got %T", resource["metadata"])
	}
	delete(metadata, "resourceVersion")
	delete(metadata, "creationTimestamp")
	delete(metadata, "managedFields")

	return resource, kind, nil
}

// validateResource checks if the resource has the required metadata fields.
// It returns an error if any required field is missing.
func validateResource(resource map[string]interface{}) error {
	metadata, ok := resource["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("resource metadata not found")
	}

	// Get the kind to determine required fields
	kind, _ := resource["kind"].(string)

	// Namespaces are cluster-scoped and don't require a namespace field
	var requiredFields []string
	if kind == "Namespace" {
		requiredFields = []string{"name"}
	} else {
		requiredFields = []string{"name", "namespace"}
	}

	missingFields := []string{}

	for _, field := range requiredFields {
		if _, ok := metadata[field]; !ok {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing metadata fields: %v", missingFields)
	}

	return nil
}

// getResourceIdentifiers extracts the name and namespace from the resource's metadata.
// It returns the name, namespace, and an error if type assertions fail.
func getResourceIdentifiers(resource map[string]interface{}) (string, string, error) {
	metadata, ok := resource["metadata"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("expected metadata field to be map[string]interface{}, got %T", resource["metadata"])
	}
	name, _ := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)

	// For cluster-scoped resources like namespaces, namespace will be empty
	if namespace == "" && resource["kind"] == "Namespace" {
		namespace = "cluster-scoped"
	}

	return name, namespace, nil
}
