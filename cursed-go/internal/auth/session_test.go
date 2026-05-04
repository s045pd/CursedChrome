package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestNewManager_ShortSecret(t *testing.T) {
	if _, err := NewManager("short"); err == nil {
		t.Error("expected error for short secret")
	}
}

func TestSession_RoundTrip(t *testing.T) {
	m, err := NewManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	uid := uuid.New()
	s := Session{UserID: uid, IssuedAt: time.Now().Unix(), ExpiresAt: time.Now().Add(time.Hour).Unix()}
	enc, err := m.Encode(s)
	if err != nil {
		t.Fatal(err)
	}
	out, err := m.Decode(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.UserID != uid {
		t.Errorf("UserID mismatch: %v vs %v", out.UserID, uid)
	}
}

func TestSession_Expired(t *testing.T) {
	m, _ := NewManager(testSecret)
	s := Session{UserID: uuid.New(), IssuedAt: 1, ExpiresAt: 2}
	enc, err := m.Encode(s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Decode(enc); err == nil {
		t.Error("expected expired error")
	}
}

func TestIssueAndFromRequest(t *testing.T) {
	m, _ := NewManager(testSecret)
	uid := uuid.New()

	rr := httptest.NewRecorder()
	if err := m.Issue(rr, uid); err != nil {
		t.Fatal(err)
	}

	cookie := rr.Result().Cookies()
	if len(cookie) == 0 {
		t.Fatal("no cookie set")
	}
	if cookie[0].Name != SessionCookieName {
		t.Errorf("cookie name = %s, want %s", cookie[0].Name, SessionCookieName)
	}
	if !cookie[0].HttpOnly {
		t.Error("cookie must be HttpOnly")
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(cookie[0])
	got, err := m.FromRequest(r)
	if err != nil {
		t.Fatalf("FromRequest: %v", err)
	}
	if got.UserID != uid {
		t.Errorf("UserID = %v, want %v", got.UserID, uid)
	}
}

func TestFromRequest_NoCookie(t *testing.T) {
	m, _ := NewManager(testSecret)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := m.FromRequest(r); err != ErrNoSession {
		t.Errorf("err = %v, want ErrNoSession", err)
	}
}

func TestFromRequest_TamperedCookie(t *testing.T) {
	m, _ := NewManager(testSecret)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "tampered"})
	if _, err := m.FromRequest(r); err != ErrNoSession {
		t.Errorf("err = %v, want ErrNoSession", err)
	}
}

func TestClear_SetsExpiredCookie(t *testing.T) {
	m, _ := NewManager(testSecret)
	rr := httptest.NewRecorder()
	m.Clear(rr)
	c := rr.Result().Cookies()
	if len(c) == 0 {
		t.Fatal("no cookie set")
	}
	if !strings.Contains(rr.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Errorf("expected Max-Age=0, got %s", rr.Header().Get("Set-Cookie"))
	}
}

func TestSessionContext_RoundTrip(t *testing.T) {
	s := Session{UserID: uuid.New()}
	ctx := WithSession(t.Context(), s)
	got, ok := SessionFromContext(ctx)
	if !ok {
		t.Fatal("session not found in context")
	}
	if got.UserID != s.UserID {
		t.Errorf("UserID mismatch")
	}
}
