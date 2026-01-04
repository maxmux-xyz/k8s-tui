// Package app provides the Bubble Tea application model and UI logic
// for the Kubernetes TUI pod manager.
package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxime/k8s-tui/internal/k8s"
	"github.com/maxime/k8s-tui/internal/model"
	"github.com/maxime/k8s-tui/internal/ui"
)

// Log streaming message types
type logLineMsg struct {
	line k8s.LogLine
}

type logStreamStartedMsg struct {
	container string
}

type logStreamErrorMsg struct {
	err error
}

type logStreamEndedMsg struct{}

// Messages for async operations
type k8sClientReadyMsg struct {
	client *k8s.Client
	err    error
}

type podsLoadedMsg struct {
	pods []k8s.PodInfo
	err  error
}

type namespacesLoadedMsg struct {
	namespaces []k8s.NamespaceInfo
	err        error
}

type contextsLoadedMsg struct {
	contexts       []k8s.ContextInfo
	currentContext string
	err            error
}

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

	// K8s client
	k8sClient *k8s.Client
	k8sErr    error

	// Data
	pods       []k8s.PodInfo
	namespaces []k8s.NamespaceInfo
	contexts   []k8s.ContextInfo

	// Loading states
	loadingK8s        bool
	loadingPods       bool
	loadingNamespaces bool

	// Selected indices
	selectedPodIndex       int
	selectedNamespaceIndex int
	selectedContextIndex   int

	// Log streaming state
	logView           ui.LogViewModel
	logCancel         context.CancelFunc
	logChan           <-chan k8s.LogLine
	logStreamActive   bool
	selectedContainer string
}

// New creates a new application model with default state
func New() Model {
	return Model{
		view:       model.ViewPodList,
		prevView:   model.ViewPodList,
		keys:       ui.DefaultKeyMap(),
		help:       help.New(),
		showHelp:   false,
		loadingK8s: true,
		logView:    ui.NewLogViewModel(),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return m.initK8sClient
}

// initK8sClient initializes the Kubernetes client
func (m Model) initK8sClient() tea.Msg {
	client, err := k8s.NewClient()
	return k8sClientReadyMsg{client: client, err: err}
}

// loadPods fetches pods from the current namespace
func (m Model) loadPods() tea.Msg {
	if m.k8sClient == nil {
		return podsLoadedMsg{err: fmt.Errorf("k8s client not initialized")}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pods, err := m.k8sClient.ListPods(ctx, "")
	return podsLoadedMsg{pods: pods, err: err}
}

// loadNamespaces fetches namespaces from the cluster
func (m Model) loadNamespaces() tea.Msg {
	if m.k8sClient == nil {
		return namespacesLoadedMsg{err: fmt.Errorf("k8s client not initialized")}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	namespaces, err := m.k8sClient.ListNamespaces(ctx)
	return namespacesLoadedMsg{namespaces: namespaces, err: err}
}

// loadContexts loads available contexts
func (m Model) loadContexts() tea.Msg {
	if m.k8sClient == nil {
		return contextsLoadedMsg{err: fmt.Errorf("k8s client not initialized")}
	}

	contexts := m.k8sClient.ListContexts()
	return contextsLoadedMsg{
		contexts:       contexts,
		currentContext: m.k8sClient.CurrentContext(),
	}
}

// logStreamChanMsg carries the log channel after stream creation
type logStreamChanMsg struct {
	logChan <-chan k8s.LogLine
}

// initLogStream prepares and starts log streaming for the selected pod
func (m *Model) initLogStream() tea.Cmd {
	if m.k8sClient == nil {
		return func() tea.Msg {
			return logStreamErrorMsg{err: fmt.Errorf("k8s client not initialized")}
		}
	}

	if m.selectedPodIndex >= len(m.pods) {
		return func() tea.Msg {
			return logStreamErrorMsg{err: fmt.Errorf("no pod selected")}
		}
	}

	pod := m.pods[m.selectedPodIndex]

	// Stop any existing stream
	m.stopLogStream()

	// Determine container to use
	container := m.selectedContainer
	if container == "" && len(pod.Containers) > 0 {
		container = pod.Containers[0].Name
	}

	// Set up log view
	m.logView.Clear()
	m.logView.SetPodInfo(pod.Namespace, pod.Name, container)
	m.logView.SetState(ui.LogViewStateStreaming)
	m.selectedContainer = container

	// Create context for this stream
	ctx, cancel := context.WithCancel(context.Background())
	m.logCancel = cancel
	m.logStreamActive = true

	// Capture values for closure
	namespace := pod.Namespace
	podName := pod.Name
	client := m.k8sClient

	return func() tea.Msg {
		opts := k8s.LogOptions{
			Namespace: namespace,
			Pod:       podName,
			Container: container,
			Follow:    true,
			TailLines: 100, // Start with last 100 lines
		}

		logChan, err := client.StreamLogs(ctx, opts)
		if err != nil {
			return logStreamErrorMsg{err: err}
		}

		// Return the channel so we can store it
		return logStreamChanMsg{logChan: logChan}
	}
}

// waitForNextLogLine waits for the next line from an existing channel
func waitForNextLogLine(logChan <-chan k8s.LogLine) tea.Cmd {
	if logChan == nil {
		return nil
	}
	return func() tea.Msg {
		line, ok := <-logChan
		if !ok {
			return logStreamEndedMsg{}
		}
		if line.Error != nil {
			return logStreamErrorMsg{err: line.Error}
		}
		return logLineMsg{line: line}
	}
}

// stopLogStream stops the current log stream
func (m *Model) stopLogStream() {
	if m.logCancel != nil {
		m.logCancel()
		m.logCancel = nil
	}
	m.logChan = nil
	m.logStreamActive = false
	m.logView.SetState(ui.LogViewStateEnded)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.logView.SetSize(msg.Width, msg.Height-4) // Reserve space for header/footer
		m.ready = true
		return m, nil

	case k8sClientReadyMsg:
		m.loadingK8s = false
		if msg.err != nil {
			m.k8sErr = msg.err
			return m, nil
		}
		m.k8sClient = msg.client
		m.loadingPods = true
		// Load pods and contexts after client is ready
		return m, tea.Batch(m.loadPods, m.loadContexts)

	case podsLoadedMsg:
		m.loadingPods = false
		if msg.err != nil {
			m.k8sErr = msg.err
			return m, nil
		}
		m.pods = msg.pods
		m.k8sErr = nil
		return m, nil

	case namespacesLoadedMsg:
		m.loadingNamespaces = false
		if msg.err != nil {
			m.k8sErr = msg.err
			return m, nil
		}
		m.namespaces = msg.namespaces
		// Find and select current namespace
		for i, ns := range m.namespaces {
			if ns.IsCurrent {
				m.selectedNamespaceIndex = i
				break
			}
		}
		return m, nil

	case contextsLoadedMsg:
		if msg.err != nil {
			m.k8sErr = msg.err
			return m, nil
		}
		m.contexts = msg.contexts
		// Find and select current context
		for i, ctx := range m.contexts {
			if ctx.IsCurrent {
				m.selectedContextIndex = i
				break
			}
		}
		return m, nil

	case logStreamChanMsg:
		// Store the channel and start reading
		m.logChan = msg.logChan
		m.logView.SetState(ui.LogViewStateStreaming)
		return m, waitForNextLogLine(m.logChan)

	case logStreamStartedMsg:
		m.selectedContainer = msg.container
		m.logView.SetState(ui.LogViewStateStreaming)
		return m, nil

	case logLineMsg:
		if msg.line.Error != nil {
			m.logView.SetError(msg.line.Error.Error())
			m.logStreamActive = false
			return m, nil
		}
		m.logView.AddLine(msg.line.Content)
		// Continue reading if stream is active
		if m.logStreamActive && m.view == model.ViewLogs && m.logChan != nil {
			return m, waitForNextLogLine(m.logChan)
		}
		return m, nil

	case logStreamErrorMsg:
		m.logView.SetError(msg.err.Error())
		m.logStreamActive = false
		return m, nil

	case logStreamEndedMsg:
		m.logView.SetState(ui.LogViewStateEnded)
		m.logStreamActive = false
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
	case model.ViewLogs:
		return m.handleLogViewKeys(msg)
	case model.ViewNamespaceSelector:
		return m.handleNamespaceSelectorKeys(msg)
	case model.ViewContextSelector:
		return m.handleContextSelectorKeys(msg)
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

	// From log view, stop streaming and go back
	if m.view == model.ViewLogs {
		m.stopLogStream()
		m.view = model.ViewPodList
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
	case key.Matches(msg, m.keys.Up):
		if m.selectedPodIndex > 0 {
			m.selectedPodIndex--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.selectedPodIndex < len(m.pods)-1 {
			m.selectedPodIndex++
		}
		return m, nil

	case key.Matches(msg, m.keys.Logs):
		if len(m.pods) > 0 {
			m.view = model.ViewLogs
			m.selectedContainer = "" // Reset to use first container
			cmd := m.initLogStream()
			return m, cmd
		}
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
		m.loadingNamespaces = true
		return m, m.loadNamespaces

	case key.Matches(msg, m.keys.Context):
		m.prevView = m.view
		m.view = model.ViewContextSelector
		return m, m.loadContexts

	case key.Matches(msg, m.keys.Refresh):
		m.loadingPods = true
		return m, m.loadPods
	}

	return m, nil
}

// handleNamespaceSelectorKeys handles keys for namespace selection
func (m Model) handleNamespaceSelectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.selectedNamespaceIndex > 0 {
			m.selectedNamespaceIndex--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.selectedNamespaceIndex < len(m.namespaces)-1 {
			m.selectedNamespaceIndex++
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		if m.selectedNamespaceIndex < len(m.namespaces) {
			ns := m.namespaces[m.selectedNamespaceIndex]
			m.k8sClient.SetNamespace(ns.Name)
			m.view = m.prevView
			m.loadingPods = true
			return m, m.loadPods
		}
		return m, nil
	}

	return m, nil
}

// handleContextSelectorKeys handles keys for context selection
func (m Model) handleContextSelectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Up):
		if m.selectedContextIndex > 0 {
			m.selectedContextIndex--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.selectedContextIndex < len(m.contexts)-1 {
			m.selectedContextIndex++
		}
		return m, nil

	case key.Matches(msg, m.keys.Enter):
		if m.selectedContextIndex < len(m.contexts) {
			ctx := m.contexts[m.selectedContextIndex]
			if err := m.k8sClient.SwitchContext(ctx.Name); err != nil {
				m.k8sErr = err
				return m, nil
			}
			m.view = m.prevView
			m.loadingPods = true
			return m, tea.Batch(m.loadPods, m.loadContexts)
		}
		return m, nil
	}

	return m, nil
}

// handleLogViewKeys handles keys specific to the log view
func (m Model) handleLogViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.logView.ScrollDown(1)
		return m, nil

	case "k", "up":
		m.logView.ScrollUp(1)
		return m, nil

	case "g":
		m.logView.GotoTop()
		return m, nil

	case "G":
		m.logView.GotoBottom()
		return m, nil

	case "f", "F":
		m.logView.ToggleFollow()
		return m, nil

	case "pgdown", " ":
		m.logView.PageDown()
		return m, nil

	case "pgup":
		m.logView.PageUp()
		return m, nil
	}

	// Pass to log view for viewport handling
	var cmd tea.Cmd
	m.logView, cmd = m.logView.Update(msg)
	return m, cmd
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

// viewPodList renders the pod list view
func (m Model) viewPodList() string {
	var b strings.Builder

	// Header
	b.WriteString("K8s Pod Manager")
	if m.k8sClient != nil {
		b.WriteString(fmt.Sprintf(" | Context: %s | Namespace: %s",
			m.k8sClient.CurrentContext(),
			m.k8sClient.CurrentNamespace()))
	}
	b.WriteString("\n\n")

	// Error state
	if m.k8sErr != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n\n", m.k8sErr))
		b.WriteString("Press 'r' to retry, 'c' to change context, 'n' to change namespace")
		return b.String()
	}

	// Loading state
	if m.loadingK8s {
		b.WriteString("Connecting to Kubernetes cluster...")
		return b.String()
	}

	if m.loadingPods {
		b.WriteString("Loading pods...")
		return b.String()
	}

	// Empty state
	if len(m.pods) == 0 {
		b.WriteString("No pods found in this namespace.\n\n")
		b.WriteString("Press 'n' to switch namespace, 'c' to switch context")
		return b.String()
	}

	// Pod list header
	b.WriteString(fmt.Sprintf("%-40s %-12s %-8s %-10s %-15s\n",
		"NAME", "STATUS", "READY", "RESTARTS", "AGE"))
	b.WriteString(strings.Repeat("-", 85) + "\n")

	// Pod list
	for i := range m.pods {
		pod := &m.pods[i]
		prefix := "  "
		if i == m.selectedPodIndex {
			prefix = "> "
		}

		age := formatAge(pod.Age)
		b.WriteString(fmt.Sprintf("%s%-38s %-12s %-8s %-10d %-15s\n",
			prefix,
			truncate(pod.Name, 38),
			pod.Status,
			pod.Ready,
			pod.Restarts,
			age))
	}

	b.WriteString("\n")
	b.WriteString("Press 'l' for logs, 'e' for exec, 'f' for files, 'r' to refresh")

	return b.String()
}

func (m Model) viewLogs() string {
	if m.selectedPodIndex >= len(m.pods) {
		return "K8s Pod Manager > Logs\n\n[No pod selected]\n\nPress 'esc' to go back"
	}

	var b strings.Builder

	// Header with context info
	b.WriteString("K8s Pod Manager > Logs")
	if m.k8sClient != nil {
		b.WriteString(fmt.Sprintf(" | Context: %s", m.k8sClient.CurrentContext()))
	}
	b.WriteString("\n")

	// Log view content
	b.WriteString(m.logView.View())

	// Help text
	b.WriteString("\n")
	b.WriteString("j/k: scroll | g/G: top/bottom | f: toggle follow | esc: back")

	return b.String()
}

func (m Model) viewExec() string {
	if m.selectedPodIndex < len(m.pods) {
		pod := m.pods[m.selectedPodIndex]
		return fmt.Sprintf("K8s Pod Manager > Exec > %s\n\n[Command Execution View - Coming Soon]\n\nPress 'esc' to go back", pod.Name)
	}
	return "K8s Pod Manager > Exec\n\n[No pod selected]\n\nPress 'esc' to go back"
}

func (m Model) viewFiles() string {
	if m.selectedPodIndex < len(m.pods) {
		pod := m.pods[m.selectedPodIndex]
		return fmt.Sprintf("K8s Pod Manager > Files > %s\n\n[File Browser View - Coming Soon]\n\nPress 'esc' to go back", pod.Name)
	}
	return "K8s Pod Manager > Files\n\n[No pod selected]\n\nPress 'esc' to go back"
}

func (m Model) viewNamespaceSelector() string {
	var b strings.Builder

	b.WriteString("Select Namespace\n\n")

	if m.loadingNamespaces {
		b.WriteString("Loading namespaces...")
		return b.String()
	}

	if len(m.namespaces) == 0 {
		b.WriteString("No namespaces found.\n")
		b.WriteString("\nPress 'esc' to cancel")
		return b.String()
	}

	for i, ns := range m.namespaces {
		prefix := "  "
		if i == m.selectedNamespaceIndex {
			prefix = "> "
		}
		current := ""
		if ns.IsCurrent {
			current = " (current)"
		}
		b.WriteString(fmt.Sprintf("%s%s%s\n", prefix, ns.Name, current))
	}

	b.WriteString("\nPress 'enter' to select, 'esc' to cancel")

	return b.String()
}

func (m Model) viewContextSelector() string {
	var b strings.Builder

	b.WriteString("Select Context\n\n")

	if len(m.contexts) == 0 {
		b.WriteString("No contexts found.\n")
		b.WriteString("\nPress 'esc' to cancel")
		return b.String()
	}

	for i, ctx := range m.contexts {
		prefix := "  "
		if i == m.selectedContextIndex {
			prefix = "> "
		}
		current := ""
		if ctx.IsCurrent {
			current = " (current)"
		}
		b.WriteString(fmt.Sprintf("%s%s%s\n", prefix, ctx.Name, current))
		b.WriteString(fmt.Sprintf("    Cluster: %s, Namespace: %s\n", ctx.Cluster, ctx.Namespace))
	}

	b.WriteString("\nPress 'enter' to select, 'esc' to cancel")

	return b.String()
}

func (m Model) viewHelp() string {
	return "Help\n\n" + m.help.View(m.keys) + "\n\nPress any key to close"
}

// CurrentView returns the current view state (used for testing).
func (m Model) CurrentView() model.ViewState {
	return m.view
}

// IsReady returns whether the app has been initialized with window dimensions.
func (m Model) IsReady() bool {
	return m.ready
}

// ShowingHelp returns whether the help overlay is currently displayed.
func (m Model) ShowingHelp() bool {
	return m.showHelp
}

// Width returns the current window width.
func (m Model) Width() int {
	return m.width
}

// Height returns the current window height.
func (m Model) Height() int {
	return m.height
}

// Pods returns the current list of pods.
func (m Model) Pods() []k8s.PodInfo {
	return m.pods
}

// SelectedPodIndex returns the index of the currently selected pod.
func (m Model) SelectedPodIndex() int {
	return m.selectedPodIndex
}

// K8sError returns the most recent Kubernetes client error, if any.
func (m Model) K8sError() error {
	return m.k8sErr
}

// Helper functions

func formatAge(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
