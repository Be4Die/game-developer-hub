package domain

// ServiceType определяет тип управляемого сервиса.
type ServiceType uint8

const (
	ServiceTypeUnspecified ServiceType = iota
	ServiceTypePostgres
	ServiceTypeRedis
	ServiceTypeMySQL
	ServiceTypeMinIO
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
}
