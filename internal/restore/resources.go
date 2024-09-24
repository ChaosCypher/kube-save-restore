package restore

import (
	"encoding/json"
	"fmt"

	"context"

	"github.com/chaoscypher/k8s-backup-restore/internal/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// applyResource applies the resource to the Kubernetes cluster based on its kind.
func applyResource(client *kubernetes.Client, resource map[string]interface{}, kind, namespace string) error {
	adjustedData, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("error marshaling adjusted resource: %v", err)
	}

	switch kind {
	case "Deployment":
		return applyDeployment(client, adjustedData, namespace)
	case "Service":
		return applyService(client, adjustedData, namespace)
	case "ConfigMap":
		return applyConfigMap(client, adjustedData, namespace)
	case "Secret":
		return applySecret(client, adjustedData, namespace)
	case "StatefulSet":
		return applyStatefulSet(client, adjustedData, namespace)
	default:
		return fmt.Errorf("unsupported resource kind: %s", kind)
	}
}

// applyDeployment applies a Deployment resource to the Kubernetes cluster.
func applyDeployment(client *kubernetes.Client, data []byte, namespace string) error {
	var deployment appsv1.Deployment
	if err := json.Unmarshal(data, &deployment); err != nil {
		return fmt.Errorf("error unmarshaling deployment: %v", err)
	}
	_, err := client.Clientset.AppsV1().Deployments(namespace).Update(context.TODO(), &deployment, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.AppsV1().Deployments(namespace).Create(context.TODO(), &deployment, metav1.CreateOptions{})
	}
	return err
}

// applyService applies a Service resource to the Kubernetes cluster.
func applyService(client *kubernetes.Client, data []byte, namespace string) error {
	var service corev1.Service
	if err := json.Unmarshal(data, &service); err != nil {
		return fmt.Errorf("error unmarshaling service: %v", err)
	}
	_, err := client.Clientset.CoreV1().Services(namespace).Update(context.TODO(), &service, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().Services(namespace).Create(context.TODO(), &service, metav1.CreateOptions{})
	}
	return err
}

// applyConfigMap applies a ConfigMap resource to the Kubernetes cluster.
func applyConfigMap(client *kubernetes.Client, data []byte, namespace string) error {
	var configMap corev1.ConfigMap
	if err := json.Unmarshal(data, &configMap); err != nil {
		return fmt.Errorf("error unmarshaling configmap: %v", err)
	}
	_, err := client.Clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), &configMap, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &configMap, metav1.CreateOptions{})
	}
	return err
}

// applySecret applies a Secret resource to the Kubernetes cluster.
func applySecret(client *kubernetes.Client, data []byte, namespace string) error {
	var secret corev1.Secret
	if err := json.Unmarshal(data, &secret); err != nil {
		return fmt.Errorf("error unmarshaling secret: %v", err)
	}
	_, err := client.Clientset.CoreV1().Secrets(namespace).Update(context.TODO(), &secret, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().Secrets(namespace).Create(context.TODO(), &secret, metav1.CreateOptions{})
	}
	return err
}

// applyStatefulSet applies a StatefulSet resource to the Kubernetes cluster.
func applyStatefulSet(client *kubernetes.Client, data []byte, namespace string) error {
	var statefulSet appsv1.StatefulSet
	if err := json.Unmarshal(data, &statefulSet); err != nil {
		return fmt.Errorf("error unmarshaling stateful set: %v", err)
	}
	_, err := client.Clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), &statefulSet, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.AppsV1().StatefulSets(namespace).Create(context.TODO(), &statefulSet, metav1.CreateOptions{})
	}
	return err
}
