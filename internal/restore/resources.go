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
	// Marshal the resource into JSON format
	adjustedData, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("error marshaling adjusted resource: %v", err)
	}

	// Switch based on the kind of resource and call the appropriate function
	switch kind {
	case "Deployment":
		return applyDeployment(client, adjustedData, namespace)
	case "Service":
		return applyService(client, adjustedData, namespace)
	case "ConfigMap":
		return applyConfigMap(client, adjustedData, namespace)
	case "Secret":
		return applySecret(client, adjustedData, namespace)
	default:
		return fmt.Errorf("unsupported resource kind: %s", kind)
	}
}

// applyDeployment applies a Deployment resource to the Kubernetes cluster.
func applyDeployment(client *kubernetes.Client, data []byte, namespace string) error {
	var deployment appsv1.Deployment
	// Unmarshal the JSON data into a Deployment object
	if err := json.Unmarshal(data, &deployment); err != nil {
		return fmt.Errorf("error unmarshaling deployment: %v", err)
	}
	// Try to update the Deployment, if it does not exist, create it
	_, err := client.Clientset.AppsV1().Deployments(namespace).Update(context.TODO(), &deployment, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.AppsV1().Deployments(namespace).Create(context.TODO(), &deployment, metav1.CreateOptions{})
	}
	return err
}

// applyService applies a Service resource to the Kubernetes cluster.
func applyService(client *kubernetes.Client, data []byte, namespace string) error {
	var service corev1.Service
	// Unmarshal the JSON data into a Service object
	if err := json.Unmarshal(data, &service); err != nil {
		return fmt.Errorf("error unmarshaling service: %v", err)
	}
	// Try to update the Service, if it does not exist, create it
	_, err := client.Clientset.CoreV1().Services(namespace).Update(context.TODO(), &service, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().Services(namespace).Create(context.TODO(), &service, metav1.CreateOptions{})
	}
	return err
}

// applyConfigMap applies a ConfigMap resource to the Kubernetes cluster.
func applyConfigMap(client *kubernetes.Client, data []byte, namespace string) error {
	var configMap corev1.ConfigMap
	// Unmarshal the JSON data into a ConfigMap object
	if err := json.Unmarshal(data, &configMap); err != nil {
		return fmt.Errorf("error unmarshaling configmap: %v", err)
	}
	// Try to update the ConfigMap, if it does not exist, create it
	_, err := client.Clientset.CoreV1().ConfigMaps(namespace).Update(context.TODO(), &configMap, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &configMap, metav1.CreateOptions{})
	}
	return err
}

// applySecret applies a Secret resource to the Kubernetes cluster.
func applySecret(client *kubernetes.Client, data []byte, namespace string) error {
	var secret corev1.Secret
	// Unmarshal the JSON data into a Secret object
	if err := json.Unmarshal(data, &secret); err != nil {
		return fmt.Errorf("error unmarshaling secret: %v", err)
	}
	// Try to update the Secret, if it does not exist, create it
	_, err := client.Clientset.CoreV1().Secrets(namespace).Update(context.TODO(), &secret, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().Secrets(namespace).Create(context.TODO(), &secret, metav1.CreateOptions{})
	}
	return err
}
