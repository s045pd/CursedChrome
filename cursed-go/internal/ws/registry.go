package ws

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Registry tracks live bot WebSocket sessions. Lookups are by browser_id
// (string used by extensions) or by UUID (used by API). It also owns the
// pending-RPC table — message_id -> reply channel — replacing the
// REQUEST_TABLE NodeCache from server.js.
type Registry struct {
	mu       sync.RWMutex
	byBrowser map[string]*Session
	byBot    map[uuid.UUID]*Session
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		byBrowser: make(map[string]*Session),
		byBot:    make(map[uuid.UUID]*Session),
	}
}

// Register inserts a session. If a previous session existed for the same
// browser_id it is closed.
func (r *Registry) Register(s *Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, ok := r.byBrowser[s.BrowserID]; ok && old != s {
		_ = old.close()
	}
	r.byBrowser[s.BrowserID] = s
	if s.BotID != uuid.Nil {
		r.byBot[s.BotID] = s
	}
}

// Unregister removes a session if and only if it is still the current one.
func (r *Registry) Unregister(s *Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cur, ok := r.byBrowser[s.BrowserID]; ok && cur == s {
		delete(r.byBrowser, s.BrowserID)
	}
	if cur, ok := r.byBot[s.BotID]; ok && cur == s {
		delete(r.byBot, s.BotID)
	}
}

// ByBrowserID returns the session for a browser_id, or nil.
func (r *Registry) ByBrowserID(id string) *Session {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byBrowser[id]
}

// ByBotID returns the session for a bot UUID, or nil.
func (r *Registry) ByBotID(id uuid.UUID) *Session {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.byBot[id]
}

// Count returns the number of live sessions.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byBrowser)
}

// pendingRPC tracks a single in-flight server->bot RPC call.
type pendingRPC struct {
	id       string
	deadline time.Time
	reply    chan Envelope
}
