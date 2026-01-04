# k8s-tui

Go TUI for Kubernetes pod management using Bubble Tea. Replaces clunky Streamlit UI with keyboard-driven terminal app featuring live log streaming.

## Repo Structure

```
main.go              # Entry point
internal/
  app/app.go         # Bubble Tea model, Update/View logic
  k8s/               # Kubernetes client-go wrapper
    client.go        # Client init, kubeconfig handling
    pods.go          # Pod listing, status
    logs.go          # Log streaming (StreamLogs)
    context.go       # K8s context switching
    namespace.go     # Namespace operations
  model/types.go     # Shared types (PodInfo, PodStatus)
  ui/                # UI components
    keys.go          # Keybindings
    logs.go          # Log viewer component
```

## Commands

```bash
go run .             # Run app
go test ./...        # Run all tests
make test            # Run tests with coverage
make lint            # Run linters
make lint-fix        # Run linters and auto-fix
go build -o kpm .    # Build binary
```

## Testing

All packages have `*_test.go` files. Tests use mocks/fakes - no real cluster needed.

## Docs

- `docs/plans/` - Plans before implementation
- `docs/implemented/` - Completed work summaries
- `docs/human_test/` - Manual testing checklists for new features
- Format: `YYYY-MM-DD-HHMM-<title>.md`
