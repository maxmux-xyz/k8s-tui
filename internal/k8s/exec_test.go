package k8s

import (
	"testing"
)

func TestExecOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    ExecOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: ExecOptions{
				Namespace: "default",
				Pod:       "my-pod",
				Container: "main",
				Command:   []string{"ls", "-la"},
			},
			wantErr: false,
		},
		{
			name: "valid options without container",
			opts: ExecOptions{
				Namespace: "default",
				Pod:       "my-pod",
				Command:   []string{"pwd"},
			},
			wantErr: false,
		},
		{
			name: "missing namespace",
			opts: ExecOptions{
				Pod:     "my-pod",
				Command: []string{"ls"},
			},
			wantErr: true,
		},
		{
			name: "missing pod",
			opts: ExecOptions{
				Namespace: "default",
				Command:   []string{"ls"},
			},
			wantErr: true,
		},
		{
			name: "missing command",
			opts: ExecOptions{
				Namespace: "default",
				Pod:       "my-pod",
			},
			wantErr: true,
		},
		{
			name: "empty command",
			opts: ExecOptions{
				Namespace: "default",
				Pod:       "my-pod",
				Command:   []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecOptions_CommandString(t *testing.T) {
	tests := []struct {
		name    string
		command []string
		want    string
	}{
		{
			name:    "single command",
			command: []string{"ls"},
			want:    "ls",
		},
		{
			name:    "command with args",
			command: []string{"ls", "-la", "/tmp"},
			want:    "ls -la /tmp",
		},
		{
			name:    "empty command",
			command: []string{},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := ExecOptions{Command: tt.command}
			if got := opts.CommandString(); got != tt.want {
				t.Errorf("CommandString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecResult(t *testing.T) {
	// Test that ExecResult fields work as expected
	result := ExecResult{
		Stdout:   "hello world\n",
		Stderr:   "",
		ExitCode: 0,
		Error:    nil,
	}

	if result.Stdout != "hello world\n" {
		t.Errorf("Stdout = %v, want %v", result.Stdout, "hello world\n")
	}
	if result.Stderr != "" {
		t.Errorf("Stderr = %v, want empty", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", result.ExitCode)
	}
	if result.Error != nil {
		t.Errorf("Error = %v, want nil", result.Error)
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want []string
	}{
		{
			name: "simple command",
			cmd:  "ls",
			want: []string{"ls"},
		},
		{
			name: "command with args",
			cmd:  "ls -la /tmp",
			want: []string{"ls", "-la", "/tmp"},
		},
		{
			name: "command with quoted arg",
			cmd:  `echo "hello world"`,
			want: []string{"echo", "hello world"},
		},
		{
			name: "command with multiple quoted args",
			cmd:  `grep "foo bar" "file name.txt"`,
			want: []string{"grep", "foo bar", "file name.txt"},
		},
		{
			name: "empty string",
			cmd:  "",
			want: nil,
		},
		{
			name: "whitespace only",
			cmd:  "   ",
			want: nil,
		},
		{
			name: "extra whitespace",
			cmd:  "  ls   -la   /tmp  ",
			want: []string{"ls", "-la", "/tmp"},
		},
		{
			name: "mixed quoted and unquoted",
			cmd:  `cat "my file.txt" | grep pattern`,
			want: []string{"cat", "my file.txt", "|", "grep", "pattern"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCommand(tt.cmd)
			if len(got) != len(tt.want) {
				t.Errorf("ParseCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseCommand(%q)[%d] = %v, want %v", tt.cmd, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// Note: Testing the actual Exec method requires a real Kubernetes cluster
// or integration tests, as the SPDY executor is difficult to mock.
// The Client.Exec method is tested via manual/integration testing.
