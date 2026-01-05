package k8s

import (
	"testing"
)

func TestFileOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    FileOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: FileOptions{
				Namespace: "default",
				Pod:       "my-pod",
				Path:      "/app",
			},
			wantErr: false,
		},
		{
			name: "missing namespace",
			opts: FileOptions{
				Pod:  "my-pod",
				Path: "/app",
			},
			wantErr: true,
		},
		{
			name: "missing pod",
			opts: FileOptions{
				Namespace: "default",
				Path:      "/app",
			},
			wantErr: true,
		},
		{
			name: "missing path",
			opts: FileOptions{
				Namespace: "default",
				Pod:       "my-pod",
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

func TestParseLsOutput(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		wantCount int
		wantFirst string
		wantIsDir bool
		wantErr   bool
	}{
		{
			name: "typical directory listing",
			output: `total 24
drwxr-xr-x 3 root root 4096 Jan  1 12:00 .
drwxr-xr-x 5 root root 4096 Jan  1 12:00 ..
-rw-r--r-- 1 root root 1234 Jan  1 12:00 file.txt
drwxr-xr-x 2 root root 4096 Jan  1 12:00 subdir`,
			wantCount: 4,
			wantFirst: ".",
			wantIsDir: true,
			wantErr:   false,
		},
		{
			name:      "single file",
			output:    `-rw-r--r-- 1 root root 1234 Jan  1 12:00 file.txt`,
			wantCount: 1,
			wantFirst: "file.txt",
			wantIsDir: false,
			wantErr:   false,
		},
		{
			name:      "symlink",
			output:    `lrwxrwxrwx 1 root root 10 Jan  1 12:00 link -> /target`,
			wantCount: 1,
			wantFirst: "link",
			wantIsDir: false,
			wantErr:   false,
		},
		{
			name:      "empty output",
			output:    "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "only total line",
			output:    `total 0`,
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := ParseLsOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLsOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(entries) != tt.wantCount {
				t.Errorf("ParseLsOutput() got %d entries, want %d", len(entries), tt.wantCount)
				return
			}
			if tt.wantCount > 0 {
				if entries[0].Name != tt.wantFirst {
					t.Errorf("ParseLsOutput() first entry name = %q, want %q", entries[0].Name, tt.wantFirst)
				}
				if entries[0].IsDir != tt.wantIsDir {
					t.Errorf("ParseLsOutput() first entry IsDir = %v, want %v", entries[0].IsDir, tt.wantIsDir)
				}
			}
		})
	}
}

func TestParseLsLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    FileInfo
		wantErr bool
	}{
		{
			name: "regular file",
			line: "-rw-r--r-- 1 root root 1234 Jan  1 12:00 myfile.txt",
			want: FileInfo{
				Name:        "myfile.txt",
				IsDir:       false,
				IsSymlink:   false,
				Size:        1234,
				Permissions: "-rw-r--r--",
				Owner:       "root",
				Group:       "root",
				ModTime:     "Jan 1 12:00",
			},
			wantErr: false,
		},
		{
			name: "directory",
			line: "drwxr-xr-x 2 app app 4096 Dec 15 09:30 config",
			want: FileInfo{
				Name:        "config",
				IsDir:       true,
				IsSymlink:   false,
				Size:        4096,
				Permissions: "drwxr-xr-x",
				Owner:       "app",
				Group:       "app",
				ModTime:     "Dec 15 09:30",
			},
			wantErr: false,
		},
		{
			name: "symlink",
			line: "lrwxrwxrwx 1 root root 10 Jan  1 12:00 link -> /etc/target",
			want: FileInfo{
				Name:        "link",
				IsDir:       false,
				IsSymlink:   true,
				Size:        10,
				Permissions: "lrwxrwxrwx",
				Owner:       "root",
				Group:       "root",
				ModTime:     "Jan 1 12:00",
				LinkTarget:  "/etc/target",
			},
			wantErr: false,
		},
		{
			name: "file with year instead of time",
			line: "-rw-r--r-- 1 root root 5678 Jan  1  2023 oldfile.txt",
			want: FileInfo{
				Name:        "oldfile.txt",
				IsDir:       false,
				Size:        5678,
				Permissions: "-rw-r--r--",
				Owner:       "root",
				Group:       "root",
				ModTime:     "Jan 1 2023",
			},
			wantErr: false,
		},
		{
			name:    "invalid line",
			line:    "not a valid ls line",
			wantErr: true,
		},
		{
			name:    "empty line",
			line:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLsLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLsLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("parseLsLine() Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.IsDir != tt.want.IsDir {
				t.Errorf("parseLsLine() IsDir = %v, want %v", got.IsDir, tt.want.IsDir)
			}
			if got.IsSymlink != tt.want.IsSymlink {
				t.Errorf("parseLsLine() IsSymlink = %v, want %v", got.IsSymlink, tt.want.IsSymlink)
			}
			if got.Size != tt.want.Size {
				t.Errorf("parseLsLine() Size = %d, want %d", got.Size, tt.want.Size)
			}
			if got.Permissions != tt.want.Permissions {
				t.Errorf("parseLsLine() Permissions = %q, want %q", got.Permissions, tt.want.Permissions)
			}
			if got.Owner != tt.want.Owner {
				t.Errorf("parseLsLine() Owner = %q, want %q", got.Owner, tt.want.Owner)
			}
			if got.LinkTarget != tt.want.LinkTarget {
				t.Errorf("parseLsLine() LinkTarget = %q, want %q", got.LinkTarget, tt.want.LinkTarget)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{0, "0"},
		{512, "512"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1572864, "1.5M"},
		{1073741824, "1.0G"},
		{1610612736, "1.5G"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatSize(tt.size); got != tt.want {
				t.Errorf("FormatSize(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		base string
		name string
		want string
	}{
		{"/", "file.txt", "/file.txt"},
		{"/app", "config", "/app/config"},
		{"/app/", "config", "/app/config"},
		{"", "file.txt", "/file.txt"},
		{"/app/logs", "app.log", "/app/logs/app.log"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := JoinPath(tt.base, tt.name); got != tt.want {
				t.Errorf("JoinPath(%q, %q) = %q, want %q", tt.base, tt.name, got, tt.want)
			}
		})
	}
}

func TestParentPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/", "/"},
		{"/app", "/"},
		{"/app/config", "/app"},
		{"/app/config/", "/app"},
		{"/a/b/c/d", "/a/b/c"},
		{"", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := ParentPath(tt.path); got != tt.want {
				t.Errorf("ParentPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestFileInfoSymlink(t *testing.T) {
	// Test parsing symlink with arrow
	line := "lrwxrwxrwx 1 root root 15 Jan  1 12:00 current -> /app/v1.2.3"
	entry, err := parseLsLine(line)
	if err != nil {
		t.Fatalf("parseLsLine() error = %v", err)
	}

	if !entry.IsSymlink {
		t.Error("expected IsSymlink to be true")
	}
	if entry.Name != "current" {
		t.Errorf("Name = %q, want %q", entry.Name, "current")
	}
	if entry.LinkTarget != "/app/v1.2.3" {
		t.Errorf("LinkTarget = %q, want %q", entry.LinkTarget, "/app/v1.2.3")
	}
}

func TestParseLsOutputWithSpacesInFilename(t *testing.T) {
	// Files with spaces are tricky - our simple parser may not handle them perfectly
	// but we should at least not crash
	output := `-rw-r--r-- 1 root root 1234 Jan  1 12:00 file with spaces.txt`
	entries, err := ParseLsOutput(output)
	if err != nil {
		t.Fatalf("ParseLsOutput() error = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
		return
	}

	// The filename should include the spaces
	if entries[0].Name != "file with spaces.txt" {
		t.Errorf("Name = %q, want %q", entries[0].Name, "file with spaces.txt")
	}
}
