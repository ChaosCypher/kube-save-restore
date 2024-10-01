package diff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
)

// DiffManager handles the diff process between multiple backup directories.
type DiffManager struct {
	backupDirs []string
	logger     logger.LoggerInterface
}

// NewDiffManager creates a new DiffManager instance.
func NewDiffManager(backupDirs []string, logger logger.LoggerInterface) *DiffManager {
	return &DiffManager{
		backupDirs: backupDirs,
		logger:     logger,
	}
}

// PerformDiff executes the diff operation and returns the report.
func (dm *DiffManager) PerformDiff(ctx context.Context) (*DiffReport, error) {
	dm.logger.Info("Starting diff operation")

	resourceMaps := make(map[string]map[string]*ResourceWrapper)

	for _, dir := range dm.backupDirs {
		dm.logger.Infof("Processing backup directory: %s", dir)
		resources, err := LoadBackupDirectory(dir)
		if err != nil {
			return nil, fmt.Errorf("error loading backup directory %s: %v", dir, err)
		}
		resourceMaps[dir] = resources
	}

	comparer := NewComparer(resourceMaps, dm.logger)
	report := comparer.Compare()

	dm.logger.Info("Diff operation completed")
	return report, nil
}

// LoadBackupDirectory loads all resources from a backup directory.
func LoadBackupDirectory(dir string) (map[string]*ResourceWrapper, error) {
	resources := make(map[string]*ResourceWrapper)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			resource, err := LoadResourceFromFile(path)
			if err != nil {
				return err
			}
			key := GenerateResourceKey(resource.Metadata.Kind, resource.Metadata.Namespace, resource.Metadata.Name)
			resources[key] = resource
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}

// GenerateResourceKey creates a unique key for a resource based on kind, namespace, and name.
func GenerateResourceKey(kind, namespace, name string) string {
	return fmt.Sprintf("%s/%s/%s", kind, namespace, name)
}

// Add this method to the ResourceWrapper struct
func (rw *ResourceWrapper) UnmarshalJSON(data []byte) error {
	type Alias ResourceWrapper
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(rw),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	rw.convertFloatsToInts(rw.Resource)
	return nil
}

func (rw *ResourceWrapper) convertFloatsToInts(m map[string]interface{}) {
	for k, v := range m {
		switch val := v.(type) {
		case float64:
			if float64(int(val)) == val {
				m[k] = int(val)
			}
		case map[string]interface{}:
			rw.convertFloatsToInts(val)
		case []interface{}:
			for _, item := range val {
				if subMap, ok := item.(map[string]interface{}); ok {
					rw.convertFloatsToInts(subMap)
				}
			}
		}
	}
}

func (r *DiffReport) Print() {
	var keys []string
	for k := range r.Resources {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("Resource: %s\n", key)
		backups := r.Resources[key]
		for backup, resource := range backups {
			if resource != nil {
				fmt.Printf("  Backup: %s - Exists\n", backup)
			}
		}
		if missing, ok := r.Missing[key]; ok {
			for _, backup := range missing {
				fmt.Printf("  Backup: %s - Missing\n", backup)
			}
		}
		fmt.Println()
	}
}