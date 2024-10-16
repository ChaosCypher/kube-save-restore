package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeClaimLister defines the methods to list persistent volume claims
type PersistentVolumeClaimLister interface {
	ListPersistentVolumeClaims(ctx context.Context, namespace string) (*v1.PersistentVolumeClaimList, error)
}

// ListPersistentVolumeClaims lists all persistent volume claims in the specified namespace
func (c *Client) ListPersistentVolumeClaims(ctx context.Context, namespace string) (*v1.PersistentVolumeClaimList, error) {
	return c.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
}
