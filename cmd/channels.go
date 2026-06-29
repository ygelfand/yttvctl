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
	return &cobra.Command{
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
			channels, err := epg.Fetch(ctx, innertube.New(&cfg.Creds))
			if err != nil {
				return err
			}
			for _, ch := range channels {
				fmt.Printf("%-20s  %s  %s\n", ch.Name, ch.LiveVideoID, ch.CurrentTitle)
			}
			return nil
		},
	}
}
