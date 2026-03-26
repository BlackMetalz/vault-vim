package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BlackMetalz/vault-vim/internal/vault"
)

type view int

const (
	viewMounts view = iota
	viewBrowser
	viewSecret
)

type browserReloadMsg struct{}

type AppModel struct {
	client   *vault.Client
	current  view
	mounts   mountsModel
	browser  browserModel
	secret   secretModel
	editor   editorOverlay
	showHelp bool
	version  string
	width    int
	height   int
}

func NewApp(client *vault.Client, version string) AppModel {
	return AppModel{
		client:  client,
		version: version,
		current: viewMounts,
		mounts:  newMountsModel(client),
		editor:  newEditorOverlay(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.mounts.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propagate to all sub-models
		m.mounts.width = msg.Width
		m.mounts.height = msg.Height
		m.browser.width = msg.Width
		m.browser.height = msg.Height
		m.secret.width = msg.Width
		m.secret.height = msg.Height
		m.editor.width = msg.Width
		m.editor.height = msg.Height
		return m, nil
	}

	// Help modal: dismiss on any key
	if m.showHelp {
		if _, ok := msg.(tea.KeyMsg); ok {
			m.showHelp = false
			return m, nil
		}
	}

	// Toggle help modal with ?
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "?" {
		m.showHelp = true
		return m, nil
	}

	// If editor is active, route all messages to editor
	if m.editor.active {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			var cmd tea.Cmd
			m.editor, cmd = m.editor.Update(keyMsg)
			return m, cmd
		}
		// Route non-key messages (paste, blink, etc.) to textarea/input
		var cmd tea.Cmd
		m.editor, cmd = m.editor.UpdateMsg(msg)
		return m, cmd
	}

	// Handle editor results
	if msg, ok := msg.(editorDoneMsg); ok {
		return m.handleEditorDone(msg)
	}

	// Handle view transitions
	switch msg := msg.(type) {
	case mountSelectedMsg:
		m.current = viewBrowser
		m.browser = newBrowserModel(m.client, msg.mount.Path)
		m.browser.width = m.width
		m.browser.height = m.height
		return m, m.browser.Init()

	case secretSelectedMsg:
		m.current = viewSecret
		m.secret = newSecretModel(m.client, msg.mount, msg.path)
		m.secret.width = m.width
		m.secret.height = m.height
		return m, m.secret.Init()

	case browserBackMsg:
		m.current = viewMounts
		return m, m.mounts.Init()

	case secretBackMsg:
		m.current = viewBrowser
		return m, m.browser.loadItems()

	case browserReloadMsg:
		return m, m.browser.loadItems()

	case startEditMsg:
		m.editor = m.editor.startEdit(msg.key, msg.value)
		return m, nil

	case startAddMsg:
		m.editor = m.editor.startAdd()
		return m, nil

	case startDeleteMsg:
		m.editor = m.editor.startDelete(msg.key)
		return m, nil

	case startRenameMsg:
		m.editor = m.editor.startRename(msg.key)
		return m, nil

	case browserDeleteMsg:
		m.editor = m.editor.startDeletePath(msg.mount, msg.path, msg.isDir)
		return m, nil

	case newSecretMsg:
		m.editor = m.editor.startNewSecret(msg.mount, msg.path)
		return m, nil

	case newFolderMsg:
		m.editor = m.editor.startNewFolder(msg.mount, msg.path)
		return m, nil
	}

	// Route to current view
	var cmd tea.Cmd
	switch m.current {
	case viewMounts:
		m.mounts, cmd = m.mounts.Update(msg)
	case viewBrowser:
		m.browser, cmd = m.browser.Update(msg)
	case viewSecret:
		m.secret, cmd = m.secret.Update(msg)
	}

	return m, cmd
}

func (m AppModel) handleEditorDone(msg editorDoneMsg) (tea.Model, tea.Cmd) {
	switch msg.action {
	case "cancel":
		return m, nil
	case "edit":
		mount := m.secret.mount
		path := m.secret.path
		client := m.client
		updates := map[string]string{msg.key: msg.value}
		return m, func() tea.Msg {
			err := client.PatchSecret(mount, path, updates)
			return secretUpdatedMsg{err: err}
		}
	case "add":
		mount := m.secret.mount
		path := m.secret.path
		client := m.client
		updates := map[string]string{msg.key: msg.value}
		return m, func() tea.Msg {
			err := client.PatchSecret(mount, path, updates)
			return secretUpdatedMsg{err: err}
		}
	case "delete":
		mount := m.secret.mount
		path := m.secret.path
		key := msg.key
		client := m.client
		return m, func() tea.Msg {
			err := client.DeleteKey(mount, path, key)
			return secretUpdatedMsg{err: err}
		}
	case "rename":
		// msg.key = new name, msg.value = old name
		mount := m.secret.mount
		path := m.secret.path
		newKey := msg.key
		oldKey := msg.value
		client := m.client
		return m, func() tea.Msg {
			err := client.RenameKey(mount, path, oldKey, newKey)
			return secretUpdatedMsg{err: err}
		}
	case "newsecret":
		// msg.key = mount, msg.value = full path
		// Navigate to secret view and auto-open add key editor
		mount := msg.key
		path := msg.value
		m.current = viewSecret
		m.secret = newSecretModel(m.client, mount, path)
		m.secret.width = m.width
		m.secret.height = m.height
		m.editor = m.editor.startAdd()
		return m, m.secret.Init()

	case "newfolder":
		// msg.key = mount, msg.value = folder path
		// Just navigate into the folder path — it'll show as empty
		// Folder appears in Vault once a secret is created inside
		mount := msg.key
		folderPath := msg.value
		m.current = viewBrowser
		m.browser = newBrowserModel(m.client, mount)
		m.browser.width = m.width
		m.browser.height = m.height
		// Set the path segments
		parts := strings.Split(strings.TrimSuffix(folderPath, "/"), "/")
		for _, p := range parts {
			if p != "" {
				m.browser.path = append(m.browser.path, p)
			}
		}
		return m, m.browser.loadItems()
	case "deletepath":
		// msg.key = mount, msg.value = path
		mount := msg.key
		path := msg.value
		client := m.client
		return m, func() tea.Msg {
			err := client.DeleteSecret(mount, path)
			if err != nil {
				return secretUpdatedMsg{err: err}
			}
			return browserReloadMsg{}
		}
	}
	return m, nil
}

func (m AppModel) View() string {
	var mainView string

	switch m.current {
	case viewMounts:
		mainView = m.mounts.View()
	case viewBrowser:
		mainView = m.browser.View()
	case viewSecret:
		mainView = m.secret.View()
	}

	// Overlay editor if active
	if m.editor.active {
		editorView := m.editor.View()
		overlay := lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			editorView,
		)
		return overlay
	}

	// Help modal overlay
	if m.showHelp {
		helpView := renderHelpModal(m.current, m.width)
		overlay := lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			helpView,
		)
		return overlay
	}

	// Version bottom-right
	ver := lipgloss.NewStyle().Foreground(colorMuted).Render(m.version)
	verLine := lipgloss.PlaceHorizontal(m.width, lipgloss.Right, ver)
	return mainView + "\n" + verLine
}
