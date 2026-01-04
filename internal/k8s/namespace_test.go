package k8s

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClient_ListNamespaces(t *testing.T) {
	// Create fake clientset with test namespaces
	now := time.Now()
	namespaces := []runtime.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "default",
				CreationTimestamp: metav1.Time{Time: now.Add(-24 * time.Hour)},
			},
			Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "kube-system",
				CreationTimestamp: metav1.Time{Time: now.Add(-48 * time.Hour)},
			},
			Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "production",
				CreationTimestamp: metav1.Time{Time: now.Add(-12 * time.Hour)},
			},
			Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
	}

	fakeClient := fake.NewClientset(namespaces...)

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()
	result, err := client.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 namespaces, got %d", len(result))
	}

	// Verify sorted order
	expectedOrder := []string{"default", "kube-system", "production"}
	for i, ns := range result {
		if ns.Name != expectedOrder[i] {
			t.Errorf("expected namespace %d to be %q, got %q", i, expectedOrder[i], ns.Name)
		}
	}

	// Verify current namespace is marked
	for _, ns := range result {
		if ns.Name == "default" {
			if !ns.IsCurrent {
				t.Error("expected 'default' to be marked as current")
			}
		} else {
			if ns.IsCurrent {
				t.Errorf("expected %q to not be marked as current", ns.Name)
			}
		}
	}

	// Verify status
	for _, ns := range result {
		if ns.Status != "Active" {
			t.Errorf("expected status 'Active', got %q", ns.Status)
		}
	}

	// Verify ages are reasonable (within a small tolerance)
	for _, ns := range result {
		if ns.Age < 0 {
			t.Errorf("expected positive age, got %v", ns.Age)
		}
	}
}

func TestClient_ListNamespaces_Empty(t *testing.T) {
	fakeClient := fake.NewClientset()

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()
	result, err := client.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 namespaces, got %d", len(result))
	}
}

func TestClient_NamespaceExists(t *testing.T) {
	namespaces := []runtime.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "existing-ns"},
		},
	}

	fakeClient := fake.NewClientset(namespaces...)

	client := &Client{
		clientset: fakeClient,
	}

	ctx := context.Background()

	// Test existing namespace
	exists, err := client.NamespaceExists(ctx, "existing-ns")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected namespace to exist")
	}

	// Test non-existing namespace
	exists, err = client.NamespaceExists(ctx, "non-existing-ns")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected namespace to not exist")
	}
}

func TestClient_namespacesToInfo_MarksCurrentNamespace(t *testing.T) {
	now := time.Now()
	namespaces := []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "ns-a",
				CreationTimestamp: metav1.Time{Time: now},
			},
			Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "ns-b",
				CreationTimestamp: metav1.Time{Time: now},
			},
			Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
		},
	}

	client := &Client{currentNamespace: "ns-b"}

	result := client.namespacesToInfo(namespaces)

	for _, ns := range result {
		if ns.Name == "ns-b" && !ns.IsCurrent {
			t.Error("expected ns-b to be marked as current")
		}
		if ns.Name == "ns-a" && ns.IsCurrent {
			t.Error("expected ns-a to not be marked as current")
		}
	}
}
