package domain

import (
	"context"
	"io"
)

// NodeClient определяет интерфейс gRPC-клиента для управления нодой.
// Абстрагирует конкретную реализацию gRPC-подключения к game-server-node.
// apiKey — токен авторизации ноды (передается как "authorization: Bearer <apiKey>").
type NodeClient interface {
	// LoadImage загружает Docker-образ на ноду. Метаданные отправляются первым сообщением,
	// далее — чанки данных. Возвращает подтверждение с размером загруженного образа.
	LoadImage(ctx context.Context, nodeAddress, apiKey string, metadata ImageMetadata, chunks io.Reader) (*ImageLoadResult, error)

	// StartInstance запускает экземпляр игрового сервера на ноде.
	StartInstance(ctx context.Context, nodeAddress, apiKey string, req StartInstanceRequest) (*StartInstanceResult, error)

	// StopInstance выполняет graceful остановку экземпляра.
	StopInstance(ctx context.Context, nodeAddress, apiKey string, instanceID int64, timeoutSec uint32) error

	// StreamLogs открывает поток журналов инстанса. Читатель должен закрыть stream
	// для освобождения ресурсов.
	StreamLogs(ctx context.Context, nodeAddress, apiKey string, req StreamLogsRequest) (LogStream, error)

	// GetNodeInfo запрашивает статические характеристики ноды.
	GetNodeInfo(ctx context.Context, nodeAddress, apiKey string) (*NodeInfo, error)

	// Heartbeat получает текущую загруженность ноды.
	Heartbeat(ctx context.Context, nodeAddress, apiKey string) (*ResourceUsage, error)

	// ListInstances возвращает все экземпляры на ноде.
	ListInstances(ctx context.Context, nodeAddress, apiKey string) ([]*Instance, error)

	// GetInstance возвращает экземпляр по идентификатору.
	GetInstance(ctx context.Context, nodeAddress, apiKey string, instanceID int64) (*Instance, error)

	// GetInstanceUsage возвращает потребление ресурсов конкретным инстансом.
	GetInstanceUsage(ctx context.Context, nodeAddress, apiKey string, instanceID int64) (*ResourceUsage, error)
}

// LogStream представляет поток журнальных записей от ноды.
type LogStream interface {
	// Recv возвращает следующую журнальную запись. Возвращает io.EOF при завершении потока.
	Recv() (*LogEntry, error)

	// Close закрывает поток.
	Close() error
}

// ImageMetadata описывает загружаемый Docker-образ.
type ImageMetadata struct {
	GameID   int64
	ImageTag string
}

// ImageLoadResult описывает результат загрузки Docker-образа.
type ImageLoadResult struct {
	ImageTag  string
	SizeBytes uint64
}

// StartInstanceRequest содержит параметры запуска экземпляра на ноде.
type StartInstanceRequest struct {
	GameID           int64
	Name             string
	Protocol         Protocol
	InternalPort     uint32
	PortAllocation   PortAllocation
	MaxPlayers       uint32
	DeveloperPayload map[string]string
	EnvVars          map[string]string
	Args             []string
	ResourceLimits   *ResourceLimits
}

// StartInstanceResult содержит результат запуска экземпляра.
type StartInstanceResult struct {
	InstanceID int64
	HostPort   uint32
}

// StreamLogsRequest содержит параметры запроса журналов.
type StreamLogsRequest struct {
	InstanceID   int64
	FollowStdout bool
	FollowStderr bool
	Tail         uint32
}

// NodeInfo описывает статические характеристики ноды.
type NodeInfo struct {
	Region           string
	CPUCores         uint32
	TotalMemoryBytes uint64
	TotalDiskBytes   uint64
	NetworkBandwidth uint64
	AgentVersion     string
}
