package k8s

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LogLine represents a single line of log output
type LogLine struct {
	Content   string
	Timestamp time.Time
	Error     error
}

// LogOptions configures how logs are streamed
type LogOptions struct {
	Namespace  string
	Pod        string
	Container  string
	Follow     bool
	TailLines  int64
	Timestamps bool
	SinceTime  *time.Time
}

// StreamLogs streams logs from a pod container and sends them to a channel.
// The channel is closed when the stream ends or the context is cancelled.
func (c *Client) StreamLogs(ctx context.Context, opts LogOptions) (<-chan LogLine, error) {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = c.currentNamespace
	}

	// Build pod log options
	podLogOpts := &corev1.PodLogOptions{
		Container:  opts.Container,
		Follow:     opts.Follow,
		Timestamps: opts.Timestamps,
	}

	if opts.TailLines > 0 {
		podLogOpts.TailLines = &opts.TailLines
	}

	if opts.SinceTime != nil {
		sinceTime := opts.SinceTime
		podLogOpts.SinceTime = &metav1.Time{Time: *sinceTime}
	}

	// Get the log stream
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(opts.Pod, podLogOpts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open log stream for pod %q: %w", opts.Pod, err)
	}

	// Create the output channel
	logChan := make(chan LogLine, 100)

	// Start goroutine to read from stream and send to channel
	go func() {
		defer close(logChan)
		defer stream.Close()

		reader := bufio.NewReader(stream)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						// Send any remaining content
						if line != "" {
							logChan <- LogLine{Content: line, Timestamp: time.Now()}
						}
						return
					}
					// Send error and exit
					logChan <- LogLine{Error: fmt.Errorf("error reading log stream: %w", err)}
					return
				}

				// Remove trailing newline
				if len(line) > 0 && line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
				}

				logChan <- LogLine{Content: line, Timestamp: time.Now()}
			}
		}
	}()

	return logChan, nil
}

// GetContainers returns the list of containers in a pod
func (c *Client) GetContainers(ctx context.Context, namespace, pod string) ([]string, error) {
	if namespace == "" {
		namespace = c.currentNamespace
	}

	podInfo, err := c.GetPod(ctx, namespace, pod)
	if err != nil {
		return nil, err
	}

	containers := make([]string, 0, len(podInfo.Containers))
	for _, c := range podInfo.Containers {
		containers = append(containers, c.Name)
	}

	return containers, nil
}

// GetFirstContainer returns the name of the first container in a pod
func (c *Client) GetFirstContainer(ctx context.Context, namespace, pod string) (string, error) {
	containers, err := c.GetContainers(ctx, namespace, pod)
	if err != nil {
		return "", err
	}

	if len(containers) == 0 {
		return "", fmt.Errorf("no containers found in pod %q", pod)
	}

	return containers[0], nil
}
