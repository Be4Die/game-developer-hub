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

	// CreateContainer создаёт контейнер и возвращает его ID.
	CreateContainer(ctx context.Context, opts ContainerOpts) (containerID string, err error)
	// StartContainer запускает существующий контейнер.
	StartContainer(ctx context.Context, containerID string) error
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
