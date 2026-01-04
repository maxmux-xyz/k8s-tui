# Plan: Enable Strict Linting and Fix Issues

**Date**: 2026-01-04
**Status**: Planned

## Overview

Enable the full strict linting configuration in golangci-lint and fix all existing issues in the codebase. This follows up on the initial golangci-lint integration which used a relaxed configuration.

## Current State

The `.golangci.yml` currently has these checks disabled:
- `SA1019` - Deprecation warnings (7 issues)
- `ST1000` - Package comments (3 issues)
- `ST1020` - Exported function comment format (1 issue)
- `errcheck` in test files (4 issues in `client_test.go`)
- `errcheck` for `io.Closer.Close` (1 issue in `logs.go`)

Total: ~16 issues to fix

## Implementation Steps

### Step 1: Fix Deprecation Warnings (SA1019)

**Files affected:**
- `internal/k8s/logs_test.go` - Replace `fake.NewSimpleClientset` with `fake.NewClientset`
- `internal/ui/logs.go` - Replace deprecated viewport methods:
  - `m.viewport.LineUp(1)` → `m.viewport.ScrollUp(1)`
  - `m.viewport.LineDown(1)` → `m.viewport.ScrollDown(1)`
  - `m.viewport.ViewUp()` → `m.viewport.PageUp()`
  - `m.viewport.ViewDown()` → `m.viewport.PageDown()`

### Step 2: Add Package Comments (ST1000)

Add package documentation comments to:
- `internal/app/app.go` - Package app provides the Bubble Tea application model
- `internal/k8s/client.go` - Package k8s provides Kubernetes client operations

### Step 3: Fix Exported Comment Format (ST1020)

**File:** `internal/app/app.go:781`

Change:
```go
// Getters for testing
func (m Model) CurrentView() model.ViewState {
```

To:
```go
// CurrentView returns the current view state (used for testing).
func (m Model) CurrentView() model.ViewState {
```

### Step 4: Fix errcheck Issues in Tests

**File:** `internal/k8s/client_test.go`

Handle error returns from `os.Setenv` and `os.Unsetenv`:
```go
// Option 1: Use t.Setenv (preferred for tests)
t.Setenv("HOME", tmpDir)

// Option 2: Check errors explicitly
if err := os.Setenv("HOME", tmpDir); err != nil {
    t.Fatal(err)
}
```

### Step 5: Fix errcheck for stream.Close

**File:** `internal/k8s/logs.go:69`

Change:
```go
defer stream.Close()
```

To:
```go
defer func() {
    if err := stream.Close(); err != nil {
        // Log error but don't fail - we're already in cleanup
    }
}()
```

Or use a helper function if logging is available.

### Step 6: Update .golangci.yml

Remove all temporary exclusions:
```yaml
version: "2"

run:
  timeout: 5m

linters:
  default: standard
  enable:
    - gosec
    - misspell
    - gocritic
    - revive
    - prealloc
    - unconvert
    - nolintlint

  settings:
    errcheck:
      check-type-assertions: true
      check-blank: true

    gocritic:
      enabled-tags:
        - diagnostic
        - style
        - performance

    revive:
      rules:
        - name: blank-imports
        - name: context-as-argument
        - name: context-keys-type
        - name: error-return
        - name: error-strings
        - name: exported
        - name: increment-decrement
        - name: var-declaration
        - name: package-comments
          disabled: true  # Keep disabled for small project

    gosec:
      excludes:
        - G104

  exclusions:
    rules:
      - path: _test\.go
        linters:
          - gosec

formatters:
  enable:
    - gofmt
    - goimports

output:
  show-stats: true
```

### Step 7: Run Full Lint and Fix Remaining Issues

1. Run `make lint` to identify any remaining issues
2. Run `make lint-fix` for auto-fixable issues
3. Manually fix remaining issues
4. Verify all tests pass

## Files to Modify

| File | Changes |
|------|---------|
| `internal/k8s/logs_test.go` | Update deprecated fake client |
| `internal/ui/logs.go` | Update deprecated viewport methods |
| `internal/app/app.go` | Add package comment, fix exported comments |
| `internal/k8s/client.go` | Add package comment |
| `internal/k8s/client_test.go` | Use t.Setenv or handle errors |
| `internal/k8s/logs.go` | Handle stream.Close error |
| `.golangci.yml` | Enable strict configuration |

## Verification

1. `make lint` passes with 0 issues
2. `make test` passes
3. `go build` succeeds
4. No `//nolint` directives added (prefer fixing over suppressing)

## Notes

- Some gocritic warnings about "hugeParam" (passing large structs by value) are intentional in Bubble Tea - the framework requires value receivers for `Update` and `View` methods
- If hugeParam warnings are too noisy, consider adding exclusion for specific functions
