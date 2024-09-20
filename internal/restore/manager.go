package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/chaoscypher/k8s-backup-restore/internal/kubernetes"
	"github.com/chaoscypher/k8s-backup-restore/internal/logger"
)

const maxConcurrency = 10

// Manager handles the restore operations.
type Manager struct {
	k8sClient *kubernetes.Client
	logger    logger.LoggerInterface
}

// NewManager creates a new restore Manager.
func NewManager(k8sClient *kubernetes.Client, logger logger.LoggerInterface) *Manager {
	return &Manager{
		k8sClient: k8sClient,
		logger:    logger,
	}
}

// PerformRestore performs the restore operation by reading resource files from the specified directory
// and applying them to the Kubernetes cluster. If dryRun is true, no changes will be made.
func (m *Manager) PerformRestore(restoreDir string, dryRun bool) error {
	m.logger.Info("Starting restore operation")

	files, err := getResourceFiles(restoreDir)
	if err != nil {
		return fmt.Errorf("error getting resource files: %v", err)
	}

	totalResources := len(files)

	if dryRun {
		m.logger.Info("Dry run mode: No resources will be created or modified")
	}

	tasks := make(chan string, totalResources)
	var wg sync.WaitGroup

	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range tasks {
				if err := m.RestoreResource(filename, dryRun); err != nil {
					m.logger.Errorf("Error restoring resource from %s: %v", filename, err)
				}
			}
		}()
	}

	for _, file := range files {
		tasks <- file
	}

	close(tasks)
	wg.Wait()

	if dryRun {
		m.logger.Infof("Dry run completed. %d resources would be restored from: %s", totalResources, restoreDir)
	} else {
		m.logger.Infof("Restore completed. %d resources restored from: %s", totalResources, restoreDir)
	}
	return nil
}

// RestoreResource restores a single resource from the specified file. If dryRun is true, no changes will be made.
func (m *Manager) RestoreResource(filename string, dryRun bool) error {
	m.logger.Debugf("Restoring resource from file: %s", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filename, err)
	}

	var rawResource map[string]interface{}
	if err := json.Unmarshal(data, &rawResource); err != nil {
		return fmt.Errorf("error unmarshaling resource: %v", err)
	}

	resource, kind := adjustResourceStructure(rawResource)

	if err := validateResource(resource); err != nil {
		return err
	}

	name, namespace := getResourceIdentifiers(resource)

	if dryRun {
		m.logger.Infof("Would restore %s: %s/%s", kind, namespace, name)
		return nil
	}

	return applyResource(m.k8sClient, resource, kind, namespace)
}
