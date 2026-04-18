// Package grpc реализует gRPC-транспорт оркестратора.
package grpc

import (
	"context"
	"errors"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ─── Converters: Domain -> Proto ─────────────────────────────────────────────

func buildToProto(b *domain.ServerBuild) *pb.ServerBuild {
	return &pb.ServerBuild{
		Id:            b.ID,
		OwnerId:       b.OwnerID,
		GameId:        b.GameID,
		BuildVersion:  b.Version,
		ImageTag:      b.ImageTag,
		Protocol:      protocolToProto(b.Protocol),
		InternalPort:  b.InternalPort,
		MaxPlayers:    b.MaxPlayers,
		FileSizeBytes: b.FileSize,
		CreatedAt:     timestamppb.New(b.CreatedAt),
	}
}

func instanceToProto(inst *domain.Instance) *pb.Instance {
	resp := &pb.Instance{
		Id:               inst.ID,
		OwnerId:          inst.OwnerID,
		GameId:           inst.GameID,
		NodeId:           inst.NodeID,
		BuildVersion:     inst.BuildVersion,
		Name:             inst.Name,
		Protocol:         protocolToProto(inst.Protocol),
		HostPort:         inst.HostPort,
		InternalPort:     inst.InternalPort,
		Status:           instanceStatusToProto(inst.Status),
		MaxPlayers:       inst.MaxPlayers,
		DeveloperPayload: inst.DeveloperPayload,
		ServerAddress:    inst.ServerAddress,
		CreatedAt:        timestamppb.New(inst.CreatedAt),
		UpdatedAt:        timestamppb.New(inst.UpdatedAt),
	}
	if inst.PlayerCount != nil {
		pc := *inst.PlayerCount
		resp.PlayerCount = &pc
	}
	if !inst.StartedAt.IsZero() {
		resp.StartedAt = timestamppb.New(inst.StartedAt)
	}
	return resp
}

func enrichedInstanceToProto(inst *service.EnrichedInstance) *pb.Instance {
	playerCount := inst.PlayerCount
	if playerCount == nil {
		playerCount = inst.Instance.PlayerCount
	}

	resp := &pb.Instance{
		Id:               inst.ID,
		OwnerId:          inst.OwnerID,
		GameId:           inst.GameID,
		NodeId:           inst.NodeID,
		BuildVersion:     inst.BuildVersion,
		Name:             inst.Name,
		Protocol:         protocolToProto(inst.Protocol),
		HostPort:         inst.HostPort,
		InternalPort:     inst.InternalPort,
		Status:           instanceStatusToProto(inst.Status),
		MaxPlayers:       inst.MaxPlayers,
		DeveloperPayload: inst.DeveloperPayload,
		ServerAddress:    inst.ServerAddress,
		CreatedAt:        timestamppb.New(inst.CreatedAt),
		UpdatedAt:        timestamppb.New(inst.UpdatedAt),
	}
	if playerCount != nil {
		pc := *playerCount
		resp.PlayerCount = &pc
	}
	if !inst.StartedAt.IsZero() {
		resp.StartedAt = timestamppb.New(inst.StartedAt)
	}
	return resp
}

func nodeToProto(n *domain.Node) *pb.Node {
	return &pb.Node{
		Id:               n.ID,
		OwnerId:          n.OwnerID,
		Address:          n.Address,
		Region:           n.Region,
		Status:           nodeStatusToProto(n.Status),
		CpuCores:         n.CPUCores,
		TotalMemoryBytes: n.TotalMemory,
		TotalDiskBytes:   n.TotalDisk,
		AgentVersion:     n.AgentVersion,
		CreatedAt:        timestamppb.New(n.CreatedAt),
		UpdatedAt:        timestamppb.New(n.UpdatedAt),
	}
}

func enrichedNodeToProto(n *service.EnrichedNode) *pb.Node {
	node := nodeToProto(&domain.Node{
		ID:           n.ID,
		OwnerID:      n.OwnerID,
		Address:      n.Address,
		Region:       n.Region,
		Status:       n.Status,
		CPUCores:     n.CPUCores,
		TotalMemory:  n.TotalMemory,
		TotalDisk:    n.TotalDisk,
		AgentVersion: n.AgentVersion,
		LastPingAt:   n.LastPingAt,
		CreatedAt:    n.CreatedAt,
		UpdatedAt:    n.UpdatedAt,
	})
	if !n.LastPingAt.IsZero() {
		node.LastPingAt = timestamppb.New(n.LastPingAt)
	}
	return node
}

func resourceUsageToProto(u *domain.ResourceUsage) *pb.ResourceUsage {
	return &pb.ResourceUsage{
		CpuUsagePercent:    u.CPUUsagePercent,
		MemoryUsedBytes:    u.MemoryUsedBytes,
		DiskUsedBytes:      u.DiskUsedBytes,
		NetworkBytesPerSec: u.NetworkBytesPerSec,
	}
}

func logEntryToProto(e *domain.LogEntry) *pb.LogEntry {
	return &pb.LogEntry{
		Timestamp: timestamppb.New(e.Timestamp),
		Source:    logSourceToProto(e.Source),
		Message:   e.Message,
	}
}

func serverEndpointToProto(ep domain.ServerEndpoint) *pb.ServerEndpoint {
	pc := ep.PlayerCount
	var playerCount uint32
	if pc != nil {
		playerCount = *pc
	}
	return &pb.ServerEndpoint{
		InstanceId:  ep.InstanceID,
		Address:     ep.Address,
		Port:        ep.Port,
		Protocol:    protocolToProto(ep.Protocol),
		PlayerCount: playerCount,
		MaxPlayers:  ep.MaxPlayers,
	}
}

// ─── Enum Converters ─────────────────────────────────────────────────────────

func protocolToProto(p domain.Protocol) pb.Protocol {
	switch p {
	case domain.ProtocolTCP:
		return pb.Protocol_PROTOCOL_TCP
	case domain.ProtocolUDP:
		return pb.Protocol_PROTOCOL_UDP
	case domain.ProtocolWebSocket:
		return pb.Protocol_PROTOCOL_WEBSOCKET
	case domain.ProtocolWebRTC:
		return pb.Protocol_PROTOCOL_WEBRTC
	default:
		return pb.Protocol_PROTOCOL_UNSPECIFIED
	}
}

func protocolFromProto(p pb.Protocol) domain.Protocol {
	switch p {
	case pb.Protocol_PROTOCOL_TCP:
		return domain.ProtocolTCP
	case pb.Protocol_PROTOCOL_UDP:
		return domain.ProtocolUDP
	case pb.Protocol_PROTOCOL_WEBSOCKET:
		return domain.ProtocolWebSocket
	case pb.Protocol_PROTOCOL_WEBRTC:
		return domain.ProtocolWebRTC
	default:
		return 0
	}
}

func instanceStatusToProto(s domain.InstanceStatus) pb.InstanceStatus {
	switch s {
	case domain.InstanceStatusStarting:
		return pb.InstanceStatus_INSTANCE_STATUS_STARTING
	case domain.InstanceStatusRunning:
		return pb.InstanceStatus_INSTANCE_STATUS_RUNNING
	case domain.InstanceStatusStopping:
		return pb.InstanceStatus_INSTANCE_STATUS_STOPPING
	case domain.InstanceStatusStopped:
		return pb.InstanceStatus_INSTANCE_STATUS_STOPPED
	case domain.InstanceStatusCrashed:
		return pb.InstanceStatus_INSTANCE_STATUS_CRASHED
	default:
		return pb.InstanceStatus_INSTANCE_STATUS_UNSPECIFIED
	}
}

func instanceStatusFromProto(s pb.InstanceStatus) domain.InstanceStatus {
	switch s {
	case pb.InstanceStatus_INSTANCE_STATUS_STARTING:
		return domain.InstanceStatusStarting
	case pb.InstanceStatus_INSTANCE_STATUS_RUNNING:
		return domain.InstanceStatusRunning
	case pb.InstanceStatus_INSTANCE_STATUS_STOPPING:
		return domain.InstanceStatusStopping
	case pb.InstanceStatus_INSTANCE_STATUS_STOPPED:
		return domain.InstanceStatusStopped
	case pb.InstanceStatus_INSTANCE_STATUS_CRASHED:
		return domain.InstanceStatusCrashed
	default:
		return 0
	}
}

func nodeStatusToProto(s domain.NodeStatus) pb.NodeStatus {
	switch s {
	case domain.NodeStatusUnauthorized:
		return pb.NodeStatus_NODE_STATUS_UNAUTHORIZED
	case domain.NodeStatusOnline:
		return pb.NodeStatus_NODE_STATUS_ONLINE
	case domain.NodeStatusOffline:
		return pb.NodeStatus_NODE_STATUS_OFFLINE
	case domain.NodeStatusMaintenance:
		return pb.NodeStatus_NODE_STATUS_MAINTENANCE
	default:
		return pb.NodeStatus_NODE_STATUS_UNSPECIFIED
	}
}

func logSourceToProto(s domain.LogSource) pb.LogSource {
	switch s {
	case domain.LogSourceStdout:
		return pb.LogSource_LOG_SOURCE_STDOUT
	case domain.LogSourceStderr:
		return pb.LogSource_LOG_SOURCE_STDERR
	default:
		return pb.LogSource_LOG_SOURCE_UNSPECIFIED
	}
}

func logSourceFromProto(s pb.LogSource) domain.LogSource {
	switch s {
	case pb.LogSource_LOG_SOURCE_STDOUT:
		return domain.LogSourceStdout
	case pb.LogSource_LOG_SOURCE_STDERR:
		return domain.LogSourceStderr
	default:
		return 0
	}
}

// domainError мапит доменные ошибки на gRPC статусы.
func domainError(err error, action string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, action+": resource not found")
	case errors.Is(err, domain.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, action+": already exists")
	case errors.Is(err, domain.ErrBuildInUse):
		return status.Error(codes.FailedPrecondition, action+": "+err.Error())
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, action+": invalid token")
	case errors.Is(err, domain.ErrNoAvailableNode):
		return status.Error(codes.ResourceExhausted, action+": "+err.Error())
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, action+": forbidden")
	default:
		return status.Error(codes.Internal, action+": "+err.Error())
	}
}

// _ используется для подавления unused import.
var _ = context.Background
