# Human Test Checklist: Command Execution Feature

**Date:** 2026-01-04
**Feature:** Basic one-shot command execution in pods

## Prerequisites

- A Kubernetes cluster accessible via kubectl
- At least one running pod in a namespace

## Test Cases

### Basic Navigation

- [ ] Start the app with `go run .`
- [ ] Select a pod using j/k keys
- [ ] Press 'e' to enter exec view
- [ ] Verify header shows pod name and namespace
- [ ] Verify input field is focused and shows placeholder
- [ ] Press Esc to return to pod list
- [ ] Verify you're back at pod list

### Command Execution

- [ ] Enter exec view for a pod
- [ ] Type `pwd` and press Enter
- [ ] Verify "[RUNNING]" status appears briefly
- [ ] Verify output shows current directory
- [ ] Verify status changes to "[COMPLETE]"

- [ ] Type `ls -la` and press Enter
- [ ] Verify directory listing appears in output

- [ ] Type `env` and press Enter
- [ ] Verify environment variables are displayed

- [ ] Type `echo "hello world"` and press Enter
- [ ] Verify "hello world" appears in output

### Error Handling

- [ ] Type a non-existent command like `nonexistent_cmd`
- [ ] Press Enter
- [ ] Verify error message appears (with [stderr] prefix)
- [ ] Verify status shows "[ERROR...]"

- [ ] Type `exit 1` (or similar failing command)
- [ ] Verify stderr output is displayed

### Command History

- [ ] Run several commands: `pwd`, `ls`, `env`
- [ ] Clear input field
- [ ] Press Up arrow
- [ ] Verify last command (`env`) appears
- [ ] Press Up arrow again
- [ ] Verify previous command (`ls`) appears
- [ ] Press Down arrow
- [ ] Verify next command (`env`) appears
- [ ] Press Down arrow again
- [ ] Verify input clears (past end of history)

### Output Scrolling

- [ ] Run a command that produces lots of output (e.g., `find /`)
- [ ] Press Tab to switch focus to output
- [ ] Use j/k or arrow keys to scroll
- [ ] Use PgUp/PgDn for page scrolling
- [ ] Press Tab to return focus to input

### Multiple Commands

- [ ] Run `pwd`
- [ ] Verify command marker (`$ pwd`) appears
- [ ] Run `ls`
- [ ] Verify new command marker appears below previous output
- [ ] Verify both outputs are visible and separated

### Edge Cases

- [ ] Try pressing Enter with empty input
- [ ] Verify nothing happens (no error)

- [ ] Try command with quotes: `echo "test with spaces"`
- [ ] Verify output shows correctly

- [ ] Press 'e' when no pods exist in namespace
- [ ] Verify exec view doesn't open (stays at pod list)

### Return Navigation

- [ ] While in exec view, press Esc
- [ ] Verify return to pod list
- [ ] Pod selection should be preserved

## Notes

- Commands have a 30 second timeout
- Uses first container if pod has multiple containers
- Output buffer limited to 5000 lines
- History limited to 50 commands
