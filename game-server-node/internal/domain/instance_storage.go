package domain

import "context"

type InstanceStorage interface {
	GetInstanceByID(ctx context.Context, id int64) (*Instance, error)
	GetInstancesByGameID(ctx context.Context, gameID int64) ([]Instance, error)
	GetInstanceByContainerID(ctx context.Context, containerID string) (*Instance, error)
	GetAllInstances(ctx context.Context) ([]Instance, error)

	RecordInstance(ctx context.Context, instance Instance) error

	DeleteInstance(ctx context.Context, id int64) error
}
