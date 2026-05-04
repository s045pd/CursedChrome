package version

import "testing"

func TestNameNotEmpty(t *testing.T) {
	if Name == "" {
		t.Fatal("Name must not be empty")
	}
}

func TestVersionNotEmpty(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must not be empty")
	}
}
