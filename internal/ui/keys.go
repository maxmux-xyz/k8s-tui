package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the application
type KeyMap struct {
	// Navigation
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding

	// Actions
	Logs    key.Binding
	Exec    key.Binding
	Files   key.Binding
	Refresh key.Binding

	// Selectors
	Namespace key.Binding
	Context   key.Binding

	// General
	Help key.Binding
	Back key.Binding
	Quit key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Logs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "logs"),
		),
		Exec: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "exec"),
		),
		Files: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "files"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Namespace: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "namespace"),
		),
		Context: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "context"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns keybindings to show in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},             // Navigation
		{k.Logs, k.Exec, k.Files},           // Actions
		{k.Namespace, k.Context, k.Refresh}, // Management
		{k.Help, k.Back, k.Quit},            // General
	}
}
