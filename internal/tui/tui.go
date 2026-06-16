// Package tui is the YouTube-TV-style live guide for yttvctl.
package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/ygelfand/yttvctl/internal/config"
)

func Run(ctx context.Context, cfg *config.Config, devCfg, devAddr string) error {
	m := newModel(ctx, cfg, devCfg, devAddr)
	p := tea.NewProgram(m, tea.WithContext(ctx))
	_, err := p.Run()
	return err
}
