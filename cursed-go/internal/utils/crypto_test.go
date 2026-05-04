package utils

import (
	"strings"
	"testing"
)

func TestSecureRandomHex_LengthAndUniqueness(t *testing.T) {
	a, err := SecureRandomHex(16)
	if err != nil {
		t.Fatal(err)
	}
	b, err := SecureRandomHex(16)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 32 {
		t.Errorf("expected 32 hex chars from 16 bytes, got %d", len(a))
	}
	if a == b {
		t.Errorf("two random strings collided unexpectedly: %q", a)
	}
}

func TestSecureRandomHex_ZeroLen(t *testing.T) {
	if _, err := SecureRandomHex(0); err == nil {
		t.Error("expected error for byteLen=0")
	}
}

func TestHashPassword_Verifies(t *testing.T) {
	hash, err := HashPassword("hunter2", 4)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("expected bcrypt hash, got %q", hash)
	}
	if !VerifyPassword(hash, "hunter2") {
		t.Error("VerifyPassword returned false for correct password")
	}
	if VerifyPassword(hash, "wrong") {
		t.Error("VerifyPassword returned true for wrong password")
	}
}

func TestHashPassword_DefaultCost(t *testing.T) {
	hash, err := HashPassword("p", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(hash, "p") {
		t.Error("default cost verification failed")
	}
}
