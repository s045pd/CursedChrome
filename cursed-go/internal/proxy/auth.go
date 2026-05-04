// Package proxy implements the bot-routed HTTP forward proxy.
//
// The proxy listens on PROXY_PORT (default 8080). Clients authenticate
// via Proxy-Authorization Basic <base64(user:pass)>; user/pass map to
// a Bot row (proxy_username/proxy_password) or fall back to the global
// default proxy bot configured via /api/v1/settings/global-proxy.
//
// Each authenticated request is forwarded to the bot via a
// SEND_REQUEST_VIA_BROWSER RPC; the bot performs the actual fetch in
// the user's browser context and returns the response body (base64).
package proxy

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/api"
	"github.com/s045pd/cursed-go/internal/db"
	"github.com/s045pd/cursed-go/internal/db/models"
)

// authCache caches successful credential lookups for 10 minutes —
// matches the Redis cache in server.js.
type authCache struct {
	mu    sync.Mutex
	items map[string]authCacheEntry
	ttl   time.Duration
}

type authCacheEntry struct {
	bot     models.Bot
	expires time.Time
}

func newAuthCache() *authCache {
	return &authCache{items: make(map[string]authCacheEntry), ttl: 10 * time.Minute}
}

func (c *authCache) get(key string) (models.Bot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expires) {
		return models.Bot{}, false
	}
	return e.bot, true
}

func (c *authCache) put(key string, b models.Bot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = authCacheEntry{bot: b, expires: time.Now().Add(c.ttl)}
}

// extractCreds parses Proxy-Authorization Basic header.
func extractCreds(h http.Header) (string, string, bool) {
	v := h.Get("Proxy-Authorization")
	if v == "" {
		return "", "", false
	}
	if !strings.HasPrefix(v, "Basic ") {
		return "", "", false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(v, "Basic "))
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// authenticate looks up a bot by proxy credentials, falling back to the
// global default if no credentials are provided. Returns ErrUnauthorized
// when no bot matches.
func authenticate(gdb *gorm.DB, cache *authCache, h http.Header) (*models.Bot, error) {
	if user, pass, ok := extractCreds(h); ok {
		key := user + ":" + pass
		if b, ok := cache.get(key); ok {
			return &b, nil
		}
		var b models.Bot
		err := gdb.Where("proxy_username = ? AND proxy_password = ?", user, pass).First(&b).Error
		if err != nil {
			return nil, ErrUnauthorized
		}
		cache.put(key, b)
		return &b, nil
	}
	// fallback: global default
	v, err := db.GetSetting(gdb, api.SettingGlobalProxyBot)
	if err != nil || v == "" {
		return nil, ErrUnauthorized
	}
	var b models.Bot
	if err := gdb.Where("id = ?", v).First(&b).Error; err != nil {
		return nil, ErrUnauthorized
	}
	return &b, nil
}

// ErrUnauthorized is returned when proxy credentials are missing or wrong.
var ErrUnauthorized = errors.New("proxy unauthorized")
