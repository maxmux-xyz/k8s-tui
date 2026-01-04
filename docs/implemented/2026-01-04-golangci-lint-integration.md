# Implementation: golangci-lint Integration

**Date**: 2026-01-04
**Status**: Completed

## Summary

Integrated golangci-lint v2 into the k8s-tui repository with:
- Configuration file (`.golangci.yml`)
- Makefile targets (`lint`, `lint-fix`)
- GitHub Actions CI workflow

## Changes Made

### Files Created

1. **`.golangci.yml`** - Linter configuration (v2 format)
   - Uses `default: standard` linters (errcheck, govet, staticcheck, ineffassign, unused)
   - Adds gosec for security checks
   - Adds misspell for spelling
   - Formatters: gofmt, goimports
   - Some checks temporarily disabled to allow incremental adoption

2. **`.github/workflows/lint.yml`** - CI workflow
   - Runs on push/PR to main branch
   - Uses official golangci-lint-action v6
   - 5 minute timeout

### Files Modified

1. **`Makefile`** - Added lint targets:
   - `make lint` - Run linter
   - `make lint-fix` - Run linter with auto-fix

2. **`CLAUDE.md`** - Documented new commands

## Configuration Notes

The golangci-lint configuration uses v2 format (requires `version: "2"`). Key differences from v1:
- Formatters (gofmt, goimports) are in separate `formatters:` section
- Exclusion rules moved to `linters.exclusions.rules`
- `linters-settings` renamed to `linters.settings`
- Some linters merged (e.g., gosimple into staticcheck)

### Temporarily Disabled Checks

To allow incremental adoption, some checks are disabled:
- `SA1019` - Deprecation warnings (deprecated API usage in tests)
- `ST1000` - Package comments
- `ST1020` - Exported function comment format
- `errcheck` in test files

## Follow-up Tasks

A separate plan should be created to:
1. Fix all existing lint issues
2. Enable stricter linting configuration
3. Add pre-commit hook for local enforcement

## Verification

```bash
make lint      # Passes with 0 issues
make lint-fix  # Auto-fixes formatting issues
make test      # All tests pass
```
