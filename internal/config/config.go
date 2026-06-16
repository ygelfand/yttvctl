package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/viper"
	"github.com/ygelfand/lib-yttv/auth"
)

type Config struct {
	Device     string `mapstructure:"device"`
	auth.Creds `mapstructure:",squash"`

	v *viper.Viper
}

// Load resolves the config via viper. Values can come from any of:
// the --config flag, $YTTV_CONFIG, $XDG_CONFIG_HOME/yttvctl/config.yaml,
// $HOME/.config/yttvctl/config.yaml, or YTTV_* env vars. A config file is
// optional — env vars alone are sufficient.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("YTTV")
	v.AutomaticEnv()
	// AutomaticEnv only binds keys viper already knows about; register the
	// expected keys so YTTV_* works without a config file present.
	for _, k := range []string{"device", "sapisid", "secure_3psid", "google_account_id"} {
		_ = v.BindEnv(k)
	}

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath("$XDG_CONFIG_HOME/yttvctl")
		v.AddConfigPath("$HOME/.config/yttvctl")
	}
	if err := v.ReadInConfig(); err != nil {
		var nf viper.ConfigFileNotFoundError
		if !errors.As(err, &nf) {
			return nil, err
		}
	}
	c := &Config{v: v}
	if err := v.Unmarshal(c); err != nil {
		return nil, fmt.Errorf("decode %s: %w", v.ConfigFileUsed(), err)
	}
	return c, nil
}

func (c *Config) Path() string { return c.v.ConfigFileUsed() }

func (c *Config) EnsureGoogleAccountID(ctx context.Context, hc *http.Client) error {
	if c.GoogleAccountID != "" {
		return nil
	}
	gid, err := c.DiscoverGoogleAccountID(ctx, hc)
	if err != nil {
		return err
	}
	c.GoogleAccountID = gid
	return nil
}
