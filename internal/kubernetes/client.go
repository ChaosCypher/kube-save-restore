package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientInterface defines the methods that a Kubernetes client should implement
type ClientInterface interface {
	ConfigMapLister
	DeploymentLister
	NamespaceLister
	SecretLister
	ServiceLister
	StatefulSetLister
}

// Client implements the ClientInterface
type Client struct {
	Clientset kubernetes.Interface
}

// NewClient creates a new Client instance
func NewClient(kubeconfigPath, context string) (*Client, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: context},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Client{Clientset: clientset}, nil
}
