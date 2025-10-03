package restore

import (
	"encoding/json"
	"fmt"

	"context"

	"github.com/chaoscypher/kube-save-restore/internal/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// applyResource applies the resource to the Kubernetes cluster based on its kind
func applyResource(client *kubernetes.Client, resource map[string]interface{}, kind, namespace string) error {
	// Marshal the resource into JSON format
	adjustedData, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("error marshaling adjusted resource: %v", err)
	}

	// Switch based on the kind of resource and call the appropriate function
	switch kind {
	case "Namespace":
		return applyNamespace(client, adjustedData)
	case "Deployment":
		return applyDeployment(client, adjustedData, namespace)
	case "Service":
		return applyService(client, adjustedData, namespace)
	case "ConfigMap":
		return applyConfigMap(client, adjustedData, namespace)
	case "Secret":
		return applySecret(client, adjustedData, namespace)
	case "ServiceAccount":
		return applyServiceAccount(client, adjustedData, namespace)
	case "StatefulSet":
		return applyStatefulSet(client, adjustedData, namespace)
	case "DaemonSet":
		return applyDaemonSet(client, adjustedData, namespace)
	case "HorizontalPodAutoscaler":
		return applyHorizontalPodAutoscalers(client, adjustedData, namespace)
	case "CronJob":
		return applyCronJob(client, adjustedData, namespace)
	case "Job":
		return applyJob(client, adjustedData, namespace)
	case "PersistentVolumeClaim":
		return applyPersistentVolumeClaim(client, adjustedData, namespace)
	case "Ingress":
		return applyIngress(client, adjustedData, namespace)
	case "Role":
		return applyRole(client, adjustedData, namespace)
	case "NetworkPolicy":
		return applyNetworkPolicy(client, adjustedData, namespace)
	default:
		return fmt.Errorf("unsupported resource kind: %s", kind)
	}
}

// applyNamespace applies a Namespace resource to the Kubernetes cluster
func applyNamespace(client *kubernetes.Client, data []byte) error {
	var namespace corev1.Namespace
	// Unmarshal the JSON data into a Namespace object
	if err := json.Unmarshal(data, &namespace); err != nil {
		return fmt.Errorf("error unmarshaling namespace: %v", err)
	}
	// Try to update the Namespace, if it does not exist, create it
	_, err := client.Clientset.CoreV1().Namespaces().Update(context.TODO(), &namespace, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().Namespaces().Create(context.TODO(), &namespace, metav1.CreateOptions{})
	}
	return err
}

// applyDeployment applies a Deployment resource to the Kubernetes cluster
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

// applyService applies a Service resource to the Kubernetes cluster
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

// applyConfigMap applies a ConfigMap resource to the Kubernetes cluster
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

// applySecret applies a Secret resource to the Kubernetes cluster
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

// applyServiceAccount applies a ServiceAccount resource to the Kubernetes cluster
func applyServiceAccount(client *kubernetes.Client, data []byte, namespace string) error {
	var serviceAccount corev1.ServiceAccount
	// Unmarshal the JSON data into a ServiceAccount object
	if err := json.Unmarshal(data, &serviceAccount); err != nil {
		return fmt.Errorf("error unmarshaling service account: %v", err)
	}
	// Try to update the ServiceAccount, if it does not exist, create it
	_, err := client.Clientset.CoreV1().ServiceAccounts(namespace).Update(context.TODO(), &serviceAccount, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().ServiceAccounts(namespace).Create(context.TODO(), &serviceAccount, metav1.CreateOptions{})
	}
	return err
}

// applyStatefulSet applies a StatefulSet resource to the Kubernetes cluster
func applyStatefulSet(client *kubernetes.Client, data []byte, namespace string) error {
	var statefulSet appsv1.StatefulSet
	// Unmarshal the JSON data into a StatefulSet object
	if err := json.Unmarshal(data, &statefulSet); err != nil {
		return fmt.Errorf("error unmarshaling stateful set: %v", err)
	}
	// Try to update the StatefulSet, if it does not exist, create it
	_, err := client.Clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), &statefulSet, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.AppsV1().StatefulSets(namespace).Create(context.TODO(), &statefulSet, metav1.CreateOptions{})
	}
	return err
}

// applyDaemonSet applies a DaemonSet resource to the Kubernetes cluster
func applyDaemonSet(client *kubernetes.Client, data []byte, namespace string) error {
	var daemonSet appsv1.DaemonSet
	// Unmarshal the JSON data into a DaemonSet object
	if err := json.Unmarshal(data, &daemonSet); err != nil {
		return fmt.Errorf("error unmarshaling daemon set: %v", err)
	}
	// Try to update the DaemonSet, if it does not exist, create it
	_, err := client.Clientset.AppsV1().DaemonSets(namespace).Update(context.TODO(), &daemonSet, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.AppsV1().DaemonSets(namespace).Create(context.TODO(), &daemonSet, metav1.CreateOptions{})
	}
	return err
}

// applyHorizontalPodAutoscalers applies a HorizontalPodAutoscaler resource to the Kubernetes cluster
func applyHorizontalPodAutoscalers(client *kubernetes.Client, data []byte, namespace string) error {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	// Unmarshal the JSON data into a HorizontalPodAutoscaler object
	if err := json.Unmarshal(data, &hpa); err != nil {
		return fmt.Errorf("error unmarshaling hpa: %v", err)
	}
	// Try to update the HorizontalPodAutoscaler, if it does not exist, create it
	_, err := client.Clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(context.TODO(), &hpa, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(context.TODO(), &hpa, metav1.CreateOptions{})

	}
	return err
}

// applyCronJob applies a CronJob resource to the Kubernetes cluster
func applyCronJob(client *kubernetes.Client, data []byte, namespace string) error {
	var cronJob batchv1.CronJob
	// Unmarshal the JSON data into a CronJob object
	if err := json.Unmarshal(data, &cronJob); err != nil {
		return fmt.Errorf("error unmarshaling cron job: %v", err)
	}
	// Try to update the CronJob, if it does not exist, create it
	_, err := client.Clientset.BatchV1().CronJobs(namespace).Update(context.TODO(), &cronJob, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.BatchV1().CronJobs(namespace).Create(context.TODO(), &cronJob, metav1.CreateOptions{})
	}
	return err
}

// applyPersistentVolumeClaim applies a PersistentVolumeClaim resource to the Kubernetes cluster
func applyPersistentVolumeClaim(client *kubernetes.Client, data []byte, namespace string) error {
	var pvc corev1.PersistentVolumeClaim
	// Unmarshal the JSON data into a PersistentVolumeClaim object
	if err := json.Unmarshal(data, &pvc); err != nil {
		return fmt.Errorf("error unmarshaling pvc: %v", err)
	}
	// Try to update the PersistentVolumeClaim, if it does not exist, create it
	_, err := client.Clientset.CoreV1().PersistentVolumeClaims(namespace).Update(context.TODO(), &pvc, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.CoreV1().PersistentVolumeClaims(namespace).Create(context.TODO(), &pvc, metav1.CreateOptions{})
	}
	return err
}

func applyJob(client *kubernetes.Client, data []byte, namespace string) error {
	var job batchv1.Job
	if err := json.Unmarshal(data, &job); err != nil {
		return fmt.Errorf("error unmarshaling job: %v", err)
	}
	_, err := client.Clientset.BatchV1().Jobs(namespace).Update(context.TODO(), &job, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.BatchV1().Jobs(namespace).Create(context.TODO(), &job, metav1.CreateOptions{})
	}
	return err
}

// applyIngress applies an Ingress resource to the Kubernetes cluster
func applyIngress(client *kubernetes.Client, data []byte, namespace string) error {
	var ingress networkingv1.Ingress
	// Unmarshal the JSON data into an Ingress object
	if err := json.Unmarshal(data, &ingress); err != nil {
		return fmt.Errorf("error unmarshaling ingress: %v", err)
	}
	// Try to update the Ingress, if it does not exist, create it
	_, err := client.Clientset.NetworkingV1().Ingresses(namespace).Update(context.TODO(), &ingress, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.NetworkingV1().Ingresses(namespace).Create(context.TODO(), &ingress, metav1.CreateOptions{})
	}
	return err
}

// applyRole applies a Role resource to the Kubernetes cluster
func applyRole(client *kubernetes.Client, data []byte, namespace string) error {
	var role rbacv1.Role
	// Unmarshal the JSON data into a Role object
	if err := json.Unmarshal(data, &role); err != nil {
		return fmt.Errorf("error unmarshaling role: %v", err)
	}
	// Try to update the Role, if it does not exist, create it
	_, err := client.Clientset.RbacV1().Roles(namespace).Update(context.TODO(), &role, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.RbacV1().Roles(namespace).Create(context.TODO(), &role, metav1.CreateOptions{})
	}
	return err
}

// applyNetworkPolicy applies a NetworkPolicy resource to the Kubernetes cluster
func applyNetworkPolicy(client *kubernetes.Client, data []byte, namespace string) error {
	var networkPolicy networkingv1.NetworkPolicy
	// Unmarshal the JSON data into a NetworkPolicy object
	if err := json.Unmarshal(data, &networkPolicy); err != nil {
		return fmt.Errorf("error unmarshaling network policy: %v", err)
	}
	// Try to update the NetworkPolicy, if it does not exist, create it
	_, err := client.Clientset.NetworkingV1().NetworkPolicies(namespace).Update(context.TODO(), &networkPolicy, metav1.UpdateOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = client.Clientset.NetworkingV1().NetworkPolicies(namespace).Create(context.TODO(), &networkPolicy, metav1.CreateOptions{})
	}
	return err
}
