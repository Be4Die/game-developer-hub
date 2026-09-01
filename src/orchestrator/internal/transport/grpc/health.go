package grpc

import (
	"context"
	"time"

	pb "github.com/Be4Die/game-developer-hub/protos/orchestrator/v1"
)

// HealthHandler реализует HealthService.
type HealthHandler struct {
	pb.UnimplementedHealthServiceServer
	startTime time.Time
	version   string
}

// NewHealthHandler создаёт обработчик health check.
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   version,
	}
}

// Check возвращает статус работоспособности.
func (h *HealthHandler) Check(_ context.Context, _ *pb.HealthServiceCheckRequest) (*pb.HealthServiceCheckResponse, error) {
	uptime := time.Since(h.startTime).Seconds()

	return &pb.HealthServiceCheckResponse{
		Status:        "ok",
		Version:       h.version,
		UptimeSeconds: uptime,
	}, nil
}
