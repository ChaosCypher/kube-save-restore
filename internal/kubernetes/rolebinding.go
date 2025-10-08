package kubernetes

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleBindingLister defines the method to list RoleBindings
type RoleBindingLister interface {
	ListRoleBindings(ctx context.Context, namespace string) (*rbacv1.RoleBindingList, error)
}

// ListRoleBindings lists all RoleBindings in the specified namespace
func (c *Client) ListRoleBindings(ctx context.Context, namespace string) (*rbacv1.RoleBindingList, error) {
	return c.Clientset.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
}
