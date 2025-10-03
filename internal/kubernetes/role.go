package kubernetes

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleLister defines the method to list Roles
type RoleLister interface {
	ListRoles(ctx context.Context, namespace string) (*rbacv1.RoleList, error)
}

// ListRoles lists all Roles in the specified namespace
func (c *Client) ListRoles(ctx context.Context, namespace string) (*rbacv1.RoleList, error) {
	return c.Clientset.RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
}
