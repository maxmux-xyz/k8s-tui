package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// ContextInfo contains information about a Kubernetes context
type ContextInfo struct {
	Name      string
	Cluster   string
	User      string
	Namespace string
	IsCurrent bool
}

// ListContexts returns all available contexts from the kubeconfig
func (c *Client) ListContexts() []ContextInfo {
	var contexts []ContextInfo

	for name, ctx := range c.rawConfig.Contexts {
		namespace := ctx.Namespace
		if namespace == "" {
			namespace = "default"
		}

		contexts = append(contexts, ContextInfo{
			Name:      name,
			Cluster:   ctx.Cluster,
			User:      ctx.AuthInfo,
			Namespace: namespace,
			IsCurrent: name == c.currentContext,
		})
	}

	// Sort by name for consistent ordering
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})

	return contexts
}

// SwitchContext switches to a different context and reinitializes the client
func (c *Client) SwitchContext(contextName string) error {
	// Validate context exists
	if _, exists := c.rawConfig.Contexts[contextName]; !exists {
		return fmt.Errorf("context %q not found", contextName)
	}

	// Build new config with the selected context using the stored kubeconfig path
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: c.kubeconfigPath,
	}
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: contextName,
	}

	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	// Build REST config for new context
	restConfig, err := configLoader.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build config for context %q: %w", contextName, err)
	}

	// Create new clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create client for context %q: %w", contextName, err)
	}

	// Get namespace for new context
	namespace, _, err := configLoader.Namespace()
	if err != nil {
		namespace = "default"
	}

	// Update client state
	c.clientset = clientset
	c.config = restConfig
	c.configLoader = configLoader
	c.currentContext = contextName
	c.currentNamespace = namespace

	return nil
}

// GetContextInfo returns information about a specific context
func (c *Client) GetContextInfo(contextName string) (ContextInfo, error) {
	ctx, exists := c.rawConfig.Contexts[contextName]
	if !exists {
		return ContextInfo{}, fmt.Errorf("context %q not found", contextName)
	}

	namespace := ctx.Namespace
	if namespace == "" {
		namespace = "default"
	}

	return ContextInfo{
		Name:      contextName,
		Cluster:   ctx.Cluster,
		User:      ctx.AuthInfo,
		Namespace: namespace,
		IsCurrent: contextName == c.currentContext,
	}, nil
}

// getKubeconfigPath returns the kubeconfig path from env or default location
func getKubeconfigPath() string {
	if path := os.Getenv("KUBECONFIG"); path != "" {
		return path
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

// ListContextsFromConfig lists contexts without requiring a connected client
// This is useful for initial context selection
func ListContextsFromConfig(kubeconfigPath string) ([]ContextInfo, string, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = getKubeconfigPath()
	}

	if kubeconfigPath == "" {
		return nil, "", fmt.Errorf("no kubeconfig path found")
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: kubeconfigPath,
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).RawConfig()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return contextsFromRawConfig(config, config.CurrentContext), config.CurrentContext, nil
}

// contextsFromRawConfig extracts context info from a raw config
func contextsFromRawConfig(config api.Config, currentContext string) []ContextInfo {
	var contexts []ContextInfo

	for name, ctx := range config.Contexts {
		namespace := ctx.Namespace
		if namespace == "" {
			namespace = "default"
		}

		contexts = append(contexts, ContextInfo{
			Name:      name,
			Cluster:   ctx.Cluster,
			User:      ctx.AuthInfo,
			Namespace: namespace,
			IsCurrent: name == currentContext,
		})
	}

	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})

	return contexts
}
