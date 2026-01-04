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

func createTestPodWithContainers(name, namespace string, containers []string) *corev1.Pod {
	now := time.Now()

	containerSpecs := make([]corev1.Container, 0, len(containers))
	containerStatuses := make([]corev1.ContainerStatus, 0, len(containers))

	for _, c := range containers {
		containerSpecs = append(containerSpecs, corev1.Container{Name: c})
		containerStatuses = append(containerStatuses, corev1.ContainerStatus{
			Name:  c,
			Ready: true,
			State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
		})
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: now.Add(-1 * time.Hour)},
		},
		Spec: corev1.PodSpec{
			NodeName:   "node-1",
			Containers: containerSpecs,
		},
		Status: corev1.PodStatus{
			Phase:             corev1.PodRunning,
			PodIP:             "10.0.0.1",
			ContainerStatuses: containerStatuses,
		},
	}
}

func TestClient_GetContainers(t *testing.T) {
	pods := []runtime.Object{
		createTestPodWithContainers("multi-container-pod", "default", []string{"app", "sidecar", "init"}),
	}

	fakeClient := fake.NewClientset(pods...)

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()

	containers, err := client.GetContainers(ctx, "default", "multi-container-pod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containers) != 3 {
		t.Errorf("expected 3 containers, got %d", len(containers))
	}

	// Check that all expected containers are present
	expected := map[string]bool{"app": true, "sidecar": true, "init": true}
	for _, c := range containers {
		if !expected[c] {
			t.Errorf("unexpected container %q", c)
		}
	}
}

func TestClient_GetContainers_SingleContainer(t *testing.T) {
	pods := []runtime.Object{
		createTestPodWithContainers("single-container-pod", "default", []string{"main"}),
	}

	fakeClient := fake.NewClientset(pods...)

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()

	containers, err := client.GetContainers(ctx, "default", "single-container-pod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(containers))
	}

	if containers[0] != "main" {
		t.Errorf("expected container 'main', got %q", containers[0])
	}
}

func TestClient_GetContainers_UsesCurrentNamespace(t *testing.T) {
	pods := []runtime.Object{
		createTestPodWithContainers("my-pod", "my-namespace", []string{"app"}),
	}

	fakeClient := fake.NewClientset(pods...)

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "my-namespace",
	}

	ctx := context.Background()

	// Empty namespace should use current namespace
	containers, err := client.GetContainers(ctx, "", "my-pod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(containers))
	}
}

func TestClient_GetContainers_PodNotFound(t *testing.T) {
	fakeClient := fake.NewClientset()

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()

	_, err := client.GetContainers(ctx, "default", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent pod")
	}
}

func TestClient_GetFirstContainer(t *testing.T) {
	pods := []runtime.Object{
		createTestPodWithContainers("my-pod", "default", []string{"first", "second"}),
	}

	fakeClient := fake.NewClientset(pods...)

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()

	container, err := client.GetFirstContainer(ctx, "default", "my-pod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if container != "first" {
		t.Errorf("expected 'first', got %q", container)
	}
}

func TestClient_GetFirstContainer_PodNotFound(t *testing.T) {
	fakeClient := fake.NewClientset()

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()

	_, err := client.GetFirstContainer(ctx, "default", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent pod")
	}
}

func TestLogOptions_Defaults(t *testing.T) {
	opts := LogOptions{
		Namespace: "default",
		Pod:       "my-pod",
		Container: "main",
	}

	if opts.Follow != false {
		t.Error("expected Follow to default to false")
	}

	if opts.TailLines != 0 {
		t.Error("expected TailLines to default to 0")
	}

	if opts.Timestamps != false {
		t.Error("expected Timestamps to default to false")
	}

	if opts.SinceTime != nil {
		t.Error("expected SinceTime to default to nil")
	}
}

func TestLogLine_Error(t *testing.T) {
	line := LogLine{
		Error: context.Canceled,
	}

	if line.Error == nil {
		t.Error("expected error to be set")
	}

	if line.Content != "" {
		t.Error("expected content to be empty when error is set")
	}
}

// Note: StreamLogs is difficult to unit test with the fake clientset because
// it doesn't support the GetLogs().Stream() API. Integration tests should be
// used to verify streaming behavior against a real cluster.
//
// The streaming implementation:
// 1. Opens a log stream using the K8s API
// 2. Reads lines from the stream in a goroutine
// 3. Sends lines to a channel
// 4. Closes the channel when the stream ends or context is cancelled
// 5. Handles errors by sending a LogLine with Error field set
