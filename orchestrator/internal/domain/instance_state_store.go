package domain

import "context"

// InstanceStateStore хранит горячее состояние инстансов в KV (Valkey/Redis).
// Эфемерные данные: текущий статус, player_count, resource_usage.
// TTL обновляется при каждом heartbeat, просроченные ключи автоматически удаляются.
type InstanceStateStore interface {
	// SetStatus обновляет статус инстанса и сбрасывает TTL.
	SetStatus(ctx context.Context, instanceID int64, status InstanceStatus) error

	// GetStatus возвращает текущий статус инстанса. ErrNotFound если ключ истёк.
	GetStatus(ctx context.Context, instanceID int64) (InstanceStatus, error)

	// SetPlayerCount обновляет количество игроков и сбрасывает TTL.
	SetPlayerCount(ctx context.Context, instanceID int64, count uint32) error

	// GetPlayerCount возвращает текущее количество игроков.
	GetPlayerCount(ctx context.Context, instanceID int64) (uint32, error)

	// SetUsage обновляет метрики потребления ресурсов и сбрасывает TTL.
	SetUsage(ctx context.Context, instanceID int64, usage *ResourceUsage) error

	// GetUsage возвращает текущие метрики потребления ресурсов.
	GetUsage(ctx context.Context, instanceID int64) (*ResourceUsage, error)

	// Delete удаляет все ключи состояния инстанса.
	Delete(ctx context.Context, instanceID int64) error
}
