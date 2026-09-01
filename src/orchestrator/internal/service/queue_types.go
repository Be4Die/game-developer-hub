package service

import (
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// QueueStatusResult — результат операции с очередью.
type QueueStatusResult struct {
	Status               domain.QueueStatus
	Position             int32
	TotalInQueue         int32
	EstimatedWaitSeconds int32
	ReservedEndpoint     *domain.ServerEndpoint
	ReservedUntil        time.Time
}
