// Package http реализует HTTP-транспорт оркестратора на chi.
package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/go-chi/chi/v5"
)

// errorResponse записывает JSON-ответ с ошибкой.
func errorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := map[string]string{
		"code":    code,
		"message": message,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// badRequest записывает ответ 400 с ошибкой.
func badRequest(w http.ResponseWriter, msg string) {
	errorResponse(w, http.StatusBadRequest, "BAD_REQUEST", msg)
}

// notFound записывает ответ 404 с ошибкой.
func notFound(w http.ResponseWriter, msg string) {
	errorResponse(w, http.StatusNotFound, "NOT_FOUND", msg)
}

// conflict записывает ответ 409 с ошибкой.
func conflict(w http.ResponseWriter, msg string) {
	errorResponse(w, http.StatusConflict, "CONFLICT", msg)
}

// internalError записывает ответ 500 с ошибкой.
func internalError(w http.ResponseWriter, msg string) {
	errorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", msg)
}

// jsonOK записывает JSON-ответ с данными.
func jsonOK(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// jsonOK204 записывает ответ 204 No Content.
func jsonOK204(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// parseGameID извлекает gameId из URL-параметра.
func parseGameID(r *http.Request) (int64, error) {
	s := chi.URLParam(r, "gameId")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// parseInstanceID извлекает instanceId из URL-параметра.
func parseInstanceID(r *http.Request) (int64, error) {
	s := chi.URLParam(r, "instanceId")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// parseNodeID извлекает nodeId из URL-параметра.
func parseNodeID(r *http.Request) (int64, error) {
	s := chi.URLParam(r, "nodeId")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// decodeJSON декодирует JSON-тело запроса в target.
func decodeJSON(r *http.Request, target any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(target)
}

// handleDomainError маппит доменные ошибки на HTTP-статусы.
func handleDomainError(w http.ResponseWriter, err error, action string) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		notFound(w, action+": resource not found")
	case errors.Is(err, domain.ErrAlreadyExists):
		conflict(w, action+": already exists")
	case errors.Is(err, domain.ErrBuildInUse):
		conflict(w, action+": "+err.Error())
	case errors.Is(err, domain.ErrInvalidToken):
		errorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", action+": invalid token")
	case errors.Is(err, domain.ErrNoAvailableNode):
		conflict(w, action+": "+err.Error())
	default:
		internalError(w, action+": "+err.Error())
	}
	return true
}

// buildStatusResponse преобразует доменный статус в строку для API.
func buildStatusResponse(s domain.InstanceStatus) string {
	return s.String()
}

// nodeStatusResponse преобразует доменный статус ноды в строку для API.
func nodeStatusResponse(s domain.NodeStatus) string {
	return s.String()
}

// protocolResponse преобразует доменный протокол в строку для API.
func protocolResponse(p domain.Protocol) string {
	return p.String()
}

// logSourceResponse преобразует источник лога в строку для API.
func logSourceResponse(s domain.LogSource) string {
	return s.String()
}

// parseInstanceStatus извлекает статус из query-параметра.
func parseInstanceStatus(r *http.Request) *domain.InstanceStatus {
	s := r.URL.Query().Get("status")
	if s == "" {
		return nil
	}
	switch s {
	case "starting":
		st := domain.InstanceStatusStarting
		return &st
	case "running":
		st := domain.InstanceStatusRunning
		return &st
	case "stopping":
		st := domain.InstanceStatusStopping
		return &st
	case "stopped":
		st := domain.InstanceStatusStopped
		return &st
	case "crashed":
		st := domain.InstanceStatusCrashed
		return &st
	default:
		return nil
	}
}

// parseNodeStatus извлекает статус ноды из query-параметра.
func parseNodeStatus(r *http.Request) *domain.NodeStatus {
	s := r.URL.Query().Get("status")
	if s == "" {
		return nil
	}
	switch s {
	case "unauthorized":
		st := domain.NodeStatusUnauthorized
		return &st
	case "online":
		st := domain.NodeStatusOnline
		return &st
	case "offline":
		st := domain.NodeStatusOffline
		return &st
	case "maintenance":
		st := domain.NodeStatusMaintenance
		return &st
	default:
		return nil
	}
}

// parseProtocol преобразует строку протокола в доменный тип.
func parseProtocol(s string) domain.Protocol {
	switch s {
	case "tcp":
		return domain.ProtocolTCP
	case "udp":
		return domain.ProtocolUDP
	case "websocket":
		return domain.ProtocolWebSocket
	case "webrtc":
		return domain.ProtocolWebRTC
	default:
		return 0
	}
}
