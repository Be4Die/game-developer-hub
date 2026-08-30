package domain

import (
	"context"
	"io"
)

// BuildStorage определяет контракт хранилища архивов сборок веб-игр.
type BuildStorage interface {
	SaveArchiveStream(ctx context.Context, projectID int64, version string, src io.Reader) (archivePath string, sizeBytes int64, err error)
	SaveArchive(ctx context.Context, projectID int64, version string, data []byte) (archivePath string, sizeBytes int64, err error)
	GetArchivePath(projectID int64, version string) string
	DeleteBuild(projectID int64, version string) error
	DeleteProject(projectID int64) error
}

// MediaStorage определяет контракт хранилища промо-материалов (иконки, обложки, видео).
type MediaStorage interface {
	SaveMediaStream(ctx context.Context, projectID int64, mediaType string, src io.Reader) (filePath string, err error)
	SaveMedia(ctx context.Context, projectID int64, mediaType string, data []byte) (filePath string, err error)
	DeleteMedia(projectID int64, mediaType string) error
	SnapshotMediaForRelease(ctx context.Context, projectID int64, version string, srcPath string, mediaType string) (filePath string, err error)
}

// Deployer управляет жизненным циклом распаковки и публикации сборок веб-игр в целевые окружения.
type Deployer interface {
	DeployDev(ctx context.Context, projectID int64, version string, archivePath string) (*DeploymentResult, error)
	DeployProd(ctx context.Context, projectID int64, version string, archivePath string) (*DeploymentResult, error)
	UndeployProd(ctx context.Context, projectID int64) error
	DeleteVersion(ctx context.Context, projectID int64, version string) error
	DeleteProject(ctx context.Context, projectID int64) error
}
