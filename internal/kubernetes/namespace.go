package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceLister defines the method to list Namespaces
type NamespaceLister interface {
	ListNamespaces(ctx context.Context) ([]string, error)
}

// ListNamespaces lists all namespaces in the cluster
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

// GetNamespace gets a namespace by name
func (c *Client) GetNamespace(ctx context.Context, namespace string) (*corev1.Namespace, error) {
	return c.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
}
