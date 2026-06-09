package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/Be4Die/game-developer-hub/protos/chat/v1"
	"github.com/Be4Die/game-developer-hub/chat/internal/domain"
	"github.com/Be4Die/game-developer-hub/chat/internal/service"
)

func getUserIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		if val := ctx.Value("user_id"); val != nil {
			if s, ok := val.(string); ok {
				return s
			}
		}
		return ""
	}
	if vals := md.Get("x-user-id"); len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func getUserNameFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		if val := ctx.Value("user_name"); val != nil {
			if s, ok := val.(string); ok {
				return s
			}
		}
		return "User"
	}
	if vals := md.Get("x-user-name"); len(vals) > 0 {
		return vals[0]
	}
	return "User"
}

func getUserRoleFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		if val := ctx.Value("user_role"); val != nil {
			if s, ok := val.(string); ok {
				return s
			}
		}
		return ""
	}
	if vals := md.Get("x-user-role"); len(vals) > 0 {
		return vals[0]
	}
	return ""
}

type ChatHandler struct {
	pb.UnimplementedChatServiceServer
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	senderID := getUserIDFromContext(ctx)
	senderName := getUserNameFromContext(ctx)
	senderRole := getUserRoleFromContext(ctx)

	h.chatService.Log().Info("SendMessage", "senderID", senderID, "senderName", senderName, "senderRole", senderRole)

	if senderID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	msg, err := h.chatService.SendMessage(ctx, req.GetConversationId(), senderID, senderName, senderRole, req.GetContent())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SendMessageResponse{
		Message: convertMessage(msg),
	}, nil
}

func (h *ChatHandler) GetConversations(ctx context.Context, req *pb.GetConversationsRequest) (*pb.GetConversationsResponse, error) {
	userID := getUserIDFromContext(ctx)

	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	conversations, err := h.chatService.GetConversations(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbConversations := make([]*pb.Conversation, len(conversations))
	for i, conv := range conversations {
		pbConversations[i] = convertConversation(&conv, userID)
	}

	return &pb.GetConversationsResponse{
		Conversations: pbConversations,
	}, nil
}

func (h *ChatHandler) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	messages, total, err := h.chatService.GetMessages(ctx, req.GetConversationId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	h.chatService.Log().Info("GetMessages", "count", len(messages))
	for _, msg := range messages {
		h.chatService.Log().Info("message from DB", "id", msg.ID, "sender_role", msg.SenderRole)
	}

	pbMessages := make([]*pb.Message, len(messages))
	for i, msg := range messages {
		h.chatService.Log().Info("converting message", "id", msg.ID, "sender_role", msg.SenderRole)
		pbMessages[i] = convertMessage(&msg)
		h.chatService.Log().Info("converted message", "id", pbMessages[i].Id, "sender_role", pbMessages[i].SenderRole)
	}

	return &pb.GetMessagesResponse{
		Messages: pbMessages,
		Total:    int32(total),
	}, nil
}

func (h *ChatHandler) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	userID := getUserIDFromContext(ctx)

	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if err := h.chatService.MarkAsRead(ctx, req.GetConversationId(), userID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.MarkAsReadResponse{}, nil
}

func (h *ChatHandler) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	userID := getUserIDFromContext(ctx)

	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	count, err := h.chatService.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetUnreadCountResponse{
		TotalUnread: int32(count),
	}, nil
}

func (h *ChatHandler) CreateConversation(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error) {
	userID := getUserIDFromContext(ctx)
	userName := getUserNameFromContext(ctx)

	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	participantName := req.GetParticipantName()
	if participantName == "" {
		participantName = req.GetParticipantId()
	}

	conv, err := h.chatService.CreateConversation(ctx, userID, userName, req.GetParticipantId(), participantName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateConversationResponse{
		Conversation: convertConversation(conv, userID),
	}, nil
}

func convertConversation(conv *domain.Conversation, viewerID string) *pb.Conversation {
	participantID := conv.ParticipantID
	participantName := conv.ParticipantName
	if viewerID == conv.ParticipantID {
		participantID = conv.UserID
		participantName = conv.UserName
		if participantName == "" {
			participantName = "Разработчик"
		}
	}
	return &pb.Conversation{
		Id:              conv.ID,
		ParticipantId:   participantID,
		ParticipantName: participantName,
		LastMessage:     conv.LastMessage,
		LastMessageAt:   conv.LastMessageAt.Unix(),
		UnreadCount:     int32(conv.UnreadCount),
	}
}

func convertMessage(msg *domain.Message) *pb.Message {
	return &pb.Message{
		Id:             msg.ID,
		ConversationId: msg.ConversationID,
		SenderId:       msg.SenderID,
		SenderName:     msg.SenderName,
		SenderRole:     msg.SenderRole,
		Content:        msg.Content,
		CreatedAt:      msg.CreatedAt.Unix(),
		IsRead:         msg.IsRead,
	}
}
