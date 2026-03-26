package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/BlackMetalz/vault-vim/internal/vault"
)

type browserModel struct {
	client    *vault.Client
	mount     string
	path      []string // path segments
	items     []string
	cursor    int
	filter    string
	filtered  []string
	filtering bool
	err       error
	width     int
	height    int
}

type browserLoadedMsg struct {
	items []string
	err   error
}

type secretSelectedMsg struct {
	mount string
	path  string
}

type browserBackMsg struct{}

func newBrowserModel(client *vault.Client, mount string) browserModel {
	return browserModel{
		client: client,
		mount:  mount,
	}
}

func (m browserModel) loadItems() tea.Cmd {
	mount := m.mount
	path := m.currentPath()
	return func() tea.Msg {
		items, err := m.client.ListSecrets(mount, path)
		return browserLoadedMsg{items: items, err: err}
	}
}

func (m browserModel) currentPath() string {
	if len(m.path) == 0 {
		return ""
	}
	return strings.Join(m.path, "/") + "/"
}

func (m browserModel) Init() tea.Cmd {
	return m.loadItems()
}

func (m browserModel) Update(msg tea.Msg) (browserModel, tea.Cmd) {
	switch msg := msg.(type) {
	case browserLoadedMsg:
		m.err = msg.err
		m.items = msg.items
		m.filtered = msg.items
		m.cursor = 0
		m.filter = ""
		m.filtering = false
		return m, nil

	case tea.KeyMsg:
		if m.filtering {
			return m.updateFilter(msg)
		}
		return m.updateNormal(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m browserModel) updateNormal(msg tea.KeyMsg) (browserModel, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.filtered) > 0 {
			selected := m.filtered[m.cursor]
			if strings.HasSuffix(selected, "/") {
				// It's a directory, drill in
				m.path = append(m.path, strings.TrimSuffix(selected, "/"))
				m.cursor = 0
				return m, m.loadItems()
			}
			// It's a secret, open detail view
			fullPath := m.currentPath() + selected
			return m, func() tea.Msg {
				return secretSelectedMsg{
					mount: m.mount,
					path:  fullPath,
				}
			}
		}
	case "esc":
		if len(m.path) > 0 {
			m.path = m.path[:len(m.path)-1]
			m.cursor = 0
			return m, m.loadItems()
		}
		return m, func() tea.Msg { return browserBackMsg{} }
	case "/":
		m.filtering = true
	case "n":
		// Create new secret at current path
		mount := m.mount
		path := m.currentPath()
		return m, func() tea.Msg {
			return newSecretMsg{mount: mount, path: path}
		}
	case "N":
		// Create new folder at current path
		mount := m.mount
		path := m.currentPath()
		return m, func() tea.Msg {
			return newFolderMsg{mount: mount, path: path}
		}
	case "ctrl+d":
		// Delete selected secret or folder
		if len(m.filtered) > 0 {
			selected := m.filtered[m.cursor]
			mount := m.mount
			path := m.currentPath() + selected
			return m, func() tea.Msg {
				return browserDeleteMsg{mount: mount, path: path, isDir: strings.HasSuffix(selected, "/")}
			}
		}
	}
	return m, nil
}

func (m browserModel) updateFilter(msg tea.KeyMsg) (browserModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filtering = false
		m.filter = ""
		m.filtered = m.items
		m.cursor = 0
	case "enter":
		m.filtering = false
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.applyFilter()
		}
	default:
		if len(msg.String()) == 1 {
			m.filter += msg.String()
			m.applyFilter()
		}
	}
	return m, nil
}

func (m *browserModel) applyFilter() {
	if m.filter == "" {
		m.filtered = m.items
		m.cursor = 0
		return
	}
	var filtered []string
	for _, item := range m.items {
		if strings.Contains(strings.ToLower(item), strings.ToLower(m.filter)) {
			filtered = append(filtered, item)
		}
	}
	m.filtered = filtered
	m.cursor = 0
}

func (m browserModel) breadcrumbs() []string {
	parts := []string{"Mounts", m.mount}
	parts = append(parts, m.path...)
	return parts
}

func (m browserModel) View() string {
	if m.err != nil {
		return errorStyle.Render("Error: " + m.err.Error())
	}

	if m.items == nil {
		return titleStyle.Render("Loading...")
	}

	var b strings.Builder

	// Title + breadcrumb
	b.WriteString(titleStyle.Render("vault-vim") + "\n")
	b.WriteString(renderBreadcrumb(m.breadcrumbs()) + "\n\n")

	// List
	for i, item := range m.filtered {
		isDir := strings.HasSuffix(item, "/")
		if i == m.cursor {
			if isDir {
				b.WriteString(selectedDirItemStyle.Render(" "+item) + "\n")
			} else {
				b.WriteString(selectedItemStyle.Render(" "+item) + "\n")
			}
		} else {
			if isDir {
				b.WriteString(dirItemStyle.Render(item) + "\n")
			} else {
				b.WriteString(itemStyle.Render(item) + "\n")
			}
		}
	}

	if len(m.filtered) == 0 {
		b.WriteString(itemStyle.Render("  (empty)") + "\n")
	}

	// Filter bar
	if m.filtering {
		b.WriteString("\n" + filterPromptStyle.Render("/") + filterTextStyle.Render(m.filter+"_") + "\n")
	}

	// Fill space
	lines := strings.Count(b.String(), "\n")
	for i := lines; i < m.height-5; i++ {
		b.WriteString("\n")
	}

	// Help
	b.WriteString(renderHelp(browserHelp, m.width))

	return b.String()
}

type newSecretMsg struct {
	mount string
	path  string
}

type browserDeleteMsg struct {
	mount string
	path  string
	isDir bool
}

type newFolderMsg struct {
	mount string
	path  string
}
