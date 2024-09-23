package kubernetes

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentLister defines the method to list Deployments
type DeploymentLister interface {
	ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error)
}

// ListDeployments lists all Deployments in the specified namespace
func (c *Client) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	return c.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
}
