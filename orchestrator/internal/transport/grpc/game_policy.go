package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// GamePolicyHandler реализует GamePolicyService.
type GamePolicyHandler struct {
	pb.UnimplementedGamePolicyServiceServer
	policyService *service.GamePolicyService
}

// NewGamePolicyHandler создаёт обработчик политик.
func NewGamePolicyHandler(svc *service.GamePolicyService) *GamePolicyHandler {
	return &GamePolicyHandler{policyService: svc}
}

// Get возвращает политику игры.
func (h *GamePolicyHandler) Get(ctx context.Context, req *pb.GamePolicyServiceGetRequest) (*pb.GamePolicyServiceGetResponse, error) {
	policy, err := h.policyService.Get(ctx, req.GetGameId())
	if err != nil {
		return nil, domainError(err, "get game policy")
	}

	return &pb.GamePolicyServiceGetResponse{Policy: gamePolicyToProto(policy)}, nil
}

// Set создаёт или обновляет политику игры.
func (h *GamePolicyHandler) Set(ctx context.Context, req *pb.GamePolicyServiceSetRequest) (*pb.GamePolicyServiceSetResponse, error) {
	policy := gamePolicyFromProto(req)
	policy, err := h.policyService.Set(ctx, policy)
	if err != nil {
		return nil, domainError(err, "set game policy")
	}

	return &pb.GamePolicyServiceSetResponse{Policy: gamePolicyToProto(policy)}, nil
}
