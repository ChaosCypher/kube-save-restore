package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Client implements the ClientInterface
type Client struct {
	Clientset kubernetes.Interface
}

// NewClient creates a new Client instance
// kubeconfigPath is the path to the kubeconfig file
// context is the Kubernetes context to use
func NewClient(kubeconfigPath, context string) (*Client, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: context},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Client{Clientset: clientset}, nil
}

// ListConfigMaps lists all ConfigMaps in the specified namespace
// ctx is the context for the request
// namespace is the namespace to list ConfigMaps from
func (c *Client) ListConfigMaps(ctx context.Context, namespace string) (*v1.ConfigMapList, error) {
	return c.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
}

// ListDeployments lists all Deployments in the specified namespace
// ctx is the context for the request
// namespace is the namespace to list Deployments from
func (c *Client) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	return c.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
}

// ListNamespaces lists all namespaces in the cluster
// ctx is the context for the request
func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var namespaceList []string
	for _, ns := range namespaces.Items {
		namespaceList = append(namespaceList, ns.Name)
	}
	return namespaceList, nil
}

// ListSecrets lists all Secrets in the specified namespace
// ctx is the context for the request
// namespace is the namespace to list Secrets from
func (c *Client) ListSecrets(ctx context.Context, namespace string) (*v1.SecretList, error) {
	return c.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
}

// ListServices lists all Services in the specified namespace
// ctx is the context for the request
// namespace is the namespace to list Services from
func (c *Client) ListServices(ctx context.Context, namespace string) (*v1.ServiceList, error) {
	return c.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
}
