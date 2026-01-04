package model

// ViewState represents the current view/screen of the application
type ViewState int

// View state constants for application navigation.
const (
	ViewPodList           ViewState = iota // Main pod list view
	ViewLogs                               // Log streaming view
	ViewExec                               // Command execution view
	ViewFiles                              // File browser view
	ViewNamespaceSelector                  // Namespace selection overlay
	ViewContextSelector                    // Context selection overlay
	ViewHelp                               // Help overlay
)

// String returns a human-readable name for the view state
func (v ViewState) String() string {
	switch v {
	case ViewPodList:
		return "Pod List"
	case ViewLogs:
		return "Logs"
	case ViewExec:
		return "Exec"
	case ViewFiles:
		return "Files"
	case ViewNamespaceSelector:
		return "Namespace Selector"
	case ViewContextSelector:
		return "Context Selector"
	case ViewHelp:
		return "Help"
	default:
		return "Unknown"
	}
}

// IsOverlay returns true if this view is displayed as an overlay
func (v ViewState) IsOverlay() bool {
	switch v {
	case ViewNamespaceSelector, ViewContextSelector, ViewHelp:
		return true
	default:
		return false
	}
}
