package domain

import "context"

// ProjectRepo определяет контракт репозитория базовых сущностей игровых проектов.
type ProjectRepo interface {
	Create(ctx context.Context, p *Project) (int64, error)
	Get(ctx context.Context, id int64) (*Project, error)
	ListByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*Project, error)
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

// ModerationRepo определяет контракт репозитория тикетов модерации.
type ModerationRepo interface {
	CreateTicket(ctx context.Context, t *ModerationTicket) (int64, error)
	GetTicket(ctx context.Context, id int64) (*ModerationTicket, error)
	GetLatestTicketByProject(ctx context.Context, projectID int64) (*ModerationTicket, error)
	ListTickets(ctx context.Context, status *ModerationStatus, limit, offset int) ([]*ModerationTicket, error)
	ResolveTicket(ctx context.Context, id int64, status ModerationStatus, rejectionReason, moderatorID string) error
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
