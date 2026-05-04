package utils

import (
	"log/slog"
	"os"
)

// NewLogger returns a JSON-structured slog.Logger writing to stdout.
// Centralizing this makes it easy to swap implementations later.
func NewLogger() *slog.Logger {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(h)
}
