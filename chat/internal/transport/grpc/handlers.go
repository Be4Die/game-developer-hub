package transport

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Be4Die/game-developer-hub/chat/internal/app"
	"github.com/Be4Die/game-developer-hub/chat/internal/domain"
	pb "github.com/Be4Die/game-developer-hub/protos/chat/v1"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	service *app.ChatService
}

func NewChatServer(service *app.ChatService) *ChatServer {
	return &ChatServer{service: service}
}

func (s *ChatServer) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	chatType := domain.ChatTypeFromProto(req.Type)
	chat, err := s.service.CreateChat(ctx, userID, chatType, req.Title, req.ParticipantIds)
	if err != nil {
		log.Printf("ERROR CreateChat: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateChatResponse{Chat: chat.ToProto()}, nil
}

func (s *ChatServer) GetChat(ctx context.Context, req *pb.GetChatRequest) (*pb.GetChatResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	chat, err := s.service.GetChat(ctx, userID, req.ChatId)
	if err != nil {
		if err == domain.ErrChatNotFound {
			return nil, status.Error(codes.NotFound, "chat not found")
		}
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetChatResponse{Chat: chat.ToProto()}, nil
}

func (s *ChatServer) ListChats(ctx context.Context, req *pb.ListChatsRequest) (*pb.ListChatsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	chats, err := s.service.ListChats(ctx, userID, int(req.PageSize), 0)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbChats []*pb.Chat
	for _, c := range chats {
		pbChats = append(pbChats, c.ToProto())
	}

	return &pb.ListChatsResponse{
		Chats:         pbChats,
		NextPageToken: "",
	}, nil
}

func (s *ChatServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	msg, err := s.service.SendMessage(ctx, userID, req.ChatId, req.Content)
	if err != nil {
		if err == domain.ErrChatNotFound {
			return nil, status.Error(codes.NotFound, "chat not found")
		}
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SendMessageResponse{Message: msg.ToProto()}, nil
}

func (s *ChatServer) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	messages, err := s.service.GetMessages(ctx, userID, req.ChatId, int(req.PageSize), 0)
	if err != nil {
		if err == domain.ErrChatNotFound {
			return nil, status.Error(codes.NotFound, "chat not found")
		}
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbMessages []*pb.Message
	for _, m := range messages {
		pbMessages = append(pbMessages, m.ToProto())
	}

	return &pb.GetMessagesResponse{
		Messages:      pbMessages,
		NextPageToken: "",
	}, nil
}

func (s *ChatServer) AddParticipant(ctx context.Context, req *pb.AddParticipantRequest) (*pb.AddParticipantResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	chat, err := s.service.AddParticipant(ctx, userID, req.ChatId, req.UserId)
	if err != nil {
		if err == domain.ErrChatNotFound {
			return nil, status.Error(codes.NotFound, "chat not found")
		}
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddParticipantResponse{Chat: chat.ToProto()}, nil
}

func (s *ChatServer) RemoveParticipant(ctx context.Context, req *pb.RemoveParticipantRequest) (*pb.RemoveParticipantResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	chat, err := s.service.RemoveParticipant(ctx, userID, req.ChatId, req.UserId)
	if err != nil {
		if err == domain.ErrChatNotFound {
			return nil, status.Error(codes.NotFound, "chat not found")
		}
		if err == domain.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RemoveParticipantResponse{Chat: chat.ToProto()}, nil
}
