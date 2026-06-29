package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ygelfand/lib-yttv/cast"
)

type devicePicker struct {
	devices  []cast.Device
	statuses map[string]*cast.Status // keyed by cast.Device.ID()
	idx      int
	loading  bool
	errorMsg string
}

type deviceChosenMsg struct{ dev cast.Device }
type devicePickerClosedMsg struct{}

func newDevicePicker() *devicePicker {
	return &devicePicker{loading: true}
}

func (p *devicePicker) setDevices(devs []cast.Device) {
	p.devices = devs
	p.loading = false
	if p.idx >= len(devs) {
		p.idx = max(len(devs)-1, 0)
	}
}

func (p *devicePicker) setStatuses(s map[string]*cast.Status) {
	p.statuses = s
}

// nowPlaying returns a short "what's playing" suffix for a device, or "".
func (p *devicePicker) nowPlaying(d cast.Device) string {
	st := p.statuses[d.ID()]
	if st == nil || st.Idle {
		return ""
	}
	if st.Media != nil && st.Media.Title != "" {
		return st.Media.Title
	}
	return st.AppName
}

func (p *devicePicker) selectByName(name string) {
	if name == "" {
		return
	}
	want := strings.ToLower(name)
	for i, d := range p.devices {
		if strings.ToLower(d.Name) == want {
			p.idx = i
			return
		}
	}
}

func (p *devicePicker) setError(err error) {
	p.loading = false
	p.errorMsg = err.Error()
}

func (p *devicePicker) update(msg tea.Msg) (tea.Cmd, bool) {
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		switch m.String() {
		case "j", "down":
			if len(p.devices) > 0 {
				p.idx = (p.idx + 1) % len(p.devices)
			}
			return nil, true
		case "k", "up":
			if len(p.devices) > 0 {
				p.idx = (p.idx - 1 + len(p.devices)) % len(p.devices)
			}
			return nil, true
		case "enter":
			if len(p.devices) == 0 {
				return nil, true
			}
			dev := p.devices[p.idx]
			return func() tea.Msg { return deviceChosenMsg{dev: dev} }, true
		case "esc", "d", "q":
			return func() tea.Msg { return devicePickerClosedMsg{} }, true
		}
		return nil, true
	}
	return nil, false
}

func (p *devicePicker) layer(termW, termH int) *lipgloss.Layer {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render("Cast device"))
	b.WriteString("\n\n")

	switch {
	case p.loading:
		b.WriteString(styleMuted.Render("Discovering Chromecasts on the LAN…"))
	case p.errorMsg != "":
		b.WriteString(styleError.Render(p.errorMsg))
	case len(p.devices) == 0:
		b.WriteString(styleMuted.Render("No Chromecast devices found. Press [esc] to close."))
	default:
		for i, d := range p.devices {
			if i == p.idx {
				b.WriteString(styleCellSelected.Render("▸ " + d.Name))
			} else {
				b.WriteString("  ")
				b.WriteString(d.Name)
			}
			if np := p.nowPlaying(d); np != "" {
				b.WriteString(styleMuted.Render("  ▶ " + np))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(styleMuted.Render("[j/k] move  [enter] select  [esc/d] close"))

	overlay := styleOverlay.Render(b.String())
	w := lipgloss.Width(overlay)
	h := lipgloss.Height(overlay)
	x := max((termW-w)/2, 0)
	y := max((termH-h)/2, 0)
	return lipgloss.NewLayer(overlay).X(x).Y(y).Z(20)
}
