package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientInterface defines the methods that a Kubernetes client should implement.
// This interface allows for easier testing and mocking of the Kubernetes client.
type ClientInterface interface {
	ConfigMapLister
	DeploymentLister
	NamespaceLister
	SecretLister
	ServiceLister
	StatefulSetLister
	HorizontalPodAutoscalerLister
	CronJobLister
	PersistentVolumeClaimLister
	SwitchContext(context string) error // Allows dynamic switching of Kubernetes contexts.
}

// Client implements the ClientInterface and provides access to the Kubernetes API.
type Client struct {
	Clientset      kubernetes.Interface // The Kubernetes clientset for interacting with the API.
	kubeconfigPath string               // Path to the kubeconfig file used for configuring the client.
}

// ConfigModifier is a function type that modifies a rest.Config.
// This allows customization of the Kubernetes client configuration, such as setting QPS and Burst values.
type ConfigModifier func(*rest.Config)

// DefaultConfigModifier sets default QPS and Burst values for the Kubernetes client configuration.
// These values control the rate limits for API requests.
func DefaultConfigModifier(config *rest.Config) {
	config.QPS = 50.0  // Set queries per second (QPS) to 50.
	config.Burst = 100 // Set burst limit to 100.
}

// NewClient creates a new Client instance configured to interact with a Kubernetes cluster.
//
// Parameters:
// - kubeconfigPath: The path to the kubeconfig file.
// - context: The Kubernetes context to use (can be empty for default context).
// - modifier: A function to modify the default client configuration.
//
// Returns:
// - A pointer to a new Client instance.
// - An error if there was an issue creating the client.
func NewClient(kubeconfigPath, context string, modifier ConfigModifier) (*Client, error) {
	// Load the kubeconfig file and apply any context overrides.
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: context},
	)

	// Build the Kubernetes REST configuration from the loader.
	config, err := loader.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Apply any custom modifications to the configuration (e.g., QPS and Burst).
	if modifier != nil {
		modifier(config)
	}

	// Create a new Kubernetes clientset using the REST configuration.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Client{
		Clientset:      clientset,
		kubeconfigPath: kubeconfigPath,
	}, nil
}

// SwitchContext dynamically switches the Kubernetes context for the client.
//
// Parameters:
// - context: The name of the new context to switch to.
//
// Returns:
// - An error if there was an issue switching contexts or creating a new clientset.
func (c *Client) SwitchContext(context string) error {
	// Create a new config loader with the specified context.
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: c.kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: context},
	)

	// Load the new REST configuration from the loader.
	config, err := loader.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig for context %s: %w", context, err)
	}

	// Create a new clientset using the updated REST configuration.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset for context %s: %w", context, err)
	}

	// Update the existing Client instance with the new clientset.
	c.Clientset = clientset
	return nil
}
