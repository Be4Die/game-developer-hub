package domain

import "context"

// NodeStateStore хранит горячее состояние нод в KV (Valkey/Redis).
// Эфемерные данные: heartbeat, usage. TTL обновляется при каждом ping.
type NodeStateStore interface {
	// UpdateHeartbeat обновляет время последнего ping ноды и сбрасывает TTL.
	UpdateHeartbeat(ctx context.Context, nodeID int64, usage *ResourceUsage) error

	// GetUsage возвращает текущую загруженность ноды. ErrNotFound если ключ истёк.
	GetUsage(ctx context.Context, nodeID int64) (*ResourceUsage, error)

	// GetActiveInstanceCount возвращает количество активных инстансов на ноде.
	GetActiveInstanceCount(ctx context.Context, nodeID int64) (uint32, error)

	// SetActiveInstanceCount обновляет счётчик активных инстансов.
	SetActiveInstanceCount(ctx context.Context, nodeID int64, count uint32) error

	// Delete удаляет все ключи состояния ноды.
	Delete(ctx context.Context, nodeID int64) error
}
