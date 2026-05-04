package auth

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
)

// RequireSession is HTTP middleware that rejects requests without a valid session cookie.
// On success it stores the session in the request context so handlers can read it.
func (m *Manager) RequireSession(gdb *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, err := m.FromRequest(r)
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "session required")
				return
			}

			// Ensure the user still exists in DB (handles deleted/disabled users).
			var n int64
			if err := gdb.Model(&models.User{}).Where("id = ?", s.UserID).Count(&n).Error; err != nil {
				writeAuthError(w, http.StatusInternalServerError, "auth lookup failed")
				return
			}
			if n == 0 {
				writeAuthError(w, http.StatusUnauthorized, "user not found")
				return
			}

			ctx := WithSession(r.Context(), s)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SecurityHeaders adds the same set of headers that api-server.js writes
// (X-XSS-Protection, X-Content-Type-Options, X-Frame-Options) and a strict
// CSP for API responses.
func SecurityHeaders(strictCSP bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-XSS-Protection", "1; mode=block")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "deny")
			if strictCSP {
				h.Set("Content-Security-Policy", "default-src 'none'")
			} else {
				h.Set("Content-Security-Policy",
					"default-src 'none'; "+
						"script-src 'self' 'unsafe-eval'; "+
						"style-src 'self' 'unsafe-inline'; "+
						"img-src 'self' data:; "+
						"font-src 'self' data:; "+
						"connect-src 'self'")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORS adds permissive CORS headers; mirrors api-server.js behavior of
// allowing credentialed requests from any origin (not strictly safe but
// matches existing GUI deployment).
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Credentials", "true")
		h.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"success":false,"error":"` + msg + `"}`))
}
