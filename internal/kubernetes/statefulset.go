package kubernetes

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSetLister defines the method to list StatefulSets
type StatefulSetLister interface {
	ListStatefulSets(ctx context.Context, namespace string) (*appsv1.StatefulSetList, error)
}

// ListStatefulSets lists all StatefulSets in the specified namespace
func (c *Client) ListStatefulSets(ctx context.Context, namespace string) (*appsv1.StatefulSetList, error) {
	return c.Clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
}
