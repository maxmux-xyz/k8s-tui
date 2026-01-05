package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxime/k8s-tui/internal/k8s"
)

func TestNewFileBrowserModel(t *testing.T) {
	m := NewFileBrowserModel()

	if m.currentPath != "/" {
		t.Errorf("currentPath = %q, want %q", m.currentPath, "/")
	}
	if m.state != FileBrowserStateIdle {
		t.Errorf("state = %v, want %v", m.state, FileBrowserStateIdle)
	}
	if len(m.entries) != 0 {
		t.Errorf("entries length = %d, want 0", len(m.entries))
	}
	if len(m.pathHistory) != 0 {
		t.Errorf("pathHistory length = %d, want 0", len(m.pathHistory))
	}
}

func TestFileBrowserModel_SetPodInfo(t *testing.T) {
	m := NewFileBrowserModel()
	m.SetPodInfo("default", "my-pod", "main")

	if m.namespace != "default" {
		t.Errorf("namespace = %q, want %q", m.namespace, "default")
	}
	if m.pod != "my-pod" {
		t.Errorf("pod = %q, want %q", m.pod, "my-pod")
	}
	if m.container != "main" {
		t.Errorf("container = %q, want %q", m.container, "main")
	}
}

func TestFileBrowserModel_SetState(t *testing.T) {
	m := NewFileBrowserModel()

	states := []FileBrowserState{
		FileBrowserStateLoading,
		FileBrowserStateReady,
		FileBrowserStateError,
		FileBrowserStateViewingFile,
	}

	for _, state := range states {
		m.SetState(state)
		if m.State() != state {
			t.Errorf("State() = %v, want %v", m.State(), state)
		}
	}
}

func TestFileBrowserModel_SetError(t *testing.T) {
	m := NewFileBrowserModel()
	m.SetError("something went wrong")

	if m.state != FileBrowserStateError {
		t.Errorf("state = %v, want %v", m.state, FileBrowserStateError)
	}
	if m.errorMsg != "something went wrong" {
		t.Errorf("errorMsg = %q, want %q", m.errorMsg, "something went wrong")
	}
}

func TestFileBrowserModel_SetEntries(t *testing.T) {
	m := NewFileBrowserModel()
	entries := []k8s.FileInfo{
		{Name: "..", IsDir: true},
		{Name: "config", IsDir: true},
		{Name: "app.log", IsDir: false, Size: 1234},
	}

	m.SetEntries(entries)

	if len(m.entries) != 3 {
		t.Errorf("entries length = %d, want 3", len(m.entries))
	}
	if m.state != FileBrowserStateReady {
		t.Errorf("state = %v, want %v", m.state, FileBrowserStateReady)
	}
	if m.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0", m.selectedIndex)
	}
}

func TestFileBrowserModel_Navigation(t *testing.T) {
	m := NewFileBrowserModel()
	entries := []k8s.FileInfo{
		{Name: "..", IsDir: true},
		{Name: "config", IsDir: true},
		{Name: "app.log", IsDir: false},
		{Name: "readme.md", IsDir: false},
	}
	m.SetEntries(entries)

	// Test NavigateDown
	m.NavigateDown()
	if m.selectedIndex != 1 {
		t.Errorf("after NavigateDown: selectedIndex = %d, want 1", m.selectedIndex)
	}

	m.NavigateDown()
	m.NavigateDown()
	if m.selectedIndex != 3 {
		t.Errorf("after 3x NavigateDown: selectedIndex = %d, want 3", m.selectedIndex)
	}

	// Should not go beyond last entry
	m.NavigateDown()
	if m.selectedIndex != 3 {
		t.Errorf("NavigateDown past end: selectedIndex = %d, want 3", m.selectedIndex)
	}

	// Test NavigateUp
	m.NavigateUp()
	if m.selectedIndex != 2 {
		t.Errorf("after NavigateUp: selectedIndex = %d, want 2", m.selectedIndex)
	}

	// Test GotoTop
	m.GotoTop()
	if m.selectedIndex != 0 {
		t.Errorf("after GotoTop: selectedIndex = %d, want 0", m.selectedIndex)
	}

	// Should not go below 0
	m.NavigateUp()
	if m.selectedIndex != 0 {
		t.Errorf("NavigateUp past start: selectedIndex = %d, want 0", m.selectedIndex)
	}

	// Test GotoBottom
	m.GotoBottom()
	if m.selectedIndex != 3 {
		t.Errorf("after GotoBottom: selectedIndex = %d, want 3", m.selectedIndex)
	}
}

func TestFileBrowserModel_SelectedEntry(t *testing.T) {
	m := NewFileBrowserModel()

	// Empty entries
	entry := m.SelectedEntry()
	if entry != nil {
		t.Error("SelectedEntry should be nil for empty entries")
	}

	// With entries
	entries := []k8s.FileInfo{
		{Name: "file1.txt"},
		{Name: "file2.txt"},
	}
	m.SetEntries(entries)

	entry = m.SelectedEntry()
	if entry == nil {
		t.Fatal("SelectedEntry should not be nil")
	}
	if entry.Name != "file1.txt" {
		t.Errorf("SelectedEntry().Name = %q, want %q", entry.Name, "file1.txt")
	}

	m.NavigateDown()
	entry = m.SelectedEntry()
	if entry.Name != "file2.txt" {
		t.Errorf("after NavigateDown: SelectedEntry().Name = %q, want %q", entry.Name, "file2.txt")
	}
}

func TestFileBrowserModel_PathHistory(t *testing.T) {
	m := NewFileBrowserModel()

	// Push some paths
	m.currentPath = "/app"
	m.PushPath()
	m.currentPath = "/app/logs"
	m.PushPath()
	m.currentPath = "/app/logs/today"

	if len(m.pathHistory) != 2 {
		t.Errorf("pathHistory length = %d, want 2", len(m.pathHistory))
	}

	// Pop paths
	if !m.PopPath() {
		t.Error("PopPath should return true")
	}
	if m.currentPath != "/app/logs" {
		t.Errorf("after PopPath: currentPath = %q, want %q", m.currentPath, "/app/logs")
	}

	if !m.PopPath() {
		t.Error("PopPath should return true")
	}
	if m.currentPath != "/app" {
		t.Errorf("after 2nd PopPath: currentPath = %q, want %q", m.currentPath, "/app")
	}

	// Pop when empty
	if m.PopPath() {
		t.Error("PopPath should return false when history is empty")
	}
}

func TestFileBrowserModel_NavigateToParent(t *testing.T) {
	m := NewFileBrowserModel()
	m.currentPath = "/app/logs"

	parent := m.NavigateToParent()
	if parent != "/app" {
		t.Errorf("NavigateToParent returned %q, want %q", parent, "/app")
	}
	if m.currentPath != "/app" {
		t.Errorf("currentPath = %q, want %q", m.currentPath, "/app")
	}

	// Check history was pushed
	if len(m.pathHistory) != 1 {
		t.Errorf("pathHistory length = %d, want 1", len(m.pathHistory))
	}
	if m.pathHistory[0] != "/app/logs" {
		t.Errorf("pathHistory[0] = %q, want %q", m.pathHistory[0], "/app/logs")
	}

	// Navigate to parent from root
	m.currentPath = "/"
	parent = m.NavigateToParent()
	if parent != "" {
		t.Errorf("NavigateToParent from root returned %q, want empty", parent)
	}
}

func TestFileBrowserModel_NavigateToEntry(t *testing.T) {
	m := NewFileBrowserModel()
	m.currentPath = "/app"

	entries := []k8s.FileInfo{
		{Name: "..", IsDir: true},
		{Name: ".", IsDir: true},
		{Name: "config", IsDir: true},
		{Name: "app.log", IsDir: false},
	}
	m.SetEntries(entries)

	// Test ".." navigation
	path, isFile := m.NavigateToEntry()
	if path != "/" {
		t.Errorf("NavigateToEntry for '..' returned path %q, want %q", path, "/")
	}
	if isFile {
		t.Error("'..' should not be a file")
	}

	// Reset path and test "."
	m.currentPath = "/app"
	m.selectedIndex = 1
	path, _ = m.NavigateToEntry()
	if path != "" {
		t.Errorf("NavigateToEntry for '.' returned path %q, want empty", path)
	}

	// Test directory navigation
	m.currentPath = "/app"
	m.selectedIndex = 2
	path, isFile = m.NavigateToEntry()
	if path != "/app/config" {
		t.Errorf("NavigateToEntry for 'config' returned path %q, want %q", path, "/app/config")
	}
	if isFile {
		t.Error("'config' should not be a file")
	}
	if m.currentPath != "/app/config" {
		t.Errorf("currentPath after dir navigation = %q, want %q", m.currentPath, "/app/config")
	}

	// Test file selection
	m.currentPath = "/app"
	m.selectedIndex = 3
	path, isFile = m.NavigateToEntry()
	if path != "/app/app.log" {
		t.Errorf("NavigateToEntry for 'app.log' returned path %q, want %q", path, "/app/app.log")
	}
	if !isFile {
		t.Error("'app.log' should be a file")
	}
}

func TestFileBrowserModel_SetFileContent(t *testing.T) {
	m := NewFileBrowserModel()
	m.SetSize(80, 24) // Initialize viewport

	content := "line1\nline2\nline3"
	m.SetFileContent("test.txt", content)

	if m.state != FileBrowserStateViewingFile {
		t.Errorf("state = %v, want %v", m.state, FileBrowserStateViewingFile)
	}
	if m.viewingFile != "test.txt" {
		t.Errorf("viewingFile = %q, want %q", m.viewingFile, "test.txt")
	}
	if !m.IsViewingFile() {
		t.Error("IsViewingFile should return true")
	}
}

func TestFileBrowserModel_ExitFileView(t *testing.T) {
	m := NewFileBrowserModel()
	m.SetSize(80, 24)
	m.SetFileContent("test.txt", "content")

	m.ExitFileView()

	if m.state != FileBrowserStateReady {
		t.Errorf("state = %v, want %v", m.state, FileBrowserStateReady)
	}
	if m.viewingFile != "" {
		t.Errorf("viewingFile = %q, want empty", m.viewingFile)
	}
	if m.IsViewingFile() {
		t.Error("IsViewingFile should return false")
	}
}

func TestFileBrowserModel_Clear(t *testing.T) {
	m := NewFileBrowserModel()
	m.currentPath = "/app/logs"
	m.pathHistory = []string{"/app"}
	m.SetEntries([]k8s.FileInfo{{Name: "file.txt"}})
	m.selectedIndex = 5
	m.SetError("some error")

	m.Clear()

	if m.currentPath != "/" {
		t.Errorf("currentPath = %q, want %q", m.currentPath, "/")
	}
	if len(m.pathHistory) != 0 {
		t.Errorf("pathHistory length = %d, want 0", len(m.pathHistory))
	}
	if len(m.entries) != 0 {
		t.Errorf("entries length = %d, want 0", len(m.entries))
	}
	if m.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0", m.selectedIndex)
	}
	if m.state != FileBrowserStateIdle {
		t.Errorf("state = %v, want %v", m.state, FileBrowserStateIdle)
	}
	if m.errorMsg != "" {
		t.Errorf("errorMsg = %q, want empty", m.errorMsg)
	}
}

func TestFileBrowserModel_Update_Navigation(t *testing.T) {
	m := NewFileBrowserModel()
	m.SetSize(80, 24)
	m.SetEntries([]k8s.FileInfo{
		{Name: "file1"},
		{Name: "file2"},
		{Name: "file3"},
	})

	// Test 'j' key
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.selectedIndex != 1 {
		t.Errorf("after 'j': selectedIndex = %d, want 1", m.selectedIndex)
	}

	// Test 'k' key
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.selectedIndex != 0 {
		t.Errorf("after 'k': selectedIndex = %d, want 0", m.selectedIndex)
	}

	// Test 'G' key (go to bottom)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	if m.selectedIndex != 2 {
		t.Errorf("after 'G': selectedIndex = %d, want 2", m.selectedIndex)
	}

	// Test 'g' key (go to top)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if m.selectedIndex != 0 {
		t.Errorf("after 'g': selectedIndex = %d, want 0", m.selectedIndex)
	}
}

func TestFileBrowserModel_View_DirectoryListing(t *testing.T) {
	m := NewFileBrowserModel()
	m.SetSize(80, 24)
	m.SetPodInfo("default", "my-pod", "main")
	m.SetEntries([]k8s.FileInfo{
		{Name: "..", IsDir: true},
		{Name: "config", IsDir: true, Permissions: "drwxr-xr-x"},
		{Name: "app.log", IsDir: false, Size: 12345, Permissions: "-rw-r--r--"},
	})

	view := m.View()

	// Check header
	if !strings.Contains(view, "default/my-pod/main") {
		t.Error("view should contain pod info")
	}

	// Check path
	if !strings.Contains(view, "Path: /") {
		t.Error("view should contain path")
	}

	// Check entries are shown
	if !strings.Contains(view, "config") {
		t.Error("view should contain 'config' directory")
	}
	if !strings.Contains(view, "app.log") {
		t.Error("view should contain 'app.log' file")
	}

	// Check status line
	if !strings.Contains(view, "[READY]") {
		t.Error("view should contain [READY] status")
	}
}

func TestFileBrowserModel_View_FileContent(t *testing.T) {
	m := NewFileBrowserModel()
	m.SetSize(80, 24)
	m.SetPodInfo("default", "my-pod", "main")
	m.currentPath = "/app"
	m.SetFileContent("test.txt", "Hello, World!")

	view := m.View()

	// Check it's in file viewing mode
	if !strings.Contains(view, "[VIEWING]") {
		t.Error("view should contain [VIEWING] status")
	}

	// Check file path is shown
	if !strings.Contains(view, "/app/test.txt") {
		t.Error("view should contain file path")
	}
}

func TestFileBrowserModel_View_Error(t *testing.T) {
	m := NewFileBrowserModel()
	m.SetSize(80, 24)
	m.SetError("permission denied")

	view := m.View()

	if !strings.Contains(view, "Error: permission denied") {
		t.Error("view should contain error message")
	}
}

func TestFileBrowserModel_View_Loading(t *testing.T) {
	m := NewFileBrowserModel()
	m.SetSize(80, 24)
	m.SetState(FileBrowserStateLoading)

	view := m.View()

	if !strings.Contains(view, "Loading") {
		t.Error("view should contain 'Loading'")
	}
}

func TestFileBrowserState_String(t *testing.T) {
	tests := []struct {
		state FileBrowserState
		want  string
	}{
		{FileBrowserStateIdle, "Idle"},
		{FileBrowserStateLoading, "Loading"},
		{FileBrowserStateReady, "Ready"},
		{FileBrowserStateError, "Error"},
		{FileBrowserStateViewingFile, "Viewing"},
		{FileBrowserState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMaxFilePreviewBytes(t *testing.T) {
	bytes := MaxFilePreviewBytes()
	if bytes != 100*1024 {
		t.Errorf("MaxFilePreviewBytes() = %d, want %d", bytes, 100*1024)
	}
}
