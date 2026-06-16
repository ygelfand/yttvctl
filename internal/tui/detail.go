package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ygelfand/lib-yttv/epg"
)

type detail struct {
	channel string
	airing  epg.Airing
	images  *imageCache
}

const (
	detailThumbW = 30
	detailThumbH = 18
)

type detailCloseMsg struct{}

func newDetail(channel string, a epg.Airing, images *imageCache) *detail {
	return &detail{channel: channel, airing: a, images: images}
}

func (d *detail) fetchThumb(ctx context.Context) tea.Cmd {
	if d.airing.ThumbnailURL == "" {
		return nil
	}
	return d.images.fetch(ctx, "thumb|"+d.airing.VideoID, d.airing.ThumbnailURL, detailThumbW, detailThumbH)
}

func (d *detail) update(msg tea.Msg) (tea.Cmd, bool) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "esc", "q":
			return func() tea.Msg { return detailCloseMsg{} }, true
		}
		return nil, true // swallow all keys while open (except c, which the model handles before this)
	}
	return nil, false
}

func (d *detail) layer(termW, termH int) *lipgloss.Layer {
	overlay := d.body()
	w := lipgloss.Width(overlay)
	h := lipgloss.Height(overlay)
	x := max((termW-w)/2, 0)
	y := max((termH-h)/2, 0)
	return lipgloss.NewLayer(overlay).X(x).Y(y).Z(30)
}

func (d *detail) body() string {
	a := d.airing
	thumb, _ := d.images.get("thumb|"+a.VideoID, detailThumbW, detailThumbH)

	header := lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render(d.channel)

	title := a.Title
	if title == "" {
		title = "—"
	}
	titleLine := lipgloss.NewStyle().Bold(true).Foreground(colorText).Render(title)

	timeRange := fmt.Sprintf("%s – %s",
		time.UnixMilli(a.BeginTimeMs).Local().Format("Mon 3:04 PM"),
		time.UnixMilli(a.EndTimeMs).Local().Format("3:04 PM"),
	)
	if a.IsLive {
		timeRange = styleLiveBadge.Render("LIVE NOW") + "  " + timeRange
	} else {
		timeRange = styleMuted.Render(timeRange)
	}

	subtitle := ""
	if a.Subtitle != "" {
		subtitle = styleMuted.Render(a.Subtitle)
	}

	synopsis := wrapText(a.Synopsis, 60)
	if synopsis == "" {
		synopsis = styleMuted.Render("No description.")
	}

	hint := "[esc] back"
	if a.IsLive {
		hint = "[c] cast   " + hint
	}

	left := lipgloss.NewStyle().Width(64).Render(lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		titleLine,
		subtitle,
		"",
		timeRange,
		"",
		synopsis,
		"",
		hint,
	))

	body := left
	if thumb != "" {
		thumbBlock := lipgloss.NewStyle().MarginRight(2).Render(thumb)
		body = lipgloss.JoinHorizontal(lipgloss.Top, thumbBlock, left)
	}

	return styleOverlay.Render(body)
}

func wrapText(s string, width int) string {
	if s == "" {
		return ""
	}
	words := strings.Fields(s)
	var lines []string
	cur := ""
	for _, w := range words {
		if cur == "" {
			cur = w
			continue
		}
		if lipgloss.Width(cur)+1+lipgloss.Width(w) > width {
			lines = append(lines, cur)
			cur = w
		} else {
			cur += " " + w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return strings.Join(lines, "\n")
}
