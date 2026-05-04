package config

import (
	"strings"
	"testing"
)

func TestLoad_MissingDatabaseHost(t *testing.T) {
	t.Setenv("DATABASE_HOST", "")
	t.Setenv("DATABASE_NAME", "x")
	t.Setenv("DATABASE_USER", "x")
	t.Setenv("REDIS_HOST", "x")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_HOST is missing")
	}
	if !strings.Contains(err.Error(), "DATABASE_HOST") {
		t.Fatalf("error should mention DATABASE_HOST, got: %v", err)
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	t.Setenv("DATABASE_HOST", "db")
	t.Setenv("DATABASE_NAME", "cursedchrome")
	t.Setenv("DATABASE_USER", "cursedchrome")
	t.Setenv("DATABASE_PASSWORD", "secret")
	t.Setenv("REDIS_HOST", "redis")
	// Ensure no overrides
	t.Setenv("API_PORT", "")
	t.Setenv("WS_PORT", "")
	t.Setenv("PROXY_PORT", "")
	t.Setenv("BCRYPT_ROUNDS", "")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.APIPort != DefaultAPIPort {
		t.Errorf("APIPort = %d, want %d", c.APIPort, DefaultAPIPort)
	}
	if c.WSPort != DefaultWSPort {
		t.Errorf("WSPort = %d, want %d", c.WSPort, DefaultWSPort)
	}
	if c.ProxyPort != DefaultProxyPort {
		t.Errorf("ProxyPort = %d, want %d", c.ProxyPort, DefaultProxyPort)
	}
	if c.BcryptRounds != DefaultBcryptRounds {
		t.Errorf("BcryptRounds = %d, want %d", c.BcryptRounds, DefaultBcryptRounds)
	}
	if c.DatabasePort != DefaultDatabasePort {
		t.Errorf("DatabasePort = %d, want %d", c.DatabasePort, DefaultDatabasePort)
	}
}

func TestLoad_OverridesApplied(t *testing.T) {
	t.Setenv("DATABASE_HOST", "db")
	t.Setenv("DATABASE_NAME", "cursedchrome")
	t.Setenv("DATABASE_USER", "cursedchrome")
	t.Setenv("DATABASE_PASSWORD", "secret")
	t.Setenv("REDIS_HOST", "redis")
	t.Setenv("API_PORT", "9118")
	t.Setenv("BCRYPT_ROUNDS", "12")

	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.APIPort != 9118 {
		t.Errorf("APIPort = %d, want 9118", c.APIPort)
	}
	if c.BcryptRounds != 12 {
		t.Errorf("BcryptRounds = %d, want 12", c.BcryptRounds)
	}
}

func TestDSN(t *testing.T) {
	c := &Config{
		DatabaseHost:     "db",
		DatabasePort:     5432,
		DatabaseUser:     "u",
		DatabasePassword: "p",
		DatabaseName:     "n",
	}
	got := c.DSN()
	wantParts := []string{"host=db", "port=5432", "user=u", "password=p", "dbname=n", "sslmode=disable"}
	for _, p := range wantParts {
		if !strings.Contains(got, p) {
			t.Errorf("DSN() = %q, missing %q", got, p)
		}
	}
}

func TestEnvInt_Fallback(t *testing.T) {
	t.Setenv("X_PORT_TEST", "not-a-number")
	if got := envInt("X_PORT_TEST", 99); got != 99 {
		t.Errorf("envInt fallback = %d, want 99", got)
	}
}
