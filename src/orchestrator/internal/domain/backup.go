package domain

import (
	"context"
	"time"
)

// ServiceBackup представляет запись о резервной копии сервиса.
type ServiceBackup struct {
	ID          int64
	BackupID    string
	NodeID      int64
	ServiceID   int64
	ServiceName string
	ServiceType ServiceType
	FileName    string
	SizeBytes   uint64
	Checksum    string
	BackupType  BackupType
	Status      BackupStatus
	CreatedAt   time.Time
}

// BackupRepo определяет контракт работы с метаданными бэкапов в БД.
type BackupRepo interface {
	Create(ctx context.Context, b *ServiceBackup) error
	GetByID(ctx context.Context, backupID string) (*ServiceBackup, error)
	ListByService(ctx context.Context, nodeID int64, serviceName string) ([]*ServiceBackup, error)
	Update(ctx context.Context, b *ServiceBackup) error
	Delete(ctx context.Context, backupID string) error
}
