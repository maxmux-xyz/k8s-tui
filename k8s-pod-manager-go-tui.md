# K8s Pod Manager - Go TUI Implementation Plan

## Overview

Replace the Streamlit-based `backend/scripts/k8s_pod_manager.py` with a standalone Go terminal application using Bubble Tea. This addresses Streamlit's limitations: random refreshes, clunky state management, and inability to stream logs.

## Goals

1. **Live log streaming** - `kubectl logs -f` with real-time output
2. **Responsive UI** - no page refreshes, instant feedback
3. **Single binary** - no Python/deps, easy distribution
4. **Keyboard-driven** - vim-like navigation, efficient workflow

## Tech Stack

| Component | Library | Purpose |
|-----------|---------|---------|
| TUI Framework | [bubbletea](https://github.com/charmbracelet/bubbletea) | Elm-architecture TUI |
| Styling | [lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| Tables | [bubbles/table](https://github.com/charmbracelet/bubbles) | Pod list display |
| Text Input | [bubbles/textinput](https://github.com/charmbracelet/bubbles) | Command input, path input |
| Viewport | [bubbles/viewport](https://github.com/charmbracelet/bubbles) | Scrollable log viewer |
| K8s Client | [client-go](https://github.com/kubernetes/client-go) | Native k8s API access |

## UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ K8s Pod Manager    Context: prod-cluster    Namespace: temporal │
├─────────────────────────────┬───────────────────────────────────┤
│ PODS                        │ LOGS (streaming)                  │
│ ─────────────────────────── │ ─────────────────────────────────│
│ > worker-abc-123   Running  │ 2024-01-03 10:00:01 Starting...  │
│   worker-def-456   Running  │ 2024-01-03 10:00:02 Connected    │
│   api-server-789   Running  │ 2024-01-03 10:00:03 Processing   │
│   scheduler-012    Pending  │ 2024-01-03 10:00:04 Done         │
│                             │ ...                               │
│                             │                                   │
├─────────────────────────────┴───────────────────────────────────┤
│ [l]ogs [e]xec [f]iles [n]amespace [c]ontext [r]efresh [q]uit   │
└─────────────────────────────────────────────────────────────────┘
```

### Views

1. **Main View** - Pod list + log viewer (split pane)
2. **File Browser View** - Directory listing with file preview
3. **Exec View** - Command input + output display
4. **Namespace Selector** - Fuzzy-searchable namespace list
5. **Context Selector** - Switch k8s contexts

## Features

### 1. Pod Management
- List pods in selected namespace
- Show status with color indicators (Running=green, Pending=yellow, Failed=red)
- Auto-refresh pod list every 10s (configurable)
- Manual refresh with `r` key

### 2. Log Streaming (Primary Feature)
- Stream logs with `kubectl logs -f` equivalent via client-go
- Scroll through log history
- Toggle follow mode (auto-scroll vs manual)
- Filter logs by search term
- Support multi-container pods (container selector)

### 3. Command Execution
- Execute arbitrary commands in pod
- Display stdout/stderr with color coding
- Command history (up/down arrows)
- Quick commands: shell (`/bin/sh`), env, ps

### 4. File Browser
- Navigate directories with arrow keys
- View file contents in scrollable viewport
- Show file metadata (size, permissions)
- Copy file path to clipboard

### 5. Context/Namespace Management
- Switch k8s contexts without leaving app
- Fuzzy search namespaces
- Remember last used namespace per context

## Implementation Steps

### Phase 1: Project Setup & Core Structure
**Files to create:**
```
tools/k8s-pod-manager/
├── main.go                 # Entry point
├── go.mod                  # Module definition
├── go.sum                  # Dependencies
├── internal/
│   ├── app/
│   │   └── app.go          # Main application model
│   ├── k8s/
│   │   ├── client.go       # K8s client wrapper
│   │   ├── pods.go         # Pod operations
│   │   ├── logs.go         # Log streaming
│   │   └── exec.go         # Command execution
│   ├── ui/
│   │   ├── styles.go       # Lipgloss styles
│   │   ├── keys.go         # Keybindings
│   │   ├── pods.go         # Pod list component
│   │   ├── logs.go         # Log viewer component
│   │   ├── files.go        # File browser component
│   │   └── exec.go         # Exec component
│   └── model/
│       └── types.go        # Shared types
└── Makefile                # Build commands
```

**Tasks:**
- [ ] Initialize Go module
- [ ] Set up Bubble Tea application skeleton
- [ ] Define main model with view state enum
- [ ] Create basic keybinding handler

### Phase 2: K8s Client Integration
**Tasks:**
- [ ] Initialize client-go with kubeconfig
- [ ] Implement context listing/switching
- [ ] Implement namespace listing
- [ ] Implement pod listing with status parsing
- [ ] Handle connection errors gracefully

### Phase 3: Pod List UI
**Tasks:**
- [ ] Create styled table for pod list
- [ ] Add status color indicators
- [ ] Implement keyboard navigation (j/k or arrows)
- [ ] Add namespace selector dropdown
- [ ] Implement auto-refresh with ticker

### Phase 4: Log Streaming (Core Feature)
**Tasks:**
- [ ] Implement log streaming via client-go PodLogs API
- [ ] Create scrollable viewport for logs
- [ ] Add follow mode toggle (F key)
- [ ] Support container selection for multi-container pods
- [ ] Add log search/filter (/ key)
- [ ] Handle stream reconnection on errors

### Phase 5: Command Execution
**Tasks:**
- [ ] Implement exec via client-go remotecommand
- [ ] Create input field for commands
- [ ] Display output in viewport
- [ ] Add command history
- [ ] Support interactive shell spawning (optional - may need pty)

### Phase 6: File Browser
**Tasks:**
- [ ] Implement directory listing via exec
- [ ] Create file list UI with icons
- [ ] Add navigation (enter=open, backspace=parent)
- [ ] Implement file content viewer
- [ ] Add path input for direct navigation

### Phase 7: Polish & Distribution
**Tasks:**
- [ ] Add help overlay (? key)
- [ ] Implement error toasts/notifications
- [ ] Add loading indicators
- [ ] Create Makefile with build targets
- [ ] Add goreleaser config for releases
- [ ] Write README with usage instructions

## Keybindings

| Key | Action |
|-----|--------|
| `j/↓` | Move down in list |
| `k/↑` | Move up in list |
| `Enter` | Select item / Open directory |
| `l` | View logs for selected pod |
| `e` | Exec into selected pod |
| `f` | Open file browser |
| `n` | Change namespace |
| `c` | Change context |
| `r` | Refresh pod list |
| `F` | Toggle follow mode (logs) |
| `/` | Search/filter |
| `Esc` | Back / Cancel |
| `q` | Quit |
| `?` | Show help |

## Data Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Bubble Tea │────▶│   K8s Client │────▶│  Kubernetes  │
│   (UI Loop)  │◀────│   (client-go)│◀────│    API       │
└──────────────┘     └──────────────┘     └──────────────┘
       │
       │ Msg/Cmd
       ▼
┌──────────────┐
│    Model     │
│  - pods[]    │
│  - logs[]    │
│  - view      │
└──────────────┘
```

## Log Streaming Implementation

```go
// Simplified log streaming approach
func (c *Client) StreamLogs(ctx context.Context, namespace, pod, container string) (<-chan string, error) {
    req := c.clientset.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{
        Follow:    true,
        Container: container,
    })

    stream, err := req.Stream(ctx)
    if err != nil {
        return nil, err
    }

    lines := make(chan string)
    go func() {
        defer close(lines)
        defer stream.Close()
        scanner := bufio.NewScanner(stream)
        for scanner.Scan() {
            lines <- scanner.Text()
        }
    }()

    return lines, nil
}
```

## Build & Run

```bash
# Development
cd tools/k8s-pod-manager
go run .

# Build binary
go build -o kpm .

# Install globally
go install .

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o kpm-linux-amd64 .
GOOS=darwin GOARCH=arm64 go build -o kpm-darwin-arm64 .
```

## Dependencies

```go
// go.mod
module github.com/nebari/k8s-pod-manager

go 1.22

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.9.1
    k8s.io/client-go v0.29.0
    k8s.io/api v0.29.0
)
```

## Success Criteria

1. **Log streaming works flawlessly** - no lag, no disconnects, scrollable history
2. **Instant response** - no UI freezes or delays
3. **Single binary** - `./kpm` just works with existing kubeconfig
4. **Intuitive navigation** - vim users feel at home
5. **Handles errors gracefully** - connection issues shown clearly, auto-reconnect

## Future Enhancements (Out of Scope)

- Multiple pod log streaming (tail multiple pods)
- Log export to file
- Pod YAML viewer/editor
- Resource usage graphs (CPU/memory)
- Port forwarding management
- Deployment scaling controls
