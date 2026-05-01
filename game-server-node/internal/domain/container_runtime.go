package domain

import (
	"context"
	"io"
	"time"
)

// ContainerRuntime управляет жизненным циклом контейнеров.
type ContainerRuntime interface {
	// LoadImage загружает образ из потока данных.
	LoadImage(ctx context.Context, imageTag string, data io.Reader) error

	// BuildImage собирает образ из исходного архива (zip/tar.gz).
	// internalPort — порт, который слушает процесс внутри контейнера.
	BuildImage(ctx context.Context, imageTag string, internalPort uint32, archive io.Reader) error

	// CreateContainer создаёт контейнер (без запуска). Возвращает ID контейнера.
	CreateContainer(ctx context.Context, opts ContainerOpts) (string, error)

	// StartContainer запускает существующий контейнер.
	StartContainer(ctx context.Context, containerID string) error

	// GetHostPort возвращает реальный хост-порт, который Docker присвоил контейнеру
	// после его запуска. Для динамических портов (HostPort=0) это фактический
	// опубликованный порт; для статических — запрошенный порт.
	GetHostPort(ctx context.Context, containerID string, internalPort uint32) (uint32, error)

	// StopContainer останавливает контейнер с заданным таймаутом.
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	// RemoveContainer удаляет контейнер безвозвратно.
	RemoveContainer(ctx context.Context, containerID string) error

	// ContainerLogs возвращает поток логов контейнера.
	ContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error)
	// ContainerStats возвращает текущие метрики использования ресурсов.
	ContainerStats(ctx context.Context, containerID string) (ResourcesUsage, error)
}

// ContainerOpts задаёт параметры создания контейнера.
type ContainerOpts struct {
	ImageTag     string
	InternalPort uint32
	HostPort     uint32
	EnvVars      map[string]string
	Args         []string
	CPUMillis    *uint32
	MemoryBytes  *uint64
}
