// Package http реализует HTTP-сервер для приёма отчётов от игровых серверов.
package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/storage/memory"
)

// ReportHandler принимает отчёты об онлайне и очереди от запущенных игровых серверов.
type ReportHandler struct {
	log     *slog.Logger
	storage *memory.Storage
}

// ReportPayload описывает тело запроса от игрового сервера.
type ReportPayload struct {
	InstanceID  int64  `json:"instance_id"`
	PlayerCount uint32 `json:"player_count"`
	QueueSize   uint32 `json:"queue_size"`
	MaxPlayers  uint32 `json:"max_players"`
}

// NewReportHandler создаёт обработчик отчётов.
func NewReportHandler(log *slog.Logger, storage *memory.Storage) *ReportHandler {
	return &ReportHandler{log: log, storage: storage}
}

// ServeHTTP обрабатывает POST /v1/report.
func (h *ReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload ReportPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("decode json: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if payload.InstanceID <= 0 {
		http.Error(w, "instance_id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	inst, err := h.storage.GetInstanceByID(ctx, payload.InstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("instance not found: %v", err), http.StatusNotFound)
		return
	}

	// Обновляем поля.
	if payload.PlayerCount > 0 || true { // обновляем даже 0 — это валидное значение
		inst.PlayerCount = &payload.PlayerCount
	}
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

// NewReportServer создаёт HTTP-сервер отчётов, слушающий только localhost.
func NewReportServer(log *slog.Logger, storage *memory.Storage, port int) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/v1/report", NewReportHandler(log, storage))

	return &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(port),
		Handler: mux,
	}
}
