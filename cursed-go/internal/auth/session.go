package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
)

// SessionCookieName matches the name used by api-server.js client-sessions.
const SessionCookieName = "session"

// SessionTTL mirrors the 7-day duration used by client-sessions.
const SessionTTL = 7 * 24 * time.Hour

// Session is the payload stored in the cookie. Keeping it small lets us
// avoid persisting sessions in the database — same model as Node version.
type Session struct {
	UserID    uuid.UUID `json:"u"`
	IssuedAt  int64     `json:"i"`
	ExpiresAt int64     `json:"e"`
}

// Manager signs and verifies session cookies using HMAC + symmetric encryption.
type Manager struct {
	sc *securecookie.SecureCookie
}

// NewManager creates a manager from a hex-encoded secret (>= 32 chars).
// Half of the secret is used for HMAC, the other half for AES.
func NewManager(secret string) (*Manager, error) {
	if len(secret) < 32 {
		return nil, errors.New("session secret too short (need >= 32 chars)")
	}
	hashKey := []byte(secret)[:32]
	blockKey := []byte(secret)
	if len(blockKey) >= 64 {
		blockKey = blockKey[32:64]
	} else {
		// pad if shorter than 64 by repeating
		pad := make([]byte, 32)
		copy(pad, []byte(secret))
		blockKey = pad
	}
	sc := securecookie.New(hashKey, blockKey)
	sc.MaxAge(int(SessionTTL.Seconds()))
	return &Manager{sc: sc}, nil
}

// Encode produces a signed+encrypted cookie value.
func (m *Manager) Encode(s Session) (string, error) {
	return m.sc.Encode(SessionCookieName, s)
}

// Decode parses and validates a cookie value. Returns an error if
// invalid, tampered, or expired.
func (m *Manager) Decode(value string) (Session, error) {
	var s Session
	if err := m.sc.Decode(SessionCookieName, value, &s); err != nil {
		return Session{}, err
	}
	if time.Now().Unix() > s.ExpiresAt {
		return Session{}, errors.New("session expired")
	}
	return s, nil
}

// Issue writes a fresh session cookie for the given user.
func (m *Manager) Issue(w http.ResponseWriter, userID uuid.UUID) error {
	now := time.Now()
	s := Session{
		UserID:    userID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(SessionTTL).Unix(),
	}
	enc, err := m.Encode(s)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    enc,
		Path:     "/",
		Expires:  now.Add(SessionTTL),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// Clear writes an immediately-expiring cookie to log a user out.
func (m *Manager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// FromRequest extracts a session from request cookies. Returns
// ErrNoSession if no cookie / invalid / expired.
var ErrNoSession = errors.New("no valid session")

func (m *Manager) FromRequest(r *http.Request) (Session, error) {
	c, err := r.Cookie(SessionCookieName)
	if err != nil {
		return Session{}, ErrNoSession
	}
	s, err := m.Decode(c.Value)
	if err != nil {
		return Session{}, ErrNoSession
	}
	return s, nil
}

// ctxKey is unexported to prevent external collisions.
type ctxKey int

const sessionCtxKey ctxKey = 1

// WithSession stores a session in context.
func WithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey, s)
}

// SessionFromContext retrieves the session previously stored.
func SessionFromContext(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionCtxKey).(Session)
	return s, ok
}
