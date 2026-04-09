package grpc

import (
	"errors"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/game_server_node/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// instanceToProto преобразует доменный инстанс в protobuf-сообщение.
func instanceToProto(inst *domain.Instance) *pb.Instance {
	p := &pb.Instance{
		InstanceId:       inst.ID,
		Name:             inst.Name,
		GameId:           inst.GameID,
		BuildVersion:     inst.BuildVersion,
		Port:             inst.Port,
		Protocol:         protocolToProto(inst.Protocol),
		Status:           statusToProto(inst.Status),
		MaxPlayers:       inst.MaxPlayers,
		DeveloperPayload: inst.DeveloperPayload,
		StartedAt:        timestamppb.New(inst.StartedAt),
	}

	if inst.PlayerCount != nil {
		p.PlayerCount = inst.PlayerCount
	}

	return p
}

// protocolToProto преобразует доменный Protocol в protobuf Protocol.
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

// statusToProto преобразует доменный InstanceStatus в protobuf InstanceStatus.
func statusToProto(s domain.InstanceStatus) pb.InstanceStatus {
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

// --- proto → domain ---

// protoToProtocol преобразует protobuf Protocol в доменный Protocol.
func protoToProtocol(p pb.Protocol) domain.Protocol {
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

// protoToPortStrategy преобразует protobuf PortAllocation в доменный PortStrategy.
// При nil возвращает стратегию Any.
func protoToPortStrategy(pa *pb.PortAllocation) domain.PortStrategy {
	if pa == nil {
		return domain.PortStrategy{Any: true}
	}

	switch s := pa.GetStrategy().(type) {
	case *pb.PortAllocation_Any:
		return domain.PortStrategy{Any: true}
	case *pb.PortAllocation_Exact:
		return domain.PortStrategy{Exact: s.Exact}
	case *pb.PortAllocation_Range:
		return domain.PortStrategy{
			Range: &domain.PortRange{
				Min: s.Range.GetMinPort(),
				Max: s.Range.GetMaxPort(),
			},
		}
	default:
		return domain.PortStrategy{Any: true}
	}
}

// domainErrToStatus преобразует доменную ошибку в gRPC status.
func domainErrToStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Errorf(codes.NotFound, "%v", err)
	case errors.Is(err, domain.ErrAlreadyExists):
		return status.Errorf(codes.AlreadyExists, "%v", err)
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
