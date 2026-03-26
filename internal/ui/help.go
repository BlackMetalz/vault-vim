package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpBinding struct {
	key  string
	desc string
}

func renderHelp(bindings []helpBinding, width int) string {
	var parts []string
	for _, b := range bindings {
		parts = append(parts,
			helpKeyStyle.Render(b.key)+
				helpDescStyle.Render(" "+b.desc))
	}

	// Arrange in rows, fitting to width
	sep := helpSepStyle.String()
	var rows []string
	var current []string
	currentLen := 0

	for _, p := range parts {
		pLen := lipgloss.Width(p)
		if currentLen > 0 && currentLen+lipgloss.Width(sep)+pLen > width-2 {
			rows = append(rows, strings.Join(current, sep))
			current = nil
			currentLen = 0
		}
		current = append(current, p)
		if currentLen > 0 {
			currentLen += lipgloss.Width(sep)
		}
		currentLen += pLen
	}
	if len(current) > 0 {
		rows = append(rows, strings.Join(current, sep))
	}

	border := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorMuted).
		Padding(0, 1).
		Width(width - 2)

	return border.Render(strings.Join(rows, "\n"))
}

var mountsHelp = []helpBinding{
	{"enter", "select"},
	{"/", "filter"},
	{"?", "help"},
	{"q", "quit"},
}

var browserHelp = []helpBinding{
	{"enter", "open"},
	{"n", "new secret"},
	{"N", "new folder"},
	{"ctrl+d", "delete"},
	{"esc", "back"},
	{"/", "filter"},
	{"?", "help"},
	{"q", "quit"},
}

var secretHelp = []helpBinding{
	{"e", "edit value"},
	{"a", "add key"},
	{"i", "import JSON"},
	{"r", "rename key"},
	{"ctrl+d", "delete key"},
	{"s", "reveal/hide"},
	{"S", "reveal/hide all"},
	{"esc", "back"},
	{"?", "help"},
	{"q", "quit"},
}

var editorHelp = []helpBinding{
	{"enter", "confirm"},
	{"esc", "cancel"},
}

func renderHelpModal(current view, width int) string {
	modalW := 60
	if width < 64 {
		modalW = width - 4
	}

	title := lipgloss.NewStyle().
		Background(colorPrimary).
		Foreground(colorBg).
		Bold(true).
		Padding(0, 1).
		Render(" KEYBOARD SHORTCUTS ")

	var sections []string

	// Global
	sections = append(sections, renderHelpSection("Global", []helpBinding{
		{"?", "Toggle this help"},
		{"q", "Quit"},
		{"up/down", "Move up/down"},
		{"/", "Filter"},
		{"esc", "Go back"},
	}, modalW-4))

	// Context-specific
	_ = current // show all sections regardless of current view

	sections = append(sections, renderHelpSection("Browser", []helpBinding{
		{"enter", "Open path / secret"},
		{"n", "Create new secret"},
		{"N", "Create new folder"},
		{"ctrl+d", "Delete secret / folder"},
	}, modalW-4))

	sections = append(sections, renderHelpSection("Secret View", []helpBinding{
		{"e", "Edit selected value"},
		{"a", "Add new key"},
		{"i", "Import from JSON (overwrite)"},
		{"r", "Rename key"},
		{"ctrl+d", "Delete key"},
		{"s", "Reveal/hide selected"},
		{"S", "Reveal/hide all"},
	}, modalW-4))

	// Editor
	sections = append(sections, renderHelpSection("Editor", []helpBinding{
		{"ctrl+s / F2", "Save"},
		{"esc", "Cancel"},
		{"enter", "New line"},
		{"ctrl+d", "Delete line"},
		{"ctrl+a", "Go to line start"},
		{"ctrl+e", "Go to line end"},
		{"arrows", "Move cursor"},
	}, modalW-4))

	footer := lipgloss.NewStyle().
		Foreground(colorMuted).
		Italic(true).
		Render("Press any key to close")

	content := title + "\n\n" +
		strings.Join(sections, "\n\n") +
		"\n\n" + footer

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Width(modalW)

	return box.Render(content)
}

func renderHelpSection(title string, bindings []helpBinding, width int) string {
	header := lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true).
		Render(title)

	var rows []string
	for _, b := range bindings {
		key := helpKeyStyle.Width(14).Render(b.key)
		desc := helpDescStyle.Render(b.desc)
		rows = append(rows, "  "+key+" "+desc)
	}

	return header + "\n" + strings.Join(rows, "\n")
}
