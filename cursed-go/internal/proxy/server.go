package proxy

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/s045pd/cursed-go/internal/api"
)

// Server is the HTTP forward proxy.
type Server struct {
	DB     *gorm.DB
	RPC    api.BotRPC
	Logger *slog.Logger
	cache  *authCache
}

// New creates a Server.
func New(gdb *gorm.DB, rpc api.BotRPC, logger *slog.Logger) *Server {
	return &Server{DB: gdb, RPC: rpc, Logger: logger, cache: newAuthCache()}
}

// Handler returns an http.Handler implementing the forward proxy.
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.serve)
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	bot, err := authenticate(s.DB, s.cache, r.Header)
	if err != nil {
		w.Header().Set("Proxy-Authenticate", `Basic realm="cursed-proxy"`)
		http.Error(w, "Proxy Authentication Required", http.StatusProxyAuthRequired)
		return
	}

	if r.Method == http.MethodConnect {
		s.handleConnect(w, r, bot.BrowserID)
		return
	}
	s.handleHTTP(w, r, bot.BrowserID)
}

// handleHTTP forwards a regular request via the bot's browser.
func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request, browserID string) {
	bodyBytes, _ := io.ReadAll(io.LimitReader(r.Body, 32<<20))
	bodyB64 := base64.StdEncoding.EncodeToString(bodyBytes)

	headers := map[string]string{}
	for k, v := range r.Header {
		if isHopByHop(k) {
			continue
		}
		headers[k] = strings.Join(v, ", ")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	resp, err := s.RPC.CallBot(ctx, browserID, "SEND_REQUEST_VIA_BROWSER", map[string]any{
		"method":  r.Method,
		"url":     fullURL(r),
		"headers": headers,
		"body":    bodyB64,
	})
	if err != nil {
		if errors.Is(err, api.ErrBotOffline) {
			http.Error(w, "Bot offline", http.StatusBadGateway)
			return
		}
		http.Error(w, err.Error(), http.StatusGatewayTimeout)
		return
	}

	// Bot reply shape: {status, headers{}, body(base64)}
	status, _ := resp["status"].(float64)
	if status == 0 {
		status = 200
	}
	if hdrs, ok := resp["headers"].(map[string]any); ok {
		for k, v := range hdrs {
			if isHopByHop(k) || strings.EqualFold(k, "content-encoding") {
				continue
			}
			w.Header().Set(k, fmt.Sprint(v))
		}
	}
	w.WriteHeader(int(status))
	if body, ok := resp["body"].(string); ok && body != "" {
		decoded, err := base64.StdEncoding.DecodeString(body)
		if err == nil {
			_, _ = w.Write(decoded)
		} else {
			_, _ = w.Write([]byte(body))
		}
	}
}

// handleConnect handles CONNECT method (HTTPS) by hijacking the
// connection. We DO NOT attempt MITM — we tunnel raw bytes through the
// bot via a (future) streamed RPC. For MVP we reject with 502 if the
// bot doesn't support streaming, matching the most common usage where
// callers go through HTTP only.
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request, browserID string) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking unsupported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	host := r.URL.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}
	// Direct dial; the bot is informed but not actually used as exit node
	// for CONNECT (matches Node.js behavior of HTTPS pass-through).
	_, _ = browserID, "" // reserved for future bot-tunnelled CONNECT
	upstream, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer upstream.Close()
	_, _ = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))

	tunnel(clientConn, upstream)
}

func tunnel(a, b net.Conn) {
	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(a, b); done <- struct{}{} }()
	go func() { _, _ = io.Copy(b, a); done <- struct{}{} }()
	<-done
}

func fullURL(r *http.Request) string {
	if r.URL.IsAbs() {
		return r.URL.String()
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + r.URL.RequestURI()
}

func isHopByHop(name string) bool {
	switch strings.ToLower(name) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
		"te", "trailer", "transfer-encoding", "upgrade":
		return true
	}
	return false
}
