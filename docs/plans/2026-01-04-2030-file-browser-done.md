# Plan: File Browser (Phase 6)

**Date:** 2026-01-04
**Feature:** Phase 6 - File browser for navigating pod filesystem

## Overview

Implement a file browser allowing users to navigate directories inside a pod, view file contents, and see file metadata. This uses `exec` under the hood to run `ls` and `cat` commands.

## Architecture Decisions

The file browser will use exec-based commands (`ls -la`, `cat`) rather than trying to copy files. This is simpler and follows kubectl patterns.

**Approach:**
1. Use existing `k8s.Exec()` to run `ls -la <path>` for directory listing
2. Parse ls output into structured file entries
3. Use `cat` to read file contents (with size limits)
4. Display in a list + preview split or single-pane view

## Files to Create

### 1. `internal/k8s/files.go`
File operations using exec commands.

```go
// FileInfo represents a file or directory entry
type FileInfo struct {
    Name        string
    IsDir       bool
    Size        int64
    Permissions string
    Owner       string
    Group       string
    ModTime     string
}

// FileOptions configures file operations
type FileOptions struct {
    Namespace string
    Pod       string
    Container string
    Path      string
}

// ListDir lists directory contents using ls -la
func (c *Client) ListDir(ctx context.Context, opts FileOptions) ([]FileInfo, error)

// ReadFile reads file contents using cat (with size limit)
func (c *Client) ReadFile(ctx context.Context, opts FileOptions, maxBytes int) (string, error)

// StatFile gets file info for a single path
func (c *Client) StatFile(ctx context.Context, opts FileOptions) (*FileInfo, error)
```

Implementation notes:
- `ListDir` runs `ls -la <path>` and parses output
- Handle parsing edge cases (spaces in filenames, symlinks)
- `ReadFile` uses `head -c <maxBytes>` to limit output
- Return appropriate errors for not found, permission denied

### 2. `internal/k8s/files_test.go`
Tests for file operations.

- Test `parseLsOutput()` with various ls -la formats
- Test edge cases: symlinks, special chars, spaces in names
- Test FileInfo struct creation
- Test FileOptions validation

### 3. `internal/ui/files.go`
UI component for file browser view.

```go
// FileBrowserState represents the current state
type FileBrowserState int
const (
    FileBrowserStateIdle FileBrowserState = iota
    FileBrowserStateLoading
    FileBrowserStateReady
    FileBrowserStateError
    FileBrowserStateViewingFile
)

// FileBrowserModel manages the file browser UI
type FileBrowserModel struct {
    // Current location
    currentPath string
    pathHistory []string  // For backspace navigation

    // Directory contents
    entries       []k8s.FileInfo
    selectedIndex int

    // File preview
    previewContent string
    previewScroll  int
    viewport       viewport.Model

    // State
    state    FileBrowserState
    errorMsg string

    // Pod info
    namespace string
    pod       string
    container string

    // Dimensions
    width  int
    height int
    ready  bool
}
```

Features:
- Directory listing with file type icons (📁 / 📄)
- Size and permissions display
- Navigation: Enter=open, Backspace=parent
- File content preview in viewport
- Path display showing current location

### 4. `internal/ui/files_test.go`
Tests for file browser UI component.

- Test initialization
- Test navigation (up/down selection)
- Test path history
- Test state transitions
- Test view rendering

## Files to Modify

### 5. `internal/app/app.go`
Add file browser integration to main app.

**New message types:**
```go
type dirLoadedMsg struct {
    entries []k8s.FileInfo
    path    string
    err     error
}

type fileContentMsg struct {
    content string
    path    string
    err     error
}
```

**New model fields:**
```go
filesView   ui.FileBrowserModel
filesCancel context.CancelFunc
```

**New methods:**
- `initFileBrowser() tea.Cmd` - Initialize file browser for selected pod
- `loadDirectory(path string) tea.Cmd` - Load directory contents
- `loadFileContent(path string) tea.Cmd` - Load file for preview
- `handleFilesViewKeys(msg tea.KeyMsg)` - Handle navigation in files view

**Update flow:**
1. User presses 'f' on pod -> switch to ViewFiles, load root dir (`/`)
2. User navigates with j/k, presses Enter on dir -> load that directory
3. User presses Enter on file -> load content preview
4. User presses Backspace -> go to parent directory
5. User presses Esc -> return to pod list

### 6. `internal/app/app_test.go`
Add tests for file browser flow.

- Test transition to files view
- Test files view requires pods (like logs/exec)
- Test back navigation

### 7. `internal/ui/keys.go`
Add new keybindings for file browser.

```go
// File browser specific
OpenFile    key.Binding  // Enter - open dir or view file
ParentDir   key.Binding  // Backspace - go to parent
CopyPath    key.Binding  // y - copy path (future enhancement)
```

## Implementation Steps

### Step 1: K8s file operations layer
1. Create `internal/k8s/files.go` with FileInfo, FileOptions, ListDir, ReadFile
2. Implement `parseLsOutput()` helper function
3. Create `internal/k8s/files_test.go` with parsing tests
4. Run `make test` to verify

### Step 2: UI component
1. Create `internal/ui/files.go` with FileBrowserModel
2. Implement View() with directory listing format
3. Implement Update() with navigation handling
4. Create `internal/ui/files_test.go`
5. Run `make test` to verify

### Step 3: App integration
1. Add message types to app.go
2. Add filesView field to Model, initialize in New()
3. Implement initFileBrowser() command
4. Implement loadDirectory() command
5. Implement handleFilesViewKeys()
6. Update viewFiles() to render FileBrowserModel
7. Add tests to app_test.go
8. Run `make test && make lint` to verify

### Step 4: Manual testing
1. Run app with real cluster
2. Select pod, press 'f'
3. Navigate directories with j/k, Enter, Backspace
4. View file contents
5. Test error handling (permission denied, not found)

## Keybindings (Files View)

| Key | Action |
|-----|--------|
| `j/↓` | Move selection down |
| `k/↑` | Move selection up |
| `Enter` | Open directory / View file |
| `Backspace` | Go to parent directory |
| `g` | Go to top of list |
| `G` | Go to bottom of list |
| `pgup/pgdn` | Page up/down in file preview |
| `Esc` | Back to pod list |

## UI Layout

```
K8s Pod Manager > Files | Context: prod-cluster
Path: /app/logs
────────────────────────────────────────────────
  📁 ..
> 📁 config/           4096  drwxr-xr-x  root
  📁 data/             4096  drwxr-xr-x  root
  📄 app.log           2.3M  -rw-r--r--  app
  📄 config.yaml       1.2K  -rw-r--r--  root
  📄 README.md          845  -rw-r--r--  root

────────────────────────────────────────────────
[READY] 5 items | Enter: open | Backspace: parent | Esc: back
```

When viewing a file:
```
K8s Pod Manager > Files | Context: prod-cluster
File: /app/logs/app.log (2.3M)
────────────────────────────────────────────────
2024-01-04 10:00:01 INFO Starting application
2024-01-04 10:00:02 INFO Connected to database
2024-01-04 10:00:03 INFO Listening on port 8080
...

────────────────────────────────────────────────
[VIEWING] Line 1-50 of 1234 | Esc/Backspace: back to list
```

## Dependencies

No new dependencies needed - using existing:
- `github.com/charmbracelet/bubbles` (viewport for file preview)
- Existing `k8s.Exec()` for running commands

## Out of Scope (Future Enhancements)

- File download/copy to local machine
- File editing
- Binary file detection/handling
- Syntax highlighting for code files
- File search within directory
- Copy path to clipboard (requires external tool)

## Testing Notes

- Unit tests cover parsing and UI logic
- Integration tests would need a real pod (manual testing)
- Test with various file types and permissions
- Test with empty directories
- Test with large directories (pagination might be needed later)
