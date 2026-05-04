package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
)

func setupSettingsAPI(t *testing.T) *SettingsAPI {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AutoMigrate(&models.Setting{}); err != nil {
		t.Fatal(err)
	}
	return &SettingsAPI{DB: g}
}

func TestGlobalProxy_DefaultEmpty(t *testing.T) {
	a := setupSettingsAPI(t)
	rr := httptest.NewRecorder()
	a.GetGlobalProxy(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestGlobalProxy_SetThenGet(t *testing.T) {
	a := setupSettingsAPI(t)
	id := uuid.New().String()
	body, _ := json.Marshal(setGlobalProxyReq{BotID: id})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.SetGlobalProxy(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("set status=%d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	a.GetGlobalProxy(rr2, httptest.NewRequest(http.MethodGet, "/", nil))
	var resp envelope
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp)
	if resp.Result != id {
		t.Errorf("result=%v want %s", resp.Result, id)
	}
}

func TestGlobalProxy_InvalidUUID(t *testing.T) {
	a := setupSettingsAPI(t)
	body, _ := json.Marshal(setGlobalProxyReq{BotID: "not-a-uuid"})
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.SetGlobalProxy(rr, r)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status=%d", rr.Code)
	}
}

func TestGlobalProxy_Update(t *testing.T) {
	a := setupSettingsAPI(t)
	first := uuid.New().String()
	second := uuid.New().String()

	for _, v := range []string{first, second} {
		body, _ := json.Marshal(setGlobalProxyReq{BotID: v})
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		a.SetGlobalProxy(rr, r)
		if rr.Code != http.StatusOK {
			t.Fatalf("set status=%d", rr.Code)
		}
	}

	rr := httptest.NewRecorder()
	a.GetGlobalProxy(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	var resp envelope
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Result != second {
		t.Errorf("expected upsert to %s, got %v", second, resp.Result)
	}
}
