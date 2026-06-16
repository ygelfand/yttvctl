package tui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	yttv "github.com/ygelfand/lib-yttv"
	"github.com/ygelfand/lib-yttv/cast"
	"github.com/ygelfand/lib-yttv/discover"
	"github.com/ygelfand/lib-yttv/epg"
)

type epgLoadedMsg struct {
	channels []epg.Channel
}

type devicesLoadedMsg struct {
	devices []cast.Device
}

type castDoneMsg struct {
	channel string
	device  string
}

type stopDoneMsg struct {
	device string
}

type errMsg struct{ err error }

func loadEPG(ctx context.Context, sess *yttv.Session) tea.Cmd {
	return func() tea.Msg {
		ch, err := sess.Channels(ctx)
		if err != nil {
			return errMsg{fmt.Errorf("load EPG: %w", err)}
		}
		return epgLoadedMsg{channels: ch}
	}
}

func discoverDevices(parent context.Context, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()
		devs, err := discover.Discover(ctx, timeout)
		if err != nil {
			return errMsg{fmt.Errorf("discover devices: %w", err)}
		}
		return devicesLoadedMsg{devices: devs}
	}
}

func castChannel(parent context.Context, sess *yttv.Session, dev cast.Device, channelName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(parent, 60*time.Second)
		defer cancel()
		if err := sess.Cast(ctx, dev, channelName); err != nil {
			return errMsg{fmt.Errorf("cast %q to %s: %w", channelName, dev.Name, err)}
		}
		return castDoneMsg{channel: channelName, device: dev.Name}
	}
}

func stopCast(parent context.Context, dev cast.Device) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(parent, 15*time.Second)
		defer cancel()
		recv, err := cast.Connect(ctx, dev)
		if err != nil {
			return errMsg{fmt.Errorf("stop: connect %s: %w", dev.Name, err)}
		}
		defer recv.Close()
		if err := recv.Stop(ctx); err != nil {
			return errMsg{fmt.Errorf("stop %s: %w", dev.Name, err)}
		}
		return stopDoneMsg{device: dev.Name}
	}
}
