package kubernetes

import (
	"context"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HPA defines the methods to list HPAs
type HPA interface {
	ListHPAs(ctx context.Context, namespace string) (*autoscalingv2.HorizontalPodAutoscalerList, error)
}

// ListHPAs lists all HPAs in the specified namespace
func (c *Client) ListHorizontalPodAutoscalers(ctx context.Context, namespace string) (*autoscalingv2.HorizontalPodAutoscalerList, error) {
	return c.Clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
}
