package api

import (
	"encoding/json"
	"net/http"

	"github.com/s045pd/cursed-go/internal/version"
)

// HealthHandler returns 200 with a constant JSON body. Mirrors Node.js
// api-server.js GET /health.
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// VersionHandler returns build name and version. New endpoint introduced
// by the Go rewrite for ops-side smoke testing.
func VersionHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    version.Name,
		"version": version.Version,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
