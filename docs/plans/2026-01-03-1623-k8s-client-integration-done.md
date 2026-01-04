# Plan: K8s Client Integration (Phase 2)

**Date:** 2026-01-03 16:23
**Status:** Draft

## Overview

Integrate client-go to connect to Kubernetes clusters. This is foundational work that enables all K8s operations (pod listing, log streaming, exec, file browsing).

## Current State

- Bubble Tea skeleton complete with view states and keybindings
- No Kubernetes integration yet
- App displays placeholder UI only

## Phase 2 Goals

1. Initialize client-go with kubeconfig
2. List contexts and switch between them
3. List namespaces
4. List pods with status parsing
5. Handle connection errors gracefully

## Implementation Tasks

### Task 1: K8s Client Foundation (Self-Contained)

Create the K8s client wrapper with context and namespace support.

**Files to create:**
```
internal/k8s/
├── client.go       # Client wrapper with kubeconfig loading
├── client_test.go  # Tests with mocks
├── context.go      # Context listing/switching
├── context_test.go
├── namespace.go    # Namespace listing
└── namespace_test.go
```

**Scope:**
- Load kubeconfig from default location (`~/.kube/config`) or `KUBECONFIG` env
- List available contexts
- Get current context
- Switch context
- List namespaces in current context
- Handle errors (no kubeconfig, invalid config, connection refused)

**Tests:**
- Mock client-go interfaces for unit tests
- Test error handling paths
- Test context switching logic

**Dependencies to add:**
```go
k8s.io/client-go v0.29.0
k8s.io/api v0.29.0
```

### Task 2: Pod Listing

**Files to create:**
```
internal/k8s/
├── pods.go         # Pod listing with status
└── pods_test.go
```

**Scope:**
- List pods in a namespace
- Parse pod status (Running, Pending, Failed, etc.)
- Include container statuses
- Return structured pod data

### Task 3: Wire K8s Client to App

**Files to modify:**
```
internal/app/app.go  # Add K8s client, load data on init
```

**Scope:**
- Initialize K8s client in App.Init()
- Load current context/namespace
- Fetch pod list
- Handle loading states
- Display errors in UI

### Task 4: Real Pod List UI

**Files to create/modify:**
```
internal/ui/
├── styles.go       # Lipgloss styles (status colors)
├── pods.go         # Pod list component using bubbles/table
└── pods_test.go
```

**Scope:**
- Styled table for pod list
- Status color indicators (Running=green, Pending=yellow, Failed=red)
- Keyboard navigation (j/k)
- Selection highlighting

## Recommended Next Task: Task 1 - K8s Client Foundation

**Why this task first:**
1. **Self-contained**: No UI changes needed, pure Go/K8s work
2. **Testable**: Can fully test with mocks without real cluster
3. **Foundation**: All other K8s features depend on this
4. **Quick validation**: Can verify by printing contexts/namespaces

**Estimated scope:**
- ~200-300 lines of code
- ~150-200 lines of tests
- 3 new files + 3 test files

**Verification:**
- Unit tests pass with mocked client
- Integration test (optional) connects to real cluster
- `go build` succeeds with new dependencies

## Out of Scope for Task 1

- Pod listing (Task 2)
- UI integration (Task 3)
- Styled components (Task 4)
- Log streaming, exec, file browser (Phase 4-6)

## Testing Strategy

Since we can't assume a real K8s cluster is available during tests:

1. **Unit tests**: Mock `kubernetes.Interface` and test all logic
2. **Integration tests**: Use build tag `//go:build integration` for real cluster tests
3. **Test coverage target**: >80% for new code

## Dependencies

```go
// go.mod additions
require (
    k8s.io/client-go v0.29.0
    k8s.io/api v0.29.0
    k8s.io/apimachinery v0.29.0
)
```
