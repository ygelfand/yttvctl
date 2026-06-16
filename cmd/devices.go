package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/ygelfand/lib-yttv/discover"
)

func devicesCmd() *cobra.Command {
	var timeout time.Duration
	c := &cobra.Command{
		Use:   "devices",
		Short: "Discover Chromecasts on the local network via mDNS",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			devs, err := discover.Discover(ctx, timeout)
			if err != nil {
				return err
			}
			for _, d := range devs {
				fmt.Printf("%-30s  %s:%d\n", d.Name, d.Host, d.Port)
			}
			return nil
		},
	}
	c.Flags().DurationVar(&timeout, "timeout", 5*time.Second, "discovery window")
	return c
}
