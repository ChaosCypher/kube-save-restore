package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceLister defines the method to list Namespaces
type NamespaceLister interface {
	ListNamespaces(ctx context.Context) ([]string, error)
	GetNamespaces(ctx context.Context) (*corev1.NamespaceList, error)
}

// ListNamespaces lists all namespaces in the cluster
func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.GetNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	var namespaceList []string
	for _, ns := range namespaces.Items {
		namespaceList = append(namespaceList, ns.Name)
	}
	return namespaceList, nil
}

// GetNamespaces retrieves all namespaces in the cluster
func (c *Client) GetNamespaces(ctx context.Context) (*corev1.NamespaceList, error) {
	return c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
}
