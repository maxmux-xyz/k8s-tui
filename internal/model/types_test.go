package model

import "testing"

func TestViewState_String(t *testing.T) {
	tests := []struct {
		view     ViewState
		expected string
	}{
		{ViewPodList, "Pod List"},
		{ViewLogs, "Logs"},
		{ViewExec, "Exec"},
		{ViewFiles, "Files"},
		{ViewNamespaceSelector, "Namespace Selector"},
		{ViewContextSelector, "Context Selector"},
		{ViewHelp, "Help"},
		{ViewState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.view.String(); got != tt.expected {
				t.Errorf("ViewState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestViewState_IsOverlay(t *testing.T) {
	overlays := []ViewState{ViewNamespaceSelector, ViewContextSelector, ViewHelp}
	nonOverlays := []ViewState{ViewPodList, ViewLogs, ViewExec, ViewFiles}

	for _, v := range overlays {
		t.Run(v.String()+"_is_overlay", func(t *testing.T) {
			if !v.IsOverlay() {
				t.Errorf("%v should be an overlay", v)
			}
		})
	}

	for _, v := range nonOverlays {
		t.Run(v.String()+"_not_overlay", func(t *testing.T) {
			if v.IsOverlay() {
				t.Errorf("%v should not be an overlay", v)
			}
		})
	}
}

func TestViewState_Constants(t *testing.T) {
	// Verify the iota values are as expected
	if ViewPodList != 0 {
		t.Errorf("ViewPodList should be 0, got %d", ViewPodList)
	}
	if ViewLogs != 1 {
		t.Errorf("ViewLogs should be 1, got %d", ViewLogs)
	}
	if ViewExec != 2 {
		t.Errorf("ViewExec should be 2, got %d", ViewExec)
	}
	if ViewFiles != 3 {
		t.Errorf("ViewFiles should be 3, got %d", ViewFiles)
	}
	if ViewNamespaceSelector != 4 {
		t.Errorf("ViewNamespaceSelector should be 4, got %d", ViewNamespaceSelector)
	}
	if ViewContextSelector != 5 {
		t.Errorf("ViewContextSelector should be 5, got %d", ViewContextSelector)
	}
	if ViewHelp != 6 {
		t.Errorf("ViewHelp should be 6, got %d", ViewHelp)
	}
}
