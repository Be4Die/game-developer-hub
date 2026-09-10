package domain

import "time"

// ServiceType определяет тип управляемого сервиса.
type ServiceType uint8

const (
	ServiceTypeUnspecified ServiceType = iota
	ServiceTypePostgres
	ServiceTypeRedis
	ServiceTypeMySQL
	_ // formerly MinIO
	ServiceTypeVolume
	ServiceTypeAdminer
	ServiceTypePGAdmin
)

// DeployServiceRequest содержит параметры развертывания сервиса.
type DeployServiceRequest struct {
	ServiceType ServiceType
	Name        string
	Port        uint32
	EnvVars     map[string]string
	VolumeName  string
}

// DeployServiceResult возвращает результат создания сервиса.
type DeployServiceResult struct {
	Name          string
	ContainerID   string
	HostPort      uint32
	ConnectionURI string
	VolumePath    string
}

// ServiceInfo содержит информацию о работающем сервисе на ноде.
type ServiceInfo struct {
	Name            string
	ServiceType     ServiceType
	ContainerID     string
	Status          string
	HostPort        uint32
	VolumePath      string
	VolumeSizeBytes uint64
	AutoBackupEnabled bool
}

// BackupType определяет тип бэкапа.
type BackupType uint8

const (
	BackupTypeUnspecified BackupType = iota
	BackupTypeManual
	BackupTypeUploaded
	BackupTypeScheduled
)

// BackupStatus определяет статус создания/восстановления бэкапа.
type BackupStatus uint8

const (
	BackupStatusUnspecified BackupStatus = iota
	BackupStatusCreating
	BackupStatusReady
	BackupStatusFailed
	BackupStatusRestoring
)

// BackupInfo описывает файл резервной копии на ноде.
type BackupInfo struct {
	BackupID    string       `json:"backup_id"`
	ServiceName string       `json:"service_name"`
	FileName    string       `json:"file_name"`
	SizeBytes   uint64       `json:"size_bytes"`
	BackupType  BackupType   `json:"backup_type"`
	Status      BackupStatus `json:"status"`
	Checksum    string       `json:"checksum"`
	CreatedAt   time.Time    `json:"created_at"`
}

