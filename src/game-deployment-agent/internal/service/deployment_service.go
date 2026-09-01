package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/domain"
	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/storage/filesystem"
)

// DeploymentService реализует логику развертывания, управления симлинками и очистки сборок на агенте.
type DeploymentService struct {
	gamesBasePath     string
	urlPrefix         string
	maxUnpackedSizeMB int
	maxFilesCount     int
}

// NewDeploymentService создает новый экземпляр сервиса развертывания на агенте.
func NewDeploymentService(gamesBasePath, urlPrefix string, maxUnpackedSizeMB, maxFilesCount int) *DeploymentService {
	if urlPrefix == "" {
		urlPrefix = "/games"
	}
	return &DeploymentService{
		gamesBasePath:     gamesBasePath,
		urlPrefix:         urlPrefix,
		maxUnpackedSizeMB: maxUnpackedSizeMB,
		maxFilesCount:     maxFilesCount,
	}
}

func (s *DeploymentService) projectDir(projectID int64) string {
	return filepath.Join(s.gamesBasePath, strconv.FormatInt(projectID, 10))
}

func (s *DeploymentService) versionDir(projectID int64, version string) string {
	return filepath.Join(s.projectDir(projectID), "versions", version)
}

// DeployDevStream сохраняет временный архив из потока io.Reader, распаковывает в versions/{version} и обновляет dev symlink.
func (s *DeploymentService) DeployDevStream(ctx context.Context, projectID int64, version string, archiveReader io.Reader) (*domain.DeploymentResult, error) {
	projDir := s.projectDir(projectID)
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		return nil, fmt.Errorf("create project dir: %w", err)
	}

	tmpArchive, err := os.CreateTemp(projDir, "agent-upload-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("create temp archive: %w", err)
	}
	tmpPath := tmpArchive.Name()

	defer func() {
		_ = tmpArchive.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := io.Copy(tmpArchive, archiveReader); err != nil {
		return nil, fmt.Errorf("copy stream to temp archive: %w", err)
	}
	if err := tmpArchive.Close(); err != nil {
		return nil, fmt.Errorf("close temp archive: %w", err)
	}

	unpackedDir := s.versionDir(projectID, version)
	_ = os.RemoveAll(unpackedDir)

	maxBytes := int64(s.maxUnpackedSizeMB) * 1024 * 1024
	if err := filesystem.ExtractArchive(tmpPath, unpackedDir, maxBytes, s.maxFilesCount); err != nil {
		return nil, fmt.Errorf("extract archive: %w", err)
	}

	symlinkPath := filepath.Join(projDir, "dev")
	relTarget := filepath.Join("versions", version)

	_ = os.Remove(symlinkPath)
	if err := os.Symlink(relTarget, symlinkPath); err != nil {
		return nil, fmt.Errorf("create dev symlink: %w", err)
	}

	devURL := fmt.Sprintf("%s/%d/dev/index.html", s.urlPrefix, projectID)

	return &domain.DeploymentResult{
		URL:          devURL,
		UnpackedPath: unpackedDir,
		Success:      true,
	}, nil
}

// DeployProd переключает боевой симлинк prod -> versions/{version}.
func (s *DeploymentService) DeployProd(ctx context.Context, projectID int64, version string) (*domain.DeploymentResult, error) {
	projDir := s.projectDir(projectID)
	unpackedDir := s.versionDir(projectID, version)

	if _, err := os.Stat(unpackedDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("version %s is not deployed on agent", version)
	}

	symlinkPath := filepath.Join(projDir, "prod")
	relTarget := filepath.Join("versions", version)

	_ = os.Remove(symlinkPath)
	if err := os.Symlink(relTarget, symlinkPath); err != nil {
		return nil, fmt.Errorf("create prod symlink: %w", err)
	}

	prodURL := fmt.Sprintf("%s/%d/prod/index.html", s.urlPrefix, projectID)

	return &domain.DeploymentResult{
		URL:          prodURL,
		UnpackedPath: unpackedDir,
		Success:      true,
	}, nil
}

// UndeployProd снимает игру с публикации, удаляя симлинк prod.
func (s *DeploymentService) UndeployProd(ctx context.Context, projectID int64) error {
	symlinkPath := filepath.Join(s.projectDir(projectID), "prod")
	if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove prod symlink: %w", err)
	}
	return nil
}

// DeleteVersion удаляет директорию распакованной версии сборок (Garbage Collection).
func (s *DeploymentService) DeleteVersion(ctx context.Context, projectID int64, version string) error {
	unpackedDir := s.versionDir(projectID, version)
	if err := os.RemoveAll(unpackedDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete version dir %s: %w", unpackedDir, err)
	}
	return nil
}

// DeleteProject удаляет всю директорию проекта на агенте.
func (s *DeploymentService) DeleteProject(ctx context.Context, projectID int64) error {
	projDir := s.projectDir(projectID)
	if err := os.RemoveAll(projDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete project dir %s: %w", projDir, err)
	}
	return nil
}

// GetFreeDiskBytes возвращает количество свободных байт на диске хранилища сборок.
func (s *DeploymentService) GetFreeDiskBytes() uint64 {
	var stat syscall.Statfs_t
	dir := s.gamesBasePath
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0
	}
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0
	}
	return stat.Bavail * uint64(stat.Bsize)
}
