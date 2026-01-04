package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceInfo contains information about a Kubernetes namespace
type NamespaceInfo struct {
	Name      string
	Status    string
	Age       time.Duration
	IsCurrent bool
}

// ListNamespaces returns all namespaces in the current cluster
func (c *Client) ListNamespaces(ctx context.Context) ([]NamespaceInfo, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	return c.namespacesToInfo(namespaces.Items), nil
}

// namespacesToInfo converts namespace objects to NamespaceInfo
func (c *Client) namespacesToInfo(namespaces []corev1.Namespace) []NamespaceInfo {
	now := time.Now()
	result := make([]NamespaceInfo, 0, len(namespaces))

	for _, ns := range namespaces {
		age := now.Sub(ns.CreationTimestamp.Time)

		result = append(result, NamespaceInfo{
			Name:      ns.Name,
			Status:    string(ns.Status.Phase),
			Age:       age,
			IsCurrent: ns.Name == c.currentNamespace,
		})
	}

	// Sort alphabetically by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// NamespaceExists checks if a namespace exists in the current cluster
func (c *Client) NamespaceExists(ctx context.Context, name string) (bool, error) {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// Check if it's a "not found" error
		if isNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check namespace %q: %w", name, err)
	}
	return true, nil
}

// isNotFoundError checks if an error is a Kubernetes "not found" error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check for the standard Kubernetes not found status
	return err.Error() == "not found" ||
		(len(err.Error()) > 0 && err.Error()[0:9] == "namespace")
}
