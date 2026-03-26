package ui

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type editorState int

const (
	editorEditValue editorState = iota
	editorAddKeyName
	editorAddKeyValue
	editorConfirmDelete
	editorConfirmDeletePath
	editorRenameKey
	editorNewSecretName
	editorNewFolderName
	editorImportJSON
)

type editorOverlay struct {
	active      bool
	state       editorState
	key         string
	oldKey      string // original key name for rename
	newSecMount string // mount for new secret creation
	newSecPath  string // base path for new secret creation
	textarea    textarea.Model
	input       textinput.Model // for single-line key name input
	width       int
	height      int
}

type editorDoneMsg struct {
	action string // "edit", "add", "delete", "rename", "cancel"
	key    string
	value  string
}

func newEditorOverlay() editorOverlay {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.CharLimit = 0 // no limit
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderForeground(colorBgLight)
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(colorMuted)
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(colorHighlight)
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle().Foreground(colorMuted)
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(colorMuted)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(colorFg)
	ta.Placeholder = "Enter value..."

	ti := textinput.New()
	ti.Placeholder = "key name"
	ti.CharLimit = 256
	ti.TextStyle = lipgloss.NewStyle().Foreground(colorFg)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(colorAccent)

	return editorOverlay{
		textarea: ta,
		input:    ti,
	}
}

func (e editorOverlay) startEdit(key, value string) editorOverlay {
	e.active = true
	e.state = editorEditValue
	e.key = key
	e.textarea.SetValue(value)
	e.textarea.Focus()
	e.resizeTextarea()
	return e
}

func (e editorOverlay) startAdd() editorOverlay {
	e.active = true
	e.state = editorAddKeyName
	e.key = ""
	e.input.SetValue("")
	e.input.Focus()
	e.textarea.SetValue("")
	return e
}

func (e editorOverlay) startDelete(key string) editorOverlay {
	e.active = true
	e.state = editorConfirmDelete
	e.key = key
	e.input.SetValue("")
	e.input.Placeholder = "Type 'Confirm' to delete"
	e.input.Focus()
	return e
}

func (e editorOverlay) startDeletePath(mount, path string, isDir bool) editorOverlay {
	e.active = true
	e.state = editorConfirmDeletePath
	e.newSecMount = mount
	e.newSecPath = path
	if isDir {
		e.key = path + " (folder and all secrets inside)"
	} else {
		e.key = path
	}
	e.input.SetValue("")
	e.input.Placeholder = "Type 'Confirm' to delete"
	e.input.Focus()
	return e
}

func (e editorOverlay) startRename(key string) editorOverlay {
	e.active = true
	e.state = editorRenameKey
	e.oldKey = key
	e.key = key
	e.input.SetValue(key)
	e.input.Focus()
	e.input.CursorEnd()
	return e
}

func (e editorOverlay) startNewSecret(mount, basePath string) editorOverlay {
	e.active = true
	e.state = editorNewSecretName
	e.newSecMount = mount
	e.newSecPath = basePath
	e.key = ""
	e.input.SetValue("")
	e.input.Placeholder = "secret name (e.g. my-app)"
	e.input.Focus()
	return e
}

func (e editorOverlay) startNewFolder(mount, basePath string) editorOverlay {
	e.active = true
	e.state = editorNewFolderName
	e.newSecMount = mount
	e.newSecPath = basePath
	e.key = ""
	e.input.SetValue("")
	e.input.Placeholder = "folder name"
	e.input.Focus()
	return e
}

func (e editorOverlay) startImportJSON() editorOverlay {
	e.active = true
	e.state = editorImportJSON
	e.textarea.SetValue("")
	e.textarea.Placeholder = "Paste JSON here..."
	e.textarea.Focus()
	e.resizeTextarea()
	return e
}

func (e *editorOverlay) resizeTextarea() {
	w := e.width - 8
	if w < 30 {
		w = 30
	}
	h := e.height - 14
	if h < 5 {
		h = 5
	}
	e.textarea.SetWidth(w)
	e.textarea.SetHeight(h)
	e.input.Width = w - 4
}

func (e editorOverlay) Update(msg tea.KeyMsg) (editorOverlay, tea.Cmd) {
	switch e.state {
	case editorConfirmDelete, editorConfirmDeletePath:
		return e.updateConfirmDelete(msg)
	case editorAddKeyName:
		return e.updateKeyNameInput(msg)
	case editorRenameKey:
		return e.updateRenameInput(msg)
	case editorNewSecretName:
		return e.updateNewSecretName(msg)
	case editorNewFolderName:
		return e.updateNewFolderName(msg)
	case editorImportJSON:
		return e.updateImportJSON(msg)
	default:
		return e.updateTextarea(msg)
	}
}

func (e editorOverlay) UpdateMsg(msg tea.Msg) (editorOverlay, tea.Cmd) {
	// For non-key messages (like paste), route to textarea
	switch e.state {
	case editorEditValue, editorAddKeyValue, editorImportJSON:
		var cmd tea.Cmd
		e.textarea, cmd = e.textarea.Update(msg)
		return e, cmd
	case editorAddKeyName, editorRenameKey, editorNewSecretName, editorNewFolderName, editorConfirmDelete, editorConfirmDeletePath:
		var cmd tea.Cmd
		e.input, cmd = e.input.Update(msg)
		return e, cmd
	}
	return e, nil
}

func (e editorOverlay) updateConfirmDelete(msg tea.KeyMsg) (editorOverlay, tea.Cmd) {
	switch msg.String() {
	case "esc":
		e.active = false
		e.input.Placeholder = "key name"
		return e, func() tea.Msg {
			return editorDoneMsg{action: "cancel"}
		}
	case "enter":
		if e.input.Value() == "Confirm" {
			e.active = false
			e.input.Placeholder = "key name"
			if e.state == editorConfirmDeletePath {
				mount := e.newSecMount
				path := e.newSecPath
				return e, func() tea.Msg {
					return editorDoneMsg{action: "deletepath", key: mount, value: path}
				}
			}
			key := e.key
			return e, func() tea.Msg {
				return editorDoneMsg{action: "delete", key: key}
			}
		}
		return e, nil
	default:
		var cmd tea.Cmd
		e.input, cmd = e.input.Update(msg)
		return e, cmd
	}
}

func (e editorOverlay) updateKeyNameInput(msg tea.KeyMsg) (editorOverlay, tea.Cmd) {
	switch msg.String() {
	case "esc":
		e.active = false
		return e, func() tea.Msg {
			return editorDoneMsg{action: "cancel"}
		}
	case "enter":
		name := e.input.Value()
		if name == "" {
			return e, nil
		}
		e.key = name
		e.state = editorAddKeyValue
		e.textarea.SetValue("")
		e.textarea.Focus()
		e.resizeTextarea()
		return e, nil
	default:
		var cmd tea.Cmd
		e.input, cmd = e.input.Update(msg)
		return e, cmd
	}
}

func (e editorOverlay) updateNewSecretName(msg tea.KeyMsg) (editorOverlay, tea.Cmd) {
	switch msg.String() {
	case "esc":
		e.active = false
		e.input.Placeholder = "key name"
		return e, func() tea.Msg {
			return editorDoneMsg{action: "cancel"}
		}
	case "enter":
		name := e.input.Value()
		if name == "" {
			return e, nil
		}
		e.active = false
		e.input.Placeholder = "key name"
		mount := e.newSecMount
		path := e.newSecPath + name
		return e, func() tea.Msg {
			return editorDoneMsg{action: "newsecret", key: mount, value: path}
		}
	default:
		var cmd tea.Cmd
		e.input, cmd = e.input.Update(msg)
		return e, cmd
	}
}

func (e editorOverlay) updateNewFolderName(msg tea.KeyMsg) (editorOverlay, tea.Cmd) {
	switch msg.String() {
	case "esc":
		e.active = false
		e.input.Placeholder = "key name"
		return e, func() tea.Msg {
			return editorDoneMsg{action: "cancel"}
		}
	case "enter":
		name := e.input.Value()
		if name == "" {
			return e, nil
		}
		e.active = false
		e.input.Placeholder = "key name"
		mount := e.newSecMount
		path := e.newSecPath + name
		return e, func() tea.Msg {
			return editorDoneMsg{action: "newfolder", key: mount, value: path}
		}
	default:
		var cmd tea.Cmd
		e.input, cmd = e.input.Update(msg)
		return e, cmd
	}
}

func (e editorOverlay) updateRenameInput(msg tea.KeyMsg) (editorOverlay, tea.Cmd) {
	switch msg.String() {
	case "esc":
		e.active = false
		return e, func() tea.Msg {
			return editorDoneMsg{action: "cancel"}
		}
	case "enter":
		newName := e.input.Value()
		if newName == "" || newName == e.oldKey {
			e.active = false
			return e, func() tea.Msg {
				return editorDoneMsg{action: "cancel"}
			}
		}
		e.active = false
		oldKey := e.oldKey
		return e, func() tea.Msg {
			return editorDoneMsg{action: "rename", key: newName, value: oldKey}
		}
	default:
		var cmd tea.Cmd
		e.input, cmd = e.input.Update(msg)
		return e, cmd
	}
}

func (e editorOverlay) updateTextarea(msg tea.KeyMsg) (editorOverlay, tea.Cmd) {
	switch msg.String() {
	case "esc":
		e.active = false
		e.textarea.Blur()
		return e, func() tea.Msg {
			return editorDoneMsg{action: "cancel"}
		}
	case "ctrl+s", "f2":
		e.active = false
		e.textarea.Blur()
		value := e.textarea.Value()
		action := "edit"
		if e.state == editorAddKeyValue {
			action = "add"
		}
		return e, func() tea.Msg {
			return editorDoneMsg{action: action, key: e.key, value: value}
		}
	case "ctrl+d":
		// Delete entire current line
		val := e.textarea.Value()
		lines := splitLines(val)
		row := e.textarea.Line()
		if row >= 0 && row < len(lines) {
			lines = append(lines[:row], lines[row+1:]...)
			if len(lines) == 0 {
				lines = []string{""}
			}
			e.textarea.SetValue(joinLines(lines))
			// Move cursor to same row or last row
			if row >= len(lines) {
				row = len(lines) - 1
			}
			e.textarea.SetCursor(row)
		}
		return e, nil
	case "ctrl+k", "ctrl+u":
		// Swallow these - not used in this editor
		return e, nil
	default:
		var cmd tea.Cmd
		e.textarea, cmd = e.textarea.Update(msg)
		return e, cmd
	}
}

func (e editorOverlay) updateImportJSON(msg tea.KeyMsg) (editorOverlay, tea.Cmd) {
	switch msg.String() {
	case "esc":
		e.active = false
		e.textarea.Blur()
		return e, func() tea.Msg {
			return editorDoneMsg{action: "cancel"}
		}
	case "ctrl+s", "f2":
		raw := e.textarea.Value()
		if strings.TrimSpace(raw) == "" {
			return e, nil
		}

		// Parse JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			// Invalid JSON — set error in status, stay in editor
			e.key = "Invalid JSON: " + err.Error()
			return e, nil
		}

		// Convert all values to strings
		result := make(map[string]string)
		for k, v := range parsed {
			switch val := v.(type) {
			case string:
				result[k] = val
			default:
				// Numbers, bools, objects, arrays → JSON string
				b, _ := json.Marshal(val)
				result[k] = string(b)
			}
		}

		e.active = false
		e.textarea.Blur()
		return e, func() tea.Msg {
			return editorDoneMsg{action: "import", value: raw, key: jsonMapToString(result)}
		}
	default:
		var cmd tea.Cmd
		e.textarea, cmd = e.textarea.Update(msg)
		return e, cmd
	}
}

// jsonMapToString serializes a map for passing through editorDoneMsg.
// We use JSON encoding since editorDoneMsg only has key/value string fields.
func jsonMapToString(m map[string]string) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "\n")
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func (e editorOverlay) View() string {
	if !e.active {
		return ""
	}

	areaW := e.width - 6
	if areaW < 30 {
		areaW = 30
	}

	var title string
	var header string
	var content string

	switch e.state {
	case editorEditValue:
		title = " EDIT VALUE "
		header = keyStyle.Render("Key: ") + helpKeyStyle.Render(e.key)
		content = e.textarea.View()

	case editorAddKeyName:
		title = " NEW KEY "
		header = keyStyle.Render("Key name:")
		content = e.input.View()

	case editorAddKeyValue:
		title = " NEW KEY "
		header = keyStyle.Render("Key: ") + helpKeyStyle.Render(e.key) + "\n" +
			keyStyle.Render("Value:")
		content = e.textarea.View()

	case editorRenameKey:
		title = " RENAME KEY "
		header = keyStyle.Render("Current: ") + helpKeyStyle.Render(e.oldKey) + "\n" +
			keyStyle.Render("New name:")
		content = e.input.View()

	case editorNewSecretName:
		title = " NEW SECRET "
		header = keyStyle.Render("Path: ") + helpKeyStyle.Render(e.newSecMount+e.newSecPath) + "\n" +
			keyStyle.Render("Secret name:")
		content = e.input.View()

	case editorNewFolderName:
		title = " NEW FOLDER "
		header = keyStyle.Render("Path: ") + helpKeyStyle.Render(e.newSecMount+e.newSecPath) + "\n" +
			keyStyle.Render("Folder name:")
		content = e.input.View()

	case editorImportJSON:
		title = " IMPORT JSON "
		header = keyStyle.Render("Paste JSON object — keys become secret keys")
		if e.key != "" {
			// e.key is used for error messages
			header += "\n\n" + errorStyle.Render(e.key)
		}
		content = e.textarea.View()

	case editorConfirmDelete:
		title = " DELETE KEY "
		header = errorStyle.Render("Delete key ") + helpKeyStyle.Render(e.key) + errorStyle.Render(" ?") + "\n\n" +
			lipgloss.NewStyle().Foreground(colorMuted).Render("Type 'Confirm' and press Enter to delete")
		content = e.input.View()

	case editorConfirmDeletePath:
		title = " DELETE SECRET "
		header = errorStyle.Render("Delete ") + helpKeyStyle.Render(e.key) + errorStyle.Render(" ?") + "\n\n" +
			lipgloss.NewStyle().Foreground(colorMuted).Render("Type 'Confirm' and press Enter to delete")
		content = e.input.View()
	}

	// Title bar
	titleBar := lipgloss.NewStyle().
		Background(colorAccent).
		Foreground(colorBg).
		Bold(true).
		Padding(0, 1).
		Render(title)


	// Help line
	var helpLine string
	if e.state == editorConfirmDelete || e.state == editorConfirmDeletePath {
		helpLine = renderHelp([]helpBinding{
			{"Confirm + enter", "delete"},
			{"esc", "cancel"},
		}, areaW+4)
	} else if e.state == editorAddKeyName {
		helpLine = renderHelp([]helpBinding{
			{"enter", "next"},
			{"esc", "cancel"},
		}, areaW+4)
	} else if e.state == editorNewSecretName || e.state == editorNewFolderName {
		helpLine = renderHelp([]helpBinding{
			{"enter", "create"},
			{"esc", "cancel"},
		}, areaW+4)
	} else if e.state == editorRenameKey {
		helpLine = renderHelp([]helpBinding{
			{"enter", "confirm"},
			{"esc", "cancel"},
		}, areaW+4)
	} else {
		helpLine = renderHelp([]helpBinding{
			{"ctrl+s/F2", "save"},
			{"esc", "cancel"},
			{"enter", "new line"},
			{"ctrl+d", "delete line"},
			{"ctrl+a", "line start"},
			{"ctrl+e", "line end"},
			{"arrows", "move"},
		}, areaW+4)
	}

	// Box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(0, 1).
		Width(areaW + 4)

	inner := titleBar + "\n\n" +
		header + "\n\n" +
		content + "\n\n" +
		helpLine

	return box.Render(inner)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
