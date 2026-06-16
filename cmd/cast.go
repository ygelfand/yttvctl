package cmd

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	yttv "github.com/ygelfand/lib-yttv"
	"github.com/ygelfand/lib-yttv/cast"
	"github.com/ygelfand/lib-yttv/discover"
	"github.com/ygelfand/yttvctl/internal/config"
)

func castCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cast <channel>",
		Short: "Cast a channel by name to the configured Chromecast",
		Args:  cobra.ExactArgs(1),
		RunE:  runCast,
	}
}

func runCast(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := cfg.EnsureGoogleAccountID(cmd.Context(), nil); err != nil {
		return fmt.Errorf("discover google_account_id: %w", err)
	}
	dev, err := resolveDevice(cmd.Context(), cfg)
	if err != nil {
		return err
	}
	return yttv.New(&cfg.Creds).Cast(cmd.Context(), dev, args[0])
}

// resolveDevice returns the Cast endpoint to use. --addr wins outright;
// otherwise mDNS-discover and substring-match on friendly name from
// --device/$YTTV_DEVICE/config device.
func resolveDevice(ctx context.Context, cfg *config.Config) (cast.Device, error) {
	if deviceAddr != "" {
		return parseAddr(deviceAddr)
	}
	want := deviceName
	if want == "" {
		want = cfg.Device
	}
	if want == "" {
		return cast.Device{}, fmt.Errorf("no device (set --addr, --device, $YTTV_DEVICE, or `device:` in config)")
	}
	dctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	devs, err := discover.Discover(dctx, 5*time.Second)
	if err != nil {
		return cast.Device{}, err
	}
	wantLower := strings.ToLower(want)
	for _, d := range devs {
		if strings.Contains(strings.ToLower(d.Name), wantLower) {
			return d, nil
		}
	}
	return cast.Device{}, fmt.Errorf("device %q not found among %d discovered", want, len(devs))
}

func parseAddr(addr string) (cast.Device, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		// Bare hostname / IP without :port → assume the Cast default 8009.
		return cast.Device{Name: addr, Host: addr, Port: 8009}, nil
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return cast.Device{}, fmt.Errorf("bad port in %q: %w", addr, err)
	}
	return cast.Device{Name: addr, Host: host, Port: port}, nil
}
