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

	"github.com/s045pd/cursed-go/internal/auth"
	"github.com/s045pd/cursed-go/internal/db/models"
	"github.com/s045pd/cursed-go/internal/utils"
)

const testSecret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func setupAuthAPI(t *testing.T) *AuthAPI {
	t.Helper()
	g, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	mgr, _ := auth.NewManager(testSecret)
	return &AuthAPI{DB: g, Sessions: mgr, BcryptRounds: 4}
}

func seedUser(t *testing.T, a *AuthAPI, username, password string) uuid.UUID {
	t.Helper()
	hash, _ := utils.HashPassword(password, 4)
	u := models.User{Username: username, Password: hash}
	if err := a.DB.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	return u.ID
}

func decodeJSON(t *testing.T, r *http.Response) map[string]any {
	t.Helper()
	defer r.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestLogin_Success(t *testing.T) {
	a := setupAuthAPI(t)
	seedUser(t, a, "alice", "hunter222")

	body, _ := json.Marshal(loginReq{Username: "alice", Password: "hunter222"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.Login(rr, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
	resp := decodeJSON(t, rr.Result())
	if resp["success"] != true {
		t.Errorf("success != true: %v", resp)
	}
	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == auth.SessionCookieName {
			found = true
		}
	}
	if !found {
		t.Error("session cookie not set on success")
	}
}

func TestLogin_BadPassword(t *testing.T) {
	a := setupAuthAPI(t)
	seedUser(t, a, "alice", "hunter222")

	body, _ := json.Marshal(loginReq{Username: "alice", Password: "wrong"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.Login(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	a := setupAuthAPI(t)
	body, _ := json.Marshal(loginReq{Username: "ghost", Password: "x"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.Login(rr, r)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	a := setupAuthAPI(t)
	body, _ := json.Marshal(loginReq{})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	a.Login(rr, r)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
}

func TestLogout(t *testing.T) {
	a := setupAuthAPI(t)
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/logout", nil)
	a.Logout(rr, r)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d", rr.Code)
	}
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no cookie cleared")
	}
}

func TestMe_NoSession(t *testing.T) {
	a := setupAuthAPI(t)
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	a.Me(rr, r)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestMe_WithSession(t *testing.T) {
	a := setupAuthAPI(t)
	uid := seedUser(t, a, "alice", "hunter222")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	r = r.WithContext(auth.WithSession(r.Context(), auth.Session{UserID: uid}))
	rr := httptest.NewRecorder()
	a.Me(rr, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}
	resp := decodeJSON(t, rr.Result())
	res := resp["result"].(map[string]any)
	if res["username"] != "alice" {
		t.Errorf("username = %v", res["username"])
	}
}

func TestChangePassword_Success(t *testing.T) {
	a := setupAuthAPI(t)
	uid := seedUser(t, a, "alice", "hunter222")

	body, _ := json.Marshal(changePwdReq{NewPassword: "newhunter"})
	r := httptest.NewRequest(http.MethodPut, "/api/v1/password", bytes.NewReader(body))
	r = r.WithContext(auth.WithSession(r.Context(), auth.Session{UserID: uid}))
	rr := httptest.NewRecorder()
	a.ChangePassword(rr, r)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rr.Code, rr.Body.String())
	}

	var u models.User
	a.DB.Where("id = ?", uid).First(&u)
	if !utils.VerifyPassword(u.Password, "newhunter") {
		t.Error("new password not stored")
	}
	if u.PasswordShouldBeChanged {
		t.Error("password_should_be_changed not cleared")
	}
}

func TestChangePassword_TooShort(t *testing.T) {
	a := setupAuthAPI(t)
	uid := seedUser(t, a, "alice", "hunter222")
	body, _ := json.Marshal(changePwdReq{NewPassword: "x"})
	r := httptest.NewRequest(http.MethodPut, "/api/v1/password", bytes.NewReader(body))
	r = r.WithContext(auth.WithSession(r.Context(), auth.Session{UserID: uid}))
	rr := httptest.NewRecorder()
	a.ChangePassword(rr, r)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rr.Code)
	}
}
