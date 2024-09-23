package backup

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type KubernetesClient interface {
	ListNamespaces(ctx context.Context) ([]string, error)
	ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error)
	ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error)
	ListConfigMaps(ctx context.Context, namespace string) (*corev1.ConfigMapList, error)
	ListSecrets(ctx context.Context, namespace string) (*corev1.SecretList, error)
}
