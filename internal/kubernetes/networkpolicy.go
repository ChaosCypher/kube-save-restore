package kubernetes

import (
	"context"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkPolicyLister interface {
	ListNetworkPolicies(ctx context.Context, namespace string) (*networkingv1.NetworkPolicyList, error)
}

func (c *Client) ListNetworkPolicies(ctx context.Context, namespace string) (*networkingv1.NetworkPolicyList, error) {
	return c.Clientset.NetworkingV1().NetworkPolicies(namespace).List(ctx, metav1.ListOptions{})
}
