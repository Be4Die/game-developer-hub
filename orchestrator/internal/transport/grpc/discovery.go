package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// DiscoveryHandler реализует DiscoveryService.
type DiscoveryHandler struct {
	pb.UnimplementedDiscoveryServiceServer
	discoveryService *service.DiscoveryService
}

// NewDiscoveryHandler создаёт обработчик discovery.
func NewDiscoveryHandler(svc *service.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{discoveryService: svc}
}

// DiscoveryServiceDiscover возвращает доступные серверы для подключения.
func (h *DiscoveryHandler) DiscoveryServiceDiscover(ctx context.Context, req *pb.DiscoveryServiceDiscoverRequest) (*pb.DiscoveryServiceDiscoverResponse, error) {
	result, err := h.discoveryService.DiscoverServers(ctx, req.GetGameId(), req.GetPlayerId())
	if err != nil {
		return nil, domainError(err, "discover servers")
	}

	servers := make([]*pb.ServerEndpoint, 0, len(result.Servers))
	for _, ep := range result.Servers {
		servers = append(servers, serverEndpointToProto(ep))
	}

	return &pb.DiscoveryServiceDiscoverResponse{
		Servers: servers,
		Status:  discoveryStatusToProto(result.Status),
		Message: result.Message,
	}, nil
}
