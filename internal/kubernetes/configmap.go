package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapLister defines the method to list ConfigMaps
type ConfigMapLister interface {
	ListConfigMaps(ctx context.Context, namespace string) (*v1.ConfigMapList, error)
}

// ListConfigMaps lists all ConfigMaps in the specified namespace
func (c *Client) ListConfigMaps(ctx context.Context, namespace string) (*v1.ConfigMapList, error) {
	return c.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
}
