package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type searchMatch struct {
	rowIdx    int
	airingIdx int    // -1 if the channel name matched but no visible airing did
	label     string // pre-rendered "Channel · Show — Subtitle" for the popup
}

type searcher struct {
	guide *guide

	query   string
	matches []searchMatch
	idx     int

	origRow, origCol int
}

type searchClosedMsg struct{}

func newSearcher(g *guide) *searcher {
	return &searcher{
		guide:   g,
		origRow: g.rowIdx,
		origCol: g.colIdx,
	}
}

func (s *searcher) recompute() {
	s.matches = s.matches[:0]
	if s.query == "" {
		return
	}
	q := strings.ToLower(s.query)
	for ri, ch := range s.guide.channels {
		nameHit := strings.Contains(strings.ToLower(ch.Name), q)
		slots := s.guide.computeRowSlots(ch)
		slotHit := -1
		for si, sl := range slots {
			if strings.Contains(strings.ToLower(sl.airing.Title), q) ||
				strings.Contains(strings.ToLower(sl.airing.Subtitle), q) {
				slotHit = si
				break
			}
		}
		switch {
		case slotHit >= 0:
			a := slots[slotHit].airing
			label := ch.Name + "  ·  " + a.Title
			if a.Subtitle != "" {
				label += "  —  " + a.Subtitle
			}
			s.matches = append(s.matches, searchMatch{ri, slotHit, label})
		case nameHit:
			s.matches = append(s.matches, searchMatch{ri, -1, ch.Name})
		}
	}
	if s.idx >= len(s.matches) {
		s.idx = max(len(s.matches)-1, 0)
	}
}

func (s *searcher) applyJump() {
	if len(s.matches) == 0 {
		return
	}
	m := s.matches[s.idx]
	s.guide.rowIdx = m.rowIdx
	if m.airingIdx >= 0 {
		s.guide.colIdx = m.airingIdx
	} else {
		s.guide.snapColToLive()
	}
	// nudge scroll offset to reveal the new row
	s.guide.moveRow(0)
}

func (s *searcher) update(msg tea.Msg) (tea.Cmd, bool) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, false
	}
	switch k.String() {
	case "esc":
		s.guide.rowIdx = s.origRow
		s.guide.colIdx = s.origCol
		s.guide.moveRow(0)
		return func() tea.Msg { return searchClosedMsg{} }, true
	case "enter":
		return func() tea.Msg { return searchClosedMsg{} }, true
	case "up":
		if len(s.matches) > 0 {
			s.idx = (s.idx - 1 + len(s.matches)) % len(s.matches)
			s.applyJump()
		}
		return nil, true
	case "down":
		if len(s.matches) > 0 {
			s.idx = (s.idx + 1) % len(s.matches)
			s.applyJump()
		}
		return nil, true
	case "backspace":
		if s.query != "" {
			r := []rune(s.query)
			s.query = string(r[:len(r)-1])
			s.idx = 0
			s.recompute()
			s.applyJump()
		}
		return nil, true
	}
	// Append printable single-rune keys to the query.
	str := k.String()
	if len([]rune(str)) == 1 && str >= " " {
		s.query += str
		s.idx = 0
		s.recompute()
		s.applyJump()
	}
	return nil, true
}

func (s *searcher) layer(termW, termH int) *lipgloss.Layer {
	maxW := min(termW-4, 60)
	if maxW < 16 {
		maxW = 16
	}

	prompt := lipgloss.NewStyle().Foreground(colorAccent).Render("/") + s.query + "_"
	var status string
	switch {
	case s.query == "":
		status = styleMuted.Render("type to search")
	case len(s.matches) == 0:
		status = styleMuted.Render("no matches")
	default:
		count := styleMuted.Render(fmt.Sprintf("↑↓ %d/%d", s.idx+1, len(s.matches)))
		label := truncate(s.matches[s.idx].label, maxW-lipgloss.Width(prompt)-lipgloss.Width(count)-3)
		status = count + "  " + label
	}

	body := prompt + "  " + status
	overlay := styleOverlay.Padding(0, 1).Render(truncate(body, maxW))
	w := lipgloss.Width(overlay)
	h := lipgloss.Height(overlay)
	x := max(termW-w-1, 0)
	y := max(termH-h-2, 0)
	return lipgloss.NewLayer(overlay).X(x).Y(y).Z(20)
}
