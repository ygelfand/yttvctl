package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/ygelfand/lib-yttv/cast"
	"github.com/ygelfand/yttvctl/internal/config"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show what's playing on the configured Chromecast",
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
			st, err := cast.GetDeviceStatus(ctx, target)
			if err != nil {
				return err
			}

			if st.Idle || st.AppName == "" {
				fmt.Printf("%s: idle\n", target.Name)
				return nil
			}
			fmt.Printf("%s: %s", target.Name, st.AppName)
			if st.Media != nil {
				m := st.Media
				fmt.Printf(" — %s", m.PlayerState)
				if m.Title != "" {
					fmt.Printf(" — %s", m.Title)
				}
				if m.Subtitle != "" {
					fmt.Printf(" (%s)", m.Subtitle)
				}
			}
			fmt.Printf("  [vol %d%%%s]\n", int(st.Volume*100), mutedSuffix(st.Muted))
			return nil
		},
	}
}

func mutedSuffix(muted bool) string {
	if muted {
		return " muted"
	}
	return ""
}
