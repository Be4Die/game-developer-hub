package domain

import "context"

// ProjectRepo определяет контракт репозитория базовых сущностей игровых проектов.
type ProjectRepo interface {
	Create(ctx context.Context, p *Project) (int64, error)
	Get(ctx context.Context, id int64) (*Project, error)
	ListByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*Project, error)
	CountByOwner(ctx context.Context, ownerID string) (int, error)
	ListForUser(ctx context.Context, userID string, limit, offset int) ([]*Project, error)
	CountForUser(ctx context.Context, userID string) (int, error)
	ListPublished(ctx context.Context, limit, offset int) ([]*Project, error)
	CountPublished(ctx context.Context) (int, error)
	UpdateStatus(ctx context.Context, id int64, status ProjectStatus) error
	Delete(ctx context.Context, id int64) error
}

// DraftRepo определяет контракт репозитория черновиков проектов.
type DraftRepo interface {
	Create(ctx context.Context, d *Draft) error
	Get(ctx context.Context, projectID int64) (*Draft, error)
	Update(ctx context.Context, d *Draft) error
	UpdateActiveBuild(ctx context.Context, projectID int64, version, devURL string) error
	UpdateMedia(ctx context.Context, projectID int64, mediaType, path string) error
}

// BuildRepo определяет контракт репозитория сборок веб-игр.
type BuildRepo interface {
	Create(ctx context.Context, b *Build) error
	Get(ctx context.Context, projectID int64, version string) (*Build, error)
	ListByProject(ctx context.Context, projectID int64, limit int) ([]*Build, error)
	MarkUnpacked(ctx context.Context, projectID int64, version, unpackedPath string) error
	Delete(ctx context.Context, id int64) error
}

// ReleaseRepo определяет контракт репозитория опубликованных релизов.
type ReleaseRepo interface {
	Create(ctx context.Context, r *Release) (int64, error)
	GetActive(ctx context.Context, projectID int64) (*Release, error)
	ListByProject(ctx context.Context, projectID int64) ([]*Release, error)
	Deactivate(ctx context.Context, projectID int64) error
}

// DeploymentRepo определяет контракт репозитория истории развертываний.
type DeploymentRepo interface {
	Create(ctx context.Context, d *DeploymentRecord) error
	ListByProject(ctx context.Context, projectID int64, limit int) ([]*DeploymentRecord, error)
}
