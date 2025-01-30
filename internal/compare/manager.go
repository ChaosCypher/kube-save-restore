package compare

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/chaoscypher/kube-save-restore/internal/kubernetes"
	"github.com/chaoscypher/kube-save-restore/internal/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Manager handles the comparison of Kubernetes resources
type Manager struct {
	k8sClient *kubernetes.Client
	logger    logger.LoggerInterface
	backupDir string
}

// NewManager creates a new instance of Manager
func NewManager(k8sClient *kubernetes.Client, logger logger.LoggerInterface) *Manager {
	return &Manager{
		k8sClient: k8sClient,
		logger:    logger,
	}
}

// PerformCompare executes the comparison of resources between source and target
func (m *Manager) PerformCompare(source, target, compareType string, backupDir string, dryRun bool) error {
	m.logger.Infof("Comparing %s resources from %s to %s", compareType, source, target)

	sourceResources, err := m.getResources(source, compareType, backupDir)
	if err != nil {
		return fmt.Errorf("error getting source resources: %w", err)
	}

	targetResources, err := m.getResources(target, compareType, backupDir)
	if err != nil {
		return fmt.Errorf("error getting target resources: %w", err)
	}

	// Compare the resources and get the differences
	differences := m.compareResources(sourceResources, targetResources)

	// Report the differences found
	m.reportDifferences(differences)

	if dryRun {
		m.logger.Info("Dry run: No changes applied")
	} else {
		m.logger.Info("Comparison complete. See results above.")
	}

	return nil
}

// getResources retrieves resources either from a backup file or a cluster
func (m *Manager) getResources(location, resourceType, backupDir string) ([]runtime.Object, error) {
	if filepath.IsAbs(location) {
		return getResourcesFromBackup(location, resourceType)
	}
	return m.getResourcesFromCluster(location, resourceType)
}

// getResourcesFromBackup reads resources from backup files
func (m *Manager) getResourcesFromBackup(backupDir, resourceType string) ([]runtime.Object, error) {
	// TODO: Implement logic to read resources from backup files
	// This will depend on how your backup files are structured
	return nil, fmt.Errorf("reading from backup not implemented yet")
}

// getResourcesFromCluster fetches resources from a Kubernetes cluster
func (m *Manager) getResourcesFromCluster(context, resourceType string) ([]runtime.Object, error) {
	// Switch Kubernetes context if necessary
	if context != "" {
		if err := m.k8sClient.SwitchContext(context); err != nil {
			return nil, fmt.Errorf("error switching context: %w", err)
		}
	}

	// Fetch resources based on resourceType
	switch resourceType {
	case "deployments":
		return m.getDeployments()
	case "services":
		return m.getServices()
	case "all":
		// TODO: Implement fetching all supported resource types
		return nil, fmt.Errorf("fetching all resources not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// getDeployments fetches all deployments from the cluster
func (m *Manager) getDeployments() ([]runtime.Object, error) {
	deployments, err := m.k8sClient.Clientset.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	objects := make([]runtime.Object, len(deployments.Items))
	for i := range deployments.Items {
		objects[i] = &deployments.Items[i]
	}
	return objects, nil
}

// getServices fetches all services from the cluster
func (m *Manager) getServices() ([]runtime.Object, error) {
	// Note: Changed ClientSet to Clientset to match the correct field name
	services, err := m.k8sClient.Clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	objects := make([]runtime.Object, len(services.Items))
	for i := range services.Items {
		objects[i] = &services.Items[i]
	}
	return objects, nil
}

// compareResources compares two sets of Kubernetes resources and returns the differences
func (m *Manager) compareResources(source, target []runtime.Object) []Difference {
	var differences []Difference

	// Create a map of source resources for quick lookup
	sourceMap := make(map[string]runtime.Object)
	for _, obj := range source {
		key := getObjectKey(obj)
		sourceMap[key] = obj
	}

	// Compare target resources against source
	for _, targetObj := range target {
		key := getObjectKey(targetObj)
		sourceObj, exists := sourceMap[key]
		if !exists {
			differences = append(differences, Difference{
				Type:   "Missing",
				Object: targetObj,
			})
		} else {
			if diff := compareObjects(sourceObj, targetObj); diff != "" {
				differences = append(differences, Difference{
					Type:   "Modified",
					Object: targetObj,
					Diff:   diff,
				})
			}
			delete(sourceMap, key)
		}
	}

	// Add remaining source objects as extras
	for _, extraObj := range sourceMap {
		differences = append(differences, Difference{
			Type:   "Extra",
			Object: extraObj,
		})
	}

	return differences
}

// getObjectKey generates a unique key for a Kubernetes object
func getObjectKey(obj runtime.Object) string {
	accessor, ok := obj.(metav1.Object)
	if !ok {
		// Handle objects that do not implement metav1.Object
		return ""
	}
	return fmt.Sprintf("%s/%s/%s",
		obj.GetObjectKind().GroupVersionKind().Kind,
		accessor.GetNamespace(),
		accessor.GetName(),
	)
}

// compareObjects compares two Kubernetes objects and returns a string representation of their differences
func compareObjects(a, b runtime.Object) string {
	// TODO: Implement a detailed comparison of objects
	// This is a simplified version, you may want to use a diff library for more complex comparisons
	return fmt.Sprintf("Objects differ: %v vs %v", a, b)
}

// reportDifferences logs the differences found between resources
func (m *Manager) reportDifferences(differences []Difference) {
	for _, diff := range differences {
		switch diff.Type {
		case "Missing":
			m.logger.Infof("Resource missing in source: %s", getObjectKey(diff.Object))
		case "Extra":
			m.logger.Infof("Resource extra in source: %s", getObjectKey(diff.Object))
		case "Modified":
			m.logger.Infof("Resource modified: %s\nDiff: %s", getObjectKey(diff.Object), diff.Diff)
		}
	}
}

// Difference represents a difference found between two Kubernetes resources
type Difference struct {
	Type   string
	Object runtime.Object
	Diff   string
}
