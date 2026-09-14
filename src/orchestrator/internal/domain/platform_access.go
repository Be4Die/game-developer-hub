package domain

import (
	"context"
	"time"
)

// PlatformGrant разрешение на использование серверов платформы для игры с многомерными лимитами.
type PlatformGrant struct {
	GameID               int64
	MaxInstances         int32
	MaxTotalCPUMillis    uint32 // 0 = unlim
	MaxTotalMemoryMB     uint64 // 0 = unlim
	MaxInstanceCPUMillis uint32 // 0 = unlim
	MaxInstanceMemoryMB  uint64 // 0 = unlim
	IsActive             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// PlatformAccessRepo интерфейс хранилища разрешений и квот на серверы платформы.
type PlatformAccessRepo interface {
	SaveGrant(ctx context.Context, grant *PlatformGrant) (*PlatformGrant, error)
	RevokeGrant(ctx context.Context, gameID int64) error
	GetGrant(ctx context.Context, gameID int64) (*PlatformGrant, error)
	HasApprovedAccess(ctx context.Context, gameID int64) (bool, *PlatformGrant, error)
	CountPlatformInstances(ctx context.Context, gameID int64) (int32, error)
	GetActivePlatformUsage(ctx context.Context, gameID int64) (count int32, allocatedCPUMillis uint32, allocatedMemoryBytes uint64, err error)
}
