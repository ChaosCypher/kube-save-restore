package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistantVolumeClaimLister defines the methods to list PersistantVolumeClaims
type PersistantVolumeClaimLister interface {
	ListPersistantVolumeClaims(ctx context.Context, namespace string) (*v1.PersistentVolumeClaimList, error)
}

// ListPersistantVolumeClaims lists all PersistantVolumeClaims in the specified namespace
func (c *Client) ListPersistantVolumeClaims(ctx context.Context, namespace string) (*v1.PersistentVolumeClaimList, error) {
	return c.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
}
