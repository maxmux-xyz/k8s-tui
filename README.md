# k8s-tui

A terminal-based Kubernetes pod manager built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). This TUI provides a keyboard-driven interface for managing pods, streaming logs, executing commands, and browsing files in your Kubernetes clusters.

## Features

- **Pod Management** - List pods with status indicators, auto-refresh
- **Live Log Streaming** - Real-time log viewing with follow mode and search
- **Command Execution** - Run commands inside pods
- **File Browser** - Navigate and view files in containers
- **Context/Namespace Switching** - Quickly switch between clusters and namespaces
- **Vim-style Navigation** - Keyboard-driven workflow

## Prerequisites

- **Go 1.21+** - [Installation guide](https://go.dev/doc/install)
- **kubectl configured** - Valid kubeconfig with cluster access

## Development Setup

### 1. Clone the repository

```bash
git clone https://github.com/maxime/k8s-tui.git
cd k8s-tui
```

### 2. Install dependencies

```bash
make deps
```

Or manually:

```bash
go mod download
go mod tidy
```

### 3. Verify setup

```bash
make verify
```

## Running the Application

### Development mode

```bash
make run
```

Or directly:

```bash
go run .
```

### Build and run binary

```bash
make build
./k8s-tui
```

## Running Tests

### Run all tests

```bash
make test
```

Or directly:

```bash
go test -v ./...
```

### Run tests with coverage report

```bash
make test-coverage
```

This generates:
- `coverage.out` - Coverage data
- `coverage.html` - HTML coverage report (open in browser)

### Run tests for a specific package

```bash
go test -v ./internal/app/...
go test -v ./internal/model/...
go test -v ./internal/ui/...
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select / Open |
| `l` | View logs |
| `e` | Exec into pod |
| `f` | File browser |
| `n` | Change namespace |
| `c` | Change context |
| `r` | Refresh |
| `?` | Toggle help |
| `Esc` | Back / Cancel |
| `q` | Quit |

## Project Structure

```
k8s-tui/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── Makefile                # Build commands
├── internal/
│   ├── app/
│   │   ├── app.go          # Main Bubble Tea model
│   │   └── app_test.go
│   ├── model/
│   │   ├── types.go        # Shared types and view states
│   │   └── types_test.go
│   └── ui/
│       ├── keys.go         # Keybindings
│       └── keys_test.go
└── docs/
    ├── implemented/        # Implementation records
    └── plans/              # Planning documents
```

## Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the binary |
| `make run` | Run in development mode |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with HTML coverage report |
| `make deps` | Download and tidy dependencies |
| `make verify` | Verify dependencies |
| `make clean` | Remove build artifacts |

## License

MIT
