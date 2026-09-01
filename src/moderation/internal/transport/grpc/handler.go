package grpc

import (
	"context"
	"math"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/Be4Die/game-developer-hub/moderation/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/moderation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func clampInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

// ModerationHandler реализует gRPC-сервис ModerationServiceServer.
type ModerationHandler struct {
	pb.UnimplementedModerationServiceServer
	svc *service.ModerationService
}

// NewModerationHandler создаёт новый экземпляр ModerationHandler.
func NewModerationHandler(svc *service.ModerationService) *ModerationHandler {
	return &ModerationHandler{svc: svc}
}

// SubmitDraft принимает снимок черновика и создает заявку на модерацию.
func (h *ModerationHandler) SubmitDraft(ctx context.Context, req *pb.SubmitDraftRequest) (*pb.SubmitDraftResponse, error) {
	ownerID := req.GetOwnerId()
	if ownerID == "" {
		if uID, ok := UserIDFromContext(ctx); ok {
			ownerID = uID
		} else {
			return nil, status.Error(codes.Unauthenticated, "missing owner_id or user auth")
		}
	}

	snapshot := snapshotFromProto(req.GetSnapshot())
	modReq, err := h.svc.SubmitDraft(ctx, req.GetProjectId(), ownerID, snapshot)
	if err != nil {
		return nil, domainError(err, "submit draft")
	}

	return &pb.SubmitDraftResponse{
		Success: true,
		Request: requestToProto(modReq),
	}, nil
}

// ListRequests возвращает список заявок на модерацию с фильтрацией и пагинацией.
func (h *ModerationHandler) ListRequests(ctx context.Context, req *pb.ListModerationRequestsRequest) (*pb.ListModerationRequestsResponse, error) {
	filter := domain.RequestFilter{
		Limit:  int(req.GetLimit()),
		Offset: int(req.GetOffset()),
	}

	if req.GetStatus() != pb.RequestStatus_REQUEST_STATUS_UNSPECIFIED {
		rawStatus := int32(req.GetStatus())
		if rawStatus >= math.MinInt16 && rawStatus <= math.MaxInt16 {
			st := domain.RequestStatus(int16(rawStatus))
			filter.Status = &st
		}
	}

	if req.GetModeratorId() != "" {
		modID := req.GetModeratorId()
		filter.ModeratorID = &modID
	}

	requests, total, err := h.svc.ListRequests(ctx, filter)
	if err != nil {
		return nil, domainError(err, "list requests")
	}

	resp := &pb.ListModerationRequestsResponse{
		Requests: make([]*pb.ModerationRequest, len(requests)),
		Total:    clampInt32(total),
	}
	for i, r := range requests {
		resp.Requests[i] = requestToProto(r)
	}

	return resp, nil
}

// GetRequest возвращает заявку по ID.
func (h *ModerationHandler) GetRequest(ctx context.Context, req *pb.GetModerationRequestRequest) (*pb.GetModerationRequestResponse, error) {
	r, err := h.svc.GetRequest(ctx, req.GetRequestId())
	if err != nil {
		return nil, domainError(err, "get request")
	}
	return &pb.GetModerationRequestResponse{Request: requestToProto(r)}, nil
}

// GetLatestRequestByProject возвращает последнюю заявку проекта.
func (h *ModerationHandler) GetLatestRequestByProject(ctx context.Context, req *pb.GetLatestRequestByProjectRequest) (*pb.GetModerationRequestResponse, error) {
	r, err := h.svc.GetLatestRequestByProject(ctx, req.GetProjectId())
	if err != nil {
		return nil, domainError(err, "get latest request by project")
	}
	return &pb.GetModerationRequestResponse{Request: requestToProto(r)}, nil
}

// ClaimRequest закрепляет заявку за модератором.
func (h *ModerationHandler) ClaimRequest(ctx context.Context, req *pb.ClaimModerationRequestRequest) (*pb.ClaimModerationRequestResponse, error) {
	moderatorID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	r, err := h.svc.ClaimRequest(ctx, req.GetRequestId(), moderatorID)
	if err != nil {
		return nil, domainError(err, "claim request")
	}

	return &pb.ClaimModerationRequestResponse{Request: requestToProto(r)}, nil
}

// Approve утверждает проект и инициирует публикацию.
func (h *ModerationHandler) Approve(ctx context.Context, req *pb.ApproveModerationRequest) (*pb.ApproveModerationResponse, error) {
	moderatorID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	r, prodURL, err := h.svc.Approve(ctx, req.GetProjectId(), moderatorID, req.GetComment())
	if err != nil {
		return nil, domainError(err, "approve project")
	}

	return &pb.ApproveModerationResponse{
		Success: true,
		Request: requestToProto(r),
		ProdUrl: prodURL,
	}, nil
}

// Reject отклоняет заявку с указанием причины.
func (h *ModerationHandler) Reject(ctx context.Context, req *pb.RejectModerationRequest) (*pb.RejectModerationResponse, error) {
	moderatorID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	r, err := h.svc.Reject(ctx, req.GetProjectId(), moderatorID, req.GetReason())
	if err != nil {
		return nil, domainError(err, "reject project")
	}

	return &pb.RejectModerationResponse{
		Success: true,
		Request: requestToProto(r),
	}, nil
}

// SendMessage отправляет сообщение в чат проекта.
func (h *ModerationHandler) SendMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.SendChatMessageResponse, error) {
	senderID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	userRole := UserRoleFromContext(ctx)
	senderRole := domain.SenderRoleDeveloper
	if userRole == "moderator" || userRole == "admin" {
		senderRole = domain.SenderRoleModerator
	}

	msg, err := h.svc.SendMessage(ctx, req.GetProjectId(), senderID, senderRole, req.GetContent())
	if err != nil {
		return nil, domainError(err, "send chat message")
	}

	return &pb.SendChatMessageResponse{Message: messageToProto(msg)}, nil
}

// ListMessages возвращает историю сообщений чата проекта.
func (h *ModerationHandler) ListMessages(ctx context.Context, req *pb.ListChatMessagesRequest) (*pb.ListChatMessagesResponse, error) {
	messages, total, err := h.svc.ListMessages(ctx, req.GetProjectId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, domainError(err, "list chat messages")
	}

	resp := &pb.ListChatMessagesResponse{
		Messages: make([]*pb.ChatMessage, len(messages)),
		Total:    clampInt32(total),
	}
	for i, m := range messages {
		resp.Messages[i] = messageToProto(m)
	}

	return resp, nil
}

// ListActiveChats возвращает список активных чатов модерации.
func (h *ModerationHandler) ListActiveChats(ctx context.Context, req *pb.ListActiveChatsRequest) (*pb.ListActiveChatsResponse, error) {
	chats, total, err := h.svc.ListActiveChats(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list active chats: %v", err)
	}

	var pbChats []*pb.ChatSummary
	for _, c := range chats {
		pbChats = append(pbChats, &pb.ChatSummary{
			ProjectId:   c.ProjectID,
			LastMessage: messageToProto(c.LastMessage),
		})
	}

	return &pb.ListActiveChatsResponse{
		Chats: pbChats,
		Total: clampInt32(total),
	}, nil
}

// CloseDialog закрывает диалог модерации по проекту.
func (h *ModerationHandler) CloseDialog(ctx context.Context, req *pb.CloseDialogRequest) (*pb.CloseDialogResponse, error) {
	moderatorID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	userRole := UserRoleFromContext(ctx)
	if userRole != "moderator" && userRole != "admin" {
		return nil, status.Error(codes.PermissionDenied, "only moderators can close dialogs")
	}

	msg, err := h.svc.CloseDialog(ctx, req.GetProjectId(), moderatorID, req.GetComment())
	if err != nil {
		return nil, domainError(err, "close dialog")
	}

	return &pb.CloseDialogResponse{
		Success: true,
		Message: messageToProto(msg),
	}, nil
}
