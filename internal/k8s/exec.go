package k8s

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecOptions configures command execution in a pod
type ExecOptions struct {
	Namespace string
	Pod       string
	Container string
	Command   []string
}

// ExecResult holds the output of a command execution
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// Validate checks that the exec options are valid
func (o ExecOptions) Validate() error {
	if o.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if o.Pod == "" {
		return fmt.Errorf("pod name is required")
	}
	if len(o.Command) == 0 {
		return fmt.Errorf("command is required")
	}
	return nil
}

// CommandString returns the command as a single string for display
func (o ExecOptions) CommandString() string {
	return strings.Join(o.Command, " ")
}

// Exec executes a command in a pod and returns the result.
// This is a synchronous operation - it blocks until the command completes.
func (c *Client) Exec(ctx context.Context, opts ExecOptions) ExecResult {
	if err := opts.Validate(); err != nil {
		return ExecResult{Error: err}
	}

	// Build the exec request
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(opts.Pod).
		Namespace(opts.Namespace).
		SubResource("exec")

	// Set up exec options
	execOpts := &corev1.PodExecOptions{
		Container: opts.Container,
		Command:   opts.Command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}

	req.VersionedParams(execOpts, scheme.ParameterCodec)

	// Create the executor
	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return ExecResult{Error: fmt.Errorf("failed to create executor: %w", err)}
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer

	// Execute the command
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	result := ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		// Try to extract exit code from error message
		result.Error = err
		result.ExitCode = 1 // Default non-zero exit code for errors
	}

	return result
}

// ParseCommand splits a command string into arguments.
// It handles basic quoting with double quotes.
func ParseCommand(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}

	var args []string
	var current strings.Builder
	inQuotes := false

	for _, r := range cmd {
		switch {
		case r == '"':
			inQuotes = !inQuotes
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
