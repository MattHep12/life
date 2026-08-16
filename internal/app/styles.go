package app

import "charm.land/lipgloss/v2"

var (
	colorBackground = lipgloss.Color("#191C2A")
	colorSurface    = lipgloss.Color("#252A3D")
	colorSurfaceAlt = lipgloss.Color("#32384F")
	colorBorder     = lipgloss.Color("#59617F")
	colorPrimary    = lipgloss.Color("#A7A8FF")
	colorSecondary  = lipgloss.Color("#69E9BF")
	colorInfo       = lipgloss.Color("#65CCFF")
	colorTax        = lipgloss.Color("#FF7F96")
	colorGold       = lipgloss.Color("#FFD166")
	colorWarning    = lipgloss.Color("#FFD580")
	colorDanger     = lipgloss.Color("#FF819C")
	colorText       = lipgloss.Color("#F7F8FC")
	colorMuted      = lipgloss.Color("#AFB5CA")

	logoStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBackground).
			Background(colorPrimary).Padding(0, 1)
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorText)
	subtitleStyle = lipgloss.NewStyle().Foreground(colorMuted)
	sectionStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)
	panelStyle    = lipgloss.NewStyle().Background(colorSurface).
			Border(lipgloss.RoundedBorder()).BorderForeground(colorBorder).Padding(1, 2)
	activePanelStyle = panelStyle.BorderForeground(colorPrimary)
	selectedRowStyle = lipgloss.NewStyle().Bold(true).Foreground(colorText).
				Background(colorSurfaceAlt).Padding(0, 1)
	rowStyle        = lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 1)
	payrollRowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#DDE1F0")).
			Background(colorSurface).Padding(0, 1)
	tableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(colorInfo).
				Background(colorSurface).Padding(0, 1)
	valueStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorSecondary)
	mutedStyle    = lipgloss.NewStyle().Foreground(colorMuted)
	helpStyle     = lipgloss.NewStyle().Foreground(colorMuted).MarginTop(1)
	keyStyle      = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)
	errorStyle    = lipgloss.NewStyle().Foreground(colorDanger).Background(colorSurfaceAlt).Padding(0, 1)
	warningStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorWarning)
	inputBoxStyle = lipgloss.NewStyle().Background(colorSurfaceAlt).
			Border(lipgloss.RoundedBorder()).BorderForeground(colorPrimary).Padding(0, 1)
)
