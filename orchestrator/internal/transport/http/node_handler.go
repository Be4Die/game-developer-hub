package http

import (
	"encoding/json"
	"net/http"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
)

// NodeHandler обрабатывает HTTP-запросы для управления нодами.
type NodeHandler struct {
	nodeService *service.NodeService
}

// NewNodeHandler создаёт обработчик нод.
func NewNodeHandler(svc *service.NodeService) *NodeHandler {
	return &NodeHandler{nodeService: svc}
}

// Register обрабатывает POST /nodes.
func (h *NodeHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req json.RawMessage
	if err := decodeJSON(r, &req); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	// Пробуем ручной режим (с address).
	var manualReq struct {
		Address string `json:"address"`
		Token   string `json:"token"`
		Region  string `json:"region"`
		NodeID  *int64 `json:"node_id"`
	}
	if err := json.Unmarshal(req, &manualReq); err != nil {
		badRequest(w, "invalid request body: "+err.Error())
		return
	}

	params := service.RegisterNodeParams{
		Address: manualReq.Address,
		Token:   manualReq.Token,
		Region:  manualReq.Region,
		NodeID:  manualReq.NodeID,
	}

	node, err := h.nodeService.RegisterNode(r.Context(), params)
	if handleDomainError(w, err, "register node") {
		return
	}

	jsonOK(w, nodeResponse(node, h.nodeService), http.StatusCreated)
}

// List обрабатывает GET /nodes.
func (h *NodeHandler) List(w http.ResponseWriter, r *http.Request) {
	status := parseNodeStatus(r)

	nodes, err := h.nodeService.ListNodes(r.Context(), status)
	if handleDomainError(w, err, "list nodes") {
		return
	}

	resp := make([]any, 0, len(nodes))
	for _, n := range nodes {
		resp = append(resp, enrichedNodeResponse(n))
	}

	jsonOK(w, map[string]any{"nodes": resp}, http.StatusOK)
}

// Get обрабатывает GET /nodes/{nodeId}.
func (h *NodeHandler) Get(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		badRequest(w, "invalid nodeId")
		return
	}

	node, err := h.nodeService.GetNode(r.Context(), nodeID)
	if handleDomainError(w, err, "get node") {
		return
	}

	jsonOK(w, enrichedNodeResponse(node), http.StatusOK)
}

// Delete обрабатывает DELETE /nodes/{nodeId}.
func (h *NodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		badRequest(w, "invalid nodeId")
		return
	}

	err = h.nodeService.DeleteNode(r.Context(), nodeID)
	if handleDomainError(w, err, "delete node") {
		return
	}

	jsonOK204(w)
}

// GetUsage обрабатывает GET /nodes/{nodeId}/usage.
func (h *NodeHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		badRequest(w, "invalid nodeId")
		return
	}

	usage, err := h.nodeService.GetNodeUsage(r.Context(), nodeID)
	if handleDomainError(w, err, "get node usage") {
		return
	}

	activeCount := usage.ActiveInstanceCount

	jsonOK(w, map[string]any{
		"node_id":               nodeID,
		"usage":                 resourceUsageResponse(usage.Usage),
		"active_instance_count": activeCount,
	}, http.StatusOK)
}

// enrichedNodeResponse преобразует ноду с обогащением в API-ответ.
func enrichedNodeResponse(n *service.EnrichedNode) map[string]any {
	return map[string]any{
		"id":                    n.ID,
		"address":               n.Address,
		"region":                n.Region,
		"status":                nodeStatusResponse(n.Status),
		"cpu_cores":             n.CPUCores,
		"total_memory_bytes":    n.TotalMemory,
		"total_disk_bytes":      n.TotalDisk,
		"agent_version":         n.AgentVersion,
		"last_ping_at":          n.LastPingAt,
		"created_at":            n.CreatedAt,
		"updated_at":            n.UpdatedAt,
		"usage":                 resourceUsageResponse(n.Usage),
		"active_instance_count": n.ActiveInstanceCount,
	}
}

// nodeResponse преобразует ноду в API-ответ (без enrichment).
func nodeResponse(n *domain.Node, _ *service.NodeService) map[string]any {
	return map[string]any{
		"id":                 n.ID,
		"address":            n.Address,
		"region":             n.Region,
		"status":             nodeStatusResponse(n.Status),
		"cpu_cores":          n.CPUCores,
		"total_memory_bytes": n.TotalMemory,
		"total_disk_bytes":   n.TotalDisk,
		"agent_version":      n.AgentVersion,
		"last_ping_at":       n.LastPingAt,
		"created_at":         n.CreatedAt,
		"updated_at":         n.UpdatedAt,
	}
}
