package tui

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	gopixels "github.com/saran13raj/go-pixels"
)

// Halfcell renders 1 image pixel per terminal column and 2 per row.
const (
	logoPxWidth  = 10
	logoPxHeight = 8
)

type imageLoadedMsg struct {
	key string
}

type imageCache struct {
	mu    sync.Mutex
	views map[string]string // cache key → rendered halfblock string
	hc    *http.Client
}

func newImageCache() *imageCache {
	return &imageCache{
		views: make(map[string]string),
		hc:    &http.Client{Timeout: 10 * time.Second},
	}
}

func cacheKey(id string, w, h int) string {
	return fmt.Sprintf("%s|%d|%d", id, w, h)
}

func (c *imageCache) get(id string, w, h int) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.views[cacheKey(id, w, h)]
	return v, ok
}

func (c *imageCache) set(id string, w, h int, view string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.views[cacheKey(id, w, h)] = view
}

func (c *imageCache) fetch(ctx context.Context, id, url string, w, h int) tea.Cmd {
	if url == "" || id == "" {
		return nil
	}
	if _, ok := c.get(id, w, h); ok {
		return nil
	}
	full := url
	if strings.HasPrefix(full, "//") {
		full = "https:" + full
	}
	return func() tea.Msg {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, full, nil)
		if err != nil {
			return nil
		}
		resp, err := c.hc.Do(req)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil
		}
		img, _, err := image.Decode(bytes.NewReader(body))
		if err != nil {
			return nil
		}
		out, err := gopixels.FromImageStream(img, w, h, "halfcell", true)
		if err != nil {
			return nil
		}
		c.set(id, w, h, out)
		return imageLoadedMsg{key: id}
	}
}

func (c *imageCache) fetchLogo(ctx context.Context, channelName, url string) tea.Cmd {
	return c.fetch(ctx, channelName, url, logoPxWidth, logoPxHeight)
}

func (c *imageCache) getLogo(channelName string) (string, bool) {
	return c.get(channelName, logoPxWidth, logoPxHeight)
}
