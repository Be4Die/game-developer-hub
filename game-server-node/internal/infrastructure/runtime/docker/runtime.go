// Package docker реализует контейнерный рантайм через Docker API.
package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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

// CreateContainer создаёт контейнер с заданными параметрами.
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
				{HostPort: fmt.Sprintf("%d", opts.HostPort)},
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
	return reader, nil
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
