# Contributing

## Development Setup

```bash
# Clone
git clone https://github.com/BlackMetalz/vault-vim.git
cd vault-vim

# Start local Vault with test data
make dev-up

# Build + run against local Vault
make dev

# Run tests
make test
```

## Adding a New Feature

### Adding a new keybinding to a view

1. Add the key handler in the view's `updateNormal()` method (e.g., `secret.go`)
2. Define a new message type if it triggers a state change (e.g., `type startFooMsg struct{}`)
3. Handle the message in `app.go`'s `Update()` switch
4. If it opens the editor, add a new `editorState` in `editor.go`
5. Update the help bindings in `help.go` (both bottom bar and modal)
6. Update `README.md`

### Adding a new Vault API operation

1. Add the method to `internal/vault/client.go`
2. Use `PATCH` for partial updates, `POST` for full writes
3. For composite operations (delete key, rename key), always read-then-write
4. Handle the API call in `app.go`'s `handleEditorDone()` as an async `tea.Cmd`

### Adding a new editor state

1. Add the state constant to `editorState` enum in `editor.go`
2. Add a `startFoo()` method to initialize the editor
3. Add an `updateFoo()` method for key handling
4. Route it in `Update()` and `UpdateMsg()` switches
5. Add the view rendering in `View()` switch
6. Add help bindings for the state
7. Handle the `editorDoneMsg` action in `app.go`

## Code Conventions

- **Bubble Tea pattern**: Each view is a model with `Init()`, `Update()`, `View()` methods
- **Message passing**: Views communicate through typed messages, not direct calls
- **Async ops**: Vault API calls run as `tea.Cmd` functions, results come back as messages
- **Styles**: All in `styles.go`, using Lipgloss. Don't inline styles in view code
- **No Vault SDK**: Plain HTTP calls to keep the binary small and dependency-free

## Testing

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/ui/ -v
go test ./internal/config/ -v
```

Tests use the standard Go testing package. UI tests focus on the editor and config logic. Vault API tests require a running Vault instance (use `make dev-up`).
