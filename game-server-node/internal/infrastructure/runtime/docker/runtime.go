// Package docker реализует контейнерный рантайм через Docker API.
package docker

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

// Runtime реализует ContainerRuntime через Docker API.
// Не безопасен для конкурентного использования без внешней синхронизации.
type Runtime struct {
	cli *client.Client
	log *slog.Logger
}

// New создаёт и инициализирует Docker-клиент.
// Возвращает ошибку если демон недоступен.
func New(log *slog.Logger) (*Runtime, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker.New: %w", err)
	}

	if _, err := cli.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("docker.New: daemon not reachable: %w", err)
	}

	log.Info("connected to Docker daemon")

	return &Runtime{cli: cli, log: log}, nil
}

// LoadImage загружает Docker-образ из потока данных.
func (r *Runtime) LoadImage(ctx context.Context, imageTag string, data io.Reader) error {
	const op = "DockerRuntime.LoadImage"

	r.log.Info("loading docker image",
		slog.String("op", op),
		slog.String("image_tag", imageTag),
	)

	resp, err := r.cli.ImageLoad(ctx, data)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("%s: reading response: %w", op, err)
	}

	r.log.Info("image loaded",
		slog.String("op", op),
		slog.String("image_tag", imageTag),
	)
	return nil
}

// BuildImage собирает Docker-образ из исходного архива (zip/tar.gz).
func (r *Runtime) BuildImage(ctx context.Context, imageTag string, internalPort uint32, archive io.Reader) error {
	const op = "DockerRuntime.BuildImage"

	r.log.Info("building docker image",
		slog.String("op", op),
		slog.String("image_tag", imageTag),
		slog.Uint64("internal_port", uint64(internalPort)),
	)

	// Create temp directory for build context.
	tmpDir, err := os.MkdirTemp("", "build-context-*")
	if err != nil {
		return fmt.Errorf("%s: create temp dir: %w", op, err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	unpackDir := filepath.Join(tmpDir, "unpack")
	if err := os.MkdirAll(unpackDir, 0o750); err != nil {
		return fmt.Errorf("%s: create unpack dir: %w", op, err)
	}

	// Read all archive data.
	data, err := io.ReadAll(archive)
	if err != nil {
		return fmt.Errorf("%s: read archive: %w", op, err)
	}

	// Extract archive.
	if err := r.extractArchive(bytes.NewReader(data), unpackDir); err != nil {
		return fmt.Errorf("%s: extract archive: %w", op, err)
	}

	// Generate Dockerfile.
	dockerfile := r.generateDockerfile(unpackDir, internalPort)
	if err := os.WriteFile(filepath.Join(unpackDir, "Dockerfile"), []byte(dockerfile), 0o600); err != nil {
		return fmt.Errorf("%s: write Dockerfile: %w", op, err)
	}

	// Run docker build.
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", imageTag, unpackDir) //nolint:gosec
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: docker build: %w: %s", op, err, stderr.String())
	}

	r.log.Info("image built",
		slog.String("op", op),
		slog.String("image_tag", imageTag),
	)
	return nil
}

// hostPortString возвращает строковое представление хостового порта для Docker.
// Если порт равен 0, возвращается пустая строка — это сигнал Docker
// выделить случайный свободный порт.
func hostPortString(port uint32) string {
	if port == 0 {
		return ""
	}
	return fmt.Sprintf("%d", port)
}

// CreateContainer создаёт контейнер (не запуская его). Возвращает ID контейнера.
// Для получения присвоенного хостового порта используйте GetHostPort после старта контейнера.
func (r *Runtime) CreateContainer(ctx context.Context, opts domain.ContainerOpts) (string, error) {
	const op = "Runtime.CreateContainer"

	internalPort := nat.Port(fmt.Sprintf("%d/tcp", opts.InternalPort))

	containerConfig := &container.Config{
		Image: opts.ImageTag,
		ExposedPorts: nat.PortSet{
			internalPort: struct{}{},
		},
	}

	if len(opts.EnvVars) > 0 {
		env := make([]string, 0, len(opts.EnvVars))
		for k, v := range opts.EnvVars {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		containerConfig.Env = env
	}

	if len(opts.Args) > 0 {
		containerConfig.Cmd = opts.Args
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			internalPort: []nat.PortBinding{
				{HostPort: hostPortString(opts.HostPort)},
			},
		},
	}

	if opts.CPUMillis != nil || opts.MemoryBytes != nil {
		hostConfig.Resources = container.Resources{}

		if opts.CPUMillis != nil {
			hostConfig.CPUPeriod = 100000
			hostConfig.CPUQuota = int64(*opts.CPUMillis) * 100 //nolint:gosec // millis value is validated
		}

		if opts.MemoryBytes != nil {
			hostConfig.Memory = int64(*opts.MemoryBytes) //nolint:gosec // memory size is validated
		}
	}

	resp, err := r.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	r.log.Info("container created",
		slog.String("op", op),
		slog.String("container_id", resp.ID[:12]),
		slog.String("image", opts.ImageTag),
	)

	return resp.ID, nil
}

// StartContainer запускает остановленный контейнер.
func (r *Runtime) StartContainer(ctx context.Context, containerID string) error {
	if err := r.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("Runtime.StartContainer: %w", err)
	}
	return nil
}

// StopContainer останавливает контейнер с заданным таймаутом.
func (r *Runtime) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	timeoutSeconds := int(timeout.Seconds())

	if err := r.cli.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeoutSeconds,
	}); err != nil {
		return fmt.Errorf("Runtime.StopContainer: %w", err)
	}
	return nil
}

// RemoveContainer удаляет контейнер безвозвратно.
func (r *Runtime) RemoveContainer(ctx context.Context, containerID string) error {
	if err := r.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("Runtime.RemoveContainer: %w", err)
	}
	return nil
}

// ContainerLogs возвращает поток stdout/stderr контейнера.
// Демультиплексирует Docker multiplexed stream в чистый текстовый поток.
func (r *Runtime) ContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error) {
	reader, err := r.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: true,
	})
	if err != nil {
		return nil, fmt.Errorf("Runtime.ContainerLogs: %w", err)
	}

	pr, pw := io.Pipe()
	go func() {
		defer func() { _ = reader.Close() }()
		_, err := stdcopy.StdCopy(pw, pw, reader)
		_ = pw.CloseWithError(err)
	}()

	return pr, nil
}

// ContainerStats возвращает метрики использования ресурсов контейнера.
func (r *Runtime) ContainerStats(ctx context.Context, containerID string) (domain.ResourcesUsage, error) {
	const op = "Runtime.ContainerStats"

	resp, err := r.cli.ContainerStatsOneShot(ctx, containerID)
	if err != nil {
		return domain.ResourcesUsage{}, fmt.Errorf("%s: %w", op, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var stats container.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return domain.ResourcesUsage{}, fmt.Errorf("%s: decode: %w", op, err)
	}

	return domain.ResourcesUsage{
		CPU:     calculateCPUPercent(&stats),
		Memory:  stats.MemoryStats.Usage,
		Disk:    0,
		Network: calculateNetworkBytes(&stats),
	}, nil
}

func calculateCPUPercent(stats *container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) -
		float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) -
		float64(stats.PreCPUStats.SystemUsage)

	if systemDelta <= 0 || cpuDelta <= 0 {
		return 0.0
	}

	cpuCount := float64(stats.CPUStats.OnlineCPUs)
	if cpuCount == 0 {
		cpuCount = 1.0
	}

	return (cpuDelta / systemDelta) * cpuCount * 100.0
}

func calculateNetworkBytes(stats *container.StatsResponse) uint64 {
	var total uint64
	for _, netStats := range stats.Networks {
		total += netStats.RxBytes + netStats.TxBytes
	}
	return total
}

// GetHostPort возвращает реальный хост-порт, опубликованный для контейнера.
// Вызывать после StartContainer, когда портовые привязки активны.
// Для динамических портов (HostPort=0) Docker назначает случайный порт.
// Метод повторяет попытки inspection, так как привязка может появиться не сразу.
func (r *Runtime) GetHostPort(ctx context.Context, containerID string, internalPort uint32) (uint32, error) {
	portKey := fmt.Sprintf("%d/tcp", internalPort)

	const maxRetries = 10
	retryDelay := 200 * time.Millisecond

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		inspect, err := r.cli.ContainerInspect(ctx, containerID)
		if err != nil {
			return 0, fmt.Errorf("inspect container (attempt %d/%d): %w", i+1, maxRetries, err)
		}

		bindings, ok := inspect.NetworkSettings.Ports[nat.Port(portKey)]
		if ok && len(bindings) > 0 {
			hostPort := bindings[0].HostPort
			if hostPort == "" {
				lastErr = fmt.Errorf("HostPort empty for %s", portKey)
				time.Sleep(retryDelay)
				continue
			}

			var port uint32
			if _, err := fmt.Sscanf(hostPort, "%d", &port); err != nil {
				return 0, fmt.Errorf("parse host port %q: %w", hostPort, err)
			}
			return port, nil
		}

		lastErr = fmt.Errorf("no port binding for %s", portKey)
		time.Sleep(retryDelay)
	}

	return 0, fmt.Errorf("failed to get port after %d retries: last error: %w", maxRetries, lastErr)
}

// extractArchive распаковывает zip или tar.gz архив в целевую директорию.
func (r *Runtime) extractArchive(reader io.Reader, destDir string) error {
	buf := make([]byte, 4)
	n, err := io.ReadFull(reader, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("read header: %w", err)
	}

	// ZIP magic: PK\x03\x04
	if n == 4 && buf[0] == 0x50 && buf[1] == 0x4B && buf[2] == 0x03 && buf[3] == 0x04 {
		all, readErr := io.ReadAll(io.MultiReader(bytes.NewReader(buf), reader))
		if readErr != nil {
			return fmt.Errorf("read zip: %w", readErr)
		}
		return r.extractZip(all, destDir)
	}

	// This is tar.gz — try gzip reader first.
	gzReader, gzErr := gzip.NewReader(io.MultiReader(bytes.NewReader(buf[:n]), reader))
	if gzErr != nil {
		// Not gzip, try plain tar.
		return r.extractTar(io.MultiReader(bytes.NewReader(buf[:n]), reader), destDir)
	}
	defer func() { _ = gzReader.Close() }()

	return r.extractTar(gzReader, destDir)
}

// extractZip извлекает файлы из ZIP-архива.
func (r *Runtime) extractZip(data []byte, destDir string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	for _, f := range reader.File {
		if err := r.extractZipFile(f, destDir); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runtime) extractZipFile(f *zip.File, destDir string) error {
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

	out, err := os.Create(path) //nolint:gosec
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
func (r *Runtime) extractTar(reader io.Reader, destDir string) error {
	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Might be gzip — try to decompress.
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
func (r *Runtime) generateDockerfile(unpackDir string, internalPort uint32) string {
	executable := "server"
	_ = filepath.Walk(unpackDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode()&0o111 != 0 || strings.HasSuffix(info.Name(), ".exe") {
			executable = info.Name()
			return filepath.SkipAll
		}
		return nil
	})

	return fmt.Sprintf(`FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY . /app
RUN chmod +x /app/%s
EXPOSE %d
CMD ["./%s"]
`, executable, internalPort, executable)
}
