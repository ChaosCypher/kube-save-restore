package kubernetes

import (
	"context"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressLister interface {
	ListIngresses(ctx context.Context, namespace string) (*networkingv1.IngressList, error)
}

func (c *Client) ListIngresses(ctx context.Context, namespace string) (*networkingv1.IngressList, error) {
	return c.Clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
}
