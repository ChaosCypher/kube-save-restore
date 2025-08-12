package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAccountLister defines the method to list ServiceAccounts
type ServiceAccountLister interface {
	ListServiceAccounts(ctx context.Context, namespace string) (*v1.ServiceAccountList, error)
}

// ListServiceAccounts lists all ServiceAccounts in the specified namespace
func (c *Client) ListServiceAccounts(ctx context.Context, namespace string) (*v1.ServiceAccountList, error) {
	return c.Clientset.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
}
