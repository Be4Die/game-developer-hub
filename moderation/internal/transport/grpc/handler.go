package grpc

import (
	"context"
	"errors"

	pb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/Be4Die/game-developer-hub/moderation/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ModerationHandler struct {
	pb.UnimplementedModerationServiceServer
	service *service.ModerationService
}

func NewModerationHandler(svc *service.ModerationService) *ModerationHandler {
	return &ModerationHandler{service: svc}
}

func getUserIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if vals := md.Get("x-user-id"); len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (h *ModerationHandler) SubmitForReview(ctx context.Context, req *pb.SubmitForReviewRequest) (*pb.SubmitForReviewResponse, error) {
	developerID := getUserIDFromContext(ctx)
	if developerID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	moderation, err := h.service.SubmitForReview(ctx, req.GetGameId(), developerID, req.GetGameName(), req.GetGameDescription())
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyUnderReview) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.SubmitForReviewResponse{
		Moderation: convertToPB(moderation),
	}, nil
}

func (h *ModerationHandler) GetModerationStatus(ctx context.Context, req *pb.GetModerationStatusRequest) (*pb.GetModerationStatusResponse, error) {
	moderation, err := h.service.GetModerationStatus(ctx, req.GetGameId())
	if err != nil {
		if errors.Is(err, domain.ErrModerationNotFound) {
			return nil, status.Error(codes.NotFound, "moderation not found")
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.GetModerationStatusResponse{
		Moderation: convertToPB(moderation),
	}, nil
}

func (h *ModerationHandler) ApproveGame(ctx context.Context, req *pb.ApproveGameRequest) (*pb.ApproveGameResponse, error) {
	moderatorID := getUserIDFromContext(ctx)
	if moderatorID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	moderation, err := h.service.ApproveGame(ctx, req.GetGameId(), moderatorID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.ApproveGameResponse{
		Moderation: convertToPB(moderation),
	}, nil
}

func (h *ModerationHandler) RejectGame(ctx context.Context, req *pb.RejectGameRequest) (*pb.RejectGameResponse, error) {
	moderatorID := getUserIDFromContext(ctx)
	if moderatorID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	moderation, err := h.service.RejectGame(ctx, req.GetGameId(), moderatorID, req.GetReason())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.RejectGameResponse{
		Moderation: convertToPB(moderation),
	}, nil
}

func (h *ModerationHandler) ListPendingGames(ctx context.Context, req *pb.ListPendingGamesRequest) (*pb.ListPendingGamesResponse, error) {
	moderations, total, err := h.service.ListPendingGames(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	pbModerations := make([]*pb.GameModeration, len(moderations))
	for i, m := range moderations {
		pbModerations[i] = convertToPB(&m)
	}

	return &pb.ListPendingGamesResponse{
		Moderations: pbModerations,
		TotalCount:  int32(total),
	}, nil
}

func convertToPB(m *domain.GameModeration) *pb.GameModeration {
	var status pb.ModerationStatus
	switch m.Status {
	case domain.ModerationStatusPending:
		status = pb.ModerationStatus_MODERATION_STATUS_PENDING
	case domain.ModerationStatusApproved:
		status = pb.ModerationStatus_MODERATION_STATUS_APPROVED
	case domain.ModerationStatusRejected:
		status = pb.ModerationStatus_MODERATION_STATUS_REJECTED
	}

	var reviewedAt *timestamppb.Timestamp
	if m.ReviewedAt != nil {
		reviewedAt = timestamppb.New(*m.ReviewedAt)
	}

	return &pb.GameModeration{
		Id:              m.ID,
		GameId:          m.GameID,
		DeveloperId:     m.DeveloperID,
		GameName:        m.GameName,
		GameDescription: m.GameDescription,
		ModeratorId:     m.ModeratorID,
		Status:          status,
		RejectionReason: m.RejectionReason,
		SubmittedAt:     timestamppb.New(m.SubmittedAt),
		ReviewedAt:      reviewedAt,
	}
}
