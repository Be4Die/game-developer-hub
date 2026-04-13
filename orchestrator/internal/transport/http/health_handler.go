package http

import (
	"net/http"
	"time"
)

// HealthHandler обрабатывает эндпоинты проверки работоспособности.
type HealthHandler struct {
	startTime time.Time
	version   string
}

// NewHealthHandler создаёт обработчик health check.
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   version,
	}
}

// Check обрабатывает GET /health.
func (h *HealthHandler) Check(w http.ResponseWriter, _ *http.Request) {
	uptime := time.Since(h.startTime).Seconds()

	jsonOK(w, map[string]any{
		"status":         "ok",
		"version":        h.version,
		"uptime_seconds": uptime,
	}, http.StatusOK)
}
