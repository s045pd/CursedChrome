package api

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// BotRPC is the interface the API layer uses to talk to a connected bot.
// The concrete implementation lives in internal/ws (Stage 5).
// Defining it here lets the API package compile and be tested
// independently with a stub.
type BotRPC interface {
	// CallBot publishes an action+data to the named browser_id and waits
	// for the bot's reply (or context deadline). Returns the raw decoded
	// payload that the bot returned in `data`.
	CallBot(ctx context.Context, browserID string, action string, data map[string]any) (map[string]any, error)

	// IsBotOnline reports whether a websocket session is currently open
	// for the given bot UUID.
	IsBotOnline(botID uuid.UUID) bool
}

// ErrBotOffline is returned by BotRPC.CallBot when the target bot has no
// active websocket session.
var ErrBotOffline = errors.New("bot offline")

// stubBotRPC is the default no-op RPC used when the WS server isn't
// wired in (e.g. unit tests, smoke mode).
type stubBotRPC struct{}

func (stubBotRPC) CallBot(_ context.Context, _ string, _ string, _ map[string]any) (map[string]any, error) {
	return nil, ErrBotOffline
}
func (stubBotRPC) IsBotOnline(_ uuid.UUID) bool { return false }

// NewStubBotRPC returns a BotRPC that always reports offline.
func NewStubBotRPC() BotRPC { return stubBotRPC{} }
