package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// ─── Managed Service Backups ───────────────────────────────────────────────

// CreateServiceBackup инициирует создание резервной копии.
func (h *NodeHandler) CreateServiceBackup(ctx context.Context, req *pb.NodeServiceCreateServiceBackupRequest) (*pb.NodeServiceCreateServiceBackupResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	backup, err := h.nodeService.CreateServiceBackup(ctx, ownerID, req.GetNodeId(), req.GetServiceName())
	if err != nil {
		return nil, domainError(err, "create service backup")
	}

	return &pb.NodeServiceCreateServiceBackupResponse{Backup: serviceBackupToProto(backup)}, nil
}

// ListServiceBackups возвращает список резервных копий.
func (h *NodeHandler) ListServiceBackups(ctx context.Context, req *pb.NodeServiceListServiceBackupsRequest) (*pb.NodeServiceListServiceBackupsResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	backups, err := h.nodeService.ListServiceBackups(ctx, ownerID, req.GetNodeId(), req.GetServiceName())
	if err != nil {
		return nil, domainError(err, "list service backups")
	}

	protoBackups := make([]*pb.ServiceBackup, 0, len(backups))
	for _, b := range backups {
		protoBackups = append(protoBackups, serviceBackupToProto(b))
	}

	return &pb.NodeServiceListServiceBackupsResponse{Backups: protoBackups}, nil
}

// RestoreServiceBackup восстанавливает сервис из бэкапа.
func (h *NodeHandler) RestoreServiceBackup(ctx context.Context, req *pb.NodeServiceRestoreServiceBackupRequest) (*pb.NodeServiceRestoreServiceBackupResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	success, msg, err := h.nodeService.RestoreServiceBackup(ctx, ownerID, req.GetNodeId(), req.GetServiceName(), req.GetBackupId())
	if err != nil {
		return nil, domainError(err, "restore service backup")
	}

	return &pb.NodeServiceRestoreServiceBackupResponse{
		Success: success,
		Message: msg,
	}, nil
}

// DeleteServiceBackup удаляет бэкап.
func (h *NodeHandler) DeleteServiceBackup(ctx context.Context, req *pb.NodeServiceDeleteServiceBackupRequest) (*pb.NodeServiceDeleteServiceBackupResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	if err := h.nodeService.DeleteServiceBackup(ctx, ownerID, req.GetNodeId(), req.GetServiceName(), req.GetBackupId()); err != nil {
		return nil, domainError(err, "delete service backup")
	}

	return &pb.NodeServiceDeleteServiceBackupResponse{}, nil
}

// GetBackupTicket генерирует одноразовый защищенный тикет для скачивания бэкапа.
func (h *NodeHandler) GetBackupTicket(ctx context.Context, req *pb.NodeServiceGetBackupTicketRequest) (*pb.NodeServiceGetBackupTicketResponse, error) {
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	_, err := h.nodeService.ListServiceBackups(ctx, ownerID, req.GetNodeId(), req.GetServiceName())
	if err != nil {
		return nil, domainError(err, "get backup ticket authorization")
	}

	ticket := generateBackupTicket(req.GetNodeId(), req.GetServiceName(), req.GetBackupId(), 60*time.Second)
	return &pb.NodeServiceGetBackupTicketResponse{
		Ticket:           ticket,
		ExpiresInSeconds: 60,
	}, nil
}

// DownloadServiceBackup стримит файл бэкапа клиенту.
func (h *NodeHandler) DownloadServiceBackup(req *pb.NodeServiceDownloadServiceBackupRequest, stream pb.NodeService_DownloadServiceBackupServer) error {
	ctx := stream.Context()
	ownerID, hasUser := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	} else if !hasUser {
		// Проверяем тикет из метаданных контекста
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Errorf(codes.Unauthenticated, "authentication or ticket required")
		}
		tickets := md.Get("x-backup-ticket")
		if len(tickets) == 0 || !ValidateBackupTicket(tickets[0], req.GetNodeId(), req.GetServiceName(), req.GetBackupId()) {
			return status.Errorf(codes.PermissionDenied, "valid backup ticket or bearer token required")
		}
		// Тикет проверен, владелец был авторизован при генерации тикета
		ownerID = ""
	}

	reader, err := h.nodeService.DownloadServiceBackup(ctx, ownerID, req.GetNodeId(), req.GetServiceName(), req.GetBackupId())
	if err != nil {
		return domainError(err, "download service backup")
	}
	defer reader.Close()

	buf := make([]byte, 64*1024)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			if err := stream.Send(&pb.NodeServiceBackupChunk{Chunk: buf[:n]}); err != nil {
				return status.Errorf(codes.Internal, "stream backup chunk: %v", err)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return status.Errorf(codes.Internal, "read backup stream: %v", readErr)
		}
	}

	return nil
}

// UploadServiceBackup принимает стрим чанков бэкапа и сохраняет на ноде.
func (h *NodeHandler) UploadServiceBackup(stream pb.NodeService_UploadServiceBackupServer) error {
	ctx := stream.Context()
	ownerID, _ := GetUserID(ctx)
	if isSuperuser(ctx) {
		ownerID = ""
	}

	first, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "receive metadata: %v", err)
	}

	meta := first.GetMetadata()
	if meta == nil {
		return status.Errorf(codes.InvalidArgument, "first chunk must contain metadata")
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
			if data := chunk.GetChunk(); len(data) > 0 {
				if _, err := pw.Write(data); err != nil {
					return
				}
			}
		}
	}()

	backup, err := h.nodeService.UploadServiceBackup(ctx, ownerID, meta.GetNodeId(), meta.GetServiceName(), meta.GetFileName(), meta.GetRestoreImmediately(), pr)
	if err != nil {
		return domainError(err, "upload service backup")
	}

	return stream.SendAndClose(&pb.NodeServiceUploadBackupResponse{
		Backup:   serviceBackupToProto(backup),
		Restored: meta.GetRestoreImmediately(),
	})
}

const backupTicketSecret = "gdh-backup-secret-key-2026"

func generateBackupTicket(nodeID int64, serviceName, backupID string, ttl time.Duration) string {
	exp := time.Now().Add(ttl).Unix()
	payload := fmt.Sprintf("%d:%s:%s:%d", nodeID, serviceName, backupID, exp)
	mac := hmac.New(sha256.New, []byte(backupTicketSecret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s:%s", payload, sig)
}

// ValidateBackupTicket проверяет подпись и срок действия тикета на скачивание.
func ValidateBackupTicket(ticket string, nodeID int64, serviceName, backupID string) bool {
	parts := strings.Split(ticket, ":")
	if len(parts) != 5 {
		return false
	}
	tNodeID, _ := strconv.ParseInt(parts[0], 10, 64)
	tServiceName := parts[1]
	tBackupID := parts[2]
	exp, _ := strconv.ParseInt(parts[3], 10, 64)
	sig := parts[4]

	if tNodeID != nodeID || tServiceName != serviceName || tBackupID != backupID {
		return false
	}
	if time.Now().Unix() > exp {
		return false
	}

	payload := fmt.Sprintf("%d:%s:%s:%d", tNodeID, tServiceName, tBackupID, exp)
	mac := hmac.New(sha256.New, []byte(backupTicketSecret))
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

func serviceBackupToProto(b *domain.ServiceBackup) *pb.ServiceBackup {
	if b == nil {
		return nil
	}
	return &pb.ServiceBackup{
		Id:          b.ID,
		BackupId:    b.BackupID,
		NodeId:      b.NodeID,
		ServiceId:   b.ServiceID,
		ServiceName: b.ServiceName,
		ServiceType: pb.ServiceType(b.ServiceType),
		FileName:    b.FileName,
		SizeBytes:   int64(b.SizeBytes),
		Checksum:    b.Checksum,
		BackupType:  pb.BackupType(b.BackupType),
		Status:      pb.BackupStatus(b.Status),
		CreatedAt:   timestamppb.New(b.CreatedAt),
	}
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

func (h *NodeHandler) ToggleBackups(ctx context.Context, req *pb.NodeServiceToggleBackupsRequest) (*pb.NodeServiceToggleBackupsResponse, error) {
	callerID, ok := GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	if isSuperuser(ctx) { callerID = "" }
	node, err := h.nodeService.ToggleBackups(ctx, callerID, req.GetNodeId(), req.GetEnabled())
	if err != nil {
		return nil, domainError(err, "toggle backups")
	}
	return &pb.NodeServiceToggleBackupsResponse{Node: nodeToProto(node)}, nil
}

func (h *NodeHandler) ToggleServiceAutoBackup(ctx context.Context, req *pb.NodeServiceToggleServiceAutoBackupRequest) (*pb.NodeServiceToggleServiceAutoBackupResponse, error) {
	callerID, ok := GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	if isSuperuser(ctx) { callerID = "" }
	svc, err := h.nodeService.ToggleServiceAutoBackup(ctx, callerID, req.GetNodeId(), req.GetServiceName(), req.GetEnabled())
	if err != nil {
		return nil, domainError(err, "toggle auto backup")
	}
	return &pb.NodeServiceToggleServiceAutoBackupResponse{Service: managedServiceToProto(svc)}, nil
}
