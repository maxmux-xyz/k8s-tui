# Plan: Basic Command Execution (Exec) Feature

**Date:** 2026-01-04
**Feature:** Phase 5 - Basic one-shot command execution in pods

## Overview

Implement basic command execution allowing users to type a command, run it in the selected pod, and see stdout/stderr output. This follows the existing patterns established by log streaming.

## Files to Create

### 1. `internal/k8s/exec.go`
K8s exec functionality using client-go remotecommand.

```go
// ExecOptions configures command execution
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

// Exec executes a command in a pod and returns the result
func (c *Client) Exec(ctx context.Context, opts ExecOptions) ExecResult
```

Implementation approach:
- Use `remotecommand.NewSPDYExecutor` with REST config
- Build POST request to pod's exec subresource
- Capture stdout/stderr in bytes.Buffer
- Use `StreamWithContext` (not deprecated `Stream`)
- Return combined result struct

### 2. `internal/k8s/exec_test.go`
Tests for exec functionality.
- Test ExecOptions validation
- Test ExecResult struct
- Note: Full exec testing requires integration tests (difficult to mock SPDY)

### 3. `internal/ui/exec.go`
UI component for command execution view.

```go
// ExecViewState represents the state of the exec view
type ExecViewState int
const (
    ExecViewStateIdle ExecViewState = iota
    ExecViewStateRunning
    ExecViewStateComplete
    ExecViewStateError
)

// ExecViewModel manages the command execution UI
type ExecViewModel struct {
    input      textinput.Model  // Command input field
    output     viewport.Model   // Output display
    outputLines []string        // Output buffer
    state      ExecViewState
    // ... pod info, dimensions, etc.
}
```

Features:
- Text input for command entry (using bubbles/textinput)
- Scrollable output viewport (like log view)
- State indicator (idle/running/complete/error)
- Command history (up/down arrows) - store last N commands

### 4. `internal/ui/exec_test.go`
Tests for exec UI component.
- Test initialization
- Test command input handling
- Test output display
- Test state transitions

## Files to Modify

### 5. `internal/app/app.go`
Add exec integration to main app.

**New message types:**
```go
type execStartedMsg struct{}
type execOutputMsg struct {
    result k8s.ExecResult
}
type execErrorMsg struct {
    err error
}
```

**New model fields:**
```go
execView      ui.ExecViewModel
execCancel    context.CancelFunc
```

**New methods:**
- `initExec(command string) tea.Cmd` - Start command execution
- `handleExecViewKeys(msg tea.KeyMsg)` - Handle input in exec view
- Update `viewExec()` - Render ExecViewModel instead of placeholder

**Update flow:**
1. User presses 'e' on pod -> switch to ViewExec
2. User types command, presses Enter -> initExec() runs command
3. execOutputMsg received -> display result in viewport
4. User can run more commands or press Esc to return

### 6. `internal/app/app_test.go`
Add tests for exec flow.
- Test transition to exec view
- Test exec view key handling

## Implementation Steps

### Step 1: K8s exec layer
1. Create `internal/k8s/exec.go` with ExecOptions, ExecResult, Exec method
2. Create `internal/k8s/exec_test.go` with basic tests
3. Run `make test` to verify

### Step 2: UI component
1. Create `internal/ui/exec.go` with ExecViewModel
2. Create `internal/ui/exec_test.go`
3. Run `make test` to verify

### Step 3: App integration
1. Add message types to app.go
2. Add execView field to Model
3. Implement initExec() command
4. Implement handleExecViewKeys()
5. Update viewExec() to render component
6. Add tests to app_test.go
7. Run `make test && make lint` to verify

### Step 4: Manual testing
1. Run app with real cluster
2. Select pod, press 'e'
3. Type commands like `ls`, `pwd`, `env`
4. Verify output displays correctly
5. Verify error handling (bad command, container not found)

## Keybindings (Exec View)

| Key | Action |
|-----|--------|
| `Enter` | Execute command |
| `Up/Down` | Command history |
| `j/k` | Scroll output |
| `g/G` | Top/bottom of output |
| `Esc` | Back to pod list |

## Dependencies

No new dependencies needed - already have:
- `github.com/charmbracelet/bubbles` (textinput, viewport)
- `k8s.io/client-go` (remotecommand)

## Out of Scope

- Interactive shell (PTY-based)
- Streaming output (long-running commands)
- Tab completion
- Multi-container selection in exec view (use first container like logs)
