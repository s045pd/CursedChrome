package db

import (
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
	"github.com/s045pd/cursed-go/internal/utils"
)

// Settings keys we manage automatically. Mirrors database.js initialize_configs.
const (
	SettingSessionSecret = "SESSION_SECRET"
)

// Migrate runs AutoMigrate for all models and seeds default rows
// (admin user, SESSION_SECRET) if the database is empty.
// Returns the auto-generated admin password the *first* time the user is
// created so the operator can grab it from logs; empty string otherwise.
func Migrate(gdb *gorm.DB, logger *slog.Logger, bcryptRounds int) (adminPassword string, err error) {
	if err := gdb.AutoMigrate(models.All()...); err != nil {
		return "", fmt.Errorf("auto migrate: %w", err)
	}

	if err := ensureSessionSecret(gdb); err != nil {
		return "", err
	}

	pwd, err := ensureAdminUser(gdb, bcryptRounds)
	if err != nil {
		return "", err
	}
	if pwd != "" && logger != nil {
		logger.Warn("default admin user created", "username", "admin", "password", pwd,
			"hint", "save this password and rotate it via PUT /api/v1/password")
	}
	return pwd, nil
}

func ensureSessionSecret(gdb *gorm.DB) error {
	var existing models.Setting
	err := gdb.Where("key = ?", SettingSessionSecret).First(&existing).Error
	switch {
	case err == nil:
		return nil // already set
	case errors.Is(err, gorm.ErrRecordNotFound):
		// fall through to create
	default:
		return fmt.Errorf("query session secret: %w", err)
	}

	secret, err := utils.SecureRandomHex(32)
	if err != nil {
		return err
	}
	row := models.Setting{Key: SettingSessionSecret, Value: secret}
	if err := gdb.Create(&row).Error; err != nil {
		return fmt.Errorf("seed session secret: %w", err)
	}
	return nil
}

func ensureAdminUser(gdb *gorm.DB, bcryptRounds int) (string, error) {
	var count int64
	if err := gdb.Model(&models.User{}).Count(&count).Error; err != nil {
		return "", fmt.Errorf("count users: %w", err)
	}
	if count > 0 {
		return "", nil
	}

	plain, err := utils.SecureRandomHex(16)
	if err != nil {
		return "", err
	}
	hash, err := utils.HashPassword(plain, bcryptRounds)
	if err != nil {
		return "", err
	}
	u := models.User{
		Username:                "admin",
		Password:                hash,
		PasswordShouldBeChanged: true,
	}
	if err := gdb.Create(&u).Error; err != nil {
		return "", fmt.Errorf("create admin: %w", err)
	}
	return plain, nil
}

// GetSessionSecret returns the SESSION_SECRET value from settings table.
// Returns ErrSettingMissing if not seeded yet.
var ErrSettingMissing = errors.New("setting missing")

// GetSetting fetches a single setting value by key.
func GetSetting(gdb *gorm.DB, key string) (string, error) {
	var s models.Setting
	if err := gdb.Where("key = ?", key).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrSettingMissing
		}
		return "", err
	}
	return s.Value, nil
}
