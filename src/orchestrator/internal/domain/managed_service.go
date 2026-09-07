package domain

import (
	"context"
	"time"
)

// ManagedService представляет управляемый сервис базы данных или хранилища.
type ManagedService struct {
	ID              int64
	NodeID          int64
	OwnerID         string
	AllowedGameIDs  []int64
	ServiceType     ServiceType
	Name            string
	ContainerID     string
	HostPort        uint32
	ConnectionURI   string
	Credentials     map[string]string
	Status          ServiceStatus
	VolumePath      string
	VolumeSizeBytes uint64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ManagedServiceRepo определяет контракт хранения данных об управляемых сервисах.
type ManagedServiceRepo interface {
	Create(ctx context.Context, s *ManagedService) error
	GetByID(ctx context.Context, id int64) (*ManagedService, error)
	GetByName(ctx context.Context, nodeID int64, name string) (*ManagedService, error)
	ListByNode(ctx context.Context, nodeID int64) ([]*ManagedService, error)
	ListByOwner(ctx context.Context, ownerID string) ([]*ManagedService, error)
	ListByGame(ctx context.Context, gameID int64) ([]*ManagedService, error)
	Update(ctx context.Context, s *ManagedService) error
	Delete(ctx context.Context, id int64) error
}
