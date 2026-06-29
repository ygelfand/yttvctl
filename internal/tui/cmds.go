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

// deviceEventMsg carries one discovery event (device up/down). ok=false means
// the watch channel closed (context cancelled).
type deviceEventMsg struct {
	ev discover.Event
	ok bool
}

// statusMsg carries a pushed status update for a video device. ok=false means
// that device's listener stopped.
type statusMsg struct {
	id     string
	status *cast.Status
	ch     <-chan *cast.Status
	ok     bool
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

// waitDevice blocks for the next discovery event on ch. Re-issued after each
// event to keep the continuous watch flowing through the bubbletea loop.
func waitDevice(ch <-chan discover.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		return deviceEventMsg{ev: ev, ok: ok}
	}
}

// waitStatus blocks for the next pushed status from a device's listener.
func waitStatus(id string, ch <-chan *cast.Status) tea.Cmd {
	return func() tea.Msg {
		st, ok := <-ch
		return statusMsg{id: id, status: st, ch: ch, ok: ok}
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
