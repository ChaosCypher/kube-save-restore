package kubernetes

import (
	"context"

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
