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
	access        gameAuthorizer
}

// WithGameAccess подключает проверку доступа пользователя к игре.
func (h *GamePolicyHandler) WithGameAccess(c GameAccessChecker) *GamePolicyHandler {
	h.access = gameAuthorizer{checker: c}
	return h
}

// NewGamePolicyHandler создаёт обработчик политик.
func NewGamePolicyHandler(svc *service.GamePolicyService) *GamePolicyHandler {
	return &GamePolicyHandler{policyService: svc}
}

// Get возвращает политику игры.
func (h *GamePolicyHandler) Get(ctx context.Context, req *pb.GamePolicyServiceGetRequest) (*pb.GamePolicyServiceGetResponse, error) {
	if err := h.access.authorize(ctx, req.GetGameId()); err != nil {
		return nil, err
	}
	policy, err := h.policyService.Get(ctx, req.GetGameId())
	if err != nil {
		return nil, domainError(err, "get game policy")
	}

	return &pb.GamePolicyServiceGetResponse{Policy: gamePolicyToProto(policy)}, nil
}

// Set создаёт или обновляет политику игры.
func (h *GamePolicyHandler) Set(ctx context.Context, req *pb.GamePolicyServiceSetRequest) (*pb.GamePolicyServiceSetResponse, error) {
	if err := h.access.authorize(ctx, req.GetGameId()); err != nil {
		return nil, err
	}
	ownerID, _ := GetUserID(ctx)
	policy := gamePolicyFromProto(req)
	policy.OwnerID = ownerID
	policy, err := h.policyService.Set(ctx, policy)
	if err != nil {
		return nil, domainError(err, "set game policy")
	}

	return &pb.GamePolicyServiceSetResponse{Policy: gamePolicyToProto(policy)}, nil
}
