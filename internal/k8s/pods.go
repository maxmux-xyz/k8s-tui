package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodStatus represents the high-level status of a pod
type PodStatus string

// Pod status constants represent the high-level status of a pod.
const (
	PodStatusRunning     PodStatus = "Running"
	PodStatusPending     PodStatus = "Pending"
	PodStatusSucceeded   PodStatus = "Succeeded"
	PodStatusFailed      PodStatus = "Failed"
	PodStatusUnknown     PodStatus = "Unknown"
	PodStatusTerminating PodStatus = "Terminating"
)

// ContainerStatus represents the status of a container within a pod
type ContainerStatus struct {
	Name         string
	Ready        bool
	RestartCount int32
	State        string // Running, Waiting, Terminated
	StateReason  string // Reason for Waiting/Terminated state
}

// PodInfo contains information about a Kubernetes pod
type PodInfo struct {
	Name           string
	Namespace      string
	Status         PodStatus
	StatusMessage  string // Additional status info (e.g., reason for failure)
	Ready          string // e.g., "2/3"
	Restarts       int32
	Age            time.Duration
	IP             string
	Node           string
	Containers     []ContainerStatus
	ContainerCount int
	ReadyCount     int
}

// ListPods returns pods in the specified namespace (or current namespace if empty)
func (c *Client) ListPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %q: %w", namespace, err)
	}

	return c.podsToInfo(pods.Items), nil
}

// GetPod returns information about a specific pod
func (c *Client) GetPod(ctx context.Context, namespace, name string) (*PodInfo, error) {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %q in namespace %q: %w", name, namespace, err)
	}

	info := c.podToInfo(pod)
	return &info, nil
}

// podsToInfo converts pod objects to PodInfo slice
func (c *Client) podsToInfo(pods []corev1.Pod) []PodInfo {
	result := make([]PodInfo, 0, len(pods))

	for i := range pods {
		result = append(result, c.podToInfo(&pods[i]))
	}

	// Sort by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// podToInfo converts a single pod to PodInfo
func (c *Client) podToInfo(pod *corev1.Pod) PodInfo {
	now := time.Now()
	age := now.Sub(pod.CreationTimestamp.Time)

	// Parse container statuses
	containers, readyCount, totalRestarts := parseContainerStatuses(pod)

	// Determine pod status
	status, statusMessage := determinePodStatus(pod)

	return PodInfo{
		Name:           pod.Name,
		Namespace:      pod.Namespace,
		Status:         status,
		StatusMessage:  statusMessage,
		Ready:          fmt.Sprintf("%d/%d", readyCount, len(containers)),
		Restarts:       totalRestarts,
		Age:            age,
		IP:             pod.Status.PodIP,
		Node:           pod.Spec.NodeName,
		Containers:     containers,
		ContainerCount: len(containers),
		ReadyCount:     readyCount,
	}
}

// parseContainerStatuses extracts container status info from a pod
func parseContainerStatuses(pod *corev1.Pod) ([]ContainerStatus, int, int32) {
	containers := make([]ContainerStatus, 0, len(pod.Spec.Containers))
	var readyCount int
	var totalRestarts int32

	// Map of container names to their spec (for init containers vs regular containers)
	containerNames := make(map[string]bool)
	for i := range pod.Spec.Containers {
		containerNames[pod.Spec.Containers[i].Name] = true
	}

	for i := range pod.Status.ContainerStatuses {
		cs := &pod.Status.ContainerStatuses[i]
		state, reason := parseContainerState(cs.State)

		containers = append(containers, ContainerStatus{
			Name:         cs.Name,
			Ready:        cs.Ready,
			RestartCount: cs.RestartCount,
			State:        state,
			StateReason:  reason,
		})

		if cs.Ready {
			readyCount++
		}
		totalRestarts += cs.RestartCount
	}

	// If no status yet, create entries from spec
	if len(containers) == 0 {
		for i := range pod.Spec.Containers {
			containers = append(containers, ContainerStatus{
				Name:  pod.Spec.Containers[i].Name,
				Ready: false,
				State: "Waiting",
			})
		}
	}

	return containers, readyCount, totalRestarts
}

// parseContainerState determines the state of a container
func parseContainerState(state corev1.ContainerState) (string, string) {
	if state.Running != nil {
		return "Running", ""
	}
	if state.Waiting != nil {
		return "Waiting", state.Waiting.Reason
	}
	if state.Terminated != nil {
		return "Terminated", state.Terminated.Reason
	}
	return "Unknown", ""
}

// determinePodStatus determines the high-level status of a pod
func determinePodStatus(pod *corev1.Pod) (PodStatus, string) {
	// Check if pod is being deleted
	if pod.DeletionTimestamp != nil {
		return PodStatusTerminating, ""
	}

	// Check pod phase
	switch pod.Status.Phase {
	case corev1.PodSucceeded:
		return PodStatusSucceeded, ""
	case corev1.PodFailed:
		return PodStatusFailed, getFailureReason(pod)
	case corev1.PodPending:
		reason := getPendingReason(pod)
		return PodStatusPending, reason
	case corev1.PodRunning:
		// Check if all containers are ready
		if areAllContainersReady(pod) {
			return PodStatusRunning, ""
		}
		// Running but not all containers ready
		reason := getNotReadyReason(pod)
		return PodStatusRunning, reason
	default:
		return PodStatusUnknown, string(pod.Status.Phase)
	}
}

// areAllContainersReady checks if all containers in a pod are ready
func areAllContainersReady(pod *corev1.Pod) bool {
	if len(pod.Status.ContainerStatuses) == 0 {
		return false
	}
	for i := range pod.Status.ContainerStatuses {
		if !pod.Status.ContainerStatuses[i].Ready {
			return false
		}
	}
	return true
}

// getFailureReason extracts the reason for pod failure
func getFailureReason(pod *corev1.Pod) string {
	if pod.Status.Reason != "" {
		return pod.Status.Reason
	}
	// Check container statuses for termination reason
	for i := range pod.Status.ContainerStatuses {
		cs := &pod.Status.ContainerStatuses[i]
		if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" {
			return cs.State.Terminated.Reason
		}
	}
	return ""
}

// getPendingReason extracts the reason why a pod is pending
func getPendingReason(pod *corev1.Pod) string {
	// Check conditions
	for i := range pod.Status.Conditions {
		condition := &pod.Status.Conditions[i]
		if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
			return condition.Reason
		}
	}
	// Check container waiting reasons
	for i := range pod.Status.ContainerStatuses {
		cs := &pod.Status.ContainerStatuses[i]
		if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
			return cs.State.Waiting.Reason
		}
	}
	return ""
}

// getNotReadyReason extracts the reason why containers aren't ready
func getNotReadyReason(pod *corev1.Pod) string {
	for i := range pod.Status.ContainerStatuses {
		cs := &pod.Status.ContainerStatuses[i]
		if !cs.Ready {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				return cs.State.Waiting.Reason
			}
		}
	}
	return ""
}
