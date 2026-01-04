# Implementation: Enable Strict Linting and Fix Issues

**Date**: 2026-01-04
**Status**: Completed

## Summary

Enabled strict linting configuration in golangci-lint and fixed all existing issues in the codebase. The linter now runs with stricter checks and passes with 0 issues.

## Changes Made

### 1. Fixed Deprecation Warnings (SA1019)

- **logs_test.go, namespace_test.go, pods_test.go**: Replaced `fake.NewSimpleClientset()` with `fake.NewClientset()`
- **ui/logs.go**: Updated deprecated viewport methods:
  - `LineUp(1)` → `ScrollUp(lines)`
  - `LineDown(1)` → `ScrollDown(lines)`
  - `ViewUp()` → `PageUp()`
  - `ViewDown()` → `PageDown()`
- Removed custom `min()` function that shadowed Go 1.21+ built-in

### 2. Added Package Comments (ST1000)

- **internal/app/app.go**: Added package documentation comment
- **internal/k8s/client.go**: Added package documentation comment

### 3. Fixed Exported Comment Format (ST1020)

- **internal/app/app.go**: Fixed comment format for `CurrentView()` getter
- Added documentation comments to all exported getters: `IsReady()`, `ShowingHelp()`, `Width()`, `Height()`, `Pods()`, `SelectedPodIndex()`, `K8sError()`

### 4. Fixed errcheck Issues

- **client_test.go**: Replaced `os.Setenv`/`os.Unsetenv` with `t.Setenv()` for automatic cleanup
- **logs.go**: Added `//nolint:errcheck` for `stream.Close()` in cleanup code where error cannot be usefully handled

### 5. Fixed gocritic Issues

- **emptyStringTest**: Changed `len(str) > 0` to `str != ""` in logs.go and namespace.go
- **rangeValCopy**: Converted range loops to use indexing for large structs (pods.go, namespace.go, app.go)
- **prealloc**: Pre-allocated slices where size is known (context.go, logs_test.go, pods.go)
- **evalOrder**: Fixed evaluation order in app.go by assigning cmd before return
- **octalLiteral**: Changed `0600` to `0o600` in client_test.go

### 6. Added Comments to Exported Constants

- **pods.go**: Added block comment for PodStatus constants
- **model/types.go**: Added block comment for ViewState constants
- **ui/logs.go**: Added block comment for LogViewState constants

### 7. Updated .golangci.yml

The configuration now includes:
- Strict errcheck with type-assertion and blank checks
- gocritic with diagnostic, style, and performance tags
- Disabled checks that conflict with framework patterns:
  - `hugeParam`: Bubble Tea requires value receivers
  - `unnamedResult`: Not enforced for all functions
  - `singleCaseSwitch`: Allowed for type assertions
- revive rules for Go idioms
- gosec for security checks
- Additional linters: prealloc, unconvert, nolintlint, misspell

### 8. Refactored Pod Processing Functions

Changed pod processing functions to accept pointers instead of values to avoid large struct copies:
- `podToInfo(pod *corev1.Pod)`
- `parseContainerStatuses(pod *corev1.Pod)`
- `determinePodStatus(pod *corev1.Pod)`
- `areAllContainersReady(pod *corev1.Pod)`
- `getFailureReason(pod *corev1.Pod)`
- `getPendingReason(pod *corev1.Pod)`
- `getNotReadyReason(pod *corev1.Pod)`

## Files Modified

| File | Changes |
|------|---------|
| `.golangci.yml` | Strict configuration with additional linters |
| `internal/app/app.go` | Package comment, exported comments, rangeValCopy fix, evalOrder fix |
| `internal/k8s/client.go` | Package comment |
| `internal/k8s/client_test.go` | Use t.Setenv, octal literals |
| `internal/k8s/context.go` | prealloc fixes |
| `internal/k8s/logs.go` | nolint for Close, emptyStringTest fix |
| `internal/k8s/logs_test.go` | Deprecated fake client, prealloc |
| `internal/k8s/namespace.go` | rangeValCopy, emptyStringTest |
| `internal/k8s/namespace_test.go` | Deprecated fake client |
| `internal/k8s/pods.go` | Pointer parameters, rangeValCopy, const comments |
| `internal/k8s/pods_test.go` | Updated for pointer parameters, deprecated fake client |
| `internal/model/types.go` | Const comments |
| `internal/ui/logs.go` | Deprecated viewport methods, min shadow, const comments |

## Verification

- `make lint` passes with 0 issues
- `make test` passes (74 tests)
- `go build` succeeds

## Notes

- The `hugeParam` check is disabled globally because Bubble Tea framework requires value receivers for `Update`, `View`, and `Init` methods
- Package comments are disabled in revive since this is a small project with clear structure
- Test files have relaxed errcheck and gocritic rules since test patterns differ from production code
