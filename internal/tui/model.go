package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	yttv "github.com/ygelfand/lib-yttv"
	"github.com/ygelfand/lib-yttv/cast"
	"github.com/ygelfand/lib-yttv/epg"
	"github.com/ygelfand/yttvctl/internal/config"
)

type model struct {
	ctx     context.Context
	sess    *yttv.Session
	cfg     *config.Config
	devCfg  string
	devAddr string

	guide  *guide
	picker *devicePicker
	detail *detail
	search *searcher
	images *imageCache
	toasts *toasts

	devices []cast.Device

	width, height int

	device cast.Device
	hasDev bool

	status     string
	pickerOpen bool
	detailOpen bool
	searchOpen bool

	splashFrame int
	ready       bool
}

func newModel(ctx context.Context, cfg *config.Config, devCfg, devAddr string) *model {
	images := newImageCache()
	return &model{
		ctx:     ctx,
		sess:    yttv.New(&cfg.Creds),
		cfg:     cfg,
		devCfg:  devCfg,
		devAddr: devAddr,
		guide:   newGuide(images),
		images:  images,
		toasts:  newToasts(),
		status:  "Loading channel guide…",
	}
}

func (m *model) Init() tea.Cmd {
	if m.devAddr != "" {
		m.device = cast.Device{Name: m.devAddr, Host: m.devAddr, Port: 8009}
		m.hasDev = true
	}
	return tea.Batch(
		loadEPG(m.ctx, m.sess),
		discoverDevices(m.ctx, 5*time.Second),
		splashTick(),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.guide.setSize(m.width, max(m.height-2, 1))
		return m, nil

	case epgLoadedMsg:
		m.guide.setChannels(msg.channels)
		m.status = fmt.Sprintf("Loaded %d channels", len(msg.channels))
		m.ready = true
		return m, m.kickoffLogos(msg.channels)

	case splashTickMsg:
		if m.ready {
			return m, nil
		}
		m.splashFrame++
		return m, splashTick()

	case imageLoadedMsg:
		return m, nil

	case devicesLoadedMsg:
		m.devices = msg.devices
		if m.picker != nil {
			m.picker.setDevices(msg.devices)
			if m.hasDev {
				m.picker.selectByName(m.device.Name)
			}
		}
		if !m.hasDev && m.devCfg != "" {
			want := strings.ToLower(m.devCfg)
			for _, d := range msg.devices {
				if strings.Contains(strings.ToLower(d.Name), want) {
					m.device = d
					m.hasDev = true
					m.status = "Device: " + d.Name
					break
				}
			}
		}
		return m, nil

	case castDoneMsg:
		m.status = fmt.Sprintf("Cast %q → %s", msg.channel, msg.device)
		return m, nil

	case stopDoneMsg:
		m.status = "Stopped " + msg.device
		return m, m.toasts.push(toastStop, "Stopped "+msg.device)

	case errMsg:
		m.status = "Error"
		if m.picker != nil {
			m.picker.setError(msg.err)
		}
		return m, m.toasts.push(toastError, msg.err.Error())

	case toastExpireMsg:
		m.toasts.expire(msg.id)
		return m, nil

	case deviceChosenMsg:
		m.device = msg.dev
		m.hasDev = true
		m.pickerOpen = false
		m.picker = nil
		m.status = "Device: " + msg.dev.Name
		return m, m.toasts.push(toastInfo, "Selected "+msg.dev.Name)

	case devicePickerClosedMsg:
		m.pickerOpen = false
		m.picker = nil
		return m, nil

	case detailCloseMsg:
		m.detailOpen = false
		m.detail = nil
		return m, nil

	case searchClosedMsg:
		m.searchOpen = false
		m.search = nil
		return m, nil
	}

	if m.searchOpen && m.search != nil {
		if cmd, captured := m.search.update(msg); captured {
			return m, cmd
		}
	}
	if m.pickerOpen && m.picker != nil {
		if cmd, captured := m.picker.update(msg); captured {
			return m, cmd
		}
	}
	if m.detailOpen && m.detail != nil {
		if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "c" {
			return m, m.castFromDetail()
		}
		if cmd, captured := m.detail.update(msg); captured {
			return m, cmd
		}
	}

	if k, ok := msg.(tea.KeyPressMsg); ok {
		return m.handleKey(k)
	}
	return m, nil
}

func (m *model) handleKey(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		m.guide.moveRow(1)
	case "k", "up":
		m.guide.moveRow(-1)
	case "h", "left":
		m.guide.moveCol(-1)
	case "l", "right":
		m.guide.moveCol(1)
	case "r":
		m.status = "Refreshing…"
		return m, loadEPG(m.ctx, m.sess)
	case "d":
		return m, m.openPicker()
	case "D":
		if !m.hasDev {
			return m, nil
		}
		prev := m.device.Name
		m.device = cast.Device{}
		m.hasDev = false
		m.status = "No device"
		return m, m.toasts.push(toastInfo, "Cleared device ("+prev+")")
	case "/":
		if len(m.guide.channels) > 0 {
			m.search = newSearcher(m.guide)
			m.searchOpen = true
		}
		return m, nil
	case "enter":
		return m, m.openDetail()
	case "c":
		if !m.hasDev {
			return m, m.toasts.push(toastError, "No device selected — press d to choose one")
		}
		return m, m.castRowLive()
	case "s":
		if !m.hasDev {
			return m, m.toasts.push(toastError, "No device selected — press d to choose one")
		}
		m.status = "Stopping " + m.device.Name + "…"
		return m, stopCast(m.ctx, m.device)
	}
	return m, nil
}

func (m *model) openPicker() tea.Cmd {
	m.pickerOpen = true
	m.picker = newDevicePicker()
	if len(m.devices) > 0 {
		m.picker.setDevices(m.devices)
		if m.hasDev {
			m.picker.selectByName(m.device.Name)
		}
	}
	return discoverDevices(m.ctx, 5*time.Second)
}

func (m *model) openDetail() tea.Cmd {
	a := m.guide.selectedAiring()
	if a == nil {
		return nil
	}
	ch := m.guide.selectedChannel()
	chName := ""
	if ch != nil {
		chName = ch.Name
	}
	m.detail = newDetail(chName, *a, m.images)
	m.detailOpen = true
	return m.detail.fetchThumb(m.ctx)
}

func (m *model) castRowLive() tea.Cmd {
	ch := m.guide.selectedChannel()
	if ch == nil {
		return nil
	}
	m.status = fmt.Sprintf("Casting %q to %s…", ch.Name, m.device.Name)
	return tea.Batch(
		m.toasts.push(toastCast, fmt.Sprintf("Casting %s on %s", ch.Name, m.device.Name)),
		castChannel(m.ctx, m.sess, m.device, ch.Name),
	)
}

func (m *model) castFromDetail() tea.Cmd {
	if m.detail == nil {
		return nil
	}
	a := m.detail.airing
	if !a.IsLive {
		return m.toasts.push(toastError, "Cannot cast a future airing")
	}
	if !m.hasDev {
		return m.toasts.push(toastError, "No device selected — press d to choose one")
	}
	return m.castRowLive()
}

func (m *model) kickoffLogos(chs []epg.Channel) tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(chs))
	for _, ch := range chs {
		if c := m.images.fetchLogo(m.ctx, ch.Name, ch.StationIconURL); c != nil {
			cmds = append(cmds, c)
		}
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m *model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		return tea.View{Content: "Initializing…", AltScreen: true}
	}

	if !m.ready {
		base := renderSplash(m.width, m.height, m.status, m.splashFrame)
		layers := []*lipgloss.Layer{lipgloss.NewLayer(base).Z(0)}
		if tl := m.toasts.layer(m.width, m.height); tl != nil {
			layers = append(layers, tl)
		}
		if len(layers) == 1 {
			return tea.View{Content: base, AltScreen: true}
		}
		canvas := lipgloss.NewCanvas(m.width, m.height)
		canvas = canvas.Compose(lipgloss.NewCompositor(layers...))
		return tea.View{Content: canvas.Render(), AltScreen: true}
	}

	header := m.renderHeader()
	footer := m.renderFooter()
	body := clipToHeight(m.guide.view(), max(m.height-2, 1))
	base := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	layers := []*lipgloss.Layer{lipgloss.NewLayer(base).Z(0)}
	if m.detailOpen && m.detail != nil {
		layers = append(layers, m.detail.layer(m.width, m.height))
	} else if m.pickerOpen && m.picker != nil {
		layers = append(layers, m.picker.layer(m.width, m.height))
	}
	if m.searchOpen && m.search != nil {
		layers = append(layers, m.search.layer(m.width, m.height))
	}
	if tl := m.toasts.layer(m.width, m.height); tl != nil {
		layers = append(layers, tl)
	}

	if len(layers) == 1 {
		return tea.View{Content: base, AltScreen: true}
	}
	canvas := lipgloss.NewCanvas(m.width, m.height)
	canvas = canvas.Compose(lipgloss.NewCompositor(layers...))
	return tea.View{Content: canvas.Render(), AltScreen: true}
}

func (m *model) renderHeader() string {
	left := "yttvctl"
	if m.hasDev {
		left += "  ●  " + m.device.Name
	} else {
		left += "  ○  no device"
	}
	right := time.Now().Format("Mon 3:04 PM")
	return lipgloss.NewStyle().
		Width(m.width).
		Background(colorPanel).
		Foreground(colorText).
		Padding(0, 1).
		Render(layout3(m.width-2, left, m.status, right))
}

func (m *model) renderFooter() string {
	dev := "[d] device"
	if m.hasDev {
		dev = "[d] device  [D] clear"
	}
	hints := "[hjkl] move  [/] search  [enter] detail  [c] cast  [s] stop  " + dev + "  [r] refresh  [q] quit"
	return styleFooter.Width(m.width).Render(truncate(hints, m.width-2))
}

func layout3(width int, left, mid, right string) string {
	l := lipgloss.Width(left)
	r := lipgloss.Width(right)
	room := max(width-l-r-2, 0)
	mid = truncate(mid, room)
	pad := max(room-lipgloss.Width(mid), 0)
	return left + "  " + mid + strings.Repeat(" ", pad) + right
}

func clipToHeight(s string, h int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
