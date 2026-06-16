package tui

import (
	_ "embed"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

//go:embed splash.ans
var splashArt string

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type splashTickMsg struct{}

func splashTick() tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg { return splashTickMsg{} })
}

func renderSplash(termW, termH int, status string, frame int) string {
	art := strings.TrimRight(splashArt, "\n")

	spinner := lipgloss.NewStyle().Foreground(colorAccent).Render(spinnerFrames[frame%len(spinnerFrames)])
	statusLine := spinner + "  " + lipgloss.NewStyle().Foreground(colorText).Render(status)

	parts := []string{}
	if strings.TrimSpace(art) != "" {
		parts = append(parts, art, "")
	}
	parts = append(parts, statusLine)

	body := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return lipgloss.Place(termW, termH, lipgloss.Center, lipgloss.Center, body)
}
