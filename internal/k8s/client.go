package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Client wraps the Kubernetes clientset with context and namespace management
type Client struct {
	clientset        kubernetes.Interface
	config           *rest.Config
	rawConfig        api.Config
	configLoader     clientcmd.ClientConfig
	kubeconfigPath   string
	currentContext   string
	currentNamespace string
}

// ClientOption allows configuring the client
type ClientOption func(*clientOptions)

type clientOptions struct {
	kubeconfig string
	context    string
	namespace  string
}

// WithKubeconfig sets a custom kubeconfig path
func WithKubeconfig(path string) ClientOption {
	return func(o *clientOptions) {
		o.kubeconfig = path
	}
}

// WithContext sets the initial context
func WithContext(ctx string) ClientOption {
	return func(o *clientOptions) {
		o.context = ctx
	}
}

// WithNamespace sets the initial namespace
func WithNamespace(ns string) ClientOption {
	return func(o *clientOptions) {
		o.namespace = ns
	}
}

// NewClient creates a new Kubernetes client
func NewClient(opts ...ClientOption) (*Client, error) {
	options := &clientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Determine kubeconfig path
	kubeconfigPath := options.kubeconfig
	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("KUBECONFIG")
	}
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Check if kubeconfig exists
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig not found at %s", kubeconfigPath)
	}

	// Build config loader with overrides
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: kubeconfigPath,
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	if options.context != "" {
		configOverrides.CurrentContext = options.context
	}
	if options.namespace != "" {
		configOverrides.Context.Namespace = options.namespace
	}

	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	// Get raw config for context/namespace management
	rawConfig, err := configLoader.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Determine current context
	currentContext := rawConfig.CurrentContext
	if options.context != "" {
		currentContext = options.context
	}

	// Validate context exists
	if _, exists := rawConfig.Contexts[currentContext]; !exists && currentContext != "" {
		return nil, fmt.Errorf("context %q not found in kubeconfig", currentContext)
	}

	// Determine namespace
	namespace := options.namespace
	if namespace == "" {
		namespace, _, err = configLoader.Namespace()
		if err != nil {
			namespace = "default"
		}
	}

	// Build REST config
	restConfig, err := configLoader.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		clientset:        clientset,
		config:           restConfig,
		rawConfig:        rawConfig,
		configLoader:     configLoader,
		kubeconfigPath:   kubeconfigPath,
		currentContext:   currentContext,
		currentNamespace: namespace,
	}, nil
}

// Clientset returns the underlying kubernetes clientset
func (c *Client) Clientset() kubernetes.Interface {
	return c.clientset
}

// CurrentContext returns the current context name
func (c *Client) CurrentContext() string {
	return c.currentContext
}

// CurrentNamespace returns the current namespace
func (c *Client) CurrentNamespace() string {
	return c.currentNamespace
}

// SetNamespace changes the current namespace
func (c *Client) SetNamespace(namespace string) {
	c.currentNamespace = namespace
}

// RawConfig returns the raw kubeconfig for inspection
func (c *Client) RawConfig() api.Config {
	return c.rawConfig
}
