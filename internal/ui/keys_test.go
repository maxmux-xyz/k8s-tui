package ui

import (
	"testing"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test that all bindings are defined (not empty)
	bindings := []struct {
		name    string
		keys    []string
		binding func() []string
	}{
		{"Up", []string{"k", "up"}, func() []string { return km.Up.Keys() }},
		{"Down", []string{"j", "down"}, func() []string { return km.Down.Keys() }},
		{"Enter", []string{"enter"}, func() []string { return km.Enter.Keys() }},
		{"Logs", []string{"l"}, func() []string { return km.Logs.Keys() }},
		{"Exec", []string{"e"}, func() []string { return km.Exec.Keys() }},
		{"Files", []string{"f"}, func() []string { return km.Files.Keys() }},
		{"Refresh", []string{"r"}, func() []string { return km.Refresh.Keys() }},
		{"Namespace", []string{"n"}, func() []string { return km.Namespace.Keys() }},
		{"Context", []string{"c"}, func() []string { return km.Context.Keys() }},
		{"Help", []string{"?"}, func() []string { return km.Help.Keys() }},
		{"Back", []string{"esc"}, func() []string { return km.Back.Keys() }},
		{"Quit", []string{"q", "ctrl+c"}, func() []string { return km.Quit.Keys() }},
	}

	for _, b := range bindings {
		t.Run(b.name, func(t *testing.T) {
			keys := b.binding()
			if len(keys) == 0 {
				t.Errorf("%s binding has no keys defined", b.name)
			}

			// Check that expected keys are present
			for _, expectedKey := range b.keys {
				found := false
				for _, k := range keys {
					if k == expectedKey {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s binding should contain key %q, got %v", b.name, expectedKey, keys)
				}
			}
		})
	}
}

func TestKeyMap_ShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	shortHelp := km.ShortHelp()

	if len(shortHelp) != 2 {
		t.Errorf("ShortHelp should return 2 bindings, got %d", len(shortHelp))
	}

	// Verify it contains Help and Quit
	helpFound := false
	quitFound := false
	for _, binding := range shortHelp {
		for _, k := range binding.Keys() {
			if k == "?" {
				helpFound = true
			}
			if k == "q" {
				quitFound = true
			}
		}
	}

	if !helpFound {
		t.Error("ShortHelp should contain Help binding")
	}
	if !quitFound {
		t.Error("ShortHelp should contain Quit binding")
	}
}

func TestKeyMap_FullHelp(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	if len(fullHelp) != 4 {
		t.Errorf("FullHelp should return 4 groups, got %d", len(fullHelp))
	}

	// Verify each group has 3 bindings
	for i, group := range fullHelp {
		if len(group) != 3 {
			t.Errorf("FullHelp group %d should have 3 bindings, got %d", i, len(group))
		}
	}

	// Verify group contents
	// Group 0: Navigation (Up, Down, Enter)
	// Group 1: Actions (Logs, Exec, Files)
	// Group 2: Management (Namespace, Context, Refresh)
	// Group 3: General (Help, Back, Quit)

	expectedGroups := [][]string{
		{"k", "j", "enter"},
		{"l", "e", "f"},
		{"n", "c", "r"},
		{"?", "esc", "q"},
	}

	for i, group := range fullHelp {
		for j, binding := range group {
			found := false
			for _, k := range binding.Keys() {
				if k == expectedGroups[i][j] {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("FullHelp group %d, binding %d should contain key %q", i, j, expectedGroups[i][j])
			}
		}
	}
}

func TestKeyMap_VimBindings(t *testing.T) {
	km := DefaultKeyMap()

	// Verify vim-style navigation keys are present
	tests := []struct {
		name    string
		binding func() []string
		vimKey  string
	}{
		{"Up has k", func() []string { return km.Up.Keys() }, "k"},
		{"Down has j", func() []string { return km.Down.Keys() }, "j"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.binding()
			found := false
			for _, k := range keys {
				if k == tt.vimKey {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected vim key %q in binding, got %v", tt.vimKey, keys)
			}
		})
	}
}
