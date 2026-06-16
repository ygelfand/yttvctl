package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"github.com/ygelfand/lib-yttv/cast"
	"github.com/ygelfand/yttvctl/internal/config"
)

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop whatever's running on the configured Chromecast",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			target, err := resolveDevice(ctx, cfg)
			if err != nil {
				return err
			}
			recv, err := cast.Connect(ctx, target)
			if err != nil {
				return err
			}
			defer recv.Close()
			return recv.Stop(ctx)
		},
	}
}
