package api

import (
	"encoding/json"
	"net/http"
)

// Standard envelope used by api-server.js: {success, result?, error?}
type envelope struct {
	Success bool   `json:"success"`
	Result  any    `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

// JSONOK writes a 200 success envelope.
func JSONOK(w http.ResponseWriter, result any) {
	writeJSON(w, http.StatusOK, envelope{Success: true, Result: result})
}

// JSONErr writes an error envelope with the given status.
func JSONErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, envelope{Success: false, Error: msg})
}

// MustDecode parses a JSON body or writes 400 + returns false.
func MustDecode(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return false
	}
	return true
}
