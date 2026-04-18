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
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/config"
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
	GameID       int64
	Version      string
	Protocol     domain.Protocol
	InternalPort uint32
	MaxPlayers   uint32
	Archive      io.Reader
	ArchiveSize  int64
	// ArchiveData — альтернатива Archive, когда данные уже в памяти (gRPC).
	ArchiveData []byte
}

// UploadBuild загружает серверный билд: распаковывает архив, собирает Docker-образ,
// сохраняет его в файловое хранилище, загружает на ноду и регистрирует метаданные.
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

	// Шаг 2: проверка размера архива.
	if params.ArchiveSize > p.limits.MaxBuildSizeBytes {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: archive size %d exceeds limit %d",
			params.ArchiveSize, p.limits.MaxBuildSizeBytes)
	}

	// Создаём временную директорию для пайплайна.
	tmpDir, err := os.MkdirTemp(p.workDir, "build-pipeline-*")
	if err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Шаг 3: распаковка архива в temp-директорию.
	unpackDir := filepath.Join(tmpDir, "unpack")
	if err := os.MkdirAll(unpackDir, 0o750); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: create unpack dir: %w", err)
	}
	if err := p.extractArchive(params.Archive, unpackDir); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: extract archive: %w", err)
	}

	// Шаг 4: определение image tag.
	imageTag := fmt.Sprintf("welwise/game-%d:%s", params.GameID, params.Version)

	// Шаг 5: создание Dockerfile.
	dockerfile := p.generateDockerfile(unpackDir, params.InternalPort)
	if err := os.WriteFile(filepath.Join(unpackDir, "Dockerfile"), []byte(dockerfile), 0o600); err != nil { //nolint:gosec
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: write Dockerfile: %w", err)
	}

	// Шаг 6: сборка Docker-образа.
	if err := p.dockerBuild(ctx, unpackDir, imageTag); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: docker build: %w", err)
	}

	// Шаг 7: сохранение образа в tar-файл.
	imageTarPath := filepath.Join(tmpDir, "image.tar")
	if err := p.dockerSave(ctx, imageTag, imageTarPath); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: docker save: %w", err)
	}

	imageFileInfo, err := os.Stat(imageTarPath)
	if err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: stat image tar: %w", err)
	}

	// Шаг 8: сохранение файла образа в файловое хранилище.
	//nolint:gosec // путь сгенерирован внутри пайплайна, не пользовательский ввод
	imageFile, err := os.Open(imageTarPath)
	if err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: open image tar: %w", err)
	}
	defer func() { _ = imageFile.Close() }()

	filePath, err := p.buildFS.Save(params.GameID, params.Version, imageFile, imageFileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: save to FS: %w", err)
	}

	// Шаг 9: выбор ноды для загрузки образа.
	node, err := p.selectNode(ctx)
	if err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: select node: %w", err)
	}

	// Шаг 10: загрузка Docker-образа на ноду через gRPC.
	if err := p.loadImageToNode(ctx, node.Address, node.APIToken, params.GameID, imageTag, imageTarPath); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: load image to node: %w", err)
	}

	// Шаг 11: регистрация метаданных в PostgreSQL.
	build := &domain.ServerBuild{
		GameID:       params.GameID,
		Version:      params.Version,
		ImageTag:     imageTag,
		Protocol:     params.Protocol,
		InternalPort: params.InternalPort,
		MaxPlayers:   params.MaxPlayers,
		FileURL:      filePath,
		FileSize:     imageFileInfo.Size(),
		CreatedAt:    time.Now(),
	}

	if err := p.buildRepo.Create(ctx, build); err != nil {
		return nil, fmt.Errorf("BuildPipeline.UploadBuild: save metadata: %w", err)
	}

	return build, nil
}

// UploadBuildFromBytes загружает билд из данных в памяти (для gRPC).
func (p *BuildPipeline) UploadBuildFromBytes(ctx context.Context, params UploadBuildParams) (*domain.ServerBuild, error) {
	if len(params.ArchiveData) == 0 {
		return nil, fmt.Errorf("BuildPipeline.UploadBuildFromBytes: archive data is empty")
	}
	params.Archive = bytes.NewReader(params.ArchiveData)
	params.ArchiveSize = int64(len(params.ArchiveData))
	return p.UploadBuild(ctx, params)
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

// generateDockerfile создаёт минимальный Dockerfile для игрового сервера.
func (p *BuildPipeline) generateDockerfile(unpackDir string, internalPort uint32) string {
	executable := "server"
	_ = filepath.Walk(unpackDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode()&0o111 != 0 || strings.HasSuffix(info.Name(), ".exe") {
			executable = "./" + info.Name()
			return filepath.SkipAll
		}
		return nil
	})

	return fmt.Sprintf(`FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY . /app
EXPOSE %d
CMD [%q]
`, internalPort, executable)
}

// dockerBuild запускает docker build в указанной директории.
func (p *BuildPipeline) dockerBuild(ctx context.Context, contextDir, tag string) error {
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", tag, contextDir) //nolint:gosec
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build: %w", err)
	}
	return nil
}

// dockerSave запускает docker save для сохранения образа в файл.
func (p *BuildPipeline) dockerSave(ctx context.Context, tag, outPath string) error {
	outFile, err := os.Create(outPath) //nolint:gosec // путь генерируется внутри пайплайна
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	cmd := exec.CommandContext(ctx, "docker", "save", "-o", outPath, tag) //nolint:gosec
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker save: %w", err)
	}
	return nil
}

// loadImageToNode загружает Docker-образ на ноду через gRPC.
func (p *BuildPipeline) loadImageToNode(ctx context.Context, nodeAddress, apiKey string, gameID int64, imageTag, tarPath string) error {
	file, err := os.Open(tarPath) //nolint:gosec // путь генерируется внутри пайплайна
	if err != nil {
		return fmt.Errorf("open tar file: %w", err)
	}
	defer func() { _ = file.Close() }()

	meta := domain.ImageMetadata{
		GameID:   gameID,
		ImageTag: imageTag,
	}

	_, err = p.nodeClient.LoadImage(ctx, nodeAddress, apiKey, meta, file)
	if err != nil {
		return fmt.Errorf("load image: %w", err)
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
