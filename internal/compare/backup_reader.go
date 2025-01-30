package compare

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

func getResourcesFromBackup(backupDir, resourceType string) ([]runtime.Object, error) {
	var resources []runtime.Object

	// Define the path to the backup files for the specified resource type
	backupPath := filepath.Join(backupDir, resourceType)

	// Read all files in the backup directory for the resource type
	files, err := ioutil.ReadDir(backupPath)
	if err != nil {
		return nil, fmt.Errorf("error reading backup directory: %v", err)
	}

	decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer()

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(backupPath, file.Name())
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
		}

		obj, _, err := decoder.Decode(data, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("error decoding file %s: %v", filePath, err)
		}

		resources = append(resources, obj)
	}

	return resources, nil
}
