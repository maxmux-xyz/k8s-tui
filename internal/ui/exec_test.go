package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewExecViewModel(t *testing.T) {
	m := NewExecViewModel()

	if m.state != ExecViewStateIdle {
		t.Errorf("Initial state = %v, want %v", m.state, ExecViewStateIdle)
	}
	if len(m.outputLines) != 0 {
		t.Errorf("Initial outputLines length = %d, want 0", len(m.outputLines))
	}
	if len(m.history) != 0 {
		t.Errorf("Initial history length = %d, want 0", len(m.history))
	}
	if m.historyIndex != -1 {
		t.Errorf("Initial historyIndex = %d, want -1", m.historyIndex)
	}
	if !m.input.Focused() {
		t.Error("Input should be focused initially")
	}
}

func TestExecViewModel_SetSize(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)

	if !m.ready {
		t.Error("Model should be ready after SetSize")
	}
	if m.width != 80 {
		t.Errorf("Width = %d, want 80", m.width)
	}
	if m.height != 24 {
		t.Errorf("Height = %d, want 24", m.height)
	}
}

func TestExecViewModel_SetPodInfo(t *testing.T) {
	m := NewExecViewModel()
	m.SetPodInfo("default", "my-pod", "main")

	if m.namespace != "default" {
		t.Errorf("Namespace = %s, want default", m.namespace)
	}
	if m.pod != "my-pod" {
		t.Errorf("Pod = %s, want my-pod", m.pod)
	}
	if m.container != "main" {
		t.Errorf("Container = %s, want main", m.container)
	}
}

func TestExecViewModel_SetState(t *testing.T) {
	m := NewExecViewModel()

	states := []ExecViewState{
		ExecViewStateIdle,
		ExecViewStateRunning,
		ExecViewStateComplete,
		ExecViewStateError,
	}

	for _, state := range states {
		m.SetState(state)
		if m.State() != state {
			t.Errorf("State() = %v, want %v", m.State(), state)
		}
	}
}

func TestExecViewModel_SetError(t *testing.T) {
	m := NewExecViewModel()
	m.SetError("connection refused")

	if m.state != ExecViewStateError {
		t.Errorf("State = %v, want %v", m.state, ExecViewStateError)
	}
	if m.errorMsg != "connection refused" {
		t.Errorf("ErrorMsg = %s, want 'connection refused'", m.errorMsg)
	}
}

func TestExecViewModel_CommandHistory(t *testing.T) {
	m := NewExecViewModel()

	// Add commands to history
	m.AddToHistory("ls")
	m.AddToHistory("pwd")
	m.AddToHistory("env")

	if len(m.history) != 3 {
		t.Errorf("History length = %d, want 3", len(m.history))
	}

	// Navigate history backwards
	m.HistoryPrev()
	if m.input.Value() != "env" {
		t.Errorf("After HistoryPrev, input = %s, want env", m.input.Value())
	}

	m.HistoryPrev()
	if m.input.Value() != "pwd" {
		t.Errorf("After second HistoryPrev, input = %s, want pwd", m.input.Value())
	}

	m.HistoryPrev()
	if m.input.Value() != "ls" {
		t.Errorf("After third HistoryPrev, input = %s, want ls", m.input.Value())
	}

	// Can't go before first
	m.HistoryPrev()
	if m.input.Value() != "ls" {
		t.Errorf("After fourth HistoryPrev, input = %s, want ls (should stay)", m.input.Value())
	}

	// Navigate forward
	m.HistoryNext()
	if m.input.Value() != "pwd" {
		t.Errorf("After HistoryNext, input = %s, want pwd", m.input.Value())
	}

	m.HistoryNext()
	if m.input.Value() != "env" {
		t.Errorf("After second HistoryNext, input = %s, want env", m.input.Value())
	}

	// Past end should clear
	m.HistoryNext()
	if m.input.Value() != "" {
		t.Errorf("After HistoryNext past end, input = %s, want empty", m.input.Value())
	}
}

func TestExecViewModel_AddToHistory_NoDuplicates(t *testing.T) {
	m := NewExecViewModel()

	m.AddToHistory("ls")
	m.AddToHistory("ls")
	m.AddToHistory("ls")

	if len(m.history) != 1 {
		t.Errorf("History length = %d, want 1 (no duplicates)", len(m.history))
	}
}

func TestExecViewModel_AddToHistory_EmptyCommand(t *testing.T) {
	m := NewExecViewModel()

	m.AddToHistory("")

	if len(m.history) != 0 {
		t.Errorf("History length = %d, want 0 (empty not added)", len(m.history))
	}
}

func TestExecViewModel_AddOutput(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)

	m.AddOutput("hello world\n", false)
	m.AddOutput("error message\n", true)

	if len(m.outputLines) != 2 {
		t.Errorf("OutputLines length = %d, want 2", len(m.outputLines))
	}

	if m.outputLines[0] != "hello world" {
		t.Errorf("First line = %s, want 'hello world'", m.outputLines[0])
	}

	if !strings.HasPrefix(m.outputLines[1], "[stderr]") {
		t.Errorf("Second line = %s, should start with [stderr]", m.outputLines[1])
	}
}

func TestExecViewModel_AddOutput_Empty(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)

	m.AddOutput("", false)

	if len(m.outputLines) != 0 {
		t.Errorf("OutputLines length = %d, want 0 (empty not added)", len(m.outputLines))
	}
}

func TestExecViewModel_AddCommandMarker(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)

	m.AddCommandMarker("ls -la")

	if len(m.outputLines) < 2 {
		t.Fatalf("OutputLines length = %d, want at least 2", len(m.outputLines))
	}

	found := false
	for _, line := range m.outputLines {
		if strings.Contains(line, "$ ls -la") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Command marker not found in output")
	}
}

func TestExecViewModel_Clear(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)

	m.AddOutput("line 1\n", false)
	m.AddOutput("line 2\n", false)
	m.Clear()

	if len(m.outputLines) != 0 {
		t.Errorf("OutputLines length = %d, want 0 after Clear", len(m.outputLines))
	}
}

func TestExecViewModel_GetCommand_ClearInput(t *testing.T) {
	m := NewExecViewModel()
	m.input.SetValue("test command")

	if m.GetCommand() != "test command" {
		t.Errorf("GetCommand() = %s, want 'test command'", m.GetCommand())
	}

	m.ClearInput()

	if m.GetCommand() != "" {
		t.Errorf("After ClearInput, GetCommand() = %s, want empty", m.GetCommand())
	}
}

func TestExecViewModel_Focus(t *testing.T) {
	m := NewExecViewModel()

	// Initially focused
	if !m.IsFocused() {
		t.Error("Should be focused initially")
	}

	m.Blur()
	if m.IsFocused() {
		t.Error("Should not be focused after Blur")
	}

	m.Focus()
	if !m.IsFocused() {
		t.Error("Should be focused after Focus")
	}
}

func TestExecViewModel_View(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)
	m.SetPodInfo("default", "my-pod", "main")

	view := m.View()

	// Check that view contains expected elements
	if !strings.Contains(view, "Exec:") {
		t.Error("View should contain 'Exec:'")
	}
	if !strings.Contains(view, "my-pod") {
		t.Error("View should contain pod name")
	}
	if !strings.Contains(view, "[READY]") {
		t.Error("View should contain status [READY] when idle")
	}
}

func TestExecViewModel_View_Running(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)
	m.SetState(ExecViewStateRunning)

	view := m.View()

	if !strings.Contains(view, "[RUNNING]") {
		t.Error("View should contain [RUNNING] status")
	}
	if !strings.Contains(view, "* ") {
		t.Error("View should contain '* ' prompt when running")
	}
}

func TestExecViewModel_View_Error(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)
	m.SetError("connection failed")

	view := m.View()

	if !strings.Contains(view, "[ERROR") {
		t.Error("View should contain [ERROR status")
	}
	if !strings.Contains(view, "connection failed") {
		t.Error("View should contain error message")
	}
}

func TestExecViewState_String(t *testing.T) {
	tests := []struct {
		state ExecViewState
		want  string
	}{
		{ExecViewStateIdle, "Idle"},
		{ExecViewStateRunning, "Running"},
		{ExecViewStateComplete, "Complete"},
		{ExecViewStateError, "Error"},
		{ExecViewState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecViewModel_Update_KeyHandling(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)

	// Test tab to toggle focus
	m.input.Focus()
	msg := tea.KeyMsg{Type: tea.KeyTab}
	m, _ = m.Update(msg)

	if m.input.Focused() {
		t.Error("Input should not be focused after Tab")
	}

	m, _ = m.Update(msg)
	if !m.input.Focused() {
		t.Error("Input should be focused after second Tab")
	}
}

func TestExecViewModel_Scroll(t *testing.T) {
	m := NewExecViewModel()
	m.SetSize(80, 24)

	// Add enough lines to enable scrolling
	for i := 0; i < 50; i++ {
		m.AddOutput("test line\n", false)
	}

	// These should not panic
	m.ScrollUp(1)
	m.ScrollDown(1)
	m.PageUp()
	m.PageDown()
	m.GotoTop()
	m.GotoBottom()
}
