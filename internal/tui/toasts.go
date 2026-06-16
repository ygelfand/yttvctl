package tui

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type toastKind int

const (
	toastInfo toastKind = iota
	toastError
	toastCast
	toastStop
)

const (
	toastTTL      = 6 * time.Second
	toastWidth    = 44
	toastMaxStack = 4
)

type toast struct {
	id      uint64
	kind    toastKind
	message string
	expires time.Time
}

type toastExpireMsg struct{ id uint64 }

type toasts struct {
	items  []toast
	nextID uint64
}

func newToasts() *toasts { return &toasts{} }

func (t *toasts) push(kind toastKind, message string) tea.Cmd {
	t.nextID++
	id := t.nextID
	t.items = append(t.items, toast{
		id:      id,
		kind:    kind,
		message: message,
		expires: time.Now().Add(toastTTL),
	})
	return tea.Tick(toastTTL, func(time.Time) tea.Msg { return toastExpireMsg{id: id} })
}

func (t *toasts) expire(id uint64) {
	out := t.items[:0]
	for _, x := range t.items {
		if x.id != id {
			out = append(out, x)
		}
	}
	t.items = out
}

func (t *toasts) layer(termW, _ int) *lipgloss.Layer {
	if len(t.items) == 0 {
		return nil
	}
	start := 0
	if len(t.items) > toastMaxStack {
		start = len(t.items) - toastMaxStack
	}
	visible := t.items[start:]

	lines := make([]string, 0, len(visible)*2)
	for i := len(visible) - 1; i >= 0; i-- {
		lines = append(lines, renderToast(visible[i]))
		if i > 0 {
			lines = append(lines, "")
		}
	}
	stack := strings.Join(lines, "\n")

	x := max(termW-toastWidth-1, 0)
	return lipgloss.NewLayer(stack).X(x).Y(2).Z(50)
}

func renderToast(t toast) string {
	var prefix string
	var border color.Color
	switch t.kind {
	case toastError:
		prefix = "✖"
		border = colorAccent
	case toastCast:
		prefix = "▶"
		border = colorCast
	case toastStop:
		prefix = "■"
		border = colorCast
	default:
		prefix = "•"
		border = colorCast
	}
	head := lipgloss.NewStyle().Bold(true).Foreground(border).Render(prefix)
	body := head + "  " + wrapText(t.message, toastWidth-7)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Background(colorPanel).
		Foreground(colorText).
		Width(toastWidth-2).
		Padding(0, 1).
		Render(body)
}
