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

// Discover возвращает доступные серверы для подключения.
func (h *DiscoveryHandler) Discover(ctx context.Context, req *pb.DiscoveryServiceDiscoverRequest) (*pb.DiscoveryServiceDiscoverResponse, error) {
	endpoints, err := h.discoveryService.DiscoverServers(ctx, req.GetGameId())
	if err != nil {
		return nil, domainError(err, "discover servers")
	}

	resp := make([]*pb.ServerEndpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		resp = append(resp, serverEndpointToProto(ep))
	}

	return &pb.DiscoveryServiceDiscoverResponse{Servers: resp}, nil
}
