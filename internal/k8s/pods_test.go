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

func createTestPod(name, namespace string, phase corev1.PodPhase, ready bool) *corev1.Pod {
	now := time.Now()

	containerState := corev1.ContainerState{}
	if ready {
		containerState.Running = &corev1.ContainerStateRunning{StartedAt: metav1.Time{Time: now}}
	} else {
		containerState.Waiting = &corev1.ContainerStateWaiting{Reason: "ContainerCreating"}
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: now.Add(-1 * time.Hour)},
		},
		Spec: corev1.PodSpec{
			NodeName: "node-1",
			Containers: []corev1.Container{
				{Name: "main"},
			},
		},
		Status: corev1.PodStatus{
			Phase: phase,
			PodIP: "10.0.0.1",
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "main",
					Ready:        ready,
					RestartCount: 0,
					State:        containerState,
				},
			},
		},
	}
}

func TestClient_ListPods(t *testing.T) {
	pods := []runtime.Object{
		createTestPod("pod-alpha", "default", corev1.PodRunning, true),
		createTestPod("pod-beta", "default", corev1.PodRunning, true),
		createTestPod("pod-gamma", "kube-system", corev1.PodRunning, true),
	}

	fakeClient := fake.NewClientset(pods...)

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()

	// List pods in default namespace
	result, err := client.ListPods(ctx, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 pods in default namespace, got %d", len(result))
	}

	// Verify sorted order
	if result[0].Name != "pod-alpha" || result[1].Name != "pod-beta" {
		t.Error("pods not in expected sorted order")
	}

	// List pods in kube-system namespace
	result, err = client.ListPods(ctx, "kube-system")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 pod in kube-system, got %d", len(result))
	}
}

func TestClient_ListPods_UsesCurrentNamespace(t *testing.T) {
	pods := []runtime.Object{
		createTestPod("pod-1", "my-namespace", corev1.PodRunning, true),
	}

	fakeClient := fake.NewClientset(pods...)

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "my-namespace",
	}

	ctx := context.Background()

	// Empty namespace should use current namespace
	result, err := client.ListPods(ctx, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 pod, got %d", len(result))
	}
}

func TestClient_GetPod(t *testing.T) {
	pods := []runtime.Object{
		createTestPod("my-pod", "default", corev1.PodRunning, true),
	}

	fakeClient := fake.NewClientset(pods...)

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()

	pod, err := client.GetPod(ctx, "default", "my-pod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pod.Name != "my-pod" {
		t.Errorf("expected name 'my-pod', got %q", pod.Name)
	}
	if pod.Namespace != "default" {
		t.Errorf("expected namespace 'default', got %q", pod.Namespace)
	}
	if pod.Status != PodStatusRunning {
		t.Errorf("expected status Running, got %q", pod.Status)
	}
}

func TestClient_GetPod_NotFound(t *testing.T) {
	fakeClient := fake.NewClientset()

	client := &Client{
		clientset:        fakeClient,
		currentNamespace: "default",
	}

	ctx := context.Background()

	_, err := client.GetPod(ctx, "default", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent pod")
	}
}

func TestPodStatus_Running(t *testing.T) {
	pod := createTestPod("test", "default", corev1.PodRunning, true)

	status, _ := determinePodStatus(pod)
	if status != PodStatusRunning {
		t.Errorf("expected Running status, got %q", status)
	}
}

func TestPodStatus_Pending(t *testing.T) {
	pod := createTestPod("test", "default", corev1.PodPending, false)

	status, _ := determinePodStatus(pod)
	if status != PodStatusPending {
		t.Errorf("expected Pending status, got %q", status)
	}
}

func TestPodStatus_Failed(t *testing.T) {
	pod := createTestPod("test", "default", corev1.PodFailed, false)

	status, _ := determinePodStatus(pod)
	if status != PodStatusFailed {
		t.Errorf("expected Failed status, got %q", status)
	}
}

func TestPodStatus_Succeeded(t *testing.T) {
	pod := createTestPod("test", "default", corev1.PodSucceeded, false)

	status, _ := determinePodStatus(pod)
	if status != PodStatusSucceeded {
		t.Errorf("expected Succeeded status, got %q", status)
	}
}

func TestPodStatus_Terminating(t *testing.T) {
	now := time.Now()
	pod := createTestPod("test", "default", corev1.PodRunning, true)
	pod.DeletionTimestamp = &metav1.Time{Time: now}

	status, _ := determinePodStatus(pod)
	if status != PodStatusTerminating {
		t.Errorf("expected Terminating status, got %q", status)
	}
}

func TestPodInfo_ReadyCount(t *testing.T) {
	now := time.Now()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "multi-container",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: now},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "container-1"},
				{Name: "container-2"},
				{Name: "container-3"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "container-1", Ready: true, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
				{Name: "container-2", Ready: true, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
				{Name: "container-3", Ready: false, State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}},
			},
		},
	}

	client := &Client{}
	info := client.podToInfo(pod)

	if info.Ready != "2/3" {
		t.Errorf("expected ready '2/3', got %q", info.Ready)
	}
	if info.ReadyCount != 2 {
		t.Errorf("expected ready count 2, got %d", info.ReadyCount)
	}
	if info.ContainerCount != 3 {
		t.Errorf("expected container count 3, got %d", info.ContainerCount)
	}
}

func TestPodInfo_Restarts(t *testing.T) {
	now := time.Now()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "restarts-pod",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: now},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "container-1"},
				{Name: "container-2"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "container-1", Ready: true, RestartCount: 5, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
				{Name: "container-2", Ready: true, RestartCount: 3, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
			},
		},
	}

	client := &Client{}
	info := client.podToInfo(pod)

	if info.Restarts != 8 {
		t.Errorf("expected total restarts 8, got %d", info.Restarts)
	}
}

func TestParseContainerState(t *testing.T) {
	tests := []struct {
		name          string
		state         corev1.ContainerState
		expectedState string
	}{
		{
			name:          "running",
			state:         corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
			expectedState: "Running",
		},
		{
			name:          "waiting",
			state:         corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"}},
			expectedState: "Waiting",
		},
		{
			name:          "terminated",
			state:         corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: "Completed"}},
			expectedState: "Terminated",
		},
		{
			name:          "unknown",
			state:         corev1.ContainerState{},
			expectedState: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, _ := parseContainerState(tt.state)
			if state != tt.expectedState {
				t.Errorf("expected state %q, got %q", tt.expectedState, state)
			}
		})
	}
}
