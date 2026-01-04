# Plan: golangci-lint Integration

**Date**: 2026-01-04
**Status**: Planned

## Overview

Integrate golangci-lint into the k8s-tui repository with strict linting configuration, Makefile target, and GitHub Actions CI/CD workflow.

## Why golangci-lint?

- Industry standard meta-linter (used by Kubernetes, Prometheus, Terraform)
- Aggregates 50+ linters into a single tool
- Fast parallel execution with caching
- Single YAML configuration file

## Implementation Steps

### Step 1: Create `.golangci.yml` Configuration

Create a strict linter configuration at the project root:

```yaml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    # Essential
    - govet          # Reports suspicious constructs
    - errcheck       # Checks unchecked errors
    - staticcheck    # Gold standard static analyzer
    - gosimple       # Simplification suggestions
    - ineffassign    # Detects unused assignments
    - unused         # Finds unused code

    # Formatting
    - gofmt          # Formatting check
    - goimports      # Import ordering

    # Security
    - gosec          # Security issues

    # Style & Best Practices
    - gocritic       # Opinionated style checks
    - revive         # Fast, configurable linter (golint replacement)
    - misspell       # Spelling mistakes
    - unconvert      # Unnecessary type conversions
    - prealloc       # Slice preallocation suggestions
    - nolintlint     # Ill-formed nolint directives

linters-settings:
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
        disabled: true  # Can be noisy for small projects

  gosec:
    excludes:
      - G104  # Audit errors not checked (covered by errcheck)

issues:
  exclude-rules:
    # Test files can have longer functions and more complexity
    - path: _test\.go
      linters:
        - gocritic
        - gosec

  max-issues-per-linter: 50
  max-same-issues: 10
```

**Location**: `.golangci.yml` (project root)

### Step 2: Update Makefile

Add lint targets to the existing Makefile:

```makefile
# Add to .PHONY line:
.PHONY: build run test test-coverage clean deps verify lint lint-fix

# Add lint targets:

# Run linter
lint:
	golangci-lint run

# Run linter and fix auto-fixable issues
lint-fix:
	golangci-lint run --fix
```

**Location**: `Makefile`

### Step 3: Create GitHub Actions Workflow

Create a new workflow file for CI linting:

```yaml
name: Lint

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  golangci-lint:
    name: golangci-lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m
```

**Location**: `.github/workflows/lint.yml`

### Step 4: Fix Existing Lint Issues

After adding the configuration, run `make lint` and fix any issues:

1. Run `make lint` to see all issues
2. Run `make lint-fix` to auto-fix formatting issues
3. Manually fix remaining issues
4. Consider using `//nolint:lintername` for intentional exceptions (with explanation)

### Step 5: Update Documentation

Update `CLAUDE.md` to add lint commands:

```markdown
## Commands

```bash
go run .             # Run app
go test ./...        # Run all tests
make test            # Run tests with coverage
make lint            # Run linters
make lint-fix        # Run linters and auto-fix
go build -o kpm .    # Build binary
```
```

## Files to Create/Modify

| File | Action | Description |
|------|--------|-------------|
| `.golangci.yml` | Create | Linter configuration |
| `Makefile` | Modify | Add lint and lint-fix targets |
| `.github/workflows/lint.yml` | Create | CI workflow |
| `CLAUDE.md` | Modify | Document new commands |

## Verification

After implementation, verify:

1. `make lint` runs without errors
2. `make lint-fix` auto-fixes issues
3. GitHub Actions workflow passes on push/PR
4. All existing tests still pass (`make test`)

## Notes

- The strict configuration may surface existing issues in the codebase that need fixing
- Consider adding a pre-commit hook in the future for local enforcement
- gosec may flag false positives in test files (excluded in config)
