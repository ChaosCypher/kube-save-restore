package backup

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// KubernetesClient defines the interface for interacting with Kubernetes resources
type KubernetesClient interface {
	// ListNamespaces returns a list of all namespace names in the cluster
	ListNamespaces(ctx context.Context) ([]string, error)

	// GetNamespaces returns a list of all namespace objects in the cluster
	GetNamespaces(ctx context.Context) (*corev1.NamespaceList, error)

	// ListDeployments returns a list of all deployments in the specified namespace
	ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error)

	// ListServices returns a list of all services in the specified namespace
	ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error)

	// ListConfigMaps returns a list of all config maps in the specified namespace
	ListConfigMaps(ctx context.Context, namespace string) (*corev1.ConfigMapList, error)

	// ListDaemonSets returns a list of all daemon sets in the specified namespace
	ListDaemonSets(ctx context.Context, namespace string) (*appsv1.DaemonSetList, error)

	// ListSecrets returns a list of all secrets in the specified namespace
	ListSecrets(ctx context.Context, namespace string) (*corev1.SecretList, error)

	// ListServiceAccounts returns a list of all service accounts in the specified namespace
	ListServiceAccounts(ctx context.Context, namespace string) (*corev1.ServiceAccountList, error)

	// ListStatefulSets returns a list of all stateful sets in the specified namespace
	ListStatefulSets(ctx context.Context, namespace string) (*appsv1.StatefulSetList, error)

	// ListHorizontalPodAutoscalers returns a list of all horizontal pod autoscalers in the specified namespace
	ListHorizontalPodAutoscalers(ctx context.Context, namespace string) (*autoscalingv2.HorizontalPodAutoscalerList, error)

	// ListCronJobs returns a list of all cron jobs in the specified namespace
	ListCronJobs(ctx context.Context, namespace string) (*batchv1.CronJobList, error)

	// ListJobs returns a list of all jobs in the specified namespace
	ListJobs(ctx context.Context, namespace string) (*batchv1.JobList, error)

	// ListPersistentVolumeClaims returns a list of all persistent volume claims in the specified namespace
	ListPersistentVolumeClaims(ctx context.Context, namespace string) (*corev1.PersistentVolumeClaimList, error)

	// ListIngresses returns a list of all ingresses in the specified namespace
	ListIngresses(ctx context.Context, namespace string) (*networkingv1.IngressList, error)

	// ListRoles returns a list of all roles in the specified namespace
	ListRoles(ctx context.Context, namespace string) (*rbacv1.RoleList, error)

	// ListRoleBindings returns a list of all role bindings in the specified namespace
	ListRoleBindings(ctx context.Context, namespace string) (*rbacv1.RoleBindingList, error)

	// ListNetworkPolicies returns a list of all network policies in the specified namespace
	ListNetworkPolicies(ctx context.Context, namespace string) (*networkingv1.NetworkPolicyList, error)
}
