package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// LogViewState represents the state of the log streaming
type LogViewState int

const (
	LogViewStateIdle LogViewState = iota
	LogViewStateStreaming
	LogViewStatePaused
	LogViewStateError
	LogViewStateEnded
)

func (s LogViewState) String() string {
	switch s {
	case LogViewStateIdle:
		return "Idle"
	case LogViewStateStreaming:
		return "Streaming"
	case LogViewStatePaused:
		return "Paused"
	case LogViewStateError:
		return "Error"
	case LogViewStateEnded:
		return "Ended"
	default:
		return "Unknown"
	}
}

// LogViewModel represents the log viewing component
type LogViewModel struct {
	viewport viewport.Model

	// Log content
	lines       []string
	maxLines    int
	contentDirty bool

	// State
	state     LogViewState
	follow    bool
	pod       string
	container string
	namespace string
	errorMsg  string

	// Dimensions
	width  int
	height int
	ready  bool
}

// NewLogViewModel creates a new log view model
func NewLogViewModel() LogViewModel {
	return LogViewModel{
		lines:    make([]string, 0),
		maxLines: 10000, // Keep last 10k lines
		follow:   true,  // Start with follow mode enabled
		state:    LogViewStateIdle,
	}
}

// SetSize updates the viewport size
func (m *LogViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Reserve space for header (2 lines) and status bar (1 line)
	viewportHeight := height - 4
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	if !m.ready {
		m.viewport = viewport.New(width, viewportHeight)
		m.viewport.YPosition = 0
		m.ready = true
	} else {
		m.viewport.Width = width
		m.viewport.Height = viewportHeight
	}

	m.updateViewportContent()
}

// SetPodInfo sets the pod information for display
func (m *LogViewModel) SetPodInfo(namespace, pod, container string) {
	m.namespace = namespace
	m.pod = pod
	m.container = container
}

// SetState sets the current streaming state
func (m *LogViewModel) SetState(state LogViewState) {
	m.state = state
}

// SetError sets an error message
func (m *LogViewModel) SetError(err string) {
	m.errorMsg = err
	m.state = LogViewStateError
}

// IsFollow returns whether follow mode is enabled
func (m *LogViewModel) IsFollow() bool {
	return m.follow
}

// ToggleFollow toggles follow mode
func (m *LogViewModel) ToggleFollow() {
	m.follow = !m.follow
	if m.follow {
		// Jump to bottom when enabling follow
		m.viewport.GotoBottom()
	}
}

// AddLine adds a new log line
func (m *LogViewModel) AddLine(line string) {
	m.lines = append(m.lines, line)

	// Trim old lines if we exceed max
	if len(m.lines) > m.maxLines {
		// Remove oldest 10% of lines
		trimCount := m.maxLines / 10
		m.lines = m.lines[trimCount:]
	}

	m.contentDirty = true
}

// AddLines adds multiple log lines
func (m *LogViewModel) AddLines(lines []string) {
	for _, line := range lines {
		m.AddLine(line)
	}
}

// Clear clears all log lines
func (m *LogViewModel) Clear() {
	m.lines = make([]string, 0)
	m.contentDirty = true
	m.updateViewportContent()
}

// LineCount returns the number of log lines
func (m *LogViewModel) LineCount() int {
	return len(m.lines)
}

// State returns the current log view state
func (m *LogViewModel) State() LogViewState {
	return m.state
}

// updateViewportContent updates the viewport with current lines
func (m *LogViewModel) updateViewportContent() {
	if !m.ready {
		return
	}

	content := strings.Join(m.lines, "\n")
	m.viewport.SetContent(content)

	if m.follow {
		m.viewport.GotoBottom()
	}

	m.contentDirty = false
}

// Update handles messages for the log view
func (m LogViewModel) Update(msg tea.Msg) (LogViewModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			// Go to top
			m.viewport.GotoTop()
			m.follow = false
			return m, nil
		case "G":
			// Go to bottom
			m.viewport.GotoBottom()
			m.follow = true
			return m, nil
		}
	}

	// Update viewport content if dirty
	if m.contentDirty {
		m.updateViewportContent()
	}

	// Pass remaining messages to viewport
	m.viewport, cmd = m.viewport.Update(msg)

	// Disable follow if user scrolled up
	if !m.viewport.AtBottom() {
		m.follow = false
	}

	return m, cmd
}

// View renders the log view
func (m LogViewModel) View() string {
	if !m.ready {
		return "Initializing log view..."
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("Logs: %s/%s", m.pod, m.container)
	if m.namespace != "" {
		header = fmt.Sprintf("Logs: %s/%s/%s", m.namespace, m.pod, m.container)
	}
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", min(len(header)+10, m.width)))
	b.WriteString("\n")

	// Viewport content
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Status bar
	statusLine := m.buildStatusLine()
	b.WriteString(statusLine)

	return b.String()
}

// buildStatusLine creates the status line at the bottom
func (m LogViewModel) buildStatusLine() string {
	// State indicator
	var stateIndicator string
	switch m.state {
	case LogViewStateStreaming:
		stateIndicator = "[STREAMING]"
	case LogViewStatePaused:
		stateIndicator = "[PAUSED]"
	case LogViewStateError:
		stateIndicator = fmt.Sprintf("[ERROR: %s]", m.errorMsg)
	case LogViewStateEnded:
		stateIndicator = "[STREAM ENDED]"
	default:
		stateIndicator = "[IDLE]"
	}

	// Follow indicator
	followIndicator := ""
	if m.follow {
		followIndicator = " [FOLLOW]"
	}

	// Line count and scroll position
	scrollInfo := fmt.Sprintf(" Lines: %d | %d%%",
		len(m.lines),
		int(m.viewport.ScrollPercent()*100))

	return fmt.Sprintf("%s%s%s", stateIndicator, followIndicator, scrollInfo)
}

// ScrollUp scrolls the viewport up
func (m *LogViewModel) ScrollUp(lines int) {
	m.follow = false
	for i := 0; i < lines; i++ {
		m.viewport.LineUp(1)
	}
}

// ScrollDown scrolls the viewport down
func (m *LogViewModel) ScrollDown(lines int) {
	for i := 0; i < lines; i++ {
		m.viewport.LineDown(1)
	}
	if m.viewport.AtBottom() {
		m.follow = true
	}
}

// PageUp scrolls the viewport up one page
func (m *LogViewModel) PageUp() {
	m.follow = false
	m.viewport.ViewUp()
}

// PageDown scrolls the viewport down one page
func (m *LogViewModel) PageDown() {
	m.viewport.ViewDown()
	if m.viewport.AtBottom() {
		m.follow = true
	}
}

// GotoTop scrolls to the top of the logs
func (m *LogViewModel) GotoTop() {
	m.follow = false
	m.viewport.GotoTop()
}

// GotoBottom scrolls to the bottom of the logs
func (m *LogViewModel) GotoBottom() {
	m.follow = true
	m.viewport.GotoBottom()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
