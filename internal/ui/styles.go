package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7DC4E4") // blue
	colorSecondary = lipgloss.Color("#A6DA95") // green
	colorAccent    = lipgloss.Color("#F5A97F") // orange
	colorDanger    = lipgloss.Color("#ED8796") // red
	colorMuted     = lipgloss.Color("#6E738D") // gray
	colorBg        = lipgloss.Color("#24273A") // dark bg
	colorBgLight   = lipgloss.Color("#363A4F") // lighter bg
	colorFg        = lipgloss.Color("#CAD3F5") // light fg
	colorHighlight = lipgloss.Color("#EED49F") // yellow

	// Breadcrumb
	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			PaddingLeft(1)

	breadcrumbSep = lipgloss.NewStyle().
			Foreground(colorMuted).
			SetString(" > ")

	// List items
	itemStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorPrimary).
				Bold(true).
				PaddingLeft(1).
				PaddingRight(1)

	dirItemStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			PaddingLeft(2)

	selectedDirItemStyle = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorSecondary).
				Bold(true).
				PaddingLeft(1).
				PaddingRight(1)

	// Secret view
	keyStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			Width(30)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorFg)

	maskedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	selectedKeyStyle = lipgloss.NewStyle().
				Foreground(colorBg).
				Background(colorAccent).
				Bold(true).
				Width(30).
				PaddingLeft(1).
				PaddingRight(1)

	// Help
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	helpSepStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			SetString("  ")

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Background(colorBgLight).
			PaddingLeft(1).
			PaddingRight(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true).
			PaddingLeft(1)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			PaddingLeft(1)

	// Title
	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			PaddingLeft(1).
			PaddingBottom(1)

	// Filter
	filterPromptStyle = lipgloss.NewStyle().
				Foreground(colorHighlight).
				Bold(true)

	filterTextStyle = lipgloss.NewStyle().
			Foreground(colorFg)
)
