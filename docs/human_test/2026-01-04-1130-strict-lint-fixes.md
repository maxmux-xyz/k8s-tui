# Human Test Checklist: Strict Linting and Fixes

**Date**: 2026-01-04
**Plan**: Enable Strict Linting and Fix Issues

## Overview

This checklist verifies that the strict linting fixes don't introduce any regressions in the application functionality, particularly around log viewing and scrolling.

## Pre-requisites

- [ ] Have access to a Kubernetes cluster (minikube, kind, or remote)
- [ ] Have `kubectl` configured and working
- [ ] Build the application: `go build -o kpm .`

## Test Cases

### 1. Application Startup

- [ ] Run `./kpm` and verify it starts without errors
- [ ] Verify the pod list loads correctly
- [ ] Check that the status bar and help bar appear properly

### 2. Pod List Navigation

- [ ] Use `j`/`k` or arrow keys to navigate up/down the pod list
- [ ] Verify pod selection indicator (`>`) moves correctly
- [ ] Press `r` to refresh and verify pods reload

### 3. Log Viewing - Basic

- [ ] Select a running pod and press `l` to view logs
- [ ] Verify logs are streamed and displayed correctly
- [ ] Check the log header shows correct namespace/pod/container info
- [ ] Verify the status shows `[STREAMING]`

### 4. Log Viewing - Scrolling (Updated Viewport Methods)

These tests specifically verify the updated viewport scroll methods work correctly:

- [ ] **ScrollUp**: Press `k` or `up` arrow - logs should scroll up by one line
- [ ] **ScrollDown**: Press `j` or `down` arrow - logs should scroll down by one line
- [ ] **PageUp**: Press `pgup` - logs should scroll up by one page
- [ ] **PageDown**: Press `pgdown` or `space` - logs should scroll down by one page
- [ ] **GotoTop**: Press `g` - should jump to the top of logs
- [ ] **GotoBottom**: Press `G` - should jump to the bottom of logs

### 5. Follow Mode

- [ ] Press `G` to go to bottom and verify `[FOLLOW]` indicator appears
- [ ] When follow mode is on, new logs should auto-scroll to bottom
- [ ] Press `k` or scroll up and verify `[FOLLOW]` indicator disappears
- [ ] Press `f` to toggle follow mode on/off manually

### 6. Log View Exit

- [ ] Press `esc` to exit log view and return to pod list
- [ ] Verify log streaming stops
- [ ] Verify status changes to `[STREAM ENDED]` before returning

### 7. Namespace Switching

- [ ] Press `n` to open namespace selector
- [ ] Navigate up/down with `j`/`k`
- [ ] Press `enter` to select a namespace
- [ ] Verify pods refresh with new namespace

### 8. Context Switching

- [ ] Press `c` to open context selector
- [ ] Navigate and select a different context
- [ ] Verify pods refresh with new context

### 9. Help Overlay

- [ ] Press `?` to open help overlay
- [ ] Verify all keybindings are displayed
- [ ] Press any key to close help

### 10. Linting Verification

- [ ] Run `make lint` - should pass with 0 issues
- [ ] Run `make test` - all tests should pass
- [ ] Run `go build -o kpm .` - should compile without warnings

## Notes

- If any scrolling feels different or incorrect, report it as it may indicate an issue with the viewport method updates
- The scroll behavior should be identical to before - we just updated to non-deprecated method names

## Sign-off

- [ ] All tests pass
- [ ] Tested by: ________________
- [ ] Date: ________________
- [ ] Notes: ________________
