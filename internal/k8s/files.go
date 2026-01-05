package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FileInfo represents a file or directory entry
type FileInfo struct {
	Name        string
	IsDir       bool
	IsSymlink   bool
	Size        int64
	Permissions string
	Owner       string
	Group       string
	ModTime     string
	LinkTarget  string // For symlinks
}

// FileOptions configures file operations
type FileOptions struct {
	Namespace string
	Pod       string
	Container string
	Path      string
}

// Validate checks that the file options are valid
func (o FileOptions) Validate() error {
	if o.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if o.Pod == "" {
		return fmt.Errorf("pod name is required")
	}
	if o.Path == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

// ListDir lists directory contents using ls -la
func (c *Client) ListDir(ctx context.Context, opts FileOptions) ([]FileInfo, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Run ls -la command
	execOpts := ExecOptions{
		Namespace: opts.Namespace,
		Pod:       opts.Pod,
		Container: opts.Container,
		Command:   []string{"ls", "-la", opts.Path},
	}

	result := c.Exec(ctx, execOpts)
	if result.Error != nil {
		// Check for common errors in stderr
		if strings.Contains(result.Stderr, "No such file or directory") {
			return nil, fmt.Errorf("directory not found: %s", opts.Path)
		}
		if strings.Contains(result.Stderr, "Permission denied") {
			return nil, fmt.Errorf("permission denied: %s", opts.Path)
		}
		if strings.Contains(result.Stderr, "Not a directory") {
			return nil, fmt.Errorf("not a directory: %s", opts.Path)
		}
		return nil, fmt.Errorf("failed to list directory: %w", result.Error)
	}

	// Check stderr for errors even if command "succeeded"
	if result.Stderr != "" && strings.Contains(result.Stderr, "cannot access") {
		return nil, fmt.Errorf("cannot access: %s", opts.Path)
	}

	return ParseLsOutput(result.Stdout)
}

// ReadFile reads file contents using cat (with size limit)
func (c *Client) ReadFile(ctx context.Context, opts FileOptions, maxBytes int) (string, error) {
	if err := opts.Validate(); err != nil {
		return "", err
	}

	// Use head -c to limit output size
	var command []string
	if maxBytes > 0 {
		command = []string{"head", "-c", strconv.Itoa(maxBytes), opts.Path}
	} else {
		command = []string{"cat", opts.Path}
	}

	execOpts := ExecOptions{
		Namespace: opts.Namespace,
		Pod:       opts.Pod,
		Container: opts.Container,
		Command:   command,
	}

	result := c.Exec(ctx, execOpts)
	if result.Error != nil {
		if strings.Contains(result.Stderr, "No such file or directory") {
			return "", fmt.Errorf("file not found: %s", opts.Path)
		}
		if strings.Contains(result.Stderr, "Permission denied") {
			return "", fmt.Errorf("permission denied: %s", opts.Path)
		}
		if strings.Contains(result.Stderr, "Is a directory") {
			return "", fmt.Errorf("is a directory: %s", opts.Path)
		}
		return "", fmt.Errorf("failed to read file: %w", result.Error)
	}

	return result.Stdout, nil
}

// StatFile gets file info for a single path
func (c *Client) StatFile(ctx context.Context, opts FileOptions) (*FileInfo, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Run ls -la on the specific file
	execOpts := ExecOptions{
		Namespace: opts.Namespace,
		Pod:       opts.Pod,
		Container: opts.Container,
		Command:   []string{"ls", "-la", "-d", opts.Path},
	}

	result := c.Exec(ctx, execOpts)
	if result.Error != nil {
		if strings.Contains(result.Stderr, "No such file or directory") {
			return nil, fmt.Errorf("file not found: %s", opts.Path)
		}
		return nil, fmt.Errorf("failed to stat file: %w", result.Error)
	}

	entries, err := ParseLsOutput(result.Stdout)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("file not found: %s", opts.Path)
	}

	return &entries[0], nil
}

// ParseLsOutput parses the output of ls -la into FileInfo entries
func ParseLsOutput(output string) ([]FileInfo, error) {
	lines := strings.Split(output, "\n")
	entries := make([]FileInfo, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip "total N" line
		if strings.HasPrefix(line, "total ") {
			continue
		}

		entry, err := parseLsLine(line)
		if err != nil {
			// Skip lines we can't parse rather than failing entirely
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// parseLsLine parses a single line of ls -la output
// Format: -rw-r--r-- 1 root root 1234 Jan  1 12:00 filename
// Or with symlink: lrwxrwxrwx 1 root root 10 Jan  1 12:00 link -> target
func parseLsLine(line string) (FileInfo, error) {
	// Use regex to handle variable whitespace
	// Permissions, links, owner, group, size, month, day, time/year, name
	pattern := `^([drwxlst-]{10})\s+(\d+)\s+(\S+)\s+(\S+)\s+(\d+)\s+(\w+)\s+(\d+)\s+([\d:]+)\s+(.+)$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return FileInfo{}, fmt.Errorf("failed to parse line: %s", line)
	}

	permissions := matches[1]
	owner := matches[3]
	group := matches[4]
	size, err := strconv.ParseInt(matches[5], 10, 64)
	if err != nil {
		size = 0 // Default to 0 if parse fails
	}
	modTime := fmt.Sprintf("%s %s %s", matches[6], matches[7], matches[8])
	name := matches[9]

	entry := FileInfo{
		Permissions: permissions,
		Owner:       owner,
		Group:       group,
		Size:        size,
		ModTime:     modTime,
		IsDir:       permissions[0] == 'd',
		IsSymlink:   permissions[0] == 'l',
	}

	// Handle symlinks: name -> target
	if entry.IsSymlink {
		parts := strings.SplitN(name, " -> ", 2)
		entry.Name = parts[0]
		if len(parts) > 1 {
			entry.LinkTarget = parts[1]
		}
	} else {
		entry.Name = name
	}

	return entry, nil
}

// FormatSize formats a file size in human-readable format
func FormatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1fG", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1fM", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1fK", float64(size)/KB)
	default:
		return fmt.Sprintf("%d", size)
	}
}

// JoinPath joins path components, handling trailing slashes
func JoinPath(base, name string) string {
	base = strings.TrimSuffix(base, "/")
	if base == "" {
		base = "/"
	}
	if base == "/" {
		return "/" + name
	}
	return base + "/" + name
}

// ParentPath returns the parent directory of a path
func ParentPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	if path == "" || path == "/" {
		return "/"
	}

	lastSlash := strings.LastIndex(path, "/")
	if lastSlash <= 0 {
		return "/"
	}
	return path[:lastSlash]
}
