package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/ygelfand/lib-yttv/epg"
)

// guide is the channel × time grid. Each airing cell's width is proportional
// to its duration within the visible window.
type guide struct {
	channels []epg.Channel

	width, height int

	axisStart time.Time // floor(now → 30min)
	axisEnd   time.Time // axisStart + windowDuration
	now       time.Time

	rowIdx    int // channel row
	colIdx    int // index into the selected row's visible slot list
	rowOffset int // vertical scroll

	logos *imageCache
}

const (
	channelColWidth = 12
	rowHeight       = 5             // 4 halfblock rows + 1 name row
	windowDuration  = 2 * time.Hour // visible time span
	minCellWidth    = 4             // cells narrower than this get rendered as a sliver
)

func newGuide(logos *imageCache) *guide {
	now := time.Now()
	return &guide{
		now:       now,
		axisStart: now.Truncate(30 * time.Minute),
		axisEnd:   now.Truncate(30 * time.Minute).Add(windowDuration),
		logos:     logos,
	}
}

func (g *guide) setChannels(chs []epg.Channel) {
	g.channels = chs
	if g.rowIdx >= len(chs) {
		g.rowIdx = len(chs) - 1
	}
	if g.rowIdx < 0 {
		g.rowIdx = 0
	}
	g.refreshTime()
	g.snapColToLive()
}

func (g *guide) refreshTime() {
	g.now = time.Now()
	g.axisStart = g.now.Truncate(30 * time.Minute)
	g.axisEnd = g.axisStart.Add(windowDuration)
}

func (g *guide) setSize(w, h int) {
	g.width = w
	g.height = h
}

func (g *guide) availableCellWidth() int {
	w := g.width - channelColWidth
	if w < 1 {
		return 1
	}
	return w
}

type airingSlot struct {
	airing epg.Airing
	x0     int // start column (within the cell area)
	width  int // width in columns
}

func (g *guide) computeRowSlots(ch epg.Channel) []airingSlot {
	startMs := g.axisStart.UnixMilli()
	endMs := g.axisEnd.UnixMilli()
	avail := g.availableCellWidth()
	span := float64(endMs - startMs)
	if span <= 0 {
		return nil
	}
	slots := make([]airingSlot, 0, len(ch.Airings))
	for _, a := range ch.Airings {
		if a.EndTimeMs <= startMs || a.BeginTimeMs >= endMs {
			continue
		}
		b := max(a.BeginTimeMs, startMs)
		e := min(a.EndTimeMs, endMs)
		x0 := int(float64(b-startMs) / span * float64(avail))
		x1 := int(float64(e-startMs) / span * float64(avail))
		w := x1 - x0
		if w < 1 {
			continue
		}
		slots = append(slots, airingSlot{airing: a, x0: x0, width: w})
	}
	return slots
}

func (g *guide) selectedChannel() *epg.Channel {
	if g.rowIdx < 0 || g.rowIdx >= len(g.channels) {
		return nil
	}
	return &g.channels[g.rowIdx]
}

func (g *guide) selectedRowSlots() []airingSlot {
	ch := g.selectedChannel()
	if ch == nil {
		return nil
	}
	return g.computeRowSlots(*ch)
}

func (g *guide) selectedAiring() *epg.Airing {
	slots := g.selectedRowSlots()
	if len(slots) == 0 || g.colIdx < 0 || g.colIdx >= len(slots) {
		return nil
	}
	return &slots[g.colIdx].airing
}

func (g *guide) snapColToLive() {
	slots := g.selectedRowSlots()
	if len(slots) == 0 {
		g.colIdx = 0
		return
	}
	for i, s := range slots {
		if s.airing.IsLive {
			g.colIdx = i
			return
		}
	}
	g.colIdx = 0
}

func (g *guide) moveRow(delta int) {
	if len(g.channels) == 0 {
		return
	}
	g.rowIdx = clamp(g.rowIdx+delta, 0, len(g.channels)-1)
	slots := g.selectedRowSlots()
	g.colIdx = clamp(g.colIdx, 0, max(len(slots)-1, 0))
	visible := g.visibleRows()
	if g.rowIdx < g.rowOffset {
		g.rowOffset = g.rowIdx
	}
	if g.rowIdx >= g.rowOffset+visible {
		g.rowOffset = g.rowIdx - visible + 1
	}
}

func (g *guide) moveCol(delta int) {
	slots := g.selectedRowSlots()
	if len(slots) == 0 {
		return
	}
	g.colIdx = clamp(g.colIdx+delta, 0, len(slots)-1)
}

func (g *guide) visibleRows() int {
	const (
		axisLines    = 2
		bottomMargin = 1
	)
	avail := g.height - axisLines - bottomMargin
	if avail < rowHeight {
		return 1
	}
	return avail / rowHeight
}

func (g *guide) view() string {
	if len(g.channels) == 0 {
		return styleMuted.Render("Loading channel guide…")
	}

	visibleR := g.visibleRows()
	out := []string{g.renderTimeAxis()}
	end := min(g.rowOffset+visibleR, len(g.channels))
	for i := g.rowOffset; i < end; i++ {
		out = append(out, g.renderRow(i))
	}
	return strings.Join(out, "\n")
}

func (g *guide) renderTimeAxis() string {
	avail := g.availableCellWidth()
	span := float64(g.axisEnd.UnixMilli() - g.axisStart.UnixMilli())
	bar := []rune(strings.Repeat(" ", avail))

	for t := g.axisStart; t.Before(g.axisEnd); t = t.Add(30 * time.Minute) {
		offsetMs := float64(t.UnixMilli() - g.axisStart.UnixMilli())
		x := int(offsetMs / span * float64(avail))
		label := t.Format("3:04 PM")
		for j, r := range label {
			if x+j < len(bar) {
				bar[x+j] = r
			}
		}
	}

	nowOff := float64(g.now.UnixMilli() - g.axisStart.UnixMilli())
	nx := int(nowOff / span * float64(avail))
	nowSegment := ""
	if nx >= 0 && nx < avail {
		nowSegment = strings.Repeat(" ", nx) + styleTimeAxisNow.Render("▼")
	}

	axisStyled := styleTimeAxis.Render(string(bar))
	left := strings.Repeat(" ", channelColWidth)
	if nowSegment != "" {
		return left + nowSegment + "\n" + left + axisStyled
	}
	return left + axisStyled
}

func (g *guide) renderRow(rowIdx int) string {
	ch := g.channels[rowIdx]
	slots := g.computeRowSlots(ch)
	avail := g.availableCellWidth()

	logoView := ""
	if g.logos != nil {
		logoView, _ = g.logos.getLogo(ch.Name)
	}
	channelCell := g.renderChannelCell(ch, logoView)

	cursor := 0
	pieces := make([]string, 0, len(slots)*2+1)
	for i, s := range slots {
		if s.x0 > cursor {
			pieces = append(pieces, g.renderGap(s.x0-cursor))
			cursor = s.x0
		}
		selected := rowIdx == g.rowIdx && i == g.colIdx
		pieces = append(pieces, g.renderAiringCell(s.airing, s.width, selected))
		cursor += s.width
	}
	if cursor < avail {
		pieces = append(pieces, g.renderGap(avail-cursor))
	}
	cellsRow := lipgloss.JoinHorizontal(lipgloss.Top, pieces...)
	return lipgloss.JoinHorizontal(lipgloss.Top, channelCell, cellsRow)
}

func (g *guide) renderChannelCell(ch epg.Channel, logo string) string {
	logoBudget := rowHeight - 1

	logoLines := []string{}
	if logo != "" {
		logoLines = append(logoLines, strings.Split(strings.TrimRight(logo, "\n"), "\n")...)
	}
	if len(logoLines) > logoBudget {
		logoLines = logoLines[:logoBudget]
	}

	lines := make([]string, 0, rowHeight)
	for i := 0; i < logoBudget-len(logoLines); i++ {
		lines = append(lines, padTo("", channelColWidth))
	}
	for _, l := range logoLines {
		lines = append(lines, padTo(l, channelColWidth))
	}
	name := truncate(ch.Name, channelColWidth-2)
	lines = append(lines, padTo(styleChannelName.Render(name), channelColWidth))
	return strings.Join(lines, "\n")
}

func (g *guide) renderAiringCell(a epg.Airing, width int, selected bool) string {
	st := styleCell
	if selected {
		st = styleCellSelected
	}

	if width < minCellWidth {
		return g.renderSliver(a, width, selected)
	}

	contentWidth := max(width-2, 1)

	title := a.Title
	if title == "" {
		title = "—"
	}
	timeRange := fmt.Sprintf(
		"%s–%s",
		time.UnixMilli(a.BeginTimeMs).Local().Format("3:04"),
		time.UnixMilli(a.EndTimeMs).Local().Format("3:04"),
	)

	mid := styleMuted.Render(truncate(a.Subtitle, contentWidth))
	if a.IsLive {
		mid = g.renderProgressBar(a, contentWidth)
	}

	lines := []string{
		truncate(title, contentWidth),
		mid,
		styleMuted.Render(truncate(timeRange, contentWidth)),
	}

	body := strings.Join(lines, "\n")
	return st.Width(width).Height(rowHeight).Render(body)
}

func (g *guide) renderProgressBar(a epg.Airing, width int) string {
	if width <= 0 {
		return ""
	}
	total := float64(a.EndTimeMs - a.BeginTimeMs)
	if total <= 0 {
		return lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("━", width))
	}
	elapsed := float64(g.now.UnixMilli() - a.BeginTimeMs)
	pct := elapsed / total
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * float64(width))
	if filled < 1 && pct > 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}
	fill := lipgloss.NewStyle().Foreground(colorLive).Render(strings.Repeat("━", filled))
	tail := lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("━", width-filled))
	return fill + tail
}

func (g *guide) renderSliver(a epg.Airing, width int, selected bool) string {
	fg := colorMuted
	switch {
	case selected:
		fg = colorAccent
	case a.IsLive:
		fg = colorLive
	}
	bar := strings.Repeat("│", width)
	style := lipgloss.NewStyle().Foreground(fg)
	lines := make([]string, rowHeight)
	for i := range lines {
		lines[i] = style.Render(bar)
	}
	return strings.Join(lines, "\n")
}

func (g *guide) renderGap(width int) string {
	if width <= 0 {
		return ""
	}
	empty := strings.Repeat(" ", width)
	lines := make([]string, rowHeight)
	for i := range lines {
		lines[i] = empty
	}
	return strings.Join(lines, "\n")
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func padTo(s string, w int) string {
	d := w - lipgloss.Width(s)
	if d <= 0 {
		return s
	}
	return s + strings.Repeat(" ", d)
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	if w <= 1 {
		return string(r[:w])
	}
	return string(r[:w-1]) + "…"
}
