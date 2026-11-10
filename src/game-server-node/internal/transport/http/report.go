// Package http provides HTTP handlers for instance reports.
package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
)

// ReportHandler принимает отчёты об онлайне и очереди от запущенных игровых серверов.
type ReportHandler struct {
	log     *slog.Logger
	storage *memory.Storage
	secret  string
}

// ReportPayload описывает тело запроса от игрового сервера.
type ReportPayload struct {
	InstanceID  int64  `json:"instance_id"`
	PlayerCount uint32 `json:"player_count"`
	QueueSize   uint32 `json:"queue_size"`
	MaxPlayers  uint32 `json:"max_players"`
}

// maxReportBodyBytes ограничивает размер тела отчёта.
const maxReportBodyBytes = 4 << 10

// NewReportHandler создаёт обработчик отчётов. secret — ключ ноды, из которого
// выводятся токены инстансов (см. domain.ReportToken).
func NewReportHandler(log *slog.Logger, storage *memory.Storage, secret string) *ReportHandler {
	return &ReportHandler{log: log, storage: storage, secret: secret}
}

// ServeHTTP обрабатывает POST /v1/report.
func (h *ReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxReportBodyBytes)
	var payload ReportPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("decode json: %v", err), http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	if payload.InstanceID <= 0 {
		http.Error(w, "instance_id is required", http.StatusBadRequest)
		return
	}

	// Без проверки токена любой процесс в сети ноды мог бы подменить онлайн чужого
	// инстанса и повлиять на автоскейлинг.
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !domain.ValidReportToken(h.secret, payload.InstanceID, token) {
		http.Error(w, "invalid report token", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	inst, err := h.storage.GetInstanceByID(ctx, payload.InstanceID)
	if err != nil {
		http.Error(w, "instance not found", http.StatusNotFound)
		return
	}

	// Онлайн обновляется всегда: 0 — валидное значение.
	inst.PlayerCount = &payload.PlayerCount
	if payload.QueueSize > 0 {
		inst.QueueSize = &payload.QueueSize
	}
	if payload.MaxPlayers > 0 {
		inst.MaxPlayers = payload.MaxPlayers
	}

	if err := h.storage.RecordInstance(ctx, *inst); err != nil {
		http.Error(w, fmt.Sprintf("save instance: %v", err), http.StatusInternalServerError)
		return
	}

	h.log.Debug("instance report received",
		slog.Int64("instance_id", payload.InstanceID),
		slog.Uint64("player_count", uint64(payload.PlayerCount)),
		slog.Uint64("queue_size", uint64(payload.QueueSize)),
	)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// NewReportServer создаёт HTTP-сервер отчётов. Сервер слушает все интерфейсы, потому что
// игровые контейнеры обращаются к нему через docker-шлюз; доступ защищён токенами инстансов.
func NewReportServer(log *slog.Logger, storage *memory.Storage, port int, secret string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/v1/report", NewReportHandler(log, storage, secret))

	return &http.Server{
		Addr:              "0.0.0.0:" + strconv.Itoa(port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
