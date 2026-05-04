package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/auth"
	"github.com/s045pd/cursed-go/internal/db/models"
	"github.com/s045pd/cursed-go/internal/utils"
)

// AuthAPI groups handlers that need access to db + session manager.
type AuthAPI struct {
	DB           *gorm.DB
	Sessions     *auth.Manager
	BcryptRounds int
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResult struct {
	Username                string `json:"username"`
	PasswordShouldBeChanged bool   `json:"password_should_be_changed"`
}

// Login is POST /api/v1/login.
func (a *AuthAPI) Login(w http.ResponseWriter, r *http.Request) {
	if a.DB == nil {
		JSONErr(w, http.StatusServiceUnavailable, "database not available")
		return
	}
	var body loginReq
	if !MustDecode(w, r, &body) {
		return
	}
	if body.Username == "" || body.Password == "" {
		JSONErr(w, http.StatusBadRequest, "username and password required")
		return
	}

	var u models.User
	err := a.DB.Where("username = ?", body.Username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// constant-time delay defeats username enumeration
			time.Sleep(150 * time.Millisecond)
			JSONErr(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		JSONErr(w, http.StatusInternalServerError, "login failed")
		return
	}

	if !utils.VerifyPassword(u.Password, body.Password) {
		JSONErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := a.Sessions.Issue(w, u.ID); err != nil {
		JSONErr(w, http.StatusInternalServerError, "session issue failed")
		return
	}

	JSONOK(w, meResult{Username: u.Username, PasswordShouldBeChanged: u.PasswordShouldBeChanged})
}

// Logout is GET /api/v1/logout.
func (a *AuthAPI) Logout(w http.ResponseWriter, _ *http.Request) {
	a.Sessions.Clear(w)
	JSONOK(w, struct{}{})
}

// Me is GET /api/v1/me.
func (a *AuthAPI) Me(w http.ResponseWriter, r *http.Request) {
	s, ok := auth.SessionFromContext(r.Context())
	if !ok {
		JSONErr(w, http.StatusUnauthorized, "no session")
		return
	}
	var u models.User
	if err := a.DB.Where("id = ?", s.UserID).First(&u).Error; err != nil {
		JSONErr(w, http.StatusUnauthorized, "user not found")
		return
	}
	JSONOK(w, meResult{Username: u.Username, PasswordShouldBeChanged: u.PasswordShouldBeChanged})
}

type changePwdReq struct {
	NewPassword string `json:"new_password"`
}

// ChangePassword is PUT /api/v1/password.
func (a *AuthAPI) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var body changePwdReq
	if !MustDecode(w, r, &body) {
		return
	}
	if len(body.NewPassword) < 8 {
		JSONErr(w, http.StatusBadRequest, "password must be at least 8 chars")
		return
	}

	s, ok := auth.SessionFromContext(r.Context())
	if !ok {
		JSONErr(w, http.StatusUnauthorized, "no session")
		return
	}

	hash, err := utils.HashPassword(body.NewPassword, a.BcryptRounds)
	if err != nil {
		JSONErr(w, http.StatusInternalServerError, "hash failed")
		return
	}
	res := a.DB.Model(&models.User{}).Where("id = ?", s.UserID).
		Updates(map[string]any{"password": hash, "password_should_be_changed": false})
	if res.Error != nil {
		JSONErr(w, http.StatusInternalServerError, "update failed")
		return
	}
	JSONOK(w, struct{}{})
}

// Compile-time assertion: handlers conform to http.HandlerFunc.
var _ http.HandlerFunc = (*AuthAPI)(nil).Login
var _ uuid.UUID
