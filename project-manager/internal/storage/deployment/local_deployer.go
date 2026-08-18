package deployment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/storage/filesystem"
)

// LocalDeployer реализует domain.Deployer для локального развертывания веб-игр на сервере платформы.
// Распаковывает сборку в директорию версий и атомарно обновляет символические ссылки dev/prod.
type LocalDeployer struct {
	gamesBasePath string
	urlPrefix     string
}

// NewLocalDeployer создаёт экземпляр локального развертывателя.
func NewLocalDeployer(gamesBasePath, urlPrefix string) *LocalDeployer {
	if urlPrefix == "" {
		urlPrefix = "/games"
	}
	return &LocalDeployer{
		gamesBasePath: gamesBasePath,
		urlPrefix:     urlPrefix,
	}
}

func (d *LocalDeployer) projectDir(projectID int64) string {
	return filepath.Join(d.gamesBasePath, strconv.FormatInt(projectID, 10))
}

func (d *LocalDeployer) versionDir(projectID int64, version string) string {
	return filepath.Join(d.projectDir(projectID), "versions", version)
}

// DeployDev развёртывает сборку в dev-окружение проекта.
func (d *LocalDeployer) DeployDev(ctx context.Context, projectID int64, version string, archivePath string) (*domain.DeploymentResult, error) {
	unpackedDir := d.versionDir(projectID, version)

	// Распаковываем архив (если еще не распакован)
	if err := filesystem.ExtractArchive(archivePath, unpackedDir); err != nil {
		return nil, fmt.Errorf("local_deployer: extract dev build: %w", err)
	}

	// Переключаем symlink dev -> versions/{version}
	symlinkPath := filepath.Join(d.projectDir(projectID), "dev")
	relTarget := filepath.Join("versions", version)

	_ = os.Remove(symlinkPath) // удаляем старый симлинк если был
	if err := os.Symlink(relTarget, symlinkPath); err != nil {
		return nil, fmt.Errorf("local_deployer: create dev symlink: %w", err)
	}

	devURL := fmt.Sprintf("%s/%d/dev/index.html", d.urlPrefix, projectID)

	return &domain.DeploymentResult{
		URL:          devURL,
		UnpackedPath: unpackedDir,
		Success:      true,
	}, nil
}

// DeployProd развёртывает одобренную версию в прод-окружение для игроков.
func (d *LocalDeployer) DeployProd(ctx context.Context, projectID int64, version string, archivePath string) (*domain.DeploymentResult, error) {
	unpackedDir := d.versionDir(projectID, version)

	// Убедимся, что сборка распакована
	if _, err := os.Stat(unpackedDir); os.IsNotExist(err) {
		if err := filesystem.ExtractArchive(archivePath, unpackedDir); err != nil {
			return nil, fmt.Errorf("local_deployer: extract prod build: %w", err)
		}
	}

	// Переключаем symlink prod -> versions/{version}
	symlinkPath := filepath.Join(d.projectDir(projectID), "prod")
	relTarget := filepath.Join("versions", version)

	_ = os.Remove(symlinkPath)
	if err := os.Symlink(relTarget, symlinkPath); err != nil {
		return nil, fmt.Errorf("local_deployer: create prod symlink: %w", err)
	}

	prodURL := fmt.Sprintf("%s/%d/prod/index.html", d.urlPrefix, projectID)

	return &domain.DeploymentResult{
		URL:          prodURL,
		UnpackedPath: unpackedDir,
		Success:      true,
	}, nil
}

// UndeployProd снимает игру с публикации, удаляя продуктивный symlink.
func (d *LocalDeployer) UndeployProd(ctx context.Context, projectID int64) error {
	symlinkPath := filepath.Join(d.projectDir(projectID), "prod")
	if err := os.Remove(symlinkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local_deployer: undeploy prod: %w", err)
	}
	return nil
}

// DeleteVersion удаляет распакованную директорию конкретной версии сборки (Garbage Collection).
func (d *LocalDeployer) DeleteVersion(ctx context.Context, projectID int64, version string) error {
	unpackedDir := d.versionDir(projectID, version)
	if err := os.RemoveAll(unpackedDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local_deployer: delete version %s: %w", unpackedDir, err)
	}
	return nil
}

// DeleteProject удаляет всю директорию проекта с распакованными сборками и симлинками.
func (d *LocalDeployer) DeleteProject(ctx context.Context, projectID int64) error {
	projDir := d.projectDir(projectID)
	if err := os.RemoveAll(projDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("local_deployer: delete project dir %s: %w", projDir, err)
	}
	return nil
}
