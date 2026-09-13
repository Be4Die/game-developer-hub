package proxy

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/ingress-proxy/internal/store"
)

// Handler обрабатывает входящий игровой трафик (/game-proxy/*) для WebSocket и HTTP.
type Handler struct {
	resolver    store.RouteResolver
	log         *slog.Logger
	dialTimeout time.Duration
}

// NewHandler создаёт новый прокси-обработчик.
func NewHandler(resolver store.RouteResolver, log *slog.Logger, dialTimeout time.Duration) *Handler {
	if dialTimeout <= 0 {
		dialTimeout = 5 * time.Second
	}
	return &Handler{
		resolver:    resolver,
		log:         log,
		dialTimeout: dialTimeout,
	}
}

// ServeHTTP маршрутизирует запросы на целевые инстансы игровых серверов.
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Health check эндпоинт
	if req.URL.Path == "/healthz" || req.URL.Path == "/health" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
		return
	}

	gameID, instanceID, subpath, ok := parseProxyPath(req.URL.Path)
	if !ok {
		h.respondJSON(w, http.StatusBadRequest, "invalid game-proxy path format; expected /game-proxy/{game_id}/{instance_id}/*")
		return
	}

	targetAddr, err := h.resolver.ResolveInstanceTarget(req.Context(), instanceID)
	if err != nil {
		if errors.Is(err, store.ErrTargetNotFound) {
			h.log.Debug("instance route not found",
				slog.Int64("game_id", gameID),
				slog.Int64("instance_id", instanceID),
			)
			h.respondJSON(w, http.StatusNotFound, fmt.Sprintf("game instance %d not found or offline", instanceID))
			return
		}
		h.log.Error("failed to resolve instance target",
			slog.Int64("instance_id", instanceID),
			slog.Any("err", err),
		)
		h.respondJSON(w, http.StatusInternalServerError, "internal server error resolving instance")
		return
	}

	// Проверяем, является ли запрос WebSocket-соединением
	if isWebSocketRequest(req) {
		h.proxyWebSocket(w, req, targetAddr, subpath)
		return
	}

	// Для обычного HTTP (REST, Server-Sent Events, Long-Polling, SignalR negotiate)
	h.proxyHTTP(w, req, targetAddr, subpath)
}

func (h *Handler) proxyWebSocket(w http.ResponseWriter, req *http.Request, targetAddr, subpath string) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		h.log.Error("webserver does not support hijacking for websocket")
		h.respondJSON(w, http.StatusInternalServerError, "websocket hijacking not supported")
		return
	}

	targetConn, err := net.DialTimeout("tcp", targetAddr, h.dialTimeout)
	if err != nil {
		h.log.Warn("failed to connect to game instance tcp target",
			slog.String("target", targetAddr),
			slog.Any("err", err),
		)
		h.respondJSON(w, http.StatusBadGateway, "failed to connect to backend game instance")
		return
	}
	defer targetConn.Close()

	clientConn, brw, err := hijacker.Hijack()
	if err != nil {
		h.log.Error("failed to hijack client connection", slog.Any("err", err))
		return
	}
	defer clientConn.Close()

	// Переписываем URL путь на subpath для целевого сервера
	req.URL.Path = subpath
	req.Host = targetAddr

	// Отправляем оригинальный HTTP handshake запрос целевому серверу
	if err := req.Write(targetConn); err != nil {
		h.log.Warn("failed to write websocket request to target", slog.Any("err", err))
		return
	}

	// Если клиент успел отправить данные в буфер до hijack
	if n := brw.Reader.Buffered(); n > 0 {
		buf := make([]byte, n)
		if _, err := io.ReadFull(brw.Reader, buf); err == nil {
			_, _ = targetConn.Write(buf)
		}
	}

	// Двунаправленный стриминг байтов
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(targetConn, clientConn)
		_ = targetConn.Close()
		close(done)
	}()

	_, _ = io.Copy(clientConn, targetConn)
	_ = clientConn.Close()
	<-done
}

func (h *Handler) proxyHTTP(w http.ResponseWriter, req *http.Request, targetAddr, subpath string) {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = targetAddr
			r.URL.Path = subpath
			r.Host = targetAddr
			if _, ok := r.Header["User-Agent"]; !ok {
				r.Header.Set("User-Agent", "")
			}
		},
		FlushInterval: -1, // Стриминг ответов без буферизации (SSE, Chunked, Polling)
		ErrorHandler: func(rw http.ResponseWriter, r *http.Request, err error) {
			h.log.Warn("http reverse proxy error",
				slog.String("target", targetAddr),
				slog.String("path", r.URL.Path),
				slog.Any("err", err),
			)
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadGateway)
			_, _ = rw.Write([]byte(`{"error":"bad gateway to game instance"}`))
		},
	}

	proxy.ServeHTTP(w, req)
}

func parseProxyPath(path string) (gameID int64, instanceID int64, subpath string, ok bool) {
	path = strings.TrimPrefix(path, "/game-proxy")
	path = strings.TrimPrefix(path, "/")

	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 {
		return 0, 0, "", false
	}

	gID, err1 := strconv.ParseInt(parts[0], 10, 64)
	iID, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, "", false
	}

	sub := "/"
	if len(parts) == 3 && parts[2] != "" {
		sub = "/" + parts[2]
	}

	return gID, iID, sub, true
}

func isWebSocketRequest(req *http.Request) bool {
	if strings.EqualFold(req.Header.Get("Upgrade"), "websocket") {
		return true
	}
	connHdr := strings.ToLower(req.Header.Get("Connection"))
	return strings.Contains(connHdr, "upgrade") && strings.EqualFold(req.Header.Get("Upgrade"), "websocket")
}

func (h *Handler) respondJSON(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = fmt.Fprintf(w, `{"error":%q}`+"\n", msg)
}
