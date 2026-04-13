package http

import (
	"net/http"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
)

// DiscoveryHandler обрабатывает HTTP-запросы для discovery-эндпоинтов.
type DiscoveryHandler struct {
	discoveryService *service.DiscoveryService
}

// NewDiscoveryHandler создаёт обработчик discovery.
func NewDiscoveryHandler(svc *service.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{discoveryService: svc}
}

// Discover обрабатывает GET /games/{gameId}/servers.
func (h *DiscoveryHandler) Discover(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	endpoints, err := h.discoveryService.DiscoverServers(r.Context(), gameID)
	if handleDomainError(w, err, "discover servers") {
		return
	}

	resp := make([]any, 0, len(endpoints))
	for _, ep := range endpoints {
		pc := ep.PlayerCount
		resp = append(resp, map[string]any{
			"instance_id":  ep.InstanceID,
			"address":      ep.Address,
			"port":         ep.Port,
			"protocol":     protocolResponse(ep.Protocol),
			"player_count": pc,
			"max_players":  ep.MaxPlayers,
		})
	}

	jsonOK(w, map[string]any{"servers": resp}, http.StatusOK)
}
