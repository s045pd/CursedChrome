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
