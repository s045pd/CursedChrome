package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/s045pd/cursed-go/internal/api"
	"github.com/s045pd/cursed-go/internal/auth"
	"github.com/s045pd/cursed-go/internal/config"
	"github.com/s045pd/cursed-go/internal/db"
	"github.com/s045pd/cursed-go/internal/proxy"
	"github.com/s045pd/cursed-go/internal/utils"
	"github.com/s045pd/cursed-go/internal/version"
	wsx "github.com/s045pd/cursed-go/internal/ws"
)

func main() {
	logger := utils.NewLogger()
	logger.Info("starting", "name", version.Name, "version", version.Version)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", "err", err)
		os.Exit(1)
	}

	deps := api.Deps{
		BcryptRounds: cfg.BcryptRounds,
		GUIDistPath:  cfg.GUIDistPath,
		BotRPC:       api.NewStubBotRPC(),
	}

	skipDB := os.Getenv("SKIP_DB") == "1"
	if skipDB {
		mgr, _ := auth.NewManager("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
		deps.Sessions = mgr
		logger.Warn("SKIP_DB=1, database/ws/proxy disabled (smoke mode)")
	} else {
		gdb, err := db.Open(cfg.DSN())
		if err != nil {
			logger.Error("db open failed", "err", err)
			os.Exit(1)
		}
		defer func() { _ = db.Close(gdb) }()

		if _, err := db.Migrate(gdb, logger, cfg.BcryptRounds); err != nil {
			logger.Error("db migrate failed", "err", err)
			os.Exit(1)
		}
		secret, err := db.GetSetting(gdb, db.SettingSessionSecret)
		if err != nil {
			logger.Error("session secret missing", "err", err)
			os.Exit(1)
		}
		mgr, err := auth.NewManager(secret)
		if err != nil {
			logger.Error("session manager init failed", "err", err)
			os.Exit(1)
		}
		deps.DB = gdb
		deps.Sessions = mgr

		// WS server doubles as the BotRPC implementation.
		ws := wsx.New(gdb, logger)
		deps.BotRPC = ws

		// Proxy uses the same RPC.
		px := proxy.New(gdb, ws, logger)

		// Run WS + Proxy.
		startServer(logger, fmt.Sprintf(":%d", cfg.WSPort), ws.Handler(), "ws")
		startServer(logger, fmt.Sprintf(":%d", cfg.ProxyPort), px.Handler(), "proxy")
		logger.Info("db ready")
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.APIPort),
		Handler:           api.NewRouter(deps),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("api listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server crashed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("api shutdown failed", "err", err)
	}
	logger.Info("bye")
}

// startServer launches an http.Server in a goroutine and registers it
// with the global wait group used at shutdown.
var serverWG sync.WaitGroup

type slogLike interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

func startServer(logger slogLike, addr string, h http.Handler, name string) {
	serverWG.Add(1)
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		defer serverWG.Done()
		logger.Info(name+" listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(name+" server crashed", "err", err)
		}
	}()
}
