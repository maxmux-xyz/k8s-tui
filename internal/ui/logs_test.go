package ui

import (
	"strings"
	"testing"
)

func TestNewLogViewModel(t *testing.T) {
	m := NewLogViewModel()

	if m.follow != true {
		t.Error("expected follow mode to be enabled by default")
	}

	if m.state != LogViewStateIdle {
		t.Errorf("expected state Idle, got %v", m.state)
	}

	if len(m.lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(m.lines))
	}

	if m.maxLines != 10000 {
		t.Errorf("expected maxLines 10000, got %d", m.maxLines)
	}
}

func TestLogViewModel_AddLine(t *testing.T) {
	m := NewLogViewModel()

	m.AddLine("line 1")
	m.AddLine("line 2")
	m.AddLine("line 3")

	if m.LineCount() != 3 {
		t.Errorf("expected 3 lines, got %d", m.LineCount())
	}
}

func TestLogViewModel_AddLines(t *testing.T) {
	m := NewLogViewModel()

	lines := []string{"line 1", "line 2", "line 3", "line 4"}
	m.AddLines(lines)

	if m.LineCount() != 4 {
		t.Errorf("expected 4 lines, got %d", m.LineCount())
	}
}

func TestLogViewModel_Clear(t *testing.T) {
	m := NewLogViewModel()
	m.SetSize(80, 24)

	m.AddLine("line 1")
	m.AddLine("line 2")

	if m.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", m.LineCount())
	}

	m.Clear()

	if m.LineCount() != 0 {
		t.Errorf("expected 0 lines after clear, got %d", m.LineCount())
	}
}

func TestLogViewModel_MaxLines(t *testing.T) {
	m := NewLogViewModel()
	m.maxLines = 100 // Set smaller for testing

	// Add more than maxLines
	for i := 0; i < 150; i++ {
		m.AddLine("line")
	}

	// Should have trimmed old lines (removes 10% when exceeding)
	if m.LineCount() > 100 {
		t.Errorf("expected lines to be trimmed to around 100, got %d", m.LineCount())
	}
}

func TestLogViewModel_ToggleFollow(t *testing.T) {
	m := NewLogViewModel()
	m.SetSize(80, 24)

	// Default is follow enabled
	if !m.IsFollow() {
		t.Error("expected follow to be enabled by default")
	}

	m.ToggleFollow()

	if m.IsFollow() {
		t.Error("expected follow to be disabled after toggle")
	}

	m.ToggleFollow()

	if !m.IsFollow() {
		t.Error("expected follow to be enabled after second toggle")
	}
}

func TestLogViewModel_SetPodInfo(t *testing.T) {
	m := NewLogViewModel()
	m.SetSize(80, 24)

	m.SetPodInfo("my-namespace", "my-pod", "my-container")

	view := m.View()

	if !strings.Contains(view, "my-namespace") {
		t.Error("expected namespace in view")
	}
	if !strings.Contains(view, "my-pod") {
		t.Error("expected pod name in view")
	}
	if !strings.Contains(view, "my-container") {
		t.Error("expected container name in view")
	}
}

func TestLogViewModel_SetState(t *testing.T) {
	m := NewLogViewModel()
	m.SetSize(80, 24)

	tests := []struct {
		state    LogViewState
		expected string
	}{
		{LogViewStateIdle, "[IDLE]"},
		{LogViewStateStreaming, "[STREAMING]"},
		{LogViewStatePaused, "[PAUSED]"},
		{LogViewStateEnded, "[STREAM ENDED]"},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			m.SetState(tt.state)
			view := m.View()

			if !strings.Contains(view, tt.expected) {
				t.Errorf("expected %q in view, got: %s", tt.expected, view)
			}
		})
	}
}

func TestLogViewModel_SetError(t *testing.T) {
	m := NewLogViewModel()
	m.SetSize(80, 24)

	m.SetError("connection lost")

	if m.State() != LogViewStateError {
		t.Errorf("expected state Error, got %v", m.State())
	}

	view := m.View()
	if !strings.Contains(view, "[ERROR: connection lost]") {
		t.Errorf("expected error message in view, got: %s", view)
	}
}

func TestLogViewModel_SetSize(t *testing.T) {
	m := NewLogViewModel()

	if m.ready {
		t.Error("expected ready to be false before SetSize")
	}

	m.SetSize(100, 30)

	if !m.ready {
		t.Error("expected ready to be true after SetSize")
	}

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}

	if m.height != 30 {
		t.Errorf("expected height 30, got %d", m.height)
	}
}

func TestLogViewModel_View_NotReady(t *testing.T) {
	m := NewLogViewModel()

	view := m.View()

	if !strings.Contains(view, "Initializing") {
		t.Errorf("expected initializing message when not ready, got: %s", view)
	}
}

func TestLogViewModel_View_WithContent(t *testing.T) {
	m := NewLogViewModel()
	m.SetSize(80, 24)
	m.SetPodInfo("default", "test-pod", "main")

	m.AddLine("Log line 1")
	m.AddLine("Log line 2")
	m.AddLine("Log line 3")

	// Force viewport update
	m.updateViewportContent()

	view := m.View()

	// Check header
	if !strings.Contains(view, "Logs:") {
		t.Error("expected header in view")
	}

	// Check content is present
	if !strings.Contains(view, "Log line") {
		t.Error("expected log content in view")
	}

	// Check line count in status
	if !strings.Contains(view, "Lines: 3") {
		t.Errorf("expected 'Lines: 3' in view, got: %s", view)
	}
}

func TestLogViewModel_FollowIndicator(t *testing.T) {
	m := NewLogViewModel()
	m.SetSize(80, 24)

	// Follow should be enabled by default
	view := m.View()
	if !strings.Contains(view, "[FOLLOW]") {
		t.Error("expected [FOLLOW] indicator when follow is enabled")
	}

	m.ToggleFollow()

	view = m.View()
	if strings.Contains(view, "[FOLLOW]") {
		t.Error("expected no [FOLLOW] indicator when follow is disabled")
	}
}

func TestLogViewState_String(t *testing.T) {
	tests := []struct {
		state    LogViewState
		expected string
	}{
		{LogViewStateIdle, "Idle"},
		{LogViewStateStreaming, "Streaming"},
		{LogViewStatePaused, "Paused"},
		{LogViewStateError, "Error"},
		{LogViewStateEnded, "Ended"},
		{LogViewState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestLogViewModel_ScrollMethods(t *testing.T) {
	m := NewLogViewModel()
	m.SetSize(80, 10)

	// Add enough lines to enable scrolling
	for i := 0; i < 50; i++ {
		m.AddLine("line")
	}
	m.updateViewportContent()

	// Test GotoTop
	m.GotoTop()
	if m.IsFollow() {
		t.Error("expected follow to be disabled after GotoTop")
	}

	// Test GotoBottom
	m.GotoBottom()
	if !m.IsFollow() {
		t.Error("expected follow to be enabled after GotoBottom")
	}

	// Test ScrollUp
	m.ScrollUp(5)
	if m.IsFollow() {
		t.Error("expected follow to be disabled after ScrollUp")
	}
}
