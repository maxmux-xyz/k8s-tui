package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestKubeconfig(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://cluster-1.example.com:6443
    insecure-skip-tls-verify: true
  name: cluster-1
- cluster:
    server: https://cluster-2.example.com:6443
    insecure-skip-tls-verify: true
  name: cluster-2
- cluster:
    server: https://cluster-3.example.com:6443
    insecure-skip-tls-verify: true
  name: cluster-3
contexts:
- context:
    cluster: cluster-1
    user: user-1
    namespace: namespace-1
  name: context-alpha
- context:
    cluster: cluster-2
    user: user-2
    namespace: namespace-2
  name: context-beta
- context:
    cluster: cluster-3
    user: user-3
  name: context-gamma
current-context: context-beta
users:
- name: user-1
  user:
    token: token-1
- name: user-2
  user:
    token: token-2
- name: user-3
  user:
    token: token-3
`
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600); err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	return kubeconfigPath
}

func TestClient_ListContexts(t *testing.T) {
	kubeconfigPath := createTestKubeconfig(t)

	client, err := NewClient(WithKubeconfig(kubeconfigPath))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	contexts := client.ListContexts()

	if len(contexts) != 3 {
		t.Errorf("expected 3 contexts, got %d", len(contexts))
	}

	// Verify sorted order
	expectedOrder := []string{"context-alpha", "context-beta", "context-gamma"}
	for i, ctx := range contexts {
		if ctx.Name != expectedOrder[i] {
			t.Errorf("expected context %d to be %q, got %q", i, expectedOrder[i], ctx.Name)
		}
	}

	// Verify current context is marked
	var currentCount int
	for _, ctx := range contexts {
		if ctx.IsCurrent {
			currentCount++
			if ctx.Name != "context-beta" {
				t.Errorf("expected context-beta to be current, got %q", ctx.Name)
			}
		}
	}
	if currentCount != 1 {
		t.Errorf("expected exactly 1 current context, got %d", currentCount)
	}

	// Verify context without namespace defaults to "default"
	for _, ctx := range contexts {
		if ctx.Name == "context-gamma" && ctx.Namespace != "default" {
			t.Errorf("expected context-gamma namespace to be 'default', got %q", ctx.Namespace)
		}
	}
}

func TestClient_GetContextInfo(t *testing.T) {
	kubeconfigPath := createTestKubeconfig(t)

	client, err := NewClient(WithKubeconfig(kubeconfigPath))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	info, err := client.GetContextInfo("context-alpha")
	if err != nil {
		t.Fatalf("failed to get context info: %v", err)
	}

	if info.Name != "context-alpha" {
		t.Errorf("expected name 'context-alpha', got %q", info.Name)
	}
	if info.Cluster != "cluster-1" {
		t.Errorf("expected cluster 'cluster-1', got %q", info.Cluster)
	}
	if info.User != "user-1" {
		t.Errorf("expected user 'user-1', got %q", info.User)
	}
	if info.Namespace != "namespace-1" {
		t.Errorf("expected namespace 'namespace-1', got %q", info.Namespace)
	}
	if info.IsCurrent {
		t.Error("expected context-alpha to not be current")
	}
}

func TestClient_GetContextInfo_NotFound(t *testing.T) {
	kubeconfigPath := createTestKubeconfig(t)

	client, err := NewClient(WithKubeconfig(kubeconfigPath))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.GetContextInfo("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent context")
	}
}

func TestClient_SwitchContext(t *testing.T) {
	kubeconfigPath := createTestKubeconfig(t)

	client, err := NewClient(WithKubeconfig(kubeconfigPath))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Initial context should be context-beta
	if client.CurrentContext() != "context-beta" {
		t.Errorf("expected initial context 'context-beta', got %q", client.CurrentContext())
	}

	// Switch to context-alpha
	if err := client.SwitchContext("context-alpha"); err != nil {
		t.Fatalf("failed to switch context: %v", err)
	}

	if client.CurrentContext() != "context-alpha" {
		t.Errorf("expected context 'context-alpha' after switch, got %q", client.CurrentContext())
	}

	// Namespace should update to the new context's namespace
	if client.CurrentNamespace() != "namespace-1" {
		t.Errorf("expected namespace 'namespace-1', got %q", client.CurrentNamespace())
	}

	// Verify ListContexts reflects the change
	contexts := client.ListContexts()
	for _, ctx := range contexts {
		if ctx.Name == "context-alpha" && !ctx.IsCurrent {
			t.Error("expected context-alpha to be marked as current after switch")
		}
		if ctx.Name == "context-beta" && ctx.IsCurrent {
			t.Error("expected context-beta to not be current after switch")
		}
	}
}

func TestClient_SwitchContext_NotFound(t *testing.T) {
	kubeconfigPath := createTestKubeconfig(t)

	client, err := NewClient(WithKubeconfig(kubeconfigPath))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.SwitchContext("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent context")
	}

	// Context should remain unchanged
	if client.CurrentContext() != "context-beta" {
		t.Errorf("context should remain 'context-beta' after failed switch, got %q", client.CurrentContext())
	}
}

func TestListContextsFromConfig(t *testing.T) {
	kubeconfigPath := createTestKubeconfig(t)

	contexts, currentContext, err := ListContextsFromConfig(kubeconfigPath)
	if err != nil {
		t.Fatalf("failed to list contexts: %v", err)
	}

	if len(contexts) != 3 {
		t.Errorf("expected 3 contexts, got %d", len(contexts))
	}

	if currentContext != "context-beta" {
		t.Errorf("expected current context 'context-beta', got %q", currentContext)
	}
}

func TestListContextsFromConfig_NoKubeconfig(t *testing.T) {
	tmpDir := t.TempDir()
	nonexistentPath := filepath.Join(tmpDir, "nonexistent")

	_, _, err := ListContextsFromConfig(nonexistentPath)
	if err == nil {
		t.Error("expected error for nonexistent kubeconfig")
	}
}
