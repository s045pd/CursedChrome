package db

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
	"github.com/s045pd/cursed-go/internal/utils"
)

// SQLite in-memory backend lets us test the GORM mapping without a
// running Postgres. The Postgres-only types (jsonb, uuid) are emulated
// by GORM as text/blob, which is sufficient for unit-testing schema
// assumptions and seed logic.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return gdb
}

func TestMigrate_CreatesTablesAndSeeds(t *testing.T) {
	gdb := openTestDB(t)

	pwd, err := Migrate(gdb, nil, 4)
	if err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if pwd == "" {
		t.Fatal("expected admin password to be returned on first migrate")
	}

	// All model tables exist
	for _, m := range models.All() {
		if !gdb.Migrator().HasTable(m) {
			t.Errorf("table for %T not created", m)
		}
	}

	// Admin user exists with bcrypt hash that verifies
	var u models.User
	if err := gdb.Where("username = ?", "admin").First(&u).Error; err != nil {
		t.Fatalf("admin not created: %v", err)
	}
	if !utils.VerifyPassword(u.Password, pwd) {
		t.Error("returned password does not verify against stored hash")
	}
	if !u.PasswordShouldBeChanged {
		t.Error("admin should have password_should_be_changed=true on first creation")
	}

	// SESSION_SECRET exists
	secret, err := GetSetting(gdb, SettingSessionSecret)
	if err != nil {
		t.Fatalf("SESSION_SECRET not seeded: %v", err)
	}
	if len(secret) < 32 {
		t.Errorf("SESSION_SECRET too short: %d", len(secret))
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	gdb := openTestDB(t)

	if _, err := Migrate(gdb, nil, 4); err != nil {
		t.Fatalf("first migrate: %v", err)
	}

	pwd2, err := Migrate(gdb, nil, 4)
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if pwd2 != "" {
		t.Errorf("second migrate must NOT recreate admin, got pwd=%q", pwd2)
	}

	var n int64
	if err := gdb.Model(&models.User{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("user count after re-migrate = %d, want 1", n)
	}

	// SESSION_SECRET stable across runs
	s1, _ := GetSetting(gdb, SettingSessionSecret)
	if _, err := Migrate(gdb, nil, 4); err != nil {
		t.Fatal(err)
	}
	s2, _ := GetSetting(gdb, SettingSessionSecret)
	if s1 != s2 {
		t.Error("SESSION_SECRET changed across migrations")
	}
}

func TestGetSetting_Missing(t *testing.T) {
	gdb := openTestDB(t)
	if err := gdb.AutoMigrate(&models.Setting{}); err != nil {
		t.Fatal(err)
	}
	_, err := GetSetting(gdb, "DOES_NOT_EXIST")
	if err != ErrSettingMissing {
		t.Errorf("expected ErrSettingMissing, got %v", err)
	}
}

func TestResetBotOnlineState_FlipsTrueRowsOnly(t *testing.T) {
	gdb := openTestDB(t)
	if err := gdb.AutoMigrate(&models.Bot{}); err != nil {
		t.Fatal(err)
	}

	// Two "online" rows survived from a previous run + one already-offline row.
	// The Bot model's is_online column defaults to true and GORM omits Go
	// zero-values from INSERTs, so we have to explicitly force the offline
	// row to false via Update after Create.
	rows := []models.Bot{
		{BrowserID: "online-1", Name: "a", ProxyUsername: "u1", ProxyPassword: "p", IsOnline: true},
		{BrowserID: "online-2", Name: "b", ProxyUsername: "u2", ProxyPassword: "p", IsOnline: true},
		{BrowserID: "offline-1", Name: "c", ProxyUsername: "u3", ProxyPassword: "p"},
	}
	for i := range rows {
		if err := gdb.Create(&rows[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := gdb.Model(&models.Bot{}).Where("browser_id = ?", "offline-1").
		Update("is_online", false).Error; err != nil {
		t.Fatal(err)
	}

	cleared, err := ResetBotOnlineState(gdb)
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if cleared != 2 {
		t.Errorf("cleared = %d, want 2", cleared)
	}

	var stillOnline int64
	gdb.Model(&models.Bot{}).Where("is_online = ?", true).Count(&stillOnline)
	if stillOnline != 0 {
		t.Errorf("rows still online after reset = %d, want 0", stillOnline)
	}

	// Idempotent: calling twice is a no-op.
	cleared2, err := ResetBotOnlineState(gdb)
	if err != nil {
		t.Fatalf("reset 2nd time: %v", err)
	}
	if cleared2 != 0 {
		t.Errorf("second reset cleared = %d, want 0", cleared2)
	}
}

func TestUser_BeforeCreate_GeneratesUUID(t *testing.T) {
	gdb := openTestDB(t)
	if err := gdb.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	u := models.User{Username: "alice", Password: "x"}
	if err := gdb.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	if u.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("UUID was not auto-assigned")
	}
}
