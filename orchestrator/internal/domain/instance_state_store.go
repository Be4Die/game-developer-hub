package domain

import (
	"context"
	"time"
)

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

	// SetQueueSize обновляет размер очереди на инстансе и сбрасывает TTL.
	SetQueueSize(ctx context.Context, instanceID int64, size uint32) error

	// GetQueueSize возвращает текущий размер очереди на инстансе.
	GetQueueSize(ctx context.Context, instanceID int64) (uint32, error)

	// SetUsage обновляет метрики потребления ресурсов и сбрасывает TTL.
	SetUsage(ctx context.Context, instanceID int64, usage *ResourceUsage) error

	// GetUsage возвращает текущие метрики потребления ресурсов.
	GetUsage(ctx context.Context, instanceID int64) (*ResourceUsage, error)

	// Delete удаляет все ключи состояния инстанса.
	Delete(ctx context.Context, instanceID int64) error

	// SetZeroPlayersSince записывает timestamp начала нулевого онлайна.
	SetZeroPlayersSince(ctx context.Context, instanceID int64, t time.Time) error

	// GetZeroPlayersSince возвращает timestamp начала нулевого онлайна.
	// ErrNotFound если ключ отсутствует.
	GetZeroPlayersSince(ctx context.Context, instanceID int64) (time.Time, error)

	// DeleteZeroPlayersSince удаляет timestamp нулевого онлайна.
	DeleteZeroPlayersSince(ctx context.Context, instanceID int64) error
}
