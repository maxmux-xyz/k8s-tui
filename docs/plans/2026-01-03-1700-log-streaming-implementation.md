# Plan: Log Streaming Implementation (Phase 4)

**Date:** 2026-01-03 17:00
**Status:** Draft

## Current State Summary

### Completed (Phases 1-2)
1. **Bubble Tea Skeleton** - Complete application structure with Elm architecture
2. **K8s Client Integration** - Full client-go integration with:
   - Kubeconfig loading
   - Context listing/switching
   - Namespace listing/switching
   - Pod listing with status parsing
   - Async loading with proper state management

### Current Capabilities
- View pods in any namespace
- Navigate pods with j/k keys
- Switch namespaces (n) and contexts (c)
- Refresh pod list (r)
- Basic placeholder views for logs/exec/files

### Missing Per Original Spec (Phase 3-7)
- **Phase 3: Pod List UI** - Styled table with lipgloss (partially done - functional but unstyled)
- **Phase 4: Log Streaming** - Core feature, not started
- **Phase 5: Command Execution** - Not started
- **Phase 6: File Browser** - Not started
- **Phase 7: Polish & Distribution** - Not started

## Recommended Next Task: Log Streaming (Phase 4)

**Why Log Streaming Next:**
1. **Primary Feature** - The spec calls it "Primary Feature" - it's the main reason for building this tool
2. **Self-Contained** - Can be implemented without touching the pod list styling
3. **High Value** - Users get immediate value from being able to stream logs
4. **Well-Defined** - Clear API (client-go PodLogs), clear UI (viewport)
5. **Testable** - Can mock the log stream for unit tests

## Implementation Plan

### Task 1: Log Streaming Core (Self-Contained)

**Goal:** Stream logs from a selected pod and display them in a scrollable viewport.

**Files to Create:**
```
internal/k8s/
└── logs.go           # Log streaming via client-go
└── logs_test.go      # Tests with mocked streams

internal/ui/
└── logs.go           # Log viewport component
└── logs_test.go
```

**Files to Modify:**
```
internal/app/app.go   # Wire up log view with real streaming
```

**Scope:**
1. Implement `StreamLogs(ctx, namespace, pod, container)` returning `<-chan string`
2. Handle stream errors and reconnection
3. Create viewport component using bubbles/viewport
4. Add follow mode toggle (F key)
5. Support scrolling through log history

**Implementation Details:**

```go
// internal/k8s/logs.go
type LogStreamer struct {
    clientset kubernetes.Interface
}

type LogOptions struct {
    Namespace   string
    Pod         string
    Container   string
    Follow      bool
    TailLines   int64
    SinceTime   *time.Time
}

func (l *LogStreamer) StreamLogs(ctx context.Context, opts LogOptions) (<-chan LogLine, error)
func (l *LogStreamer) GetContainers(ctx context.Context, namespace, pod string) ([]string, error)
```

**Bubble Tea Integration:**
- New message type: `logLineMsg string`
- New message type: `logStreamErrorMsg error`
- New message type: `logStreamEndedMsg struct{}`
- Command to start streaming: `func (m Model) startLogStream() tea.Cmd`
- Handle incoming log lines in Update()

**UI Component:**
- Use `bubbles/viewport` for scrollable content
- Header showing pod name and container
- Status line showing stream state (streaming/paused/error)
- Follow mode indicator

**Keybindings for Log View:**
| Key | Action |
|-----|--------|
| `F` | Toggle follow mode |
| `j/↓` | Scroll down |
| `k/↑` | Scroll up |
| `g` | Go to top |
| `G` | Go to bottom |
| `esc` | Back to pod list |
| `tab` | Switch container (multi-container) |

### Task 2: Multi-Container Support

**Scope:**
- Detect multi-container pods
- Show container selector when pressing `l` on multi-container pod
- Remember selected container per pod

### Task 3: Log Search/Filter

**Scope:**
- Add `/` key to enter search mode
- Filter displayed lines by search term
- Highlight matches

## Recommended First Sub-Task: Task 1 (Log Streaming Core)

This is self-contained and delivers the core value. Tasks 2-3 are enhancements.

**Estimated Scope for Task 1:**
- ~150-200 lines of K8s code + ~100 lines of tests
- ~150-200 lines of UI code + ~100 lines of tests
- ~50 lines of app.go modifications

**Verification:**
1. Unit tests pass (mocked streams)
2. Can stream logs from a real pod in a cluster
3. Viewport scrolls correctly
4. Follow mode works (auto-scroll on new lines)
5. Esc returns to pod list cleanly
6. Stream cancellation works (no goroutine leaks)

## Testing Strategy

### Unit Tests
- Mock `kubernetes.Interface` for log streaming
- Test LogLine parsing
- Test stream cancellation
- Test viewport updates with log messages

### Integration Tests (Optional)
- Use `//go:build integration` tag
- Test against real cluster
- Verify stream reconnection on errors

## Out of Scope

- Styled pod list (can be a separate task)
- Exec functionality
- File browser
- Log export to file
- Multiple pod log streaming

## Dependencies

Already have all required dependencies (client-go, bubbles).

## Success Criteria

1. Press `l` on a pod -> see real-time streaming logs
2. Can scroll through log history
3. Follow mode auto-scrolls
4. Esc cleanly returns to pod list
5. No goroutine leaks when switching views
6. Error handling for stream failures
