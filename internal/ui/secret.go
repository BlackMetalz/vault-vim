package ui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BlackMetalz/vault-vim/internal/vault"
)

type secretModel struct {
	client  *vault.Client
	mount   string
	path    string
	keys    []string
	values  map[string]string
	cursor  int
	masked  map[int]bool // true = masked
	err     error
	status  string
	width   int
	height  int
}

type secretLoadedMsg struct {
	secret *vault.Secret
	err    error
}

type secretBackMsg struct{}

type secretUpdatedMsg struct {
	err error
}

// Editor modes
type editMode int

const (
	editValue editMode = iota
	addKeyName
	addKeyValue
	confirmDelete
)

type editorModel struct {
	active  bool
	mode    editMode
	key     string
	value   string
	cursor  int
}

func newSecretModel(client *vault.Client, mount, path string) secretModel {
	return secretModel{
		client: client,
		mount:  mount,
		path:   path,
		masked: make(map[int]bool),
	}
}

func (m secretModel) loadSecret() tea.Cmd {
	mount := m.mount
	path := m.path
	return func() tea.Msg {
		secret, err := m.client.GetSecret(mount, path)
		return secretLoadedMsg{secret: secret, err: err}
	}
}

func (m secretModel) Init() tea.Cmd {
	return m.loadSecret()
}

func (m secretModel) Update(msg tea.Msg) (secretModel, tea.Cmd) {
	switch msg := msg.(type) {
	case secretLoadedMsg:
		// Treat 404 as empty secret (not yet created)
		if msg.err != nil && strings.Contains(msg.err.Error(), "HTTP 404") {
			m.err = nil
			m.values = make(map[string]string)
			m.keys = []string{}
			m.masked = make(map[int]bool)
			return m, nil
		}
		m.err = msg.err
		if msg.secret != nil {
			m.values = msg.secret.Data
			m.keys = make([]string, 0, len(msg.secret.Data))
			for k := range msg.secret.Data {
				m.keys = append(m.keys, k)
			}
			sort.Strings(m.keys)
			// Initialize all as masked
			m.masked = make(map[int]bool)
			for i := range m.keys {
				m.masked[i] = true
			}
		}
		return m, nil

	case secretUpdatedMsg:
		if msg.err != nil {
			m.status = "Error: " + msg.err.Error()
		} else {
			m.status = "Updated successfully"
		}
		return m, m.loadSecret()

	case tea.KeyMsg:
		return m.updateNormal(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m secretModel) updateNormal(msg tea.KeyMsg) (secretModel, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.keys)-1 {
			m.cursor++
		}
	case "esc":
		return m, func() tea.Msg { return secretBackMsg{} }
	case "s":
		// Toggle mask for selected key
		if len(m.keys) > 0 {
			m.masked[m.cursor] = !m.masked[m.cursor]
		}
	case "S":
		// Toggle all masks
		allRevealed := true
		for i := range m.keys {
			if m.masked[i] {
				allRevealed = false
				break
			}
		}
		for i := range m.keys {
			m.masked[i] = allRevealed
		}
	case "e":
		// Start editing - handled by app model
		if len(m.keys) > 0 {
			return m, func() tea.Msg {
				return startEditMsg{
					key:   m.keys[m.cursor],
					value: m.values[m.keys[m.cursor]],
				}
			}
		}
	case "a":
		return m, func() tea.Msg {
			return startAddMsg{}
		}
	case "r":
		// Rename key
		if len(m.keys) > 0 {
			return m, func() tea.Msg {
				return startRenameMsg{key: m.keys[m.cursor]}
			}
		}
	case "ctrl+d":
		if len(m.keys) > 0 {
			return m, func() tea.Msg {
				return startDeleteMsg{key: m.keys[m.cursor]}
			}
		}
	}
	return m, nil
}

func (m secretModel) breadcrumbs() []string {
	parts := []string{"Mounts", m.mount}
	segments := strings.Split(strings.TrimSuffix(m.path, "/"), "/")
	parts = append(parts, segments...)
	return parts
}

func (m secretModel) View() string {
	if m.err != nil {
		return errorStyle.Render("Error: " + m.err.Error())
	}

	if m.values == nil {
		return titleStyle.Render("Loading secret...")
	}

	var b strings.Builder

	// Title + breadcrumb
	b.WriteString(titleStyle.Render("vault-vim") + "\n")
	b.WriteString(renderBreadcrumb(m.breadcrumbs()) + "\n\n")

	// Calculate dynamic key column width
	keyColW := 10
	for _, key := range m.keys {
		if len(key)+2 > keyColW {
			keyColW = len(key) + 2
		}
	}
	if keyColW > 40 {
		keyColW = 40
	}

	// Value column width
	valColW := m.width - keyColW - 7 // separator + padding
	if valColW < 20 {
		valColW = 20
	}

	// Table header
	headerKey := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Width(keyColW).
		Render("KEY")
	headerVal := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render("VALUE")
	b.WriteString("  " + headerKey + " \u2502 " + headerVal + "\n")

	// Header separator
	sep := strings.Repeat("\u2500", keyColW)
	sepVal := strings.Repeat("\u2500", valColW)
	b.WriteString("  " + lipgloss.NewStyle().Foreground(colorMuted).Render(sep+"\u2500\u253c\u2500"+sepVal) + "\n")

	// KV pairs
	for i, key := range m.keys {
		val := m.values[key]

		// Format display value
		var displayVal string
		if m.masked[i] {
			displayVal = strings.Repeat("\u2022", min(len(val), 20))
		} else {
			displayVal = flattenValue(val, valColW)
		}

		// Key column (fixed width)
		keyRendered := lipgloss.NewStyle().
			Width(keyColW).
			Foreground(colorAccent).
			Bold(true).
			Render(key)

		// Separator
		sepChar := lipgloss.NewStyle().Foreground(colorMuted).Render("\u2502")

		if i == m.cursor {
			// Selected row: highlight key with background
			keyRendered = lipgloss.NewStyle().
				Width(keyColW).
				Foreground(colorBg).
				Background(colorAccent).
				Bold(true).
				Render(" " + key)

			b.WriteString(" " + keyRendered + " " + sepChar + " " + valueStyle.Render(displayVal) + "\n")
		} else {
			b.WriteString("  " + keyRendered + " " + sepChar + " ")
			if m.masked[i] {
				b.WriteString(maskedStyle.Render(displayVal) + "\n")
			} else {
				b.WriteString(valueStyle.Render(displayVal) + "\n")
			}
		}
	}

	if len(m.keys) == 0 {
		b.WriteString("  " + itemStyle.Render("(no keys)") + "\n")
	}

	// Status message
	if m.status != "" {
		b.WriteString("\n")
		if strings.HasPrefix(m.status, "Error") {
			b.WriteString(errorStyle.Render(m.status) + "\n")
		} else {
			b.WriteString(successStyle.Render(m.status) + "\n")
		}
	}

	// Fill space
	lines := strings.Count(b.String(), "\n")
	for i := lines; i < m.height-5; i++ {
		b.WriteString("\n")
	}

	// Help
	b.WriteString(renderHelp(secretHelp, m.width))

	return b.String()
}

// flattenValue collapses multiline values into a single-line preview.
// Shows the content compacted with visible markers for newlines.
func flattenValue(val string, maxLen int) string {
	if !strings.Contains(val, "\n") {
		if len(val) > maxLen {
			return val[:maxLen-3] + "..."
		}
		return val
	}
	// Replace newlines with visible symbol, collapse whitespace
	flat := strings.ReplaceAll(val, "\n", "\u21b5 ") // ↵ symbol
	flat = strings.Join(strings.Fields(flat), " ")
	if len(flat) > maxLen {
		flat = flat[:maxLen-3] + "..."
	}
	return flat
}

// Messages for editor interaction
type startEditMsg struct {
	key   string
	value string
}

type startAddMsg struct{}

type startDeleteMsg struct {
	key string
}

type startRenameMsg struct {
	key string
}
