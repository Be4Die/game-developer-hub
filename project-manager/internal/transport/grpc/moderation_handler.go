package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/project_manager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ModerationHandler реализует gRPC-сервис ModerationServiceServer.
type ModerationHandler struct {
	pb.UnimplementedModerationServiceServer
	svc        *service.ModerationService
	projectSvc *service.ProjectService
}

// NewModerationHandler создаёт новый экземпляр ModerationHandler.
func NewModerationHandler(svc *service.ModerationService, projectSvc *service.ProjectService) *ModerationHandler {
	return &ModerationHandler{
		svc:        svc,
		projectSvc: projectSvc,
	}
}

// ListTickets возвращает список тикетов модерации.
func (h *ModerationHandler) ListTickets(ctx context.Context, req *pb.ListModerationTicketsRequest) (*pb.ListModerationTicketsResponse, error) {
	var statusFilter *domain.ModerationStatus
	if req.GetStatus() != pb.ModerationStatus_MODERATION_STATUS_UNSPECIFIED {
		s := domain.ModerationStatus(req.GetStatus())
		statusFilter = &s
	}

	tickets, err := h.svc.ListTickets(ctx, statusFilter, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, domainError(err, "list moderation tickets")
	}

	resp := &pb.ListModerationTicketsResponse{Tickets: make([]*pb.ModerationTicket, len(tickets))}
	for i, t := range tickets {
		resp.Tickets[i] = ticketToProto(t)
	}
	return resp, nil
}

// GetTicket возвращает детальную информацию о тикете модерации и проекте.
func (h *ModerationHandler) GetTicket(ctx context.Context, req *pb.GetModerationTicketRequest) (*pb.GetModerationTicketResponse, error) {
	ticket, err := h.svc.GetTicket(ctx, req.GetTicketId())
	if err != nil {
		return nil, domainError(err, "get moderation ticket")
	}

	p, err := h.projectSvc.GetProject(ctx, ticket.ProjectID)
	if err != nil {
		return nil, domainError(err, "get project for ticket")
	}

	return &pb.GetModerationTicketResponse{
		Ticket:  ticketToProto(ticket),
		Project: projectToProto(p),
	}, nil
}

// GetTicketByProject возвращает тикет модерации по ID проекта.
func (h *ModerationHandler) GetTicketByProject(ctx context.Context, req *pb.GetModerationTicketByProjectRequest) (*pb.GetModerationTicketResponse, error) {
	ticket, err := h.svc.GetTicketByProject(ctx, req.GetProjectId())
	if err != nil {
		return nil, domainError(err, "get moderation ticket by project")
	}

	p, err := h.projectSvc.GetProject(ctx, ticket.ProjectID)
	if err != nil {
		return nil, domainError(err, "get project for ticket")
	}

	return &pb.GetModerationTicketResponse{
		Ticket:  ticketToProto(ticket),
		Project: projectToProto(p),
	}, nil
}

// Approve утверждает проект и развёртывает его в продуктивное окружение.
func (h *ModerationHandler) Approve(ctx context.Context, req *pb.ApproveModerationRequest) (*pb.ApproveModerationResponse, error) {
	moderatorID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing moderator id")
	}

	rel, err := h.svc.Approve(ctx, req.GetProjectId(), moderatorID, req.GetComment())
	if err != nil {
		return nil, domainError(err, "approve moderation")
	}

	return &pb.ApproveModerationResponse{
		Success: true,
		Release: releaseToProto(rel),
	}, nil
}

// Reject отклоняет проект модератором с указанием причин.
func (h *ModerationHandler) Reject(ctx context.Context, req *pb.RejectModerationRequest) (*pb.RejectModerationResponse, error) {
	moderatorID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing moderator id")
	}

	ticket, err := h.svc.Reject(ctx, req.GetProjectId(), moderatorID, req.GetReason())
	if err != nil {
		return nil, domainError(err, "reject moderation")
	}

	return &pb.RejectModerationResponse{
		Success: true,
		Ticket:  ticketToProto(ticket),
	}, nil
}
