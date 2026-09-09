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
	if isSuperuser(ctx) {
		ownerID = ""
	}

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
	if isSuperuser(ctx) {
		ownerID = ""
	}

	node, err := h.nodeService.GetNode(ctx, ownerID, req.GetNodeId())
	if err != nil {
		return nil, domainError(err, "get node")
	}

	return &pb.NodeServiceGetResponse{Node: enrichedNodeToProto(node)}, nil
}

// Delete удаляет ноду.
func (h *NodeHandler) Delete(ctx context.Context, req *pb.NodeServiceDeleteRequest) (*pb.NodeServiceDeleteResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	err := h.nodeService.DeleteNode(ctx, ownerID, req.GetNodeId())
	if err != nil {
		return nil, domainError(err, "delete node")
	}

	return &pb.NodeServiceDeleteResponse{}, nil
}

// GetUsage возвращает потребление ресурсов ноды.
func (h *NodeHandler) GetUsage(ctx context.Context, req *pb.NodeServiceGetUsageRequest) (*pb.NodeServiceGetUsageResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

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

// ListInstances возвращает список инстансов на ноде.
func (h *NodeHandler) ListInstances(ctx context.Context, req *pb.NodeServiceListInstancesRequest) (*pb.NodeServiceListInstancesResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	instances, err := h.nodeService.ListNodeInstances(ctx, ownerID, req.GetNodeId())
	if err != nil {
		return nil, domainError(err, "list node instances")
	}

	protos := make([]*pb.Instance, len(instances))
	for i, inst := range instances {
		protos[i] = instanceToProto(inst.Instance)
	}

	return &pb.NodeServiceListInstancesResponse{Instances: protos}, nil
}

// Announce обрабатывает анонсирование ноды от самой ноды.
// Не требует авторизации — нода сама заявляет о своем существовании.
func (h *NodeHandler) Announce(ctx context.Context, req *pb.NodeServiceAnnounceRequest) (*pb.NodeServiceAnnounceResponse, error) {
	params := service.AnnounceNodeParams{
		Address:          req.GetAddress(),
		AgentVersion:     req.GetAgentVersion(),
		CPUCores:         req.GetCpuCores(),
		TotalMemoryBytes: req.GetTotalMemoryBytes(),
		TotalDiskBytes:   req.GetTotalDiskBytes(),
		APIKey:           req.GetApiKey(),
	}

	if req.Region != nil {
		params.Region = *req.Region
	}

	result, err := h.nodeService.AnnounceNode(ctx, params)
	if err != nil {
		return nil, domainError(err, "announce node")
	}

	return &pb.NodeServiceAnnounceResponse{
		NodeId: result.NodeID,
	}, nil
}

// UpdateRole изменяет роль ноды (Mixed, Compute, Storage).
func (h *NodeHandler) UpdateRole(ctx context.Context, req *pb.NodeServiceUpdateRoleRequest) (*pb.NodeServiceUpdateRoleResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	node, err := h.nodeService.UpdateRole(ctx, ownerID, req.GetNodeId(), nodeRoleFromProto(req.GetRole()))
	if err != nil {
		return nil, domainError(err, "update node role")
	}

	return &pb.NodeServiceUpdateRoleResponse{Node: nodeToProto(node)}, nil
}

// CreateService разворачивает управляемый сервис хранения данных на ноде.
func (h *NodeHandler) CreateService(ctx context.Context, req *pb.NodeServiceCreateServiceRequest) (*pb.NodeServiceCreateServiceResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	params := service.CreateServiceParams{
		ServiceType:    serviceTypeFromProto(req.GetType()),
		Name:           req.GetName(),
		AllowedGameIDs: req.GetAllowedGameIds(),
		Password:       req.GetPassword(),
		DBName:         req.GetDbName(),
		Port:           req.GetPort(),
	}

	svc, err := h.nodeService.CreateService(ctx, ownerID, req.GetNodeId(), params)
	if err != nil {
		return nil, domainError(err, "create managed service")
	}

	return &pb.NodeServiceCreateServiceResponse{Service: managedServiceToProto(svc)}, nil
}

// ListServices возвращает список управляемых сервисов на ноде.
func (h *NodeHandler) ListServices(ctx context.Context, req *pb.NodeServiceListServicesRequest) (*pb.NodeServiceListServicesResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	var gameID *int64
	if req.GameId != nil {
		gameID = req.GameId
	}

	services, err := h.nodeService.ListServices(ctx, ownerID, req.GetNodeId(), gameID)
	if err != nil {
		return nil, domainError(err, "list managed services")
	}

	protoServices := make([]*pb.ManagedService, 0, len(services))
	for _, s := range services {
		protoServices = append(protoServices, managedServiceToProto(s))
	}

	return &pb.NodeServiceListServicesResponse{Services: protoServices}, nil
}

// DeleteService удаляет управляемый сервис.
func (h *NodeHandler) DeleteService(ctx context.Context, req *pb.NodeServiceDeleteServiceRequest) (*pb.NodeServiceDeleteServiceResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	if err := h.nodeService.DeleteService(ctx, ownerID, req.GetNodeId(), req.GetServiceId(), req.GetDeleteVolume()); err != nil {
		return nil, domainError(err, "delete managed service")
	}

	return &pb.NodeServiceDeleteServiceResponse{}, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func isSuperuser(ctx context.Context) bool {
	role, ok := GetUserRole(ctx)
	return ok && role == 3
}

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
