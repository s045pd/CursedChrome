package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Config holds runtime configuration loaded from environment variables.
// Mirrors the env vars used by the Node.js version so the Chrome
// extension and the GUI can connect without modification.
type Config struct {
	// Database
	DatabaseHost     string
	DatabasePort     int
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string

	// Redis
	RedisHost string
	RedisPort int

	// Auth / crypto
	BcryptRounds int

	// Server ports - keep defaults aligned with Node.js
	APIPort   int // 8118
	WSPort    int // 4343
	ProxyPort int // 8080

	// Optional external "BAK" notifier endpoint
	BakServer string

	// Path to the GUI dist directory served as static files
	GUIDistPath string
}

// Default ports/values matching Node.js server.js.
const (
	DefaultAPIPort      = 8118
	DefaultWSPort       = 4343
	DefaultProxyPort    = 8080
	DefaultDatabasePort = 5432
	DefaultRedisPort    = 6379
	DefaultBcryptRounds = 10
	DefaultGUIDistPath  = "/work/gui/dist"
)

// Load reads configuration from environment variables.
// Returns an error if required variables (DATABASE_*) are missing.
func Load() (*Config, error) {
	c := &Config{
		DatabaseHost:     os.Getenv("DATABASE_HOST"),
		DatabaseName:     os.Getenv("DATABASE_NAME"),
		DatabaseUser:     os.Getenv("DATABASE_USER"),
		DatabasePassword: os.Getenv("DATABASE_PASSWORD"),
		RedisHost:        os.Getenv("REDIS_HOST"),
		BakServer:        os.Getenv("BAK_SERVER"),
		GUIDistPath:      envOr("GUI_DIST_PATH", DefaultGUIDistPath),
		DatabasePort:     envInt("DATABASE_PORT", DefaultDatabasePort),
		RedisPort:        envInt("REDIS_PORT", DefaultRedisPort),
		BcryptRounds:     envInt("BCRYPT_ROUNDS", DefaultBcryptRounds),
		APIPort:          envInt("API_PORT", DefaultAPIPort),
		WSPort:           envInt("WS_PORT", DefaultWSPort),
		ProxyPort:        envInt("PROXY_PORT", DefaultProxyPort),
	}

	if c.DatabaseHost == "" {
		return nil, errors.New("DATABASE_HOST is required")
	}
	if c.DatabaseName == "" {
		return nil, errors.New("DATABASE_NAME is required")
	}
	if c.DatabaseUser == "" {
		return nil, errors.New("DATABASE_USER is required")
	}
	if c.RedisHost == "" {
		return nil, errors.New("REDIS_HOST is required")
	}

	return c, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// DSN returns a postgres DSN suitable for gorm.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		c.DatabaseHost, c.DatabasePort, c.DatabaseUser, c.DatabasePassword, c.DatabaseName,
	)
}
