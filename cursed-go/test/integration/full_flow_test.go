// Integration test that wires API + WS + Proxy against SQLite in-memory
// and simulates a real bot via a websocket client.
package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"nhooyr.io/websocket"

	"github.com/s045pd/cursed-go/internal/api"
	"github.com/s045pd/cursed-go/internal/auth"
	"github.com/s045pd/cursed-go/internal/db"
	"github.com/s045pd/cursed-go/internal/db/models"
	"github.com/s045pd/cursed-go/internal/proxy"
	"github.com/s045pd/cursed-go/internal/utils"
	wsx "github.com/s045pd/cursed-go/internal/ws"
)

const testSecret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

type stack struct {
	APIServer   *httptest.Server
	WSServer    *httptest.Server
	ProxyServer *httptest.Server
	DB          *gorm.DB
	AdminPwd    string
}

func setupStack(t *testing.T) *stack {
	t.Helper()
	// Use a shared in-memory database so concurrent goroutines see the
	// same schema. Each test gets its own URI so they don't collide.
	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	g, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// SQLite shared-memory cache requires single writer; cap connections.
	sqlDB, _ := g.DB()
	sqlDB.SetMaxOpenConns(1)
	pwd, err := db.Migrate(g, nil, 4)
	if err != nil {
		t.Fatal(err)
	}

	mgr, err := auth.NewManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}

	logger := utils.NewLogger()
	wsServer := wsx.New(g, logger)
	proxyServer := proxy.New(g, wsServer, logger)

	deps := api.Deps{
		DB:           g,
		Sessions:     mgr,
		BotRPC:       wsServer,
		BcryptRounds: 4,
	}
	apiHTTP := httptest.NewServer(api.NewRouter(deps))
	wsHTTP := httptest.NewServer(wsServer.Handler())
	proxyHTTP := httptest.NewServer(proxyServer.Handler())
	t.Cleanup(func() {
		apiHTTP.Close()
		wsHTTP.Close()
		proxyHTTP.Close()
	})
	return &stack{apiHTTP, wsHTTP, proxyHTTP, g, pwd}
}

// loginAdmin POSTs /api/v1/login with the seed admin credentials and
// returns the session cookie.
func loginAdmin(t *testing.T, s *stack) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"username": "admin", "password": s.AdminPwd})
	resp, err := http.Post(s.APIServer.URL+"/api/v1/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("login status %d", resp.StatusCode)
	}
	for _, c := range resp.Cookies() {
		if c.Name == auth.SessionCookieName {
			return c
		}
	}
	t.Fatal("no session cookie")
	return nil
}

func TestFlow_LoginListBots(t *testing.T) {
	s := setupStack(t)
	cookie := loginAdmin(t, s)

	req, _ := http.NewRequest("GET", s.APIServer.URL+"/api/v1/bots", nil)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("list status %d", resp.StatusCode)
	}
}

func TestFlow_BotConnectsAndPersists(t *testing.T) {
	s := setupStack(t)
	wsURL := strings.Replace(s.WSServer.URL, "http://", "ws://", 1)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// AUTH probe
	_, msg, err := conn.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var probe wsx.Envelope
	_ = json.Unmarshal(msg, &probe)
	if probe.Action != wsx.ActionAuth {
		t.Fatalf("expected AUTH, got %s", probe.Action)
	}
	browserID := uuid.NewString()
	authData, _ := json.Marshal(wsx.AuthData{BrowserID: browserID})
	reply, _ := json.Marshal(wsx.Envelope{ID: probe.ID, Action: wsx.ActionAuth, Data: authData})
	conn.Write(ctx, websocket.MessageText, reply)

	// Wait for DB row
	deadline := time.Now().Add(3 * time.Second)
	var b models.Bot
	for time.Now().Before(deadline) {
		if s.DB.Where("browser_id = ?", browserID).First(&b).Error == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if b.BrowserID != browserID {
		t.Fatal("bot not persisted")
	}

	// Send PING; DB should mark online + receive PONG.
	pingEnv, _ := json.Marshal(wsx.Envelope{ID: "ping", Action: wsx.ActionPing})
	conn.Write(ctx, websocket.MessageText, pingEnv)
	_, pongRaw, _ := conn.Read(ctx)
	var pong wsx.Envelope
	_ = json.Unmarshal(pongRaw, &pong)
	if pong.Action != wsx.ActionPong {
		t.Errorf("expected PONG, got %s", pong.Action)
	}
}

func TestFlow_ProxyWithBot(t *testing.T) {
	s := setupStack(t)

	// Connect a fake bot
	wsURL := strings.Replace(s.WSServer.URL, "http://", "ws://", 1)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	_, msg, _ := conn.Read(ctx)
	var probe wsx.Envelope
	_ = json.Unmarshal(msg, &probe)
	browserID := uuid.NewString()
	authData, _ := json.Marshal(wsx.AuthData{
		BrowserID: browserID, ProxyUsername: "u1", ProxyPassword: "p1",
	})
	reply, _ := json.Marshal(wsx.Envelope{ID: probe.ID, Action: wsx.ActionAuth, Data: authData})
	conn.Write(ctx, websocket.MessageText, reply)

	// Bot loop: respond to any SEND_REQUEST_VIA_BROWSER with a 200/hello payload
	go func() {
		for {
			_, raw, err := conn.Read(ctx)
			if err != nil {
				return
			}
			var env wsx.Envelope
			if err := json.Unmarshal(raw, &env); err != nil {
				continue
			}
			if env.Action != wsx.ActionSendRequestViaBrowser {
				continue
			}
			result, _ := json.Marshal(map[string]any{
				"status":  200,
				"headers": map[string]string{"X-Origin": "fake-bot"},
				"body":    base64.StdEncoding.EncodeToString([]byte("hi from bot")),
			})
			out, _ := json.Marshal(wsx.Envelope{
				ID:           env.ID,
				Action:       wsx.ActionSendRequestViaBrowser,
				OriginAction: env.Action,
				Data:         result,
			})
			conn.Write(ctx, websocket.MessageText, out)
		}
	}()

	// Wait for DB row + registry entry
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var b models.Bot
		if s.DB.Where("browser_id = ?", browserID).First(&b).Error == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Hit the proxy
	req, _ := http.NewRequest("GET", "http://example.com/anything", nil)
	req.URL, _ = req.URL.Parse(s.ProxyServer.URL + "/anything")
	req.Host = "example.com"
	req.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("u1:p1")))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("proxy status %d", resp.StatusCode)
	}
	if resp.Header.Get("X-Origin") != "fake-bot" {
		t.Errorf("missing forwarded header")
	}
}
