package restore

import (
	"fmt"
	"os"
	"path/filepath"
)

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

func adjustResourceStructure(rawResource map[string]interface{}) (map[string]interface{}, string) {
	var resource map[string]interface{}
	var kind string

	if rawKind, ok := rawResource["kind"].(string); ok {
		resource = rawResource["resource"].(map[string]interface{})
		resource["kind"] = rawKind
		resource["apiVersion"] = "v1"
		kind = rawKind
	} else {
		resource = rawResource
		kind = resource["kind"].(string)
	}

	metadata := resource["metadata"].(map[string]interface{})
	delete(metadata, "resourceVersion")
	delete(metadata, "creationTimestamp")
	delete(metadata, "managedFields")

	return resource, kind
}

func validateResource(resource map[string]interface{}) error {
	metadata, ok := resource["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("resource metadata not found")
	}

	requiredFields := []string{"name", "namespace"}
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

func getResourceIdentifiers(resource map[string]interface{}) (string, string) {
	metadata := resource["metadata"].(map[string]interface{})
	name, _ := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)
	return name, namespace
}
