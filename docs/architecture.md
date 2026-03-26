# Architecture

## Overview

vault-vim is a TUI (Terminal User Interface) for HashiCorp Vault KV v2 secrets engine. It is built with Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework (Elm architecture: Model-Update-View).

## Project Structure

```
vault-vim/
├── main.go                      # Entry point, loads config, starts TUI
├── internal/
│   ├── config/
│   │   └── config.go            # Vault connection config (addr + token resolution)
│   ├── vault/
│   │   └── client.go            # Vault HTTP API client
│   └── ui/
│       ├── app.go               # Root Bubble Tea model, view routing, editor dispatch
│       ├── mounts.go            # Mount selector view (first screen)
│       ├── browser.go           # Path browser view (navigate directories/secrets)
│       ├── secret.go            # Secret detail view (KV pairs table)
│       ├── editor.go            # Editor overlay (textarea for editing, textinput for names)
│       ├── help.go              # Help panel (bottom bar) + help modal (? key)
│       ├── breadcrumb.go        # Top breadcrumb bar rendering
│       └── styles.go            # All lipgloss styles and color definitions
├── docker-compose.yml           # Local dev Vault with test data
├── Makefile                     # Build/test/run commands
└── docs/                        # Documentation
```

## View Hierarchy

```
AppModel (app.go)
├── mountsModel (mounts.go)     ← first screen, lists KV v2 mounts
├── browserModel (browser.go)   ← path navigation within a mount
├── secretModel (secret.go)     ← KV pairs display for a specific secret
├── editorOverlay (editor.go)   ← modal overlay for editing/adding/renaming/deleting
└── helpModal (help.go)         ← modal overlay for keyboard shortcuts
```

The `AppModel` in `app.go` is the root. It holds a `current view` enum and routes messages to the active sub-model. Only one view is active at a time. The editor and help modal are overlays that sit on top of the current view.

## Message Flow

Bubble Tea uses an Elm-style message loop:

1. User presses a key → `tea.KeyMsg` arrives at `AppModel.Update()`
2. `AppModel` checks: is help modal open? is editor active? → routes accordingly
3. Sub-model processes the message, may return a `tea.Cmd` (async operation)
4. `tea.Cmd` runs (e.g., Vault API call), produces a result message
5. Result message flows back through `AppModel.Update()` to the sub-model

### View Transition Messages

| Message              | From       | To         | Trigger            |
|----------------------|------------|------------|--------------------|
| `mountSelectedMsg`   | mounts     | browser    | User selects mount |
| `browserBackMsg`     | browser    | mounts     | Esc at root path   |
| `secretSelectedMsg`  | browser    | secret     | User selects secret|
| `secretBackMsg`      | secret     | browser    | Esc in secret view |

### Editor Messages

| Message           | Description                                  |
|-------------------|----------------------------------------------|
| `startEditMsg`    | Open editor to edit a value                  |
| `startAddMsg`     | Open editor to add a new key                 |
| `startDeleteMsg`  | Open delete confirmation                     |
| `startRenameMsg`  | Open editor to rename a key                  |
| `newSecretMsg`    | Open editor to create a new secret path      |
| `editorDoneMsg`   | Editor completed with action + key + value   |

## Vault API Client

The client (`internal/vault/client.go`) communicates with Vault via its HTTP API. No Vault SDK dependency — just plain `net/http` + JSON.

### Authentication

Token resolution order (in `internal/config/config.go`):
1. `VAULT_TOKEN` environment variable
2. `~/.vault-token` file (written by `vault login`)

`VAULT_ADDR` defaults to `http://127.0.0.1:8200`.

### API Endpoints Used

| Operation    | Method   | Endpoint                          | Notes                          |
|--------------|----------|-----------------------------------|--------------------------------|
| List mounts  | `GET`    | `/v1/sys/mounts`                  | Filters for `kv` type v2       |
| List secrets | `LIST`   | `/v1/{mount}metadata/{path}`      | Returns keys and directories   |
| Get secret   | `GET`    | `/v1/{mount}data/{path}`          | Returns KV pairs + metadata    |
| Patch secret | `PATCH`  | `/v1/{mount}data/{path}`          | Merge-patch, no overwrite      |
| Put secret   | `POST`   | `/v1/{mount}data/{path}`          | Full write (used for delete/rename/import) |
| Delete key   | —        | Read → remove key → Put           | No native single-key delete    |
| Rename key   | —        | Read → add new → remove old → Put | No native rename               |
| Delete secret| `DELETE` | `/v1/{mount}metadata/{path}`      | Recursive for folders          |
| Import JSON  | `POST`   | `/v1/{mount}data/{path}`          | Overwrites all keys (like `vault kv put @file.json`) |

### Key Design Decision: Patch vs Put

The core problem vault-vim solves: `vault kv put` overwrites ALL keys. If you have 10 keys and `put` with 1 key, the other 9 are gone.

- **Edit value**: Uses `PATCH` with `Content-Type: application/merge-patch+json` — only updates the specified key, leaves others untouched. Falls back to read-merge-write (GET + PUT) if the Vault policy lacks `patch` capability (HTTP 403) or the secret doesn't exist yet (HTTP 404)
- **Import JSON**: Uses `PUT` — intentionally overwrites all keys, matching `vault kv put @file.json` behavior
- **Delete key**: Must use read-modify-write because Vault has no single-key delete API
- **Rename key**: Same read-modify-write pattern — copy value to new key name, delete old

## Editor Component

The editor (`editor.go`) uses `charmbracelet/bubbles/textarea` for multiline text editing and `bubbles/textinput` for single-line inputs (key names, secret names).

### Editor States

| State                | Input Type | Used For                    |
|----------------------|------------|-----------------------------|
| `editorEditValue`      | textarea   | Editing an existing value        |
| `editorAddKeyName`     | textinput  | Typing new key name              |
| `editorAddKeyValue`    | textarea   | Typing new key value             |
| `editorConfirmDelete`  | textinput  | Type "Confirm" to delete key     |
| `editorConfirmDeletePath`| textinput | Type "Confirm" to delete secret |
| `editorRenameKey`      | textinput  | Typing new key name              |
| `editorNewSecretName`  | textinput  | Typing new secret path name      |
| `editorNewFolderName`  | textinput  | Typing new folder name           |
| `editorImportJSON`     | textarea   | Paste JSON to import keys        |

## Styling

All colors and styles are defined in `styles.go` using [Lipgloss](https://github.com/charmbracelet/lipgloss). The color scheme uses Catppuccin Macchiato-inspired colors.
