package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
)

// InstanceHandler обрабатывает HTTP-запросы для управления инстансами.
type InstanceHandler struct {
	instanceService *service.InstanceService
	maxLogTailLines uint32
}

// NewInstanceHandler создаёт обработчик инстансов.
func NewInstanceHandler(svc *service.InstanceService, maxLogTailLines uint32) *InstanceHandler {
	return &InstanceHandler{instanceService: svc, maxLogTailLines: maxLogTailLines}
}

// Start обрабатывает POST /games/{gameId}/instances.
func (h *InstanceHandler) Start(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	var req struct {
		BuildVersion     string             `json:"build_version"`
		ServerMode       string             `json:"server_mode"`
		Name             string             `json:"name"`
		PortAllocation   *portAllocReq      `json:"port_allocation"`
		ResourceLimits   *resourceLimitsReq `json:"resource_limits"`
		EnvVars          map[string]string  `json:"env_vars"`
		Args             []string           `json:"args"`
		DeveloperPayload map[string]string  `json:"developer_payload"`
		MaxPlayers       *uint32            `json:"max_players"`
	}
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	if req.BuildVersion == "" {
		badRequest(w, "missing 'build_version'")
		return
	}

	params := service.StartInstanceParams{
		GameID:           gameID,
		BuildVersion:     req.BuildVersion,
		Name:             req.Name,
		EnvVars:          req.EnvVars,
		Args:             req.Args,
		DeveloperPayload: req.DeveloperPayload,
		MaxPlayers:       req.MaxPlayers,
	}

	if req.PortAllocation != nil {
		params.PortAllocation = req.PortAllocation.toDomain()
	}
	if req.ResourceLimits != nil {
		params.ResourceLimits = req.ResourceLimits.toDomain()
	}

	instance, err := h.instanceService.StartInstance(r.Context(), params)
	if handleDomainError(w, err, "start instance") {
		return
	}

	jsonOK(w, instanceResponse(instance, h.instanceService), http.StatusCreated)
}

// Stop обрабатывает DELETE /games/{gameId}/instances/{instanceId}.
func (h *InstanceHandler) Stop(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	instanceID, err := parseInstanceID(r)
	if err != nil {
		badRequest(w, "invalid instanceId")
		return
	}

	timeoutSec := uint32(30)
	if t := r.URL.Query().Get("timeout"); t != "" {
		val, pErr := strconv.ParseUint(t, 10, 32)
		if pErr == nil {
			timeoutSec = uint32(val)
		}
	}

	instance, err := h.instanceService.StopInstance(r.Context(), gameID, instanceID, timeoutSec)
	if handleDomainError(w, err, "stop instance") {
		return
	}

	jsonOK(w, map[string]any{
		"instance": instanceResponse(instance, h.instanceService),
	}, http.StatusOK)
}

// List обрабатывает GET /games/{gameId}/instances.
func (h *InstanceHandler) List(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	status := parseInstanceStatus(r)

	instances, err := h.instanceService.ListInstances(r.Context(), gameID, status)
	if handleDomainError(w, err, "list instances") {
		return
	}

	resp := make([]any, 0, len(instances))
	for _, inst := range instances {
		resp = append(resp, enrichedInstanceResponse(inst))
	}

	jsonOK(w, map[string]any{"instances": resp}, http.StatusOK)
}

// Get обрабатывает GET /games/{gameId}/instances/{instanceId}.
func (h *InstanceHandler) Get(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	instanceID, err := parseInstanceID(r)
	if err != nil {
		badRequest(w, "invalid instanceId")
		return
	}

	instance, err := h.instanceService.GetInstance(r.Context(), gameID, instanceID)
	if handleDomainError(w, err, "get instance") {
		return
	}

	jsonOK(w, enrichedInstanceResponse(instance), http.StatusOK)
}

// StreamLogs обрабатывает GET /games/{gameId}/instances/{instanceId}/logs (SSE).
func (h *InstanceHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	instanceID, err := parseInstanceID(r)
	if err != nil {
		badRequest(w, "invalid instanceId")
		return
	}

	follow := r.URL.Query().Get("follow") == "true"
	tail := uint32(100)
	if t := r.URL.Query().Get("tail"); t != "" {
		val, pErr := strconv.ParseUint(t, 10, 32)
		if pErr == nil && val > 0 {
			tail = uint32(val)
		}
	}
	if tail > h.maxLogTailLines {
		tail = h.maxLogTailLines
	}

	var source *domain.LogSource
	if s := r.URL.Query().Get("source"); s != "" {
		switch s {
		case "stdout":
			st := domain.LogSourceStdout
			source = &st
		case "stderr":
			st := domain.LogSourceStderr
			source = &st
		}
	}

	stream, err := h.instanceService.StreamInstanceLogs(r.Context(), gameID, instanceID, domain.StreamLogsRequest{
		InstanceID:   instanceID,
		FollowStdout: follow,
		FollowStderr: follow,
		Tail:         tail,
	})
	if handleDomainError(w, err, "stream logs") {
		return
	}
	defer func() { _ = stream.Close() }()

	// SSE setup.
	flusher, ok := w.(http.Flusher)
	if !ok {
		internalError(w, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		default:
		}

		entry, err := stream.Recv()
		if err != nil {
			// Конец потока или ошибка — просто выходим.
			return
		}

		if source != nil && entry.Source != *source {
			continue
		}

		fmt.Fprintf(w, "event: log\ndata: {\"timestamp\":\"%s\",\"source\":\"%s\",\"message\":\"%s\"}\n\n", //nolint:errcheck,gosec
			entry.Timestamp.Format(time.RFC3339),
			logSourceResponse(entry.Source),
			escapeSSE(entry.Message),
		)
		flusher.Flush()
	}
}

// GetUsage обрабатывает GET /games/{gameId}/instances/{instanceId}/usage.
func (h *InstanceHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	gameID, err := parseGameID(r)
	if err != nil {
		badRequest(w, "invalid gameId")
		return
	}

	instanceID, err := parseInstanceID(r)
	if err != nil {
		badRequest(w, "invalid instanceId")
		return
	}

	usage, err := h.instanceService.GetInstanceUsage(r.Context(), gameID, instanceID)
	if handleDomainError(w, err, "get instance usage") {
		return
	}

	jsonOK(w, map[string]any{
		"instance_id": instanceID,
		"usage":       resourceUsageResponse(usage),
	}, http.StatusOK)
}

// enrichedInstanceResponse преобразует инстанс с обогащением в API-ответ.
func enrichedInstanceResponse(inst *service.EnrichedInstance) map[string]any {
	playerCount := inst.PlayerCount
	if playerCount == nil {
		pc := inst.Instance.PlayerCount
		playerCount = pc
	}

	return map[string]any{
		"id":                inst.ID,
		"game_id":           inst.GameID,
		"node_id":           inst.NodeID,
		"build_version":     inst.BuildVersion,
		"name":              inst.Name,
		"protocol":          protocolResponse(inst.Protocol),
		"host_port":         inst.HostPort,
		"internal_port":     inst.InternalPort,
		"status":            buildStatusResponse(inst.Status),
		"player_count":      playerCount,
		"max_players":       inst.MaxPlayers,
		"developer_payload": inst.DeveloperPayload,
		"server_address":    inst.ServerAddress,
		"started_at":        inst.StartedAt,
		"created_at":        inst.CreatedAt,
		"updated_at":        inst.UpdatedAt,
	}
}

// instanceResponse преобразует инстанс в API-ответ (без enrichment из KV).
func instanceResponse(inst *domain.Instance, _ *service.InstanceService) map[string]any {
	return map[string]any{
		"id":                inst.ID,
		"game_id":           inst.GameID,
		"node_id":           inst.NodeID,
		"build_version":     inst.BuildVersion,
		"name":              inst.Name,
		"protocol":          protocolResponse(inst.Protocol),
		"host_port":         inst.HostPort,
		"internal_port":     inst.InternalPort,
		"status":            buildStatusResponse(inst.Status),
		"player_count":      inst.PlayerCount,
		"max_players":       inst.MaxPlayers,
		"developer_payload": inst.DeveloperPayload,
		"server_address":    inst.ServerAddress,
		"started_at":        inst.StartedAt,
		"created_at":        inst.CreatedAt,
		"updated_at":        inst.UpdatedAt,
	}
}

// resourceUsageResponse преобразует доменную ResourceUsage в API-ответ.
func resourceUsageResponse(u *domain.ResourceUsage) map[string]any {
	return map[string]any{
		"cpu_usage_percent":     u.CPUUsagePercent,
		"memory_used_bytes":     u.MemoryUsedBytes,
		"disk_used_bytes":       u.DiskUsedBytes,
		"network_bytes_per_sec": u.NetworkBytesPerSec,
	}
}

// escapeSSE экранирует символы для SSE data-поля.
func escapeSSE(s string) string {
	result := strings.ReplaceAll(s, "\\", "\\\\")
	result = strings.ReplaceAll(result, "\"", "\\\"")
	result = strings.ReplaceAll(result, "\n", "\\n")
	return result
}

// portAllocReq — структура запроса port_allocation из JSON.
type portAllocReq struct {
	Strategy string  `json:"strategy"`
	Port     *uint32 `json:"port,omitempty"`
	MinPort  *uint32 `json:"min_port,omitempty"`
	MaxPort  *uint32 `json:"max_port,omitempty"`
}

func (p *portAllocReq) toDomain() domain.PortAllocation {
	var pa domain.PortAllocation
	switch p.Strategy {
	case "any":
		pa.Any = true
	case "exact":
		if p.Port != nil {
			pa.Exact = *p.Port
		}
	case "range":
		if p.MinPort != nil && p.MaxPort != nil {
			pa.Range = &domain.PortRange{Min: *p.MinPort, Max: *p.MaxPort}
		}
	}
	return pa
}

// resourceLimitsReq — структура запроса resource_limits из JSON.
type resourceLimitsReq struct {
	CPUMillis   *uint32 `json:"cpu_millis"`
	MemoryBytes *uint64 `json:"memory_bytes"`
}

func (r *resourceLimitsReq) toDomain() *domain.ResourceLimits {
	if r == nil {
		return nil
	}
	return &domain.ResourceLimits{
		CPUMillis:   r.CPUMillis,
		MemoryBytes: r.MemoryBytes,
	}
}
