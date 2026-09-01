package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DiscoveryHandler обрабатывает gRPC-запросы к сервису обнаружения.
type DiscoveryHandler struct {
	pb.UnimplementedDiscoveryServiceServer
	svc *service.DiscoveryService
}

// NewDiscoveryHandler создаёт обработчик для сервиса обнаружения.
func NewDiscoveryHandler(svc *service.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{svc: svc}
}

// GetNodeInfo возвращает информацию об узле.
func (h *DiscoveryHandler) GetNodeInfo(
	ctx context.Context, //nolint:revive // required by gRPC interface
	req *pb.GetNodeInfoRequest, //nolint:revive // required by gRPC interface
) (*pb.GetNodeInfoResponse, error) {
	node, err := h.svc.GetNode()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get node info: %v", err)
	}

	return &pb.GetNodeInfoResponse{
		Region:                      node.Region,
		CpuCores:                    node.Resources.CPUCores,
		TotalMemoryBytes:            node.Resources.TotalMemorySize,
		TotalDiskBytes:              node.Resources.TotalDiskSpace,
		NetworkBandwidthBytesPerSec: node.Resources.NetworkBandwidth,
		StartedAt:                   timestamppb.New(node.StartedAt),
		AgentVersion:                node.Version,
	}, nil
}

// Heartbeat возвращает текущую утилизацию ресурсов узла.
func (h *DiscoveryHandler) Heartbeat(
	ctx context.Context,
	req *pb.HeartbeatRequest, //nolint:revive // required by gRPC interface
) (*pb.HeartbeatResponse, error) {
	result, err := h.svc.Heartbeat(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "heartbeat failed: %v", err)
	}

	return &pb.HeartbeatResponse{
		Usage: &pb.ResourceUsage{
			CpuUsagePercent:    result.Usage.CPU,
			MemoryUsedBytes:    result.Usage.Memory,
			DiskUsedBytes:      result.Usage.Disk,
			NetworkBytesPerSec: result.Usage.Network,
		},
		ActiveInstanceCount: result.ActiveInstanceCount,
	}, nil
}

// ListInstances возвращает список всех инстансов на узле.
func (h *DiscoveryHandler) ListInstances(
	ctx context.Context,
	req *pb.ListInstancesRequest, //nolint:revive // required by gRPC interface
) (*pb.ListInstancesResponse, error) {
	instances, err := h.svc.GetAllInstances(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list instances: %v", err)
	}

	pbInstances := make([]*pb.Instance, 0, len(instances))
	for _, inst := range instances {
		pbInstances = append(pbInstances, instanceToProto(&inst))
	}

	return &pb.ListInstancesResponse{
		Instances: pbInstances,
	}, nil
}

// GetInstance возвращает инстанс по ID. Возвращает NotFound при отсутствии.
func (h *DiscoveryHandler) GetInstance(
	ctx context.Context,
	req *pb.GetInstanceRequest,
) (*pb.GetInstanceResponse, error) {
	instance, err := h.svc.GetInstance(ctx, req.GetInstanceId())
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.GetInstanceResponse{
		Instance: instanceToProto(instance),
	}, nil
}

// ListInstancesByGame возвращает инстансы указанной игры.
func (h *DiscoveryHandler) ListInstancesByGame(
	ctx context.Context,
	req *pb.ListInstancesByGameRequest,
) (*pb.ListInstancesByGameResponse, error) {
	instances, err := h.svc.GetInstancesByGameID(ctx, req.GetGameId())
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	pbInstances := make([]*pb.Instance, 0, len(instances))
	for _, inst := range instances {
		pbInstances = append(pbInstances, instanceToProto(&inst))
	}

	return &pb.ListInstancesByGameResponse{
		Instances: pbInstances,
	}, nil
}

// GetInstanceUsage возвращает метрики использования ресурсов инстанса.
func (h *DiscoveryHandler) GetInstanceUsage(
	ctx context.Context,
	req *pb.GetInstanceUsageRequest,
) (*pb.GetInstanceUsageResponse, error) {
	usage, err := h.svc.GetInstanceUsage(ctx, req.GetInstanceId())
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.GetInstanceUsageResponse{
		InstanceId: req.GetInstanceId(),
		Usage: &pb.ResourceUsage{
			CpuUsagePercent:    usage.CPU,
			MemoryUsedBytes:    usage.Memory,
			DiskUsedBytes:      usage.Disk,
			NetworkBytesPerSec: usage.Network,
		},
	}, nil
}
