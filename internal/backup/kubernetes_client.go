package backup

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// KubernetesClient defines the interface for interacting with Kubernetes resources.
type KubernetesClient interface {
	// ListNamespaces returns a list of all namespace names in the cluster.
	ListNamespaces(ctx context.Context) ([]string, error)

	// ListDeployments returns a list of all deployments in the specified namespace.
	ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error)

	// ListServices returns a list of all services in the specified namespace.
	ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error)

	// ListConfigMaps returns a list of all config maps in the specified namespace.
	ListConfigMaps(ctx context.Context, namespace string) (*corev1.ConfigMapList, error)

	// ListSecrets returns a list of all secrets in the specified namespace.
	ListSecrets(ctx context.Context, namespace string) (*corev1.SecretList, error)
}
