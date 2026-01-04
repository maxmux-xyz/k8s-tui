# Log Viewport Update Fix

**Date:** 2026-01-04

## Problem

Log streaming was working correctly (logs were being received from Kubernetes), but no content was displayed in the viewport. The status bar showed `[STREAMING]` and the correct line count (e.g., `Lines: 88`), but the viewport area remained blank.

## Root Cause

The `LogViewModel.AddLine()` method was setting `contentDirty = true` but not actually updating the viewport content. The `updateViewportContent()` method was only called in `LogViewModel.Update()`, which is triggered by key presses.

Flow before fix:
1. `logLineMsg` received in `app.go`
2. `m.logView.AddLine(msg.line.Content)` called
3. `AddLine` appends to `m.lines` and sets `contentDirty = true`
4. Viewport content never updated until user presses a key

This meant logs were being collected but never rendered to the screen.

## Fix

Modified `AddLine()` in `internal/ui/logs.go` to call `updateViewportContent()` immediately after adding a line:

```go
func (m *LogViewModel) AddLine(line string) {
    m.lines = append(m.lines, line)

    // Trim old lines if we exceed max
    if len(m.lines) > m.maxLines {
        trimCount := m.maxLines / 10
        m.lines = m.lines[trimCount:]
    }

    m.contentDirty = true
    // Update viewport immediately so new content is visible
    m.updateViewportContent()
}
```

## Files Changed

- `internal/ui/logs.go` - Added `updateViewportContent()` call in `AddLine()`

## Debugging

Added temporary debug logging to `/tmp/k8s-tui-debug.log` to trace the issue:
- Confirmed `logStreamActive=true` and logs were being received
- Confirmed the streaming loop was continuing correctly
- Identified that viewport update was the missing piece

Debug logging code in `internal/app/app.go` should be removed after confirming the fix works in production.
