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
func (h *NodeHandler) Register(ctx context.Context, req *pb.NodeServiceRegisterRequest) (*pb.NodeServiceRegisterResponse, error) {
	ownerID, _ := GetUserID(ctx)

	params := service.RegisterNodeParams{
		OwnerID: ownerID,
	}

	switch v := req.GetMode().(type) {
	case *pb.NodeServiceRegisterRequest_Manual:
		params.Address = v.Manual.GetAddress()
		params.Token = v.Manual.GetToken()
		if v.Manual.Region != nil {
			params.Region = v.Manual.GetRegion()
		}
	case *pb.NodeServiceRegisterRequest_Authorize:
		params.NodeID = ptrInt64(v.Authorize.GetNodeId())
		params.Token = v.Authorize.GetToken()
	}

	node, err := h.nodeService.RegisterNode(ctx, params)
	if err != nil {
		return nil, domainError(err, "register node")
	}

	return &pb.NodeServiceRegisterResponse{Node: nodeToProto(node)}, nil
}

// List возвращает список всех нод пользователя.
func (h *NodeHandler) List(ctx context.Context, req *pb.NodeServiceListRequest) (*pb.NodeServiceListResponse, error) {
	ownerID, _ := GetUserID(ctx)

	var statusFilter *domain.NodeStatus
	if req.Status != nil {
		s := nodeStatusFromProto(req.GetStatus())
		statusFilter = &s
	}

	nodes, err := h.nodeService.ListNodes(ctx, ownerID, statusFilter)
	if err != nil {
		return nil, domainError(err, "list nodes")
	}

	resp := make([]*pb.Node, 0, len(nodes))
	for _, n := range nodes {
		resp = append(resp, enrichedNodeToProto(n))
	}

	return &pb.NodeServiceListResponse{Nodes: resp}, nil
}

// Get возвращает информацию о ноде.
func (h *NodeHandler) Get(ctx context.Context, req *pb.NodeServiceGetRequest) (*pb.NodeServiceGetResponse, error) {
	ownerID, _ := GetUserID(ctx)

	node, err := h.nodeService.GetNode(ctx, ownerID, req.GetNodeId())
	if err != nil {
		return nil, domainError(err, "get node")
	}

	return &pb.NodeServiceGetResponse{Node: enrichedNodeToProto(node)}, nil
}

// Delete удаляет ноду.
func (h *NodeHandler) Delete(ctx context.Context, req *pb.NodeServiceDeleteRequest) (*pb.NodeServiceDeleteResponse, error) {
	ownerID, _ := GetUserID(ctx)

	err := h.nodeService.DeleteNode(ctx, ownerID, req.GetNodeId())
	if err != nil {
		return nil, domainError(err, "delete node")
	}

	return &pb.NodeServiceDeleteResponse{}, nil
}

// GetUsage возвращает потребление ресурсов ноды.
func (h *NodeHandler) GetUsage(ctx context.Context, req *pb.NodeServiceGetUsageRequest) (*pb.NodeServiceGetUsageResponse, error) {
	ownerID, _ := GetUserID(ctx)

	usage, err := h.nodeService.GetNodeUsage(ctx, ownerID, req.GetNodeId())
	if err != nil {
		return nil, domainError(err, "get node usage")
	}

	return &pb.NodeServiceGetUsageResponse{
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
