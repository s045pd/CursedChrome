package proxy

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/api"
	"github.com/s045pd/cursed-go/internal/db/models"
)

type fakeRPC struct {
	mu    sync.Mutex
	calls int
	resp  map[string]any
	err   error
}

func (f *fakeRPC) CallBot(_ context.Context, _, _ string, _ map[string]any) (map[string]any, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.resp, f.err
}
func (f *fakeRPC) IsBotOnline(_ uuid.UUID) bool { return true }

func setupProxy(t *testing.T) (*Server, *gorm.DB, *fakeRPC) {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AutoMigrate(models.All()...); err != nil {
		t.Fatal(err)
	}
	r := &fakeRPC{resp: map[string]any{
		"status":  float64(200),
		"headers": map[string]any{"X-Test": "yes"},
		"body":    base64.StdEncoding.EncodeToString([]byte("hello")),
	}}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return New(g, r, logger), g, r
}

func TestProxy_NoCredentials(t *testing.T) {
	srv, _, _ := setupProxy(t)
	r := httptest.NewRequest(http.MethodGet, "http://example.com/x", nil)
	w := httptest.NewRecorder()
	srv.serve(w, r)
	if w.Code != http.StatusProxyAuthRequired {
		t.Errorf("status=%d, want 407", w.Code)
	}
	if w.Header().Get("Proxy-Authenticate") == "" {
		t.Error("missing Proxy-Authenticate header")
	}
}

func TestProxy_BadCredentials(t *testing.T) {
	srv, _, _ := setupProxy(t)
	r := httptest.NewRequest(http.MethodGet, "http://example.com/x", nil)
	r.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("nope:nope")))
	w := httptest.NewRecorder()
	srv.serve(w, r)
	if w.Code != http.StatusProxyAuthRequired {
		t.Errorf("status=%d, want 407", w.Code)
	}
}

func TestProxy_GoodCredentialsForwards(t *testing.T) {
	srv, gdb, rpc := setupProxy(t)
	gdb.Create(&models.Bot{
		BrowserID: "br1", ProxyUsername: "u1", ProxyPassword: "p1", LastOnline: time.Now(),
	})
	r := httptest.NewRequest(http.MethodGet, "http://example.com/x", nil)
	r.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("u1:p1")))
	w := httptest.NewRecorder()
	srv.serve(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	if w.Body.String() != "hello" {
		t.Errorf("body=%q", w.Body.String())
	}
	if w.Header().Get("X-Test") != "yes" {
		t.Errorf("missing forwarded header")
	}
	if rpc.calls != 1 {
		t.Errorf("RPC calls=%d", rpc.calls)
	}
}

func TestProxy_BotOffline(t *testing.T) {
	srv, gdb, rpc := setupProxy(t)
	rpc.err = api.ErrBotOffline
	gdb.Create(&models.Bot{BrowserID: "br1", ProxyUsername: "u", ProxyPassword: "p", LastOnline: time.Now()})

	r := httptest.NewRequest(http.MethodGet, "http://example.com/x", nil)
	r.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("u:p")))
	w := httptest.NewRecorder()
	srv.serve(w, r)
	if w.Code != http.StatusBadGateway {
		t.Errorf("status=%d, want 502", w.Code)
	}
}

func TestExtractCreds(t *testing.T) {
	for _, tc := range []struct {
		header string
		user   string
		pass   string
		ok     bool
	}{
		{"", "", "", false},
		{"Bearer xxx", "", "", false},
		{"Basic " + base64.StdEncoding.EncodeToString([]byte("a:b")), "a", "b", true},
		{"Basic " + base64.StdEncoding.EncodeToString([]byte("noColon")), "", "", false},
	} {
		h := http.Header{}
		if tc.header != "" {
			h.Set("Proxy-Authorization", tc.header)
		}
		u, p, ok := extractCreds(h)
		if ok != tc.ok || u != tc.user || p != tc.pass {
			t.Errorf("header=%q got %s/%s ok=%v want %s/%s ok=%v",
				tc.header, u, p, ok, tc.user, tc.pass, tc.ok)
		}
	}
}

func TestAuthCache_Hit(t *testing.T) {
	c := newAuthCache()
	b := models.Bot{BrowserID: "x"}
	c.put("k", b)
	got, ok := c.get("k")
	if !ok || got.BrowserID != "x" {
		t.Errorf("cache miss")
	}
}

func TestAuthCache_Expiry(t *testing.T) {
	c := newAuthCache()
	c.ttl = time.Millisecond
	c.put("k", models.Bot{})
	time.Sleep(5 * time.Millisecond)
	if _, ok := c.get("k"); ok {
		t.Error("expected expiry")
	}
}

func TestProxy_HopByHopFiltered(t *testing.T) {
	if !isHopByHop("Connection") || !isHopByHop("PROXY-AUTHORIZATION") {
		t.Error("hop-by-hop misclassified")
	}
	if isHopByHop("Content-Type") {
		t.Error("content-type wrongly classified hop-by-hop")
	}
	_ = strings.TrimSpace
}
