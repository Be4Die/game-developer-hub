package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// NodeHandler реализует NodeService.
type NodeHandler struct {
	pb.UnimplementedNodeServiceServer
	nodeService *service.NodeService
}

// NewNodeHandler создаёт обработчик нод.
func NewNodeHandler(svc *service.NodeService) *NodeHandler {
	return &NodeHandler{nodeService: svc}
}

// Register подключает вычислительную ноду.
func (h *NodeHandler) Register(ctx context.Context, req *pb.RegisterNodeRequest) (*pb.RegisterNodeResponse, error) {
	params := service.RegisterNodeParams{}

	switch v := req.GetMode().(type) {
	case *pb.RegisterNodeRequest_Manual:
		params.Address = v.Manual.GetAddress()
		params.Token = v.Manual.GetToken()
		if v.Manual.Region != nil {
			params.Region = v.Manual.GetRegion()
		}
	case *pb.RegisterNodeRequest_Authorize:
		params.NodeID = ptrInt64(v.Authorize.GetNodeId())
		params.Token = v.Authorize.GetToken()
	}

	node, err := h.nodeService.RegisterNode(ctx, params)
	if err != nil {
		return nil, domainError(err, "register node")
	}

	return &pb.RegisterNodeResponse{Node: nodeToProto(node)}, nil
}

// List возвращает список всех нод.
func (h *NodeHandler) List(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	var statusFilter *domain.NodeStatus
	if req.Status != nil {
		s := nodeStatusFromProto(req.GetStatus())
		statusFilter = &s
	}

	nodes, err := h.nodeService.ListNodes(ctx, statusFilter)
	if err != nil {
		return nil, domainError(err, "list nodes")
	}

	resp := make([]*pb.Node, 0, len(nodes))
	for _, n := range nodes {
		resp = append(resp, enrichedNodeToProto(n))
	}

	return &pb.ListNodesResponse{Nodes: resp}, nil
}

// Get возвращает информацию о ноде.
func (h *NodeHandler) Get(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	node, err := h.nodeService.GetNode(ctx, req.GetNodeId())
	if err != nil {
		return nil, domainError(err, "get node")
	}

	return &pb.GetNodeResponse{Node: enrichedNodeToProto(node)}, nil
}

// Delete удаляет ноду.
func (h *NodeHandler) Delete(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeResponse, error) {
	err := h.nodeService.DeleteNode(ctx, req.GetNodeId())
	if err != nil {
		return nil, domainError(err, "delete node")
	}

	return &pb.DeleteNodeResponse{}, nil
}

// GetUsage возвращает потребление ресурсов ноды.
func (h *NodeHandler) GetUsage(ctx context.Context, req *pb.GetNodeUsageRequest) (*pb.GetNodeUsageResponse, error) {
	usage, err := h.nodeService.GetNodeUsage(ctx, req.GetNodeId())
	if err != nil {
		return nil, domainError(err, "get node usage")
	}

	return &pb.GetNodeUsageResponse{
		NodeId:              req.GetNodeId(),
		Usage:               resourceUsageToProto(usage.Usage),
		ActiveInstanceCount: int32(usage.ActiveInstanceCount), //nolint:gosec // count не превышает int32
	}, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func nodeStatusFromProto(s pb.NodeStatus) domain.NodeStatus {
	switch s {
	case pb.NodeStatus_NODE_STATUS_UNAUTHORIZED:
		return domain.NodeStatusUnauthorized
	case pb.NodeStatus_NODE_STATUS_ONLINE:
		return domain.NodeStatusOnline
	case pb.NodeStatus_NODE_STATUS_OFFLINE:
		return domain.NodeStatusOffline
	case pb.NodeStatus_NODE_STATUS_MAINTENANCE:
		return domain.NodeStatusMaintenance
	default:
		return 0
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}
