package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DiscoveryHandler struct {
	pb.UnimplementedDiscoveryServiceServer
	svc *service.DiscoveryService
}

func NewDiscoveryHandler(svc *service.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{svc: svc}
}

func (h *DiscoveryHandler) GetNodeInfo(
	ctx context.Context,
	req *pb.GetNodeInfoRequest,
) (*pb.GetNodeInfoResponse, error) {
	// 1. Вызываем service (domain мир).
	node, err := h.svc.GetNode()
	if err != nil {
		// Конвертируем Go-ошибку в gRPC-ошибку.
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

func (h *DiscoveryHandler) Heartbeat(
	ctx context.Context,
	req *pb.HeartbeatRequest,
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

func (h *DiscoveryHandler) ListInstances(
	ctx context.Context,
	req *pb.ListInstancesRequest,
) (*pb.ListInstancesResponse, error) {
	instances, err := h.svc.GetAllInstances(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list instances: %v", err)
	}

	// Конвертируем слайс domain → слайс proto.
	pbInstances := make([]*pb.Instance, 0, len(instances))
	for _, inst := range instances {
		pbInstances = append(pbInstances, instanceToProto(&inst))
	}

	return &pb.ListInstancesResponse{
		Instances: pbInstances,
	}, nil
}

func (h *DiscoveryHandler) GetInstance(
	ctx context.Context,
	req *pb.GetInstanceRequest,
) (*pb.GetInstanceResponse, error) {
	instance, err := h.svc.GetInstance(ctx, req.GetInstanceId())
	if err != nil {
		// Проверяем тип ошибки и возвращаем правильный gRPC код.
		return nil, domainErrToStatus(err)
	}

	return &pb.GetInstanceResponse{
		Instance: instanceToProto(instance),
	}, nil
}

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
