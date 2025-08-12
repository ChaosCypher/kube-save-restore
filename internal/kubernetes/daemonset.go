package kubernetes

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSetLister defines the method to list DaemonSets
type DaemonSetLister interface {
	ListDaemonSets(ctx context.Context, namespace string) (*appsv1.DaemonSetList, error)
}

// ListDaemonSets lists all DaemonSets in the specified namespace
func (c *Client) ListDaemonSets(ctx context.Context, namespace string) (*appsv1.DaemonSetList, error) {
	return c.Clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
}
