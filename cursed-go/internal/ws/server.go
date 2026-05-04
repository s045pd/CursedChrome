package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"nhooyr.io/websocket"

	"github.com/s045pd/cursed-go/internal/api"
	"github.com/s045pd/cursed-go/internal/db/models"
)

// Server is the WebSocket bot endpoint.
type Server struct {
	db       *gorm.DB
	logger   *slog.Logger
	registry *Registry
}

// New creates a Server.
func New(db *gorm.DB, logger *slog.Logger) *Server {
	return &Server{db: db, logger: logger, registry: NewRegistry()}
}

// Registry exposes the underlying registry to the API layer (which uses
// it to satisfy api.BotRPC).
func (s *Server) Registry() *Registry { return s.registry }

// Handler returns an http.Handler that upgrades requests to WebSockets.
// Mount it at the root of the bot-facing port (default 4343).
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // origin checking handled at proxy level
			OriginPatterns:     []string{"*"},
		})
		if err != nil {
			return
		}
		// Larger frame for screenshots etc.
		conn.SetReadLimit(64 << 20) // 64 MiB

		ctx := r.Context()
		sess := NewSession(conn)
		defer func() {
			_ = sess.Close()
			s.markOffline(sess.BotID)
			s.registry.Unregister(sess)
		}()

		bot, err := s.authenticate(ctx, sess, s.logger)
		if err != nil {
			s.logger.Warn("ws auth failed", "err", err)
			return
		}
		sess.BrowserID = bot.BrowserID
		sess.BotID = bot.ID
		s.registry.Register(sess)
		s.logger.Info("ws connected", "browser", sess.BrowserID, "id", sess.BotID)

		s.readLoop(ctx, sess)
	})
}

func (s *Server) readLoop(ctx context.Context, sess *Session) {
	for {
		_, data, err := sess.conn.Read(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				s.logger.Info("ws read end", "err", err, "browser", sess.BrowserID)
			}
			return
		}
		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			s.logger.Warn("ws bad json", "err", err)
			continue
		}
		if err := s.dispatch(ctx, sess, env); err != nil {
			s.logger.Warn("ws dispatch error", "action", env.Action, "err", err)
		}
	}
}

func (s *Server) markOffline(id uuid.UUID) {
	if id == uuid.Nil {
		return
	}
	_ = s.db.Model(&models.Bot{}).Where("id = ?", id).Updates(map[string]any{
		"is_online":   false,
		"last_online": time.Now(),
	}).Error
}

// CallBot satisfies api.BotRPC.
func (s *Server) CallBot(ctx context.Context, browserID, action string, data map[string]any) (map[string]any, error) {
	sess := s.registry.ByBrowserID(browserID)
	if sess == nil {
		return nil, api.ErrBotOffline
	}
	resp, err := sess.Call(ctx, action, data)
	if err != nil {
		return nil, err
	}
	if resp.Action == "ABORT" {
		return nil, fmt.Errorf("rpc aborted")
	}
	out := map[string]any{}
	if len(resp.Data) > 0 {
		_ = json.Unmarshal(resp.Data, &out)
	}
	return out, nil
}

// IsBotOnline satisfies api.BotRPC.
func (s *Server) IsBotOnline(id uuid.UUID) bool {
	return s.registry.ByBotID(id) != nil
}
