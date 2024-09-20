package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretLister defines the method to list Secrets
type SecretLister interface {
	ListSecrets(ctx context.Context, namespace string) (*v1.SecretList, error)
}

// ListSecrets lists all Secrets in the specified namespace
func (c *Client) ListSecrets(ctx context.Context, namespace string) (*v1.SecretList, error) {
	return c.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
}
