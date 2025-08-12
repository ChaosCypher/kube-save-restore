package kubernetes

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobLister interface {
	ListJobs(ctx context.Context, namespace string) (*batchv1.JobList, error)
}

func (c *Client) ListJobs(ctx context.Context, namespace string) (*batchv1.JobList, error) {
	return c.Clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
}
