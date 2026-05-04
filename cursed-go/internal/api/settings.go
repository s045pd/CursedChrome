package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db"
	"github.com/s045pd/cursed-go/internal/db/models"
)

// Setting key for the global default proxy bot.
const SettingGlobalProxyBot = "GLOBAL_PROXY_BOT"

// SettingsAPI groups settings routes.
type SettingsAPI struct {
	DB *gorm.DB
}

type setGlobalProxyReq struct {
	BotID string `json:"bot_id"`
}

// GetGlobalProxy is GET /api/v1/settings/global-proxy
func (a *SettingsAPI) GetGlobalProxy(w http.ResponseWriter, _ *http.Request) {
	v, err := db.GetSetting(a.DB, SettingGlobalProxyBot)
	if errors.Is(err, db.ErrSettingMissing) || v == "" {
		JSONOK(w, nil)
		return
	}
	if err != nil {
		JSONErr(w, http.StatusInternalServerError, "settings lookup failed")
		return
	}
	JSONOK(w, v)
}

// SetGlobalProxy is POST /api/v1/settings/global-proxy
func (a *SettingsAPI) SetGlobalProxy(w http.ResponseWriter, r *http.Request) {
	var body setGlobalProxyReq
	if !MustDecode(w, r, &body) {
		return
	}
	// Empty body means "unset"
	if body.BotID != "" {
		if _, err := uuid.Parse(body.BotID); err != nil {
			JSONErr(w, http.StatusBadRequest, "invalid bot_id")
			return
		}
	}
	if err := upsertSetting(a.DB, SettingGlobalProxyBot, body.BotID); err != nil {
		JSONErr(w, http.StatusInternalServerError, "update failed")
		return
	}
	JSONOK(w, body.BotID)
}

// upsertSetting inserts or updates a settings row by key.
func upsertSetting(gdb *gorm.DB, key, value string) error {
	var existing models.Setting
	err := gdb.Where("key = ?", key).First(&existing).Error
	switch {
	case err == nil:
		return gdb.Model(&existing).Update("value", value).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return gdb.Create(&models.Setting{Key: key, Value: value}).Error
	default:
		return err
	}
}
