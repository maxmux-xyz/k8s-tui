package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxime/k8s-tui/internal/k8s"
)

// FileBrowserState represents the current state of the file browser
type FileBrowserState int

// File browser state constants for tracking browser status.
const (
	FileBrowserStateIdle FileBrowserState = iota
	FileBrowserStateLoading
	FileBrowserStateReady
	FileBrowserStateError
	FileBrowserStateViewingFile
)

func (s FileBrowserState) String() string {
	switch s {
	case FileBrowserStateIdle:
		return "Idle"
	case FileBrowserStateLoading:
		return "Loading"
	case FileBrowserStateReady:
		return "Ready"
	case FileBrowserStateError:
		return "Error"
	case FileBrowserStateViewingFile:
		return "Viewing"
	default:
		return "Unknown"
	}
}

const (
	maxFilePreviewBytes = 100 * 1024 // 100KB max file preview
)

// FileBrowserModel manages the file browser UI component
type FileBrowserModel struct {
	// Current location
	currentPath string
	pathHistory []string // For backspace navigation

	// Directory contents
	entries       []k8s.FileInfo
	selectedIndex int

	// File preview
	previewContent  string
	previewViewport viewport.Model
	viewingFile     string // Name of file being viewed

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

// NewFileBrowserModel creates a new file browser model
func NewFileBrowserModel() FileBrowserModel {
	return FileBrowserModel{
		currentPath: "/",
		pathHistory: make([]string, 0),
		entries:     make([]k8s.FileInfo, 0),
		state:       FileBrowserStateIdle,
	}
}

// SetSize updates the viewport size
func (m *FileBrowserModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Reserve space for header (3 lines) and status bar (2 lines)
	viewportHeight := height - 6
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	if !m.ready {
		m.previewViewport = viewport.New(width, viewportHeight)
		m.previewViewport.YPosition = 0
		m.ready = true
	} else {
		m.previewViewport.Width = width
		m.previewViewport.Height = viewportHeight
	}
}

// SetPodInfo sets the pod information for display
func (m *FileBrowserModel) SetPodInfo(namespace, pod, container string) {
	m.namespace = namespace
	m.pod = pod
	m.container = container
}

// SetState sets the current browser state
func (m *FileBrowserModel) SetState(state FileBrowserState) {
	m.state = state
}

// State returns the current file browser state
func (m *FileBrowserModel) State() FileBrowserState {
	return m.state
}

// SetError sets an error message
func (m *FileBrowserModel) SetError(err string) {
	m.errorMsg = err
	m.state = FileBrowserStateError
}

// CurrentPath returns the current directory path
func (m *FileBrowserModel) CurrentPath() string {
	return m.currentPath
}

// SetCurrentPath sets the current path
func (m *FileBrowserModel) SetCurrentPath(path string) {
	m.currentPath = path
}

// SetEntries sets the directory entries
func (m *FileBrowserModel) SetEntries(entries []k8s.FileInfo) {
	m.entries = entries
	m.selectedIndex = 0
	m.state = FileBrowserStateReady
}

// Entries returns the current directory entries
func (m *FileBrowserModel) Entries() []k8s.FileInfo {
	return m.entries
}

// SelectedIndex returns the currently selected entry index
func (m *FileBrowserModel) SelectedIndex() int {
	return m.selectedIndex
}

// SelectedEntry returns the currently selected entry, if any
func (m *FileBrowserModel) SelectedEntry() *k8s.FileInfo {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.entries) {
		return nil
	}
	return &m.entries[m.selectedIndex]
}

// SetFileContent sets the file content for preview
func (m *FileBrowserModel) SetFileContent(filename, content string) {
	m.viewingFile = filename
	m.previewContent = content
	m.previewViewport.SetContent(content)
	m.previewViewport.GotoTop()
	m.state = FileBrowserStateViewingFile
}

// ViewingFile returns the name of the file currently being viewed
func (m *FileBrowserModel) ViewingFile() string {
	return m.viewingFile
}

// IsViewingFile returns whether we're currently viewing a file
func (m *FileBrowserModel) IsViewingFile() bool {
	return m.state == FileBrowserStateViewingFile
}

// ExitFileView exits file viewing mode and returns to directory listing
func (m *FileBrowserModel) ExitFileView() {
	m.state = FileBrowserStateReady
	m.viewingFile = ""
	m.previewContent = ""
}

// NavigateUp moves selection up
func (m *FileBrowserModel) NavigateUp() {
	if m.selectedIndex > 0 {
		m.selectedIndex--
	}
}

// NavigateDown moves selection down
func (m *FileBrowserModel) NavigateDown() {
	if m.selectedIndex < len(m.entries)-1 {
		m.selectedIndex++
	}
}

// GotoTop moves to the first entry
func (m *FileBrowserModel) GotoTop() {
	m.selectedIndex = 0
}

// GotoBottom moves to the last entry
func (m *FileBrowserModel) GotoBottom() {
	if len(m.entries) > 0 {
		m.selectedIndex = len(m.entries) - 1
	}
}

// PageUp moves selection up by a page
func (m *FileBrowserModel) PageUp() {
	if m.state == FileBrowserStateViewingFile {
		m.previewViewport.PageUp()
		return
	}
	// Move by half the visible height
	pageSize := m.height / 2
	if pageSize < 1 {
		pageSize = 1
	}
	m.selectedIndex -= pageSize
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

// PageDown moves selection down by a page
func (m *FileBrowserModel) PageDown() {
	if m.state == FileBrowserStateViewingFile {
		m.previewViewport.PageDown()
		return
	}
	pageSize := m.height / 2
	if pageSize < 1 {
		pageSize = 1
	}
	m.selectedIndex += pageSize
	if m.selectedIndex >= len(m.entries) {
		m.selectedIndex = len(m.entries) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

// PushPath pushes current path to history before navigating
func (m *FileBrowserModel) PushPath() {
	m.pathHistory = append(m.pathHistory, m.currentPath)
}

// PopPath goes back to previous path in history
func (m *FileBrowserModel) PopPath() bool {
	if len(m.pathHistory) == 0 {
		return false
	}
	m.currentPath = m.pathHistory[len(m.pathHistory)-1]
	m.pathHistory = m.pathHistory[:len(m.pathHistory)-1]
	return true
}

// NavigateToParent navigates to parent directory
func (m *FileBrowserModel) NavigateToParent() string {
	parent := k8s.ParentPath(m.currentPath)
	if parent != m.currentPath {
		m.PushPath()
		m.currentPath = parent
		return parent
	}
	return ""
}

// NavigateToEntry returns the path to navigate to for the selected entry
// Returns empty string if no navigation should occur
func (m *FileBrowserModel) NavigateToEntry() (path string, isFile bool) {
	entry := m.SelectedEntry()
	if entry == nil {
		return "", false
	}

	// Handle ".." entry
	if entry.Name == ".." {
		return m.NavigateToParent(), false
	}

	// Skip "." entry
	if entry.Name == "." {
		return "", false
	}

	newPath := k8s.JoinPath(m.currentPath, entry.Name)

	if entry.IsDir {
		m.PushPath()
		m.currentPath = newPath
		return newPath, false
	}

	// It's a file
	return newPath, true
}

// Clear resets the file browser state
func (m *FileBrowserModel) Clear() {
	m.entries = make([]k8s.FileInfo, 0)
	m.selectedIndex = 0
	m.currentPath = "/"
	m.pathHistory = make([]string, 0)
	m.previewContent = ""
	m.viewingFile = ""
	m.errorMsg = ""
	m.state = FileBrowserStateIdle
}

// Update handles messages for the file browser
func (m FileBrowserModel) Update(msg tea.Msg) (FileBrowserModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keys differently based on state
		if m.state == FileBrowserStateViewingFile {
			switch msg.String() {
			case "j", "down":
				m.previewViewport.ScrollDown(1)
			case "k", "up":
				m.previewViewport.ScrollUp(1)
			case "g":
				m.previewViewport.GotoTop()
			case "G":
				m.previewViewport.GotoBottom()
			case "pgdown", " ":
				m.previewViewport.PageDown()
			case "pgup":
				m.previewViewport.PageUp()
			}
			// Esc/Backspace handled by app.go
			return m, nil
		}

		// Directory listing navigation
		switch msg.String() {
		case "j", "down":
			m.NavigateDown()
		case "k", "up":
			m.NavigateUp()
		case "g":
			m.GotoTop()
		case "G":
			m.GotoBottom()
		case "pgdown", " ":
			m.PageDown()
		case "pgup":
			m.PageUp()
		}
	}

	// Update viewport if viewing file
	if m.state == FileBrowserStateViewingFile {
		var vpCmd tea.Cmd
		m.previewViewport, vpCmd = m.previewViewport.Update(msg)
		cmds = append(cmds, vpCmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the file browser view
func (m FileBrowserModel) View() string {
	if !m.ready {
		return "Initializing file browser..."
	}

	if m.state == FileBrowserStateViewingFile {
		return m.viewFileContent()
	}

	return m.viewDirectoryListing()
}

// viewDirectoryListing renders the directory listing
func (m FileBrowserModel) viewDirectoryListing() string {
	var b strings.Builder

	// Header
	header := fmt.Sprintf("Files: %s/%s", m.pod, m.container)
	if m.namespace != "" {
		header = fmt.Sprintf("Files: %s/%s/%s", m.namespace, m.pod, m.container)
	}
	b.WriteString(header)
	b.WriteString("\n")

	// Path
	b.WriteString(fmt.Sprintf("Path: %s", m.currentPath))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", min(m.width, 80)))
	b.WriteString("\n")

	// Handle different states
	switch m.state {
	case FileBrowserStateLoading:
		b.WriteString("Loading...")
		return b.String()

	case FileBrowserStateError:
		b.WriteString(fmt.Sprintf("Error: %s", m.errorMsg))
		b.WriteString("\n\nPress 'backspace' to go back or 'esc' to exit")
		return b.String()

	case FileBrowserStateIdle:
		b.WriteString("No directory loaded")
		return b.String()
	}

	// Empty directory
	if len(m.entries) == 0 {
		b.WriteString("(empty directory)")
		b.WriteString("\n\nPress 'backspace' to go back or 'esc' to exit")
		return b.String()
	}

	// Calculate how many entries we can show
	availableHeight := m.height - 7 // header, path, separator, status lines
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Determine visible range (scrolling)
	start := 0
	if m.selectedIndex >= availableHeight {
		start = m.selectedIndex - availableHeight + 1
	}
	end := start + availableHeight
	if end > len(m.entries) {
		end = len(m.entries)
	}

	// Directory entries
	for i := start; i < end; i++ {
		entry := m.entries[i]
		prefix := "  "
		if i == m.selectedIndex {
			prefix = "> "
		}

		// Icon
		var icon string
		switch {
		case entry.IsDir:
			icon = "D "
		case entry.IsSymlink:
			icon = "L "
		default:
			icon = "F "
		}

		// Format entry line
		name := entry.Name
		if entry.IsSymlink && entry.LinkTarget != "" {
			name = fmt.Sprintf("%s -> %s", entry.Name, entry.LinkTarget)
		}

		// Truncate long names
		maxNameLen := m.width - 30
		if maxNameLen < 20 {
			maxNameLen = 20
		}
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		// Size (right-aligned)
		sizeStr := ""
		if !entry.IsDir {
			sizeStr = k8s.FormatSize(entry.Size)
		}

		// Permissions
		perms := ""
		if entry.Permissions != "" {
			perms = entry.Permissions
		}

		b.WriteString(fmt.Sprintf("%s%s%-*s  %6s  %s\n",
			prefix,
			icon,
			maxNameLen,
			name,
			sizeStr,
			perms,
		))
	}

	// Status bar
	b.WriteString(strings.Repeat("-", min(m.width, 80)))
	b.WriteString("\n")
	statusLine := m.buildStatusLine()
	b.WriteString(statusLine)

	return b.String()
}

// viewFileContent renders the file content preview
func (m FileBrowserModel) viewFileContent() string {
	var b strings.Builder

	// Header
	header := fmt.Sprintf("Files: %s/%s", m.pod, m.container)
	if m.namespace != "" {
		header = fmt.Sprintf("Files: %s/%s/%s", m.namespace, m.pod, m.container)
	}
	b.WriteString(header)
	b.WriteString("\n")

	// File path
	filePath := k8s.JoinPath(m.currentPath, m.viewingFile)
	b.WriteString(fmt.Sprintf("File: %s", filePath))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", min(m.width, 80)))
	b.WriteString("\n")

	// File content viewport
	b.WriteString(m.previewViewport.View())
	b.WriteString("\n")

	// Status bar
	b.WriteString(strings.Repeat("-", min(m.width, 80)))
	b.WriteString("\n")
	scrollPercent := int(m.previewViewport.ScrollPercent() * 100)
	b.WriteString(fmt.Sprintf("[VIEWING] %d%% | j/k: scroll | Backspace/Esc: back to list", scrollPercent))

	return b.String()
}

// buildStatusLine creates the status line at the bottom
func (m FileBrowserModel) buildStatusLine() string {
	var stateIndicator string
	switch m.state {
	case FileBrowserStateLoading:
		stateIndicator = "[LOADING]"
	case FileBrowserStateReady:
		stateIndicator = "[READY]"
	case FileBrowserStateError:
		stateIndicator = "[ERROR]"
	default:
		stateIndicator = "[IDLE]"
	}

	itemCount := fmt.Sprintf(" %d items", len(m.entries))
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.entries) {
		itemCount = fmt.Sprintf(" %d/%d", m.selectedIndex+1, len(m.entries))
	}

	return fmt.Sprintf("%s%s | Enter: open | Backspace: parent | Esc: back", stateIndicator, itemCount)
}

// MaxFilePreviewBytes returns the maximum bytes to read for file preview
func MaxFilePreviewBytes() int {
	return maxFilePreviewBytes
}
