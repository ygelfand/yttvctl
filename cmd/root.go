// Package cmd holds the cobra command tree for yttvctl.
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	configPath string
	deviceName string
	deviceAddr string
	logLevel   string
)

// BuildInfo carries ldflags-injected build metadata from main.
type BuildInfo struct {
	Version   string
	GitCommit string
	BuildDate string
}

func Root(bi BuildInfo) *cobra.Command {
	root := &cobra.Command{
		Use:   "yttvctl",
		Short: "YouTube TV remote — browse the channel guide and cast to Chromecast",
		Long: `yttvctl is a remote for YouTube TV. It lists the live channel
guide and casts channels to Chromecast receivers on your network.

Auth: copy SAPISID and __Secure-3PSID cookies from tv.youtube.com.
Provide them via YAML (~/.config/yttvctl/config.yaml: sapisid,
secure_3psid) or env (YTTV_SAPISID, YTTV_SECURE_3PSID). Either is
sufficient; env overrides config. google_account_id is auto-discovered
each run unless you set it in config or YTTV_GOOGLE_ACCOUNT_ID.`,
		Version:      fmt.Sprintf("%s (commit %s, built %s)", bi.Version, bi.GitCommit, bi.BuildDate),
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return setupLogger(logLevel)
		},
	}
	root.PersistentFlags().StringVar(&configPath, "config", os.Getenv("YTTV_CONFIG"), "path to YAML config (defaults to $XDG_CONFIG_HOME/yttvctl/config.yaml)")
	root.PersistentFlags().StringVar(&deviceName, "device", os.Getenv("YTTV_DEVICE"), "Chromecast friendly-name substring")
	root.PersistentFlags().StringVar(&deviceAddr, "addr", os.Getenv("YTTV_ADDR"), "Chromecast host:port (skips mDNS discovery)")
	root.PersistentFlags().StringVar(&logLevel, "log-level", envOr("YTTV_LOG_LEVEL", "info"), "log level (debug, info, warn, error)")
	root.AddCommand(channelsCmd(), castCmd(), stopCmd(), statusCmd(), devicesCmd(), tuiCmd())
	return root
}

func setupLogger(level string) error {
	var l slog.Level
	if err := l.UnmarshalText([]byte(strings.ToUpper(level))); err != nil {
		return fmt.Errorf("bad --log-level %q: %w", level, err)
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})
	slog.SetDefault(slog.New(h))
	return nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
