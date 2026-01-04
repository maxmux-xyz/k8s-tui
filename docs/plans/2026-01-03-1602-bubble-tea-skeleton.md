# Plan: Initialize K8s TUI Project with Bubble Tea Skeleton

## Overview

Create the foundational Go project structure with a minimal but complete Bubble Tea application. This establishes the Elm architecture (Model-Update-View) with proper view state management, keybindings, and comprehensive unit tests.

## Files to Create

```
/Users/maxime/dev/k8s-tui/
├── go.mod                      # Module definition
├── main.go                     # Entry point
├── Makefile                    # Build commands
├── internal/
│   ├── model/
│   │   ├── types.go            # ViewState enum
│   │   └── types_test.go       # Tests
│   ├── ui/
│   │   ├── keys.go             # Keybindings
│   │   └── keys_test.go        # Tests
│   └── app/
│       ├── app.go              # Main application model
│       └── app_test.go         # Tests
```

## Implementation Steps

### 1. Initialize Go Module
- Create `go.mod` with module `github.com/maxime/k8s-tui`
- Dependencies: bubbletea, bubbles, lipgloss

### 2. Create Model Types (`internal/model/types.go`)
- Define `ViewState` enum with all views:
  - `ViewPodList` (main view)
  - `ViewLogs`, `ViewExec`, `ViewFiles` (main views)
  - `ViewNamespaceSelector`, `ViewContextSelector`, `ViewHelp` (overlays)
- Add `String()` and `IsOverlay()` methods
- Write tests for all view states

### 3. Create Keybindings (`internal/ui/keys.go`)
- Define `KeyMap` struct with all bindings:
  - Navigation: Up (k/↑), Down (j/↓), Enter
  - Actions: Logs (l), Exec (e), Files (f), Refresh (r)
  - Selectors: Namespace (n), Context (c)
  - General: Help (?), Back (esc), Quit (q)
- Implement `ShortHelp()` and `FullHelp()` for help component
- Write tests for key definitions

### 4. Create Main App Model (`internal/app/app.go`)
- Implement `tea.Model` interface (Init, Update, View)
- Handle `WindowSizeMsg` for responsive layout
- Handle key presses for view navigation
- Placeholder view methods for each view state
- Export getters for testing (`CurrentView()`, `IsReady()`, `ShowingHelp()`)
- Write comprehensive tests for:
  - Initial state
  - Window resize handling
  - Quit command
  - Help toggle
  - View navigation (l, e, f, n, c keys)
  - Back navigation (esc key)

### 5. Create Entry Point (`main.go`)
- Initialize tea.Program with `tea.WithAltScreen()`
- Handle run errors

### 6. Create Makefile
- Targets: `build`, `run`, `test`, `test-coverage`, `clean`, `deps`, `verify`

## Verification Checklist

- [ ] `go mod tidy` runs without errors
- [ ] `go build` compiles without errors
- [ ] `go test ./...` passes all tests (target >80% coverage)
- [ ] Application starts and displays placeholder UI
- [ ] `q` quits cleanly
- [ ] `?` toggles help view
- [ ] `l`, `e`, `f`, `n`, `c` navigate to views
- [ ] `esc` navigates back

## Explicitly Out of Scope

- No Kubernetes client or API calls
- No styled components (lipgloss styling)
- No actual UI components (table, viewport, text input)
- No real content in views (placeholders only)
- No log streaming, exec, or file browsing
- No auto-refresh or search/filter
- No CI/CD or release configuration
