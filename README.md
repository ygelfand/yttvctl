# yttvctl

CLI + TUI for browsing the YouTube TV channel guide and casting to Chromecast devices

![yttvctl demo](assets/demo.gif)

## Auth

Sign into `tv.youtube.com` in a Chromium-based browser. In DevTools (Cmd+Opt+I), open **Application → Storage → Cookies → `https://tv.youtube.com`** and copy the values of:

- `SAPISID`
- `__Secure-3PSID`

Provide them via either a YAML file at `$XDG_CONFIG_HOME/yttvctl/config.yaml` (defaults to `~/.config/yttvctl/config.yaml`):

```yaml
device: "Living Room TV" # optional default; override with --device or --addr
sapisid: "<paste here>"
secure_3psid: "<paste here>"
```

…or via env vars: `YTTV_SAPISID`, `YTTV_SECURE_3PSID`, `YTTV_DEVICE`, `YTTV_GOOGLE_ACCOUNT_ID`. Env overrides config; either source alone is sufficient.

`google_account_id` is auto-discovered each run by fetching `tv.youtube.com` with these cookies. Set it in your config or export `YTTV_GOOGLE_ACCOUNT_ID` to skip the per-run lookup.

## Commands

```
yttvctl channels                                  list channels currently airing
yttvctl cast <channel>                            cast a channel to the configured device
yttvctl stop                                      stop whatever is running on the device
yttvctl devices                                   mDNS-discover Chromecasts on the LAN
yttvctl tui                                       interactive terminal ui
```

## Global flags

```
--config <path>     YAML config (default $XDG_CONFIG_HOME/yttvctl/config.yaml)
--device <name>     friendly-name substring (or $YTTV_DEVICE)
--addr host:port    skip mDNS, talk to a specific Cast endpoint (or $YTTV_ADDR)
```

## Build / install

```sh
make build          # → ./bin/yttvctl
make install        # go install into $GOPATH/bin
```
