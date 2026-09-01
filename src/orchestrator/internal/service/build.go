// Package service содержит бизнес-логику оркестратора.
package service

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/infrastructure/config"
	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
)

// BuildPipeline определяет пайплайн преобразования архива билда
// в Docker-образ, загрузки его на ноду и регистрации метаданных.
type BuildPipeline struct {
	buildRepo  domain.BuildStorage
	buildFS    domain.BuildStorageFS
	nodeClient domain.NodeClient
	nodeRepo   domain.NodeRepo
	nodeState  domain.NodeStateStore
	limits     config.LimitsConfig
	workDir    string // директория для временных файлов пайплайна
}

// NewBuildPipeline создаёт пайплайн обработки билдов.
func NewBuildPipeline(
	buildRepo domain.BuildStorage,
	buildFS domain.BuildStorageFS,
	nodeClient domain.NodeClient,
	nodeRepo domain.NodeRepo,
	nodeState domain.NodeStateStore,
	limits config.LimitsConfig,
) *BuildPipeline {
	return &BuildPipeline{
		buildRepo:  buildRepo,
		buildFS:    buildFS,
		nodeClient: nodeClient,
		nodeRepo:   nodeRepo,
		nodeState:  nodeState,
		limits:     limits,
		workDir:    os.TempDir(),
	}
}

// WithWorkDir задаёт директорию для временных файлов.
func (p *BuildPipeline) WithWorkDir(dir string) *BuildPipeline {
	p.workDir = dir
	return p
}

// UploadBuildParams содержит параметры загрузки билда.
type UploadBuildParams struct {
	OwnerID      string
	GameID       int64
	Version      string
	Protocol     domain.Protocol
	InternalPort uint32
	MaxPlayers   uint32
	Archive      io.Reader
	ArchiveSize  int64
	// ArchiveData — альтернатива Archive, когда данные уже в памяти (unary gRPC).
	ArchiveData []byte
}

// UploadBuild загружает серверный билд: распаковывает архив для валидации,
// отправляет архив на ноду для сборки Docker-образа, сохраняет архив в хранилище
// и регистрирует метаданные.
// Возвращает ErrNoAvailableNode при отсутствии свободных нод.
// Превышение размера архива или лимита билдов на игру возвращает ошибку.
func (p *BuildPipeline) UploadBuild(ctx context.Context, params UploadBuildParams) (*domain.ServerBuild, error) {
	// Шаг 1: проверка лимита билдов на игру.
	count, err := p.buildRepo.CountByGame(ctx, params.GameID)
	if err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: count builds: %w", err)
	}
	if count >= p.limits.MaxBuildsPerGame {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: max builds limit reached (%d)", p.limits.MaxBuildsPerGame)
	}

	// Определяем источник данных: streaming (Archive) или unary (ArchiveData).
	var data []byte
	if params.ArchiveData != nil {
		// Unary mode — данные уже в памяти.
		data = params.ArchiveData
		params.ArchiveSize = int64(len(data))
	} else if params.Archive != nil {
		// Streaming mode — читаем весь архив для обработки.
		data, err = io.ReadAll(params.Archive)
		if err != nil {
			return nil, fmt.Errorf("BuildPipeline.UploadBuild: read archive: %w", err)
		}
		params.ArchiveSize = int64(len(data))
	} else {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: no archive data")
	}

	// Шаг 2: проверка размера архива.
	if params.ArchiveSize > p.limits.MaxBuildSizeBytes {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: archive size %d exceeds limit %d",
			params.ArchiveSize, p.limits.MaxBuildSizeBytes)
	}

	// Создаём временную директорию для валидации архива.
	tmpDir, err := os.MkdirTemp(p.workDir, "build-pipeline-*")
	if err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Шаг 3: распаковка архива для валидации.
	unpackDir := filepath.Join(tmpDir, "unpack")
	if err := os.MkdirAll(unpackDir, 0o750); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: create unpack dir: %w", err)
	}
	if err := p.extractArchive(bytes.NewReader(data), unpackDir); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: extract archive: %w", err)
	}

	// Если internal port не указан, используем значение по умолчанию.
	internalPort := params.InternalPort
	if internalPort == 0 {
		internalPort = 8080
	}

	// Шаг 4: определение image tag.
	imageTag := fmt.Sprintf("welwise/game-%d:%s", params.GameID, params.Version)

	// Шаг 5: выбор ноды для сборки и запуска.
	node, err := p.selectNode(ctx)
	if err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: select node: %w", err)
	}

	// Шаг 6: отправка архива на ноду для сборки Docker-образа.
	archiveReader := bytes.NewReader(data)
	buildMeta := domain.BuildImageMetadata{
		GameID:       params.GameID,
		ImageTag:     imageTag,
		InternalPort: internalPort,
	}
	if err := p.nodeClient.BuildImage(ctx, node.Address, node.APIToken, buildMeta, archiveReader); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: build image on node: %w", err)
	}

	// Шаг 7: сохранение исходного архива в файловое хранилище.
	filePath, err := p.buildFS.Save(params.GameID, params.Version, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: save to FS: %w", err)
	}

	// Шаг 8: регистрация метаданных в PostgreSQL.
	build := &domain.ServerBuild{
		OwnerID:      params.OwnerID,
		GameID:       params.GameID,
		Version:      params.Version,
		ImageTag:     imageTag,
		Protocol:     params.Protocol,
		InternalPort: internalPort,
		MaxPlayers:   params.MaxPlayers,
		FileURL:      filePath,
		FileSize:     int64(len(data)),
		CreatedAt:    time.Now(),
	}

	if err := p.buildRepo.Create(ctx, build); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: save metadata: %w", err)
	}

	return build, nil
}

// extractArchive распаковывает zip или tar.gz архив в целевую директорию.
func (p *BuildPipeline) extractArchive(r io.Reader, destDir string) error {
	buf := make([]byte, 4)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("read header: %w", err)
	}

	// ZIP magic: PK\x03\x04
	if n == 4 && buf[0] == 0x50 && buf[1] == 0x4B && buf[2] == 0x03 && buf[3] == 0x04 {
		all, readErr := io.ReadAll(io.MultiReader(bytes.NewReader(buf), r))
		if readErr != nil {
			return fmt.Errorf("read zip: %w", readErr)
		}
		return p.extractZip(all, destDir)
	}

	// Это tar.gz.
	gzReader, gzErr := gzip.NewReader(io.MultiReader(bytes.NewReader(buf[:n]), r))
	if gzErr != nil {
		return fmt.Errorf("create gzip reader: %w", gzErr)
	}
	defer func() { _ = gzReader.Close() }()

	return p.extractTar(gzReader, destDir)
}

// extractZip извлекает файлы из ZIP-архива.
func (p *BuildPipeline) extractZip(data []byte, destDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	for _, f := range reader.File {
		if err := p.extractZipFile(f, destDir); err != nil {
			return err
		}
	}

	return nil
}

func (p *BuildPipeline) extractZipFile(f *zip.File, destDir string) error {
	// Защита от path traversal.
	cleanName := filepath.Clean(f.Name)
	if filepath.IsAbs(cleanName) || strings.HasPrefix(cleanName, "..") {
		return fmt.Errorf("zip file %q has invalid path", f.Name)
	}

	path := filepath.Join(destDir, cleanName)
	if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(destDir)) { //nolint:gosec
		return fmt.Errorf("zip file %q escapes destination dir", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(path, 0o750)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("mkdir %q: %w", filepath.Dir(path), err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zip entry %q: %w", f.Name, err)
	}
	defer func() { _ = rc.Close() }()

	out, err := os.Create(path) //nolint:gosec // путь валидирован выше
	if err != nil {
		return fmt.Errorf("create file %q: %w", path, err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, rc); err != nil { //nolint:gosec
		return fmt.Errorf("extract %q: %w", f.Name, err)
	}

	return nil
}

// extractTar извлекает файлы из tar-архива.
func (p *BuildPipeline) extractTar(r io.Reader, destDir string) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}

		cleanName := filepath.Clean(header.Name)
		if filepath.IsAbs(cleanName) || strings.HasPrefix(cleanName, "..") {
			return fmt.Errorf("tar file %q has invalid path", header.Name)
		}

		path := filepath.Join(destDir, cleanName) //nolint:gosec
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(destDir)) {
			return fmt.Errorf("tar file %q escapes destination dir", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o750); err != nil {
				return fmt.Errorf("mkdir %q: %w", path, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
				return fmt.Errorf("mkdir %q: %w", filepath.Dir(path), err)
			}
			out, err := os.Create(path) //nolint:gosec
			if err != nil {
				return fmt.Errorf("create file %q: %w", path, err)
			}
			if _, err := io.Copy(out, tr); err != nil { //nolint:gosec
				_ = out.Close()
				return fmt.Errorf("extract %q: %w", header.Name, err)
			}
			_ = out.Close()
		}
	}

	return nil
}

// selectNode выбирает ноду с наименьшей загрузкой для размещения образа.
func (p *BuildPipeline) selectNode(ctx context.Context) (*domain.Node, error) {
	nodes, err := p.nodeRepo.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("selectNode: list nodes: %w", err)
	}

	var best *domain.Node
	bestLoad := ^uint32(0)

	for _, n := range nodes {
		if n.Status != domain.NodeStatusOnline {
			continue
		}

		load, err := p.nodeState.GetActiveInstanceCount(ctx, n.ID)
		if err != nil {
			load = 0
		}

		if load < bestLoad {
			bestLoad = load
			best = n
		}
	}

	if best == nil {
		return nil, domain.ErrNoAvailableNode
	}

	return best, nil
}

// ListBuilds возвращает список билдов игры (последние limit штук).
func (p *BuildPipeline) ListBuilds(ctx context.Context, gameID int64) ([]*domain.ServerBuild, error) {
	return p.buildRepo.ListByGame(ctx, gameID, p.limits.MaxBuildsPerGame)
}

// GetBuild возвращает билд по версии.
func (p *BuildPipeline) GetBuild(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
	return p.buildRepo.GetByVersion(ctx, gameID, version)
}

// DeleteBuild удаляет билд по версии. Проверяет отсутствие активных инстансов,
// удаляет файл из файлового хранилища и метаданные из PostgreSQL.
// Возвращает ErrBuildInUse при наличии запущенных инстансов, ErrNotFound при отсутствии билда.
func (p *BuildPipeline) DeleteBuild(ctx context.Context, gameID int64, version string) error {
	build, err := p.buildRepo.GetByVersion(ctx, gameID, version)
	if err != nil {
		return fmt.Errorf("BuildPipeline.DeleteBuild: get build: %w", err)
	}

	activeCount, err := p.buildRepo.CountActiveInstancesByBuild(ctx, build.ID)
	if err != nil {
		return fmt.Errorf("BuildPipeline.DeleteBuild: count active instances: %w", err)
	}
	if activeCount > 0 {
		return domain.ErrBuildInUse
	}

	if err := p.buildFS.Delete(gameID, version); err != nil && !errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("BuildPipeline.DeleteBuild: delete from FS: %w", err)
	}

	if err := p.buildRepo.Delete(ctx, build.ID); err != nil {
		return fmt.Errorf("BuildPipeline.DeleteBuild: delete metadata: %w", err)
	}

	return nil
}
