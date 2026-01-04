package app

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxime/k8s-tui/internal/model"
	"github.com/maxime/k8s-tui/internal/ui"
)

// Model is the main application model
type Model struct {
	// Current view state
	view     model.ViewState
	prevView model.ViewState // For returning from overlays

	// Keybindings
	keys ui.KeyMap

	// Help component
	help     help.Model
	showHelp bool

	// Window dimensions
	width  int
	height int

	// Ready indicates if the app has received initial window size
	ready bool
}

// New creates a new application model with default state
func New() Model {
	return Model{
		view:     model.ViewPodList,
		prevView: model.ViewPodList,
		keys:     ui.DefaultKeyMap(),
		help:     help.New(),
		showHelp: false,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keybindings that work in any view
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		if m.showHelp {
			m.prevView = m.view
			m.view = model.ViewHelp
		} else {
			m.view = m.prevView
		}
		return m, nil

	case key.Matches(msg, m.keys.Back):
		return m.handleBack()
	}

	// View-specific keybindings
	switch m.view {
	case model.ViewPodList:
		return m.handlePodListKeys(msg)
	case model.ViewHelp:
		// Any key except ? closes help
		m.showHelp = false
		m.view = m.prevView
		return m, nil
	}

	return m, nil
}

// handleBack handles the escape/back key
func (m Model) handleBack() (tea.Model, tea.Cmd) {
	if m.view.IsOverlay() {
		m.view = m.prevView
		m.showHelp = false
		return m, nil
	}

	// From main views, go back to pod list
	if m.view != model.ViewPodList {
		m.view = model.ViewPodList
		return m, nil
	}

	return m, nil
}

// handlePodListKeys handles keys specific to the pod list view
func (m Model) handlePodListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Logs):
		m.view = model.ViewLogs
		return m, nil

	case key.Matches(msg, m.keys.Exec):
		m.view = model.ViewExec
		return m, nil

	case key.Matches(msg, m.keys.Files):
		m.view = model.ViewFiles
		return m, nil

	case key.Matches(msg, m.keys.Namespace):
		m.prevView = m.view
		m.view = model.ViewNamespaceSelector
		return m, nil

	case key.Matches(msg, m.keys.Context):
		m.prevView = m.view
		m.view = model.ViewContextSelector
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Build the main content based on current view
	var content string
	switch m.view {
	case model.ViewPodList:
		content = m.viewPodList()
	case model.ViewLogs:
		content = m.viewLogs()
	case model.ViewExec:
		content = m.viewExec()
	case model.ViewFiles:
		content = m.viewFiles()
	case model.ViewNamespaceSelector:
		content = m.viewNamespaceSelector()
	case model.ViewContextSelector:
		content = m.viewContextSelector()
	case model.ViewHelp:
		content = m.viewHelp()
	default:
		content = "Unknown view"
	}

	// Add help bar at bottom
	helpView := m.help.View(m.keys)

	return content + "\n\n" + helpView
}

// Placeholder view methods (to be implemented in later tasks)
func (m Model) viewPodList() string {
	return "K8s Pod Manager\n\n[Pod List View - Coming Soon]\n\nPress 'l' for logs, 'e' for exec, 'f' for files"
}

func (m Model) viewLogs() string {
	return "K8s Pod Manager > Logs\n\n[Log Streaming View - Coming Soon]\n\nPress 'esc' to go back"
}

func (m Model) viewExec() string {
	return "K8s Pod Manager > Exec\n\n[Command Execution View - Coming Soon]\n\nPress 'esc' to go back"
}

func (m Model) viewFiles() string {
	return "K8s Pod Manager > Files\n\n[File Browser View - Coming Soon]\n\nPress 'esc' to go back"
}

func (m Model) viewNamespaceSelector() string {
	return "Select Namespace\n\n[Namespace Selector - Coming Soon]\n\nPress 'esc' to cancel"
}

func (m Model) viewContextSelector() string {
	return "Select Context\n\n[Context Selector - Coming Soon]\n\nPress 'esc' to cancel"
}

func (m Model) viewHelp() string {
	return "Help\n\n" + m.help.View(m.keys) + "\n\nPress any key to close"
}

// Getters for testing
func (m Model) CurrentView() model.ViewState {
	return m.view
}

func (m Model) IsReady() bool {
	return m.ready
}

func (m Model) ShowingHelp() bool {
	return m.showHelp
}

func (m Model) Width() int {
	return m.width
}

func (m Model) Height() int {
	return m.height
}
