package kubernetes

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronJobLister is an interface that lists cronjobs.
type CronJobLister interface {
	ListCronJobs(ctx context.Context, namespace string) (*batchv1.CronJobList, error)
}

// ListCronJobs lists all cronjobs in the given namespace.
func (c *Client) ListCronJobs(ctx context.Context, namespace string) (*batchv1.CronJobList, error) {
	return c.Clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
}
