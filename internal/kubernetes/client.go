package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientInterface defines the methods that a Kubernetes client should implement
type ClientInterface interface {
	ConfigMapLister
	DaemonSetLister
	DeploymentLister
	NamespaceLister
	SecretLister
	ServiceAccountLister
	ServiceLister
	StatefulSetLister
	HorizontalPodAutoscalerLister
	CronJobLister
	JobLister
	PersistentVolumeClaimLister
	IngressLister
	RoleLister
	RoleBindingLister
	NetworkPolicyLister
}

// Client implements the ClientInterface
type Client struct {
	Clientset kubernetes.Interface
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

	return &Client{Clientset: clientset}, nil
}
