package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/chaoscypher/k8s-backup-restore/internal/kubernetes"
	"github.com/chaoscypher/k8s-backup-restore/internal/utils"

	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const maxConcurrency = 10

type Manager struct{}

// NewRestoreManager creates a new instance of RestoreManager.
func NewManager() *Manager {
	return &Manager{}
}

// PerformRestore performs the restore operation by reading resource files from the specified directory
// and applying them to the Kubernetes cluster. If dryRun is true, no changes will be made.
func (rm *Manager) PerformRestore(client *kubernetes.Client, restoreDir string, dryRun bool, logger *utils.Logger) error {
	logger.Info("Starting restore operation")

	files, err := getResourceFiles(restoreDir)
	if err != nil {
		return fmt.Errorf("error getting resource files: %v", err)
	}

	totalResources := len(files)

	if dryRun {
		logger.Info("Dry run mode: No resources will be created or modified")
	}

	tasks := make(chan string, totalResources)
	var wg sync.WaitGroup

	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range tasks {
				if err := rm.RestoreResource(client, filename, dryRun, logger); err != nil {
					logger.Errorf("Error restoring resource from %s: %v", filename, err)
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
		logger.Infof("Dry run completed. %d resources would be restored from: %s", totalResources, restoreDir)
	} else {
		logger.Infof("Restore completed. %d resources restored from: %s", totalResources, restoreDir)
	}
	return nil
}

// RestoreResource restores a single resource from the specified file. If dryRun is true, no changes will be made.
func (rm *Manager) RestoreResource(client *kubernetes.Client, filename string, dryRun bool, logger *utils.Logger) error {
	logger.Debugf("Restoring resource from file: %s", filename)

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
		logger.Infof("Would restore %s: %s/%s", kind, namespace, name)
		return nil
	}

	return applyResource(client, resource, kind, namespace)
}

// getResourceFiles returns a list of JSON files in the specified directory.
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

// adjustResourceStructure adjusts the structure of the raw resource to ensure it has the correct format.
func adjustResourceStructure(rawResource map[string]interface{}) (map[string]interface{}, string) {
	var resource map[string]interface{}
	var kind string

	if rawKind, ok := rawResource["kind"].(string); ok {
		resource = rawResource["resource"].(map[string]interface{})
		resource["kind"] = rawKind
		resource["apiVersion"] = "v1" // Set a default apiVersion if not present
		kind = rawKind
	} else {
		resource = rawResource // If it's already in the correct format, use as-is
		kind = resource["kind"].(string)
	}

	metadata := resource["metadata"].(map[string]interface{})
	delete(metadata, "resourceVersion")
	delete(metadata, "creationTimestamp")
	delete(metadata, "managedFields")

	return resource, kind
}

// validateResource validates that the resource has the required metadata fields.
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

// getResourceIdentifiers returns the name and namespace of the resource.
func getResourceIdentifiers(resource map[string]interface{}) (string, string) {
	metadata := resource["metadata"].(map[string]interface{})
	name, _ := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)
	return name, namespace
}

// applyResource applies the resource to the Kubernetes cluster based on its kind.
func applyResource(client *kubernetes.Client, resource map[string]interface{}, kind, namespace string) error {
	adjustedData, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("error marshaling adjusted resource: %v", err)
	}

	switch kind {
	case "Deployment":
		return applyDeployment(client, adjustedData, namespace)
	case "Service":
		return applyService(client, adjustedData, namespace)
	case "ConfigMap":
		return applyConfigMap(client, adjustedData, namespace)
	case "Secret":
		return applySecret(client, adjustedData, namespace)
	default:
		return fmt.Errorf("unsupported resource kind: %s", kind)
	}
}

// applyDeployment applies a Deployment resource to the Kubernetes cluster.
func applyDeployment(client *kubernetes.Client, data []byte, namespace string) error {
	var deployment appsv1.Deployment
	if err := json.Unmarshal(data, &deployment); err != nil {
		return fmt.Errorf("error unmarshaling deployment: %v", err)
	}
	_, err := client.Clientset.AppsV1().Deployments(namespace).Update(context.TODO(), &deployment, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.AppsV1().Deployments(namespace).Create(context.TODO(), &deployment, metav1.CreateOptions{})
	}
	return err
}

// applyService applies a Service resource to the Kubernetes cluster.
func applyService(client *kubernetes.Client, data []byte, namespace string) error {
	var service corev1.Service
	if err := json.Unmarshal(data, &service); err != nil {
		return fmt.Errorf("error unmarshaling service: %v", err)
	}
	_, err := client.Clientset.CoreV1().Services(namespace).Update(context.TODO(), &service, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().Services(namespace).Create(context.TODO(), &service, metav1.CreateOptions{})
	}
	return err
}

// applyConfigMap applies a ConfigMap resource to the Kubernetes cluster.
func applyConfigMap(client *kubernetes.Client, data []byte, namespace string) error {
	var configMap corev1.ConfigMap
	if err := json.Unmarshal(data, &configMap); err != nil {
		return fmt.Errorf("error unmarshaling configmap: %v", err)
	}
	_, err := client.Clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), &configMap, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &configMap, metav1.CreateOptions{})
	}
	return err
}

// applySecret applies a Secret resource to the Kubernetes cluster.
func applySecret(client *kubernetes.Client, data []byte, namespace string) error {
	var secret corev1.Secret
	if err := json.Unmarshal(data, &secret); err != nil {
		return fmt.Errorf("error unmarshaling secret: %v", err)
	}
	_, err := client.Clientset.CoreV1().Secrets(namespace).Update(context.TODO(), &secret, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().Secrets(namespace).Create(context.TODO(), &secret, metav1.CreateOptions{})
	}
	return err
}
