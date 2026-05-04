package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
)

func setupProxyCredsAPI(t *testing.T) (*ProxyCredsAPI, *gorm.DB, *fakeRPC) {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AutoMigrate(models.All()...); err != nil {
		t.Fatal(err)
	}
	rpc := &fakeRPC{resp: map[string]any{"cookies": []any{"a", "b"}, "history": []any{"x"}}}
	return &ProxyCredsAPI{DB: g, RPC: rpc}, g, rpc
}

func TestVerifyProxyCredentials_Success(t *testing.T) {
	a, gdb, _ := setupProxyCredsAPI(t)
	b := models.Bot{
		BrowserID: "br1", Name: "n", ProxyUsername: "u1", ProxyPassword: "p1",
		IsOnline: true, LastOnline: time.Now(),
	}
	if err := gdb.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(proxyCredReq{Username: "u1", Password: "p1"})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.VerifyProxyCredentials(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestVerifyProxyCredentials_BadCreds(t *testing.T) {
	a, _, _ := setupProxyCredsAPI(t)
	body, _ := json.Marshal(proxyCredReq{Username: "x", Password: "y"})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.VerifyProxyCredentials(rr, r)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status=%d", rr.Code)
	}
}

func TestGetBotBrowserCookies(t *testing.T) {
	a, gdb, rpc := setupProxyCredsAPI(t)
	b := models.Bot{BrowserID: "br1", Name: "n", ProxyUsername: "u1", ProxyPassword: "p1", LastOnline: time.Now()}
	gdb.Create(&b)

	body, _ := json.Marshal(proxyCredReq{Username: "u1", Password: "p1"})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.GetBotBrowserCookies(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if len(rpc.calls) != 1 || rpc.calls[0].Action != "GET_BROWSER_COOKIE_ARRAY" {
		t.Errorf("rpc call wrong: %+v", rpc.calls)
	}
}

func TestProxyCreds_BotOffline(t *testing.T) {
	a, gdb, rpc := setupProxyCredsAPI(t)
	rpc.err = ErrBotOffline
	b := models.Bot{BrowserID: "br", ProxyUsername: "u", ProxyPassword: "p", LastOnline: time.Now()}
	gdb.Create(&b)
	body, _ := json.Marshal(proxyCredReq{Username: "u", Password: "p"})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.GetBotBrowser(rr, r)
	if rr.Code != http.StatusBadGateway {
		t.Errorf("status=%d, want 502", rr.Code)
	}
	_ = uuid.New
}
