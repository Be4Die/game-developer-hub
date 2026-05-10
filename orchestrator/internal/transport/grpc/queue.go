package grpc

import (
	"context"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/service"
	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// QueueHandler реализует QueueService.
type QueueHandler struct {
	pb.UnimplementedQueueServiceServer
	queueService *service.QueueService
}

// NewQueueHandler создаёт обработчик очереди.
func NewQueueHandler(svc *service.QueueService) *QueueHandler {
	return &QueueHandler{queueService: svc}
}

// QueueServiceJoin добавляет игрока в очередь.
func (h *QueueHandler) QueueServiceJoin(ctx context.Context, req *pb.QueueServiceJoinRequest) (*pb.QueueServiceJoinResponse, error) {
	result, err := h.queueService.Join(ctx, req.GetGameId(), req.GetPlayerId(), req.GetMode())
	if err != nil {
		return nil, domainError(err, "queue join")
	}
	return queueStatusResultToProto(result), nil
}

// QueueServiceHeartbeat обновляет heartbeat и возвращает статус.
func (h *QueueHandler) QueueServiceHeartbeat(ctx context.Context, req *pb.QueueServiceHeartbeatRequest) (*pb.QueueServiceHeartbeatResponse, error) {
	result, err := h.queueService.Heartbeat(ctx, req.GetGameId(), req.GetPlayerId())
	if err != nil {
		return nil, domainError(err, "queue heartbeat")
	}
	return heartbeatResponseToProto(result), nil
}

// QueueServiceLeave удаляет игрока из очереди.
func (h *QueueHandler) QueueServiceLeave(ctx context.Context, req *pb.QueueServiceLeaveRequest) (*pb.QueueServiceLeaveResponse, error) {
	if err := h.queueService.Leave(ctx, req.GetGameId(), req.GetPlayerId()); err != nil {
		return nil, domainError(err, "queue leave")
	}
	return &pb.QueueServiceLeaveResponse{}, nil
}

// QueueServiceStatus возвращает статус без обновления heartbeat.
func (h *QueueHandler) QueueServiceStatus(ctx context.Context, req *pb.QueueServiceStatusRequest) (*pb.QueueServiceStatusResponse, error) {
	result, err := h.queueService.Status(ctx, req.GetGameId(), req.GetPlayerId())
	if err != nil {
		return nil, domainError(err, "queue status")
	}
	return statusResponseToProto(result), nil
}

func queueStatusResultToProto(r *service.QueueStatusResult) *pb.QueueServiceJoinResponse {
	return &pb.QueueServiceJoinResponse{
		Status:               queueStatusToProto(r.Status),
		Position:             r.Position,
		TotalInQueue:         r.TotalInQueue,
		EstimatedWaitSeconds: r.EstimatedWaitSeconds,
	}
}

func heartbeatResponseToProto(r *service.QueueStatusResult) *pb.QueueServiceHeartbeatResponse {
	resp := &pb.QueueServiceHeartbeatResponse{
		Status:               queueStatusToProto(r.Status),
		Position:             r.Position,
		TotalInQueue:         r.TotalInQueue,
		EstimatedWaitSeconds: r.EstimatedWaitSeconds,
	}
	if r.ReservedEndpoint != nil {
		resp.ReservedEndpoint = serverEndpointToProto(*r.ReservedEndpoint)
		v := r.ReservedUntil.Unix()
		resp.ReservedUntilUnix = &v
	}
	return resp
}

func statusResponseToProto(r *service.QueueStatusResult) *pb.QueueServiceStatusResponse {
	resp := &pb.QueueServiceStatusResponse{
		Status:               queueStatusToProto(r.Status),
		Position:             r.Position,
		TotalInQueue:         r.TotalInQueue,
		EstimatedWaitSeconds: r.EstimatedWaitSeconds,
	}
	if r.ReservedEndpoint != nil {
		resp.ReservedEndpoint = serverEndpointToProto(*r.ReservedEndpoint)
		v := r.ReservedUntil.Unix()
		resp.ReservedUntilUnix = &v
	}
	return resp
}
