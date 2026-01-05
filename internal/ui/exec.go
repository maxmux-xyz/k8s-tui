package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// ExecViewState represents the state of the exec view
type ExecViewState int

// Exec view state constants for tracking execution status.
const (
	ExecViewStateIdle ExecViewState = iota
	ExecViewStateRunning
	ExecViewStateComplete
	ExecViewStateError
)

func (s ExecViewState) String() string {
	switch s {
	case ExecViewStateIdle:
		return "Idle"
	case ExecViewStateRunning:
		return "Running"
	case ExecViewStateComplete:
		return "Complete"
	case ExecViewStateError:
		return "Error"
	default:
		return "Unknown"
	}
}

const (
	maxHistorySize = 50
	maxOutputLines = 5000
)

// ExecViewModel represents the command execution UI component
type ExecViewModel struct {
	input    textinput.Model
	viewport viewport.Model

	// Output content
	outputLines []string

	// Command history
	history      []string
	historyIndex int

	// State
	state     ExecViewState
	pod       string
	container string
	namespace string
	errorMsg  string

	// Dimensions
	width  int
	height int
	ready  bool
}

// NewExecViewModel creates a new exec view model
func NewExecViewModel() ExecViewModel {
	ti := textinput.New()
	ti.Placeholder = "Enter command (e.g., ls -la, pwd, env)"
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60

	return ExecViewModel{
		input:        ti,
		outputLines:  make([]string, 0),
		history:      make([]string, 0),
		historyIndex: -1,
		state:        ExecViewStateIdle,
	}
}

// SetSize updates the viewport size
func (m *ExecViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Input takes 1 line, header 3 lines, status 1 line
	viewportHeight := height - 6
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	m.input.Width = width - 4 // Leave room for prompt

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
func (m *ExecViewModel) SetPodInfo(namespace, pod, container string) {
	m.namespace = namespace
	m.pod = pod
	m.container = container
}

// SetState sets the current execution state
func (m *ExecViewModel) SetState(state ExecViewState) {
	m.state = state
}

// SetError sets an error message
func (m *ExecViewModel) SetError(err string) {
	m.errorMsg = err
	m.state = ExecViewStateError
}

// State returns the current exec view state
func (m *ExecViewModel) State() ExecViewState {
	return m.state
}

// GetCommand returns the current command input
func (m *ExecViewModel) GetCommand() string {
	return m.input.Value()
}

// ClearInput clears the command input
func (m *ExecViewModel) ClearInput() {
	m.input.SetValue("")
}

// AddToHistory adds a command to the history
func (m *ExecViewModel) AddToHistory(cmd string) {
	if cmd == "" {
		return
	}

	// Don't add duplicate of last command
	if len(m.history) > 0 && m.history[len(m.history)-1] == cmd {
		return
	}

	m.history = append(m.history, cmd)

	// Trim history if too large
	if len(m.history) > maxHistorySize {
		m.history = m.history[1:]
	}

	m.historyIndex = len(m.history) // Reset to end
}

// HistoryPrev moves to the previous command in history
func (m *ExecViewModel) HistoryPrev() {
	if len(m.history) == 0 {
		return
	}

	if m.historyIndex > 0 {
		m.historyIndex--
	}

	if m.historyIndex < len(m.history) {
		m.input.SetValue(m.history[m.historyIndex])
		// Move cursor to end
		m.input.CursorEnd()
	}
}

// HistoryNext moves to the next command in history
func (m *ExecViewModel) HistoryNext() {
	if len(m.history) == 0 {
		return
	}

	if m.historyIndex < len(m.history)-1 {
		m.historyIndex++
		m.input.SetValue(m.history[m.historyIndex])
		m.input.CursorEnd()
	} else if m.historyIndex == len(m.history)-1 {
		// At end of history, clear input
		m.historyIndex = len(m.history)
		m.input.SetValue("")
	}
}

// AddOutput adds output text (stdout or stderr)
func (m *ExecViewModel) AddOutput(text string, isStderr bool) {
	if text == "" {
		return
	}

	// Split into lines
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		prefix := ""
		if isStderr {
			prefix = "[stderr] "
		}

		m.outputLines = append(m.outputLines, prefix+line)
	}

	// Trim if too large
	if len(m.outputLines) > maxOutputLines {
		trimCount := maxOutputLines / 10
		m.outputLines = m.outputLines[trimCount:]
	}

	m.updateViewportContent()
}

// AddCommandMarker adds a visual separator for a new command
func (m *ExecViewModel) AddCommandMarker(cmd string) {
	m.outputLines = append(m.outputLines,
		"",
		fmt.Sprintf("$ %s", cmd),
		strings.Repeat("-", min(len(cmd)+4, m.width-2)),
	)
	m.updateViewportContent()
}

// Clear clears all output
func (m *ExecViewModel) Clear() {
	m.outputLines = make([]string, 0)
	m.updateViewportContent()
}

// Focus sets focus on the input field
func (m *ExecViewModel) Focus() {
	m.input.Focus()
}

// Blur removes focus from the input field
func (m *ExecViewModel) Blur() {
	m.input.Blur()
}

// IsFocused returns whether the input is focused
func (m *ExecViewModel) IsFocused() bool {
	return m.input.Focused()
}

// updateViewportContent updates the viewport with current output
func (m *ExecViewModel) updateViewportContent() {
	if !m.ready {
		return
	}

	content := strings.Join(m.outputLines, "\n")
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

// Update handles messages for the exec view
func (m ExecViewModel) Update(msg tea.Msg) (ExecViewModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't handle keys when running a command
		if m.state == ExecViewStateRunning {
			return m, nil
		}

		switch msg.String() {
		case "up":
			if m.input.Focused() {
				m.HistoryPrev()
				return m, nil
			}
			// Scroll viewport
			m.viewport.ScrollUp(1)
			return m, nil

		case "down":
			if m.input.Focused() {
				m.HistoryNext()
				return m, nil
			}
			// Scroll viewport
			m.viewport.ScrollDown(1)
			return m, nil

		case "pgup":
			m.viewport.PageUp()
			return m, nil

		case "pgdown":
			m.viewport.PageDown()
			return m, nil

		case "tab":
			// Toggle focus between input and viewport
			if m.input.Focused() {
				m.input.Blur()
			} else {
				m.input.Focus()
			}
			return m, nil
		}
	}

	// Update input if focused
	if m.input.Focused() {
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		cmds = append(cmds, inputCmd)
	}

	// Update viewport
	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

// View renders the exec view
func (m ExecViewModel) View() string {
	if !m.ready {
		return "Initializing exec view..."
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("Exec: %s/%s", m.pod, m.container)
	if m.namespace != "" {
		header = fmt.Sprintf("Exec: %s/%s/%s", m.namespace, m.pod, m.container)
	}
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", min(len(header)+10, m.width)))
	b.WriteString("\n")

	// Output viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Input prompt
	prompt := "> "
	if m.state == ExecViewStateRunning {
		prompt = "* "
	}
	b.WriteString(prompt)
	b.WriteString(m.input.View())
	b.WriteString("\n")

	// Status bar
	statusLine := m.buildStatusLine()
	b.WriteString(statusLine)

	return b.String()
}

// buildStatusLine creates the status line at the bottom
func (m ExecViewModel) buildStatusLine() string {
	var stateIndicator string
	switch m.state {
	case ExecViewStateRunning:
		stateIndicator = "[RUNNING]"
	case ExecViewStateComplete:
		stateIndicator = "[COMPLETE]"
	case ExecViewStateError:
		if m.errorMsg != "" {
			stateIndicator = fmt.Sprintf("[ERROR: %s]", m.errorMsg)
		} else {
			stateIndicator = "[ERROR]"
		}
	default:
		stateIndicator = "[READY]"
	}

	// History info
	historyInfo := ""
	if len(m.history) > 0 {
		historyInfo = fmt.Sprintf(" | History: %d", len(m.history))
	}

	// Focus info
	focusInfo := " | Tab: switch focus"

	return fmt.Sprintf("%s%s%s", stateIndicator, historyInfo, focusInfo)
}

// ScrollUp scrolls the output viewport up
func (m *ExecViewModel) ScrollUp(lines int) {
	m.viewport.ScrollUp(lines)
}

// ScrollDown scrolls the output viewport down
func (m *ExecViewModel) ScrollDown(lines int) {
	m.viewport.ScrollDown(lines)
}

// PageUp scrolls the output viewport up one page
func (m *ExecViewModel) PageUp() {
	m.viewport.PageUp()
}

// PageDown scrolls the output viewport down one page
func (m *ExecViewModel) PageDown() {
	m.viewport.PageDown()
}

// GotoTop scrolls to the top of the output
func (m *ExecViewModel) GotoTop() {
	m.viewport.GotoTop()
}

// GotoBottom scrolls to the bottom of the output
func (m *ExecViewModel) GotoBottom() {
	m.viewport.GotoBottom()
}
