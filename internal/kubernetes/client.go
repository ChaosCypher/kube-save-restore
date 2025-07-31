package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var NewClientFunc = NewClient


// ClientInterface defines the methods that a Kubernetes client should implement
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
	SwitchContext(context string) error // Added for Compare logic
}

// Client implements the ClientInterface
type Client struct {
	Clientset      kubernetes.Interface
	kubeconfigPath string // Added for Compare logic
}

// ConfigModifier is a function type that modifies a rest.Config
type ConfigModifier func(*rest.Config)

// DefaultConfigModifier sets default QPS and Burst values
func DefaultConfigModifier(config *rest.Config) {
	config.QPS = 50.0
	config.Burst = 100
}

// NewClient creates a new Client instance
func NewClient(kubeconfigPath, context string, modifier ConfigModifier) (*Client, error) {
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: context},
	)

	config, err := loader.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	if modifier != nil {
		modifier(config)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Client{
		Clientset:      clientset,
		kubeconfigPath: kubeconfigPath, // Added for Compare logic
	}, nil
}

// SwitchContext dynamically switches the Kubernetes context for the client
func (c *Client) SwitchContext(context string) error {
	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: c.kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: context},
	)

	config, err := loader.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig for context %s: %w", context, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset for context %s: %w", context, err)
	}

	c.Clientset = clientset
	return nil
}

// SetClientset sets the Clientset for the Client (added for testing purposes)
func (c *Client) SetClientset(clientset kubernetes.Interface) {
	c.Clientset = clientset
}
