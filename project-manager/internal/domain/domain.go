// Package domain содержит бизнес-сущности и интерфейсы репозиториев.
package domain

import (
	"context"
	"errors"
	"time"
)

// Общие доменные ошибки.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidInput  = errors.New("invalid input")
)

// ProjectStatus статус проекта.
type ProjectStatus int16

const (
	ProjectStatusDraft     ProjectStatus = 1
	ProjectStatusPending   ProjectStatus = 2
	ProjectStatusPublished ProjectStatus = 3
	ProjectStatusRejected  ProjectStatus = 4
)

// Project бизнес-сущность проекта игры.
type Project struct {
	ID                int64
	OwnerID           string
	TitleRu           string
	TitleEn           string
	SeoRu             string
	SeoEn             string
	About             string
	Status            ProjectStatus
	IconPath          string
	CoverPath         string
	VideoPath         string
	ActiveBuildVersion string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ProjectBuild бизнес-сущность клиентского билда.
type ProjectBuild struct {
	ID        int64
	ProjectID int64
	Version   string
	FilePath  string
	FileSize  int64
	CreatedAt time.Time
}

// ProjectRepo интерфейс репозитория проектов.
type ProjectRepo interface {
	Create(ctx context.Context, p *Project) (int64, error)
	Get(ctx context.Context, id int64) (*Project, error)
	ListByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*Project, error)
	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id int64) error
}

// ProjectBuildRepo интерфейс репозитория билдов.
type ProjectBuildRepo interface {
	Create(ctx context.Context, b *ProjectBuild) error
	ListByProject(ctx context.Context, projectID int64, limit int) ([]*ProjectBuild, error)
	Get(ctx context.Context, projectID int64, version string) (*ProjectBuild, error)
	Delete(ctx context.Context, id int64) error
	DeleteOldest(ctx context.Context, projectID int64, keep int) error
}

// ProjectMediaStorage интерфейс файлового хранилища проектов.
type ProjectMediaStorage interface {
	SaveIcon(projectID int64, data []byte) (path string, err error)
	SaveCover(projectID int64, data []byte) (path string, err error)
	SaveVideo(projectID int64, data []byte) (path string, err error)
	SaveBuild(projectID int64, version string, data []byte) (path string, err error)
	DeleteBuild(projectID int64, version string) error
	DeleteProjectDir(projectID int64) error
	BuildExists(projectID int64, version string) bool
}
