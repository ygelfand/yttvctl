package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/ygelfand/yttvctl/internal/config"
	"github.com/ygelfand/yttvctl/internal/tui"
)

func tuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Interactive YouTube TV live guide + cast picker",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Silence info-level logs unless the caller opted in — they garble the altscreen.
			if !cmd.Flags().Changed("log-level") && os.Getenv("YTTV_LOG_LEVEL") == "" {
				slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			if err := cfg.EnsureGoogleAccountID(ctx, nil); err != nil {
				cancel()
				return fmt.Errorf("discover google_account_id: %w", err)
			}
			cancel()

			devCfg := deviceName
			if devCfg == "" {
				devCfg = cfg.Device
			}
			return tui.Run(cmd.Context(), cfg, devCfg, deviceAddr)
		},
	}
}
