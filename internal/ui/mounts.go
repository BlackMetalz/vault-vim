package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/BlackMetalz/vault-vim/internal/vault"
)

type mountsModel struct {
	client   *vault.Client
	mounts   []vault.Mount
	cursor   int
	filter   string
	filtered []vault.Mount
	filtering bool
	err      error
	width    int
	height   int
}

type mountsLoadedMsg struct {
	mounts []vault.Mount
	err    error
}

type mountSelectedMsg struct {
	mount vault.Mount
}

func newMountsModel(client *vault.Client) mountsModel {
	return mountsModel{
		client: client,
	}
}

func (m mountsModel) Init() tea.Cmd {
	return func() tea.Msg {
		mounts, err := m.client.ListMounts()
		return mountsLoadedMsg{mounts: mounts, err: err}
	}
}

func (m mountsModel) Update(msg tea.Msg) (mountsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case mountsLoadedMsg:
		m.err = msg.err
		m.mounts = msg.mounts
		m.filtered = msg.mounts
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

func (m mountsModel) updateNormal(msg tea.KeyMsg) (mountsModel, tea.Cmd) {
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
			return m, func() tea.Msg {
				return mountSelectedMsg{mount: m.filtered[m.cursor]}
			}
		}
	case "/":
		m.filtering = true
	}
	return m, nil
}

func (m mountsModel) updateFilter(msg tea.KeyMsg) (mountsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.filtering = false
		m.filter = ""
		m.filtered = m.mounts
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

func (m *mountsModel) applyFilter() {
	if m.filter == "" {
		m.filtered = m.mounts
		m.cursor = 0
		return
	}
	var filtered []vault.Mount
	for _, mt := range m.mounts {
		if strings.Contains(strings.ToLower(mt.Path), strings.ToLower(m.filter)) {
			filtered = append(filtered, mt)
		}
	}
	m.filtered = filtered
	m.cursor = 0
}

func (m mountsModel) View() string {
	if m.err != nil {
		return errorStyle.Render("Error: " + m.err.Error())
	}

	if m.mounts == nil {
		return titleStyle.Render("Loading mounts...")
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("vault-vim") + "\n")
	b.WriteString(renderBreadcrumb([]string{"Mounts"}) + "\n\n")

	// List
	for i, mt := range m.filtered {
		label := " " + mt.Path + " (" + mt.Type + ")"
		if i == m.cursor {
			b.WriteString(selectedDirItemStyle.Render(label) + "\n")
		} else {
			b.WriteString(dirItemStyle.Render(label) + "\n")
		}
	}

	if len(m.filtered) == 0 {
		b.WriteString(itemStyle.Render("  No KV v2 mounts found") + "\n")
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
	b.WriteString(renderHelp(mountsHelp, m.width))

	return b.String()
}
