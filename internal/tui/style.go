package tui

import "charm.land/lipgloss/v2"

var (
	colorPanel    = lipgloss.Color("#1a1a1a")
	colorBorder   = lipgloss.Color("#3a3a3a")
	colorText     = lipgloss.Color("#e6e6e6")
	colorMuted    = lipgloss.Color("#7a7a7a")
	colorAccent   = lipgloss.Color("#ff4d4f") // YouTube TV red-ish
	colorAccentFg = lipgloss.Color("#ffffff")
	colorLive     = lipgloss.Color("#ff0033")
	colorNowLine  = lipgloss.Color("#ffcc00")
	colorCast     = lipgloss.Color("#4dabf7")
)

var (
	styleFooter = lipgloss.NewStyle().
			Foreground(colorMuted).
			Background(colorPanel).
			Padding(0, 1)

	styleChannelName = lipgloss.NewStyle().
				Foreground(colorText).
				Bold(true)

	styleCell = lipgloss.NewStyle().
			Foreground(colorText).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorBorder)

	styleCellSelected = lipgloss.NewStyle().
				Foreground(colorAccentFg).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(colorAccent).
				Bold(true)

	styleLiveBadge = lipgloss.NewStyle().
			Foreground(colorAccentFg).
			Background(colorLive).
			Bold(true).
			Padding(0, 1)

	styleTimeAxis = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleTimeAxisNow = lipgloss.NewStyle().
				Foreground(colorNowLine).
				Bold(true)

	styleMuted = lipgloss.NewStyle().Foreground(colorMuted)

	styleError = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	styleOverlay = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2).
			Background(colorPanel).
			Foreground(colorText)
)
