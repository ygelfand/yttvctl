package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/ygelfand/lib-yttv/epg"
	"github.com/ygelfand/lib-yttv/innertube"
	"github.com/ygelfand/yttvctl/internal/config"
)

func channelsCmd() *cobra.Command {
	var resolveLive bool
	c := &cobra.Command{
		Use:   "channels",
		Short: "List channels from the current EPG",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			if err := cfg.EnsureGoogleAccountID(ctx, nil); err != nil {
				return fmt.Errorf("discover google_account_id: %w", err)
			}
			ic := innertube.New(&cfg.Creds)
			channels, err := epg.Fetch(ctx, ic)
			if err != nil {
				return err
			}
			for _, ch := range channels {
				vid := ch.PerAiringVideoID
				if resolveLive {
					if live, err := epg.ResolveLiveVideoID(ctx, ic, ch.PerAiringVideoID); err == nil {
						vid = live
					}
				}
				fmt.Printf("%-20s  %s  %s\n", ch.Name, vid, ch.CurrentTitle)
			}
			return nil
		},
	}
	c.Flags().BoolVar(&resolveLive, "resolve-live", false, "resolve per-airing → live channel videoId (one extra /next call per channel)")
	return c
}
