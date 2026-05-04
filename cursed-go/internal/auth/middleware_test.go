package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestRequireSession_NoCookie(t *testing.T) {
	gdb := newTestDB(t)
	mgr, _ := NewManager(testSecret)
	h := mgr.RequireSession(gdb)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, r)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestRequireSession_ValidUser(t *testing.T) {
	gdb := newTestDB(t)
	u := models.User{Username: "alice", Password: "x"}
	u.ID = uuid.New()
	if err := gdb.Create(&u).Error; err != nil {
		t.Fatal(err)
	}

	mgr, _ := NewManager(testSecret)
	rr := httptest.NewRecorder()
	if err := mgr.Issue(rr, u.ID); err != nil {
		t.Fatal(err)
	}
	cookies := rr.Result().Cookies()

	called := false
	h := mgr.RequireSession(gdb)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		s, ok := SessionFromContext(r.Context())
		if !ok {
			t.Error("session not in context")
		}
		if s.UserID != u.ID {
			t.Error("UserID mismatch in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rr2 := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		r.AddCookie(c)
	}
	h.ServeHTTP(rr2, r)
	if !called {
		t.Error("downstream handler not called")
	}
}

func TestRequireSession_DeletedUser(t *testing.T) {
	gdb := newTestDB(t)
	mgr, _ := NewManager(testSecret)
	rr := httptest.NewRecorder()
	// Issue a session for a non-existent user
	if err := mgr.Issue(rr, uuid.New()); err != nil {
		t.Fatal(err)
	}
	cookies := rr.Result().Cookies()

	h := mgr.RequireSession(gdb)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr2 := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		r.AddCookie(c)
	}
	h.ServeHTTP(rr2, r)
	if rr2.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 for deleted user", rr2.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	SecurityHeaders(false)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, r)

	wantHeaders := map[string]string{
		"X-XSS-Protection":        "1; mode=block",
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "deny",
		"Content-Security-Policy": "",
	}
	for k, v := range wantHeaders {
		got := rr.Header().Get(k)
		if got == "" {
			t.Errorf("header %s missing", k)
		}
		if v != "" && got != v {
			t.Errorf("header %s = %q, want %q", k, got, v)
		}
	}
}

func TestSecurityHeaders_StrictCSP(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/anything", nil)
	SecurityHeaders(true)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, r)
	if rr.Header().Get("Content-Security-Policy") != "default-src 'none'" {
		t.Errorf("strict CSP = %q", rr.Header().Get("Content-Security-Policy"))
	}
}

func TestCORS_Preflight(t *testing.T) {
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Set("Origin", "http://example.com")
	called := false
	CORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
	})).ServeHTTP(rr, r)
	if called {
		t.Error("preflight should short-circuit")
	}
	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Error("CORS origin echo missing")
	}
}
