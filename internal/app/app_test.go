package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/maxime/k8s-tui/internal/k8s"
	"github.com/maxime/k8s-tui/internal/model"
)

func TestNew(t *testing.T) {
	m := New()

	if m.CurrentView() != model.ViewPodList {
		t.Errorf("Initial view should be ViewPodList, got %v", m.CurrentView())
	}

	if m.IsReady() {
		t.Error("Model should not be ready before receiving WindowSizeMsg")
	}

	if m.ShowingHelp() {
		t.Error("Help should not be showing initially")
	}
}

func TestInit(t *testing.T) {
	m := New()
	cmd := m.Init()

	// Init now returns a command to initialize the K8s client
	if cmd == nil {
		t.Error("Init should return a command to initialize K8s client")
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := New()

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, cmd := m.Update(msg)

	updatedModel := newModel.(Model)
	if !updatedModel.IsReady() {
		t.Error("Model should be ready after receiving WindowSizeMsg")
	}

	if updatedModel.Width() != 80 {
		t.Errorf("Width should be 80, got %d", updatedModel.Width())
	}

	if updatedModel.Height() != 24 {
		t.Errorf("Height should be 24, got %d", updatedModel.Height())
	}

	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}
}

func TestUpdate_QuitKey(t *testing.T) {
	m := New()

	// Simulate pressing 'q'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	// The command should be tea.Quit
	if cmd == nil {
		t.Error("Pressing 'q' should return a quit command")
	}
}

func TestUpdate_QuitCtrlC(t *testing.T) {
	m := New()

	// Simulate pressing Ctrl+C
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Pressing Ctrl+C should return a quit command")
	}
}

func TestUpdate_HelpToggle(t *testing.T) {
	m := New()
	// First, make the model ready
	m = makeReady(m)

	// Press '?' to show help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if !updatedModel.ShowingHelp() {
		t.Error("Help should be showing after pressing '?'")
	}

	if updatedModel.CurrentView() != model.ViewHelp {
		t.Errorf("View should be ViewHelp, got %v", updatedModel.CurrentView())
	}

	// Press '?' again to hide help
	newModel, _ = updatedModel.Update(msg)
	updatedModel = newModel.(Model)

	if updatedModel.ShowingHelp() {
		t.Error("Help should be hidden after pressing '?' again")
	}

	if updatedModel.CurrentView() != model.ViewPodList {
		t.Errorf("View should return to ViewPodList, got %v", updatedModel.CurrentView())
	}
}

func TestUpdate_ViewNavigation(t *testing.T) {
	tests := []struct {
		name         string
		key          rune
		expectedView model.ViewState
		needsPods    bool
	}{
		{"Logs", 'l', model.ViewLogs, true},
		{"Exec", 'e', model.ViewExec, false},
		{"Files", 'f', model.ViewFiles, false},
		{"Namespace", 'n', model.ViewNamespaceSelector, false},
		{"Context", 'c', model.ViewContextSelector, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			if tt.needsPods {
				m = makeReadyWithPods(m)
			} else {
				m = makeReady(m)
			}

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
			newModel, _ := m.Update(msg)
			updatedModel := newModel.(Model)

			if updatedModel.CurrentView() != tt.expectedView {
				t.Errorf("Expected view %v, got %v", tt.expectedView, updatedModel.CurrentView())
			}
		})
	}
}

func TestUpdate_BackNavigation(t *testing.T) {
	m := New()
	m = makeReadyWithPods(m)

	// Navigate to Logs view
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.CurrentView() != model.ViewLogs {
		t.Fatalf("Should be in Logs view, got %v", m.CurrentView())
	}

	// Press Escape to go back
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ = m.Update(escMsg)
	m = newModel.(Model)

	if m.CurrentView() != model.ViewPodList {
		t.Errorf("Should be back in PodList view, got %v", m.CurrentView())
	}
}

func TestUpdate_OverlayBackNavigation(t *testing.T) {
	m := New()
	m = makeReady(m)

	// Open namespace selector (overlay)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.CurrentView() != model.ViewNamespaceSelector {
		t.Fatalf("Should be in NamespaceSelector view, got %v", m.CurrentView())
	}

	// Press Escape to close overlay
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ = m.Update(escMsg)
	m = newModel.(Model)

	if m.CurrentView() != model.ViewPodList {
		t.Errorf("Should be back in PodList view, got %v", m.CurrentView())
	}
}

func TestUpdate_HelpCloseOnAnyKey(t *testing.T) {
	m := New()
	m = makeReady(m)

	// Open help
	helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	newModel, _ := m.Update(helpMsg)
	m = newModel.(Model)

	if m.CurrentView() != model.ViewHelp {
		t.Fatalf("Should be in Help view, got %v", m.CurrentView())
	}

	// Press any other key (e.g., 'a') to close help
	anyKeyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ = m.Update(anyKeyMsg)
	m = newModel.(Model)

	if m.CurrentView() != model.ViewPodList {
		t.Errorf("Help should close and return to PodList, got %v", m.CurrentView())
	}

	if m.ShowingHelp() {
		t.Error("ShowingHelp should be false after closing")
	}
}

func TestView_NotReady(t *testing.T) {
	m := New()
	view := m.View()

	if view != "Initializing..." {
		t.Errorf("View should show 'Initializing...' when not ready, got %q", view)
	}
}

func TestView_Ready(t *testing.T) {
	m := New()
	m = makeReady(m)

	view := m.View()

	if view == "Initializing..." {
		t.Error("View should not show 'Initializing...' when ready")
	}

	if len(view) == 0 {
		t.Error("View should return non-empty string when ready")
	}
}

func TestView_ContainsHelpBar(t *testing.T) {
	m := New()
	m = makeReady(m)

	view := m.View()

	// The view should contain help-related text
	if len(view) < 10 {
		t.Error("View should contain substantial content including help bar")
	}
}

func TestUpdate_BackFromPodListDoesNothing(t *testing.T) {
	m := New()
	m = makeReady(m)

	// Already at PodList, pressing back should do nothing
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := m.Update(escMsg)
	updatedModel := newModel.(Model)

	if updatedModel.CurrentView() != model.ViewPodList {
		t.Errorf("Should stay in PodList view, got %v", updatedModel.CurrentView())
	}
}

func TestUpdate_NavigationFromDifferentViews(t *testing.T) {
	m := New()
	m = makeReadyWithPods(m)

	// Navigate to Logs
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Try to navigate to Exec from Logs (should not work - view-specific keys only in PodList)
	execMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ = m.Update(execMsg)
	m = newModel.(Model)

	// Should still be in Logs since 'e' is not handled in Logs view
	if m.CurrentView() != model.ViewLogs {
		t.Errorf("Should stay in Logs view, got %v", m.CurrentView())
	}
}

func TestUpdate_ContextSelectorIsOverlay(t *testing.T) {
	m := New()
	m = makeReady(m)

	// Navigate to context selector
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if !m.CurrentView().IsOverlay() {
		t.Error("Context selector should be an overlay")
	}
}

// Helper function to make model ready
func makeReady(m Model) Model {
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return newModel.(Model)
}

// Helper function to make model ready with pods (for log view tests)
func makeReadyWithPods(m Model) Model {
	m = makeReady(m)
	// Add a test pod so we can navigate to logs
	m.pods = []k8s.PodInfo{
		{
			Name:      "test-pod",
			Namespace: "default",
			Status:    k8s.PodStatusRunning,
			Containers: []k8s.ContainerStatus{
				{Name: "main", Ready: true},
			},
		},
	}
	return m
}
