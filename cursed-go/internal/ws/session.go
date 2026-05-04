package ws

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

// Session represents a single connected bot.
type Session struct {
	BrowserID string
	BotID     uuid.UUID
	conn      *websocket.Conn

	mu      sync.Mutex
	closed  bool
	pending map[string]*pendingRPC
}

// NewSession wraps a websocket connection with the per-session
// bookkeeping needed for RPC dispatch.
func NewSession(conn *websocket.Conn) *Session {
	return &Session{
		conn:    conn,
		pending: make(map[string]*pendingRPC),
	}
}

// SendJSON writes a json-serialized envelope to the bot.
func (s *Session) SendJSON(ctx context.Context, env Envelope) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return errors.New("session closed")
	}
	s.mu.Unlock()
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return s.conn.Write(ctx, websocket.MessageText, b)
}

// Call sends an RPC request and waits for the bot's reply (matched by
// envelope.id). Implements the server->bot half of REQUEST_TABLE in server.js.
func (s *Session) Call(ctx context.Context, action string, data any) (Envelope, error) {
	id := uuid.NewString()
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return Envelope{}, err
		}
		raw = b
	}
	env := Envelope{ID: id, Action: action, Data: raw}

	reply := make(chan Envelope, 1)
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return Envelope{}, errors.New("session closed")
	}
	s.pending[id] = &pendingRPC{id: id, reply: reply, deadline: time.Now().Add(time.Minute)}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.pending, id)
		s.mu.Unlock()
	}()

	if err := s.SendJSON(ctx, env); err != nil {
		return Envelope{}, err
	}
	select {
	case r := <-reply:
		return r, nil
	case <-ctx.Done():
		return Envelope{}, ctx.Err()
	}
}

// resolvePending matches an inbound envelope to a pending RPC.
// Returns true if a waiting caller was notified.
func (s *Session) resolvePending(env Envelope) bool {
	s.mu.Lock()
	p, ok := s.pending[env.ID]
	s.mu.Unlock()
	if !ok {
		return false
	}
	select {
	case p.reply <- env:
	default:
	}
	return true
}

func (s *Session) close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	for _, p := range s.pending {
		select {
		case p.reply <- Envelope{Action: "ABORT"}:
		default:
		}
	}
	s.pending = nil
	s.mu.Unlock()
	return s.conn.Close(websocket.StatusNormalClosure, "bye")
}

// Close shuts down the underlying connection.
func (s *Session) Close() error { return s.close() }
