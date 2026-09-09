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

	// CleanupBuildArtifacts удаляет артефакты неудачной сборки
	// (промежуточные контейнеры и dangling-образы).
	CleanupBuildArtifacts(ctx context.Context, imageTag string) error

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
	// RestartContainer перезапускает контейнер с заданным таймаутом.
	RestartContainer(ctx context.Context, containerID string, timeout time.Duration) error
	// RemoveContainer удаляет контейнер безвозвратно.
	RemoveContainer(ctx context.Context, containerID string) error

	// ContainerLogs возвращает поток логов контейнера.
	ContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error)
	// ContainerStats возвращает текущие метрики использования ресурсов.
	ContainerStats(ctx context.Context, containerID string) (ResourcesUsage, error)

	// PullImage скачивает образ из Docker Hub / реестра, если его нет локально.
	PullImage(ctx context.Context, imageTag string) error

	// EnsureNetwork создает пользовательскую bridge-сеть Docker, если она не существует.
	EnsureNetwork(ctx context.Context, networkName string) error

	// ListContainers возвращает информацию обо всех контейнерах на хосте.
	ListContainers(ctx context.Context) ([]ContainerInfo, error)

	// RemoveVolume удаляет именованный том Docker.
	RemoveVolume(ctx context.Context, volumeName string) error

	// ImageExists проверяет наличие образа локально.
	ImageExists(ctx context.Context, imageTag string) bool

	// CopyToContainer копирует одиночный файл в контейнер в указанную директорию.
	CopyToContainer(ctx context.Context, containerID, targetDir, filename string, content []byte) error

	// InspectContainer возвращает подробную информацию о контейнере (статус работы, политика рестарта, порты).
	InspectContainer(ctx context.Context, containerID string) (*ContainerDetails, error)

	// UpdateRestartPolicy обновляет политику рестарта существующего контейнера.
	UpdateRestartPolicy(ctx context.Context, containerID string, policy string) error
}

// ContainerDetails содержит расширенную информацию о состоянии контейнера.
type ContainerDetails struct {
	ID            string
	Name          string
	Running       bool
	RestartPolicy string
	Ports         map[uint32]uint32
}

// ContainerInfo содержит базовые данные о контейнере.
type ContainerInfo struct {
	ID     string
	Labels map[string]string
	Status string
}

// ContainerOpts задаёт параметры создания контейнера.
type ContainerOpts struct {
	ContainerName string
	ImageTag      string
	InternalPort  uint32
	HostPort      uint32
	RestartPolicy string
	EnvVars       map[string]string
	Args          []string
	CPUMillis     *uint32
	MemoryBytes   *uint64
	Labels        map[string]string
	Binds         []string
	Network       string
}

