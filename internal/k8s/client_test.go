package k8s

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewClient_NoKubeconfig(t *testing.T) {
	// Create a temp dir with no kubeconfig
	tmpDir := t.TempDir()

	// Override HOME to use temp dir
	originalHome := os.Getenv("HOME")
	originalKubeconfig := os.Getenv("KUBECONFIG")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("KUBECONFIG", originalKubeconfig)
	})

	os.Setenv("HOME", tmpDir)
	os.Unsetenv("KUBECONFIG")

	_, err := NewClient()
	if err == nil {
		t.Error("expected error when kubeconfig doesn't exist")
	}
}

func TestNewClient_InvalidKubeconfigPath(t *testing.T) {
	_, err := NewClient(WithKubeconfig("/nonexistent/path/kubeconfig"))
	if err == nil {
		t.Error("expected error for nonexistent kubeconfig path")
	}
}

func TestNewClient_WithValidKubeconfig(t *testing.T) {
	// Create a minimal valid kubeconfig using insecure-skip-tls-verify
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
    namespace: test-ns
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600); err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	client, err := NewClient(WithKubeconfig(kubeconfigPath))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.CurrentContext() != "test-context" {
		t.Errorf("expected context 'test-context', got %q", client.CurrentContext())
	}

	if client.CurrentNamespace() != "test-ns" {
		t.Errorf("expected namespace 'test-ns', got %q", client.CurrentNamespace())
	}
}

func TestNewClient_WithContextOverride(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: cluster-1
- cluster:
    server: https://localhost:6444
    insecure-skip-tls-verify: true
  name: cluster-2
contexts:
- context:
    cluster: cluster-1
    user: user-1
    namespace: ns-1
  name: context-1
- context:
    cluster: cluster-2
    user: user-2
    namespace: ns-2
  name: context-2
current-context: context-1
users:
- name: user-1
  user:
    token: token-1
- name: user-2
  user:
    token: token-2
`
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600); err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	client, err := NewClient(
		WithKubeconfig(kubeconfigPath),
		WithContext("context-2"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.CurrentContext() != "context-2" {
		t.Errorf("expected context 'context-2', got %q", client.CurrentContext())
	}
}

func TestNewClient_WithNamespaceOverride(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
    namespace: original-ns
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600); err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	client, err := NewClient(
		WithKubeconfig(kubeconfigPath),
		WithNamespace("override-ns"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.CurrentNamespace() != "override-ns" {
		t.Errorf("expected namespace 'override-ns', got %q", client.CurrentNamespace())
	}
}

func TestNewClient_InvalidContext(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600); err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	_, err := NewClient(
		WithKubeconfig(kubeconfigPath),
		WithContext("nonexistent-context"),
	)
	if err == nil {
		t.Error("expected error for nonexistent context")
	}
}

func TestClient_SetNamespace(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
    namespace: initial-ns
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600); err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	client, err := NewClient(WithKubeconfig(kubeconfigPath))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client.SetNamespace("new-namespace")

	if client.CurrentNamespace() != "new-namespace" {
		t.Errorf("expected namespace 'new-namespace', got %q", client.CurrentNamespace())
	}
}

func TestClient_Clientset(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")

	kubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
	if err := os.WriteFile(kubeconfigPath, []byte(kubeconfig), 0600); err != nil {
		t.Fatalf("failed to write test kubeconfig: %v", err)
	}

	client, err := NewClient(WithKubeconfig(kubeconfigPath))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.Clientset() == nil {
		t.Error("expected non-nil clientset")
	}
}
