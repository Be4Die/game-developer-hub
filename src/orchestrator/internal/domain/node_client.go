package domain

import (
	"context"
	"io"
)

// NodeClient определяет интерфейс gRPC-клиента для управления нодой.
// Абстрагирует конкретную реализацию gRPC-подключения к game-server-node.
// apiKey — токен авторизации ноды (передается как "authorization: Bearer <apiKey>").
type NodeClient interface {
	// BuildImage отправляет исходный архив на ноду для сборки Docker-образа.
	// Метаданные отправляются первым сообщением, далее — чанки архива.
	BuildImage(ctx context.Context, nodeAddress, apiKey string, metadata BuildImageMetadata, archive io.Reader) error

	// LoadImage загружает Docker-образ на ноду. Метаданные отправляются первым сообщением,
	// далее — чанки данных. Возвращает подтверждение с размером загруженного образа.
	LoadImage(ctx context.Context, nodeAddress, apiKey string, metadata ImageMetadata, chunks io.Reader) (*ImageLoadResult, error)

	// StartInstance запускает экземпляр игрового сервера на ноде.
	StartInstance(ctx context.Context, nodeAddress, apiKey string, req StartInstanceRequest) (*StartInstanceResult, error)

	// StopInstance выполняет graceful остановку экземпляра.
	StopInstance(ctx context.Context, nodeAddress, apiKey string, instanceID int64, timeoutSec uint32) error

	// RestartInstance перезапускает работающий экземпляр (docker restart).
	RestartInstance(ctx context.Context, nodeAddress, apiKey string, instanceID int64, timeoutSec uint32) error

	// StartStoppedInstance запускает остановленный экземпляр (docker start).
	StartStoppedInstance(ctx context.Context, nodeAddress, apiKey string, instanceID int64) error

	// DeleteInstance удаляет экземпляр и его контейнер.
	DeleteInstance(ctx context.Context, nodeAddress, apiKey string, instanceID int64) error

	// StreamLogs открывает поток журналов инстанса. Читатель должен закрыть stream
	// для освобождения ресурсов.
	StreamLogs(ctx context.Context, nodeAddress, apiKey string, req StreamLogsRequest) (LogStream, error)

	// GetNodeInfo запрашивает статические характеристики ноды.
	GetNodeInfo(ctx context.Context, nodeAddress, apiKey string) (*NodeInfo, error)

	// Heartbeat получает текущую загруженность ноды и количество активных инстансов.
	Heartbeat(ctx context.Context, nodeAddress, apiKey string) (*HeartbeatResult, error)

	// ListInstances возвращает все экземпляры на ноде.
	ListInstances(ctx context.Context, nodeAddress, apiKey string) ([]*Instance, error)

	// GetInstance возвращает экземпляр по идентификатору.
	GetInstance(ctx context.Context, nodeAddress, apiKey string, instanceID int64) (*Instance, error)

	// GetInstanceUsage возвращает потребление ресурсов конкретным инстансом.
	GetInstanceUsage(ctx context.Context, nodeAddress, apiKey string, instanceID int64) (*ResourceUsage, error)

	// DeployService разворачивает управляемый сервис хранения данных на ноде.
	DeployService(ctx context.Context, nodeAddress, apiKey string, req DeployServiceRequest) (*DeployServiceResult, error)

	// RemoveService удаляет управляемый сервис с ноды.
	RemoveService(ctx context.Context, nodeAddress, apiKey string, name string, deleteVolume bool) error

	// StopService останавливает контейнер управляемого сервиса без удаления тома.
	StopService(ctx context.Context, nodeAddress, apiKey, name string) error

	// StartService запускает ранее остановленный управляемый сервис.
	StartService(ctx context.Context, nodeAddress, apiKey, name string) (uint32, string, error)

	// ListServices возвращает список управляемых сервисов на ноде.
	ListServices(ctx context.Context, nodeAddress, apiKey string) ([]ServiceInfo, error)
	
	// ToggleServiceAutoBackup включает или отключает авторасписание для сервиса.
	ToggleServiceAutoBackup(ctx context.Context, nodeAddress, apiKey, serviceName string, enabled bool) error

	// CreateServiceBackup инициирует создание бэкапа сервиса на ноде.
	CreateServiceBackup(ctx context.Context, nodeAddress, apiKey, serviceName string) (*ServiceBackup, error)

	// ListServiceBackups возвращает список бэкапов сервиса с ноды.
	ListServiceBackups(ctx context.Context, nodeAddress, apiKey, serviceName string) ([]*ServiceBackup, error)

	// RestoreServiceBackup восстанавливает сервис из бэкапа на ноде.
	RestoreServiceBackup(ctx context.Context, nodeAddress, apiKey, serviceName, backupID string) error

	// DeleteServiceBackup удаляет бэкап на ноде.
	DeleteServiceBackup(ctx context.Context, nodeAddress, apiKey, serviceName, backupID string) error

	// DownloadServiceBackup открывает поток скачивания бэкапа с ноды.
	DownloadServiceBackup(ctx context.Context, nodeAddress, apiKey, serviceName, backupID string) (io.ReadCloser, error)

	// UploadServiceBackup загружает бэкап на ноду через стрим.
	UploadServiceBackup(ctx context.Context, nodeAddress, apiKey, serviceName, fileName string, restoreImmediately bool, r io.Reader) (*ServiceBackup, error)
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

// BuildImageMetadata описывает параметры сборки Docker-образа на ноде.
type BuildImageMetadata struct {
	GameID       int64
	ImageTag     string
	InternalPort uint32
}

// ImageLoadResult описывает результат загрузки Docker-образа.
type ImageLoadResult struct {
	ImageTag  string
	SizeBytes uint64
}

// StartInstanceRequest содержит параметры запуска экземпляра на ноде.
type StartInstanceRequest struct {
	GameID           int64
	InstanceID       int64 // Если != 0 — передаём ноде, иначе нода генерирует сама
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

// HeartbeatResult содержит результат heartbeat от ноды.
type HeartbeatResult struct {
	Usage               *ResourceUsage
	ActiveInstanceCount uint32
}

// DeployServiceRequest содержит параметры развертывания управляемого сервиса на ноде.
type DeployServiceRequest struct {
	ServiceType ServiceType
	Name        string
	Port        uint32
	EnvVars     map[string]string
	VolumeName  string
}

// DeployServiceResult содержит результат развертывания сервиса.
type DeployServiceResult struct {
	Name          string
	ContainerID   string
	HostPort      uint32
	ConnectionURI string
	VolumePath    string
}

// ServiceInfo содержит информацию о сервисе на ноде.
type ServiceInfo struct {
	Name            string
	ServiceType     ServiceType
	ContainerID     string
	Status          string
	HostPort        uint32
	VolumePath      string
	VolumeSizeBytes uint64
}

