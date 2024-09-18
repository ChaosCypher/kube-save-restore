package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"k8s-backup-restore/internal/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const maxConcurrency = 10

type KubernetesClient interface {
	ListNamespaces(ctx context.Context) ([]string, error)
	ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error)
	ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error)
	ListConfigMaps(ctx context.Context, namespace string) (*corev1.ConfigMapList, error)
	ListSecrets(ctx context.Context, namespace string) (*corev1.SecretList, error)
}

type BackupManager struct {
	client    KubernetesClient
	backupDir string
	dryRun    bool
	logger    *utils.Logger
}

// NewBackupManager creates a new BackupManager instance.
func NewBackupManager(client KubernetesClient, backupDir string, dryRun bool, logger *utils.Logger) *BackupManager {
	return &BackupManager{
		client:    client,
		backupDir: backupDir,
		dryRun:    dryRun,
		logger:    logger,
	}
}

// PerformBackup initiates the backup process for all namespaces.
func (bm *BackupManager) PerformBackup(ctx context.Context) error {
	bm.logger.Info("Starting backup operation")

	namespaces, err := bm.client.ListNamespaces(ctx)
	if err != nil {
		return fmt.Errorf("error listing namespaces: %v", err)
	}

	totalResources := bm.countResources(ctx, namespaces)

	if bm.dryRun {
		bm.logger.Info("Dry run mode: No files will be written")
	}

	tasks := make(chan backupTask, totalResources)
	var wg sync.WaitGroup

	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go bm.worker(ctx, &wg, tasks)
	}

	bm.enqueueTasks(namespaces, tasks)

	close(tasks)
	wg.Wait()

	bm.logCompletionMessage(totalResources)
	return nil
}

// countResources counts the total number of resources across all namespaces.
func (bm *BackupManager) countResources(ctx context.Context, namespaces []string) int {
	total := 0
	for _, ns := range namespaces {
		deployments, _ := bm.client.ListDeployments(ctx, ns)
		services, _ := bm.client.ListServices(ctx, ns)
		configMaps, _ := bm.client.ListConfigMaps(ctx, ns)
		secrets, _ := bm.client.ListSecrets(ctx, ns)

		total += len(deployments.Items) + len(services.Items) + len(configMaps.Items) + len(secrets.Items)
	}
	return total
}

// worker processes backup tasks from the tasks channel.
func (bm *BackupManager) worker(ctx context.Context, wg *sync.WaitGroup, tasks <-chan backupTask) {
	defer wg.Done()
	for task := range tasks {
		if err := bm.backupResource(ctx, task.resourceType, task.namespace); err != nil {
			bm.logger.Errorf("Error backing up resource: %v", err)
		}
	}
}

// enqueueTasks adds backup tasks for each resource type in each namespace to the tasks channel.
func (bm *BackupManager) enqueueTasks(namespaces []string, tasks chan<- backupTask) {
	for _, ns := range namespaces {
		tasks <- backupTask{resourceType: "deployments", namespace: ns}
		tasks <- backupTask{resourceType: "services", namespace: ns}
		tasks <- backupTask{resourceType: "configmaps", namespace: ns}
		tasks <- backupTask{resourceType: "secrets", namespace: ns}
	}
}

// logCompletionMessage logs a message indicating the completion of the backup process.
func (bm *BackupManager) logCompletionMessage(totalResources int) {
	if bm.dryRun {
		bm.logger.Infof("Dry run completed. %d resources would be backed up to: %s", totalResources, bm.backupDir)
	} else {
		bm.logger.Infof("Backup completed. %d resources saved to: %s", totalResources, bm.backupDir)
	}
}

type backupTask struct {
	resourceType string
	namespace    string
}

// backupResource backs up a specific type of resource in a given namespace.
func (bm *BackupManager) backupResource(ctx context.Context, resourceType, namespace string) error {
	var err error
	switch resourceType {
	case "deployments":
		err = bm.backupDeployments(ctx, namespace)
	case "services":
		err = bm.backupServices(ctx, namespace)
	case "configmaps":
		err = bm.backupConfigMaps(ctx, namespace)
	case "secrets":
		err = bm.backupSecrets(ctx, namespace)
	default:
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return err
}

// backupDeployments backs up all deployments in a given namespace.
func (bm *BackupManager) backupDeployments(ctx context.Context, namespace string) error {
	deployments, err := bm.client.ListDeployments(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing deployments in namespace %s: %v", namespace, err)
	}

	for _, deployment := range deployments.Items {
		filename := filepath.Join(bm.backupDir, namespace, "deployments", deployment.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup deployment: %s/%s", namespace, deployment.Name)
		} else {
			if err := bm.saveResource(deployment, "Deployment", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupServices backs up all services in a given namespace.
func (bm *BackupManager) backupServices(ctx context.Context, namespace string) error {
	services, err := bm.client.ListServices(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing services in namespace %s: %v", namespace, err)
	}

	for _, service := range services.Items {
		filename := filepath.Join(bm.backupDir, namespace, "services", service.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup service: %s/%s", namespace, service.Name)
		} else {
			if err := bm.saveResource(service, "Service", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupConfigMaps backs up all configmaps in a given namespace.
func (bm *BackupManager) backupConfigMaps(ctx context.Context, namespace string) error {
	configMaps, err := bm.client.ListConfigMaps(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing configmaps in namespace %s: %v", namespace, err)
	}

	for _, configMap := range configMaps.Items {
		filename := filepath.Join(bm.backupDir, namespace, "configmaps", configMap.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup configmap: %s/%s", namespace, configMap.Name)
		} else {
			if err := bm.saveResource(configMap, "ConfigMap", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// backupSecrets backs up all secrets in a given namespace.
func (bm *BackupManager) backupSecrets(ctx context.Context, namespace string) error {
	secrets, err := bm.client.ListSecrets(ctx, namespace)
	if err != nil {
		return fmt.Errorf("error listing secrets in namespace %s: %v", namespace, err)
	}

	for _, secret := range secrets.Items {
		filename := filepath.Join(bm.backupDir, namespace, "secrets", secret.Name+".json")
		if bm.dryRun {
			bm.logger.Infof("Would backup secret: %s/%s", namespace, secret.Name)
		} else {
			if err := bm.saveResource(secret, "Secret", filename); err != nil {
				return err
			}
		}
	}

	return nil
}

// saveResource saves a resource to a file in JSON format.
func (bm *BackupManager) saveResource(resource interface{}, kind, filename string) error {
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

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	bm.logger.Debugf("Saved resource to file: %s", filename)
	return nil
}
