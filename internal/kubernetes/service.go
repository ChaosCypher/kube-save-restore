package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceLister defines the method to list Services
type ServiceLister interface {
	ListServices(ctx context.Context, namespace string) (*v1.ServiceList, error)
}

// ListServices lists all Services in the specified namespace
func (c *Client) ListServices(ctx context.Context, namespace string) (*v1.ServiceList, error) {
	return c.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
}
