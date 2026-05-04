package utils

// BotDefaultSwitchConfig mirrors utils.js BOT_DEFAULT_SWITCH_CONFIG.
// Keep keys identical so the Chrome extension and the GUI keep working
// without any changes after the Go rewrite is deployed.
var BotDefaultSwitchConfig = map[string]bool{
	"SYNC":                 true,
	"SYNC_HUGE":            true,
	"REALTIME_IMG":         true,
	"NOTIFICATION":         true,
	"PERSISTENT_RECORDING": false,
	"PERSISTENT_KEYBOARD":  false,
}

// BotDefaultDataConfig mirrors utils.js BOT_DEFAULT_DATA_CONFIG.
var BotDefaultDataConfig = map[string]any{
	"RECORDING_SECONDS": 30,
	"MONITOR_DOMAINS":   []string{},
}
