// Package domain определяет бизнес-модели и интерфейсы ядра оркестратора.
package domain

// ─── InstanceStatus ──────────────────────────────────────────────────────────

// InstanceStatus описывает состояние экземпляра игрового сервера.
type InstanceStatus uint8

// Состояния экземпляра сервера.
const (
	InstanceStatusStarting InstanceStatus = iota + 1
	InstanceStatusRunning
	InstanceStatusStopping
	InstanceStatusStopped
	InstanceStatusCrashed
)

func (s InstanceStatus) String() string {
	switch s {
	case InstanceStatusStarting:
		return "starting"
	case InstanceStatusRunning:
		return "running"
	case InstanceStatusStopping:
		return "stopping"
	case InstanceStatusStopped:
		return "stopped"
	case InstanceStatusCrashed:
		return "crashed"
	default:
		return "unknown"
	}
}

// ─── Protocol ────────────────────────────────────────────────────────────────

// Protocol определяет сетевой протокол игрового сервера.
type Protocol uint8

// Сетевые протоколы.
const (
	ProtocolTCP Protocol = iota + 1
	ProtocolUDP
	ProtocolWebSocket
	ProtocolWebRTC
)

func (p Protocol) String() string {
	switch p {
	case ProtocolTCP:
		return "tcp"
	case ProtocolUDP:
		return "udp"
	case ProtocolWebSocket:
		return "websocket"
	case ProtocolWebRTC:
		return "webrtc"
	default:
		return "unknown"
	}
}

// ─── NodeStatus ──────────────────────────────────────────────────────────────

// NodeStatus описывает состояние вычислительной ноды.
type NodeStatus uint8

// Состояния ноды.
const (
	NodeStatusUnauthorized NodeStatus = iota + 1
	NodeStatusOnline
	NodeStatusOffline
	NodeStatusMaintenance
)

func (s NodeStatus) String() string {
	switch s {
	case NodeStatusUnauthorized:
		return "unauthorized"
	case NodeStatusOnline:
		return "online"
	case NodeStatusOffline:
		return "offline"
	case NodeStatusMaintenance:
		return "maintenance"
	default:
		return "unknown"
	}
}

// ─── LogSource ───────────────────────────────────────────────────────────────

// LogSource определяет источник строки журнала.
type LogSource uint8

// Источники журнальных записей.
const (
	LogSourceStdout LogSource = iota + 1
	LogSourceStderr
)

func (s LogSource) String() string {
	switch s {
	case LogSourceStdout:
		return "stdout"
	case LogSourceStderr:
		return "stderr"
	default:
		return "unknown"
	}
}

// ─── QueueStatus ─────────────────────────────────────────────────────────────

// QueueStatus описывает состояние игрока в очереди.
type QueueStatus uint8

// Состояния очереди.
const (
	QueueStatusWaiting QueueStatus = iota + 1
	QueueStatusReserved
	QueueStatusExpired
)

func (s QueueStatus) String() string {
	switch s {
	case QueueStatusWaiting:
		return "waiting"
	case QueueStatusReserved:
		return "reserved"
	case QueueStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// ─── QueueEventType ──────────────────────────────────────────────────────────

// QueueEventType описывает тип события в аудит-логе очереди.
type QueueEventType uint8

// Типы событий очереди.
const (
	QueueEventJoin QueueEventType = iota + 1
	QueueEventReserved
	QueueEventConnected
	QueueEventTimeout
	QueueEventLeave
	QueueEventCancel
)

// ─── NodeRole ────────────────────────────────────────────────────────────────

// NodeRole определяет роль вычислительной ноды.
type NodeRole uint8

const (
	NodeRoleUnspecified NodeRole = iota
	NodeRoleMixed
	NodeRoleCompute
	NodeRoleStorage
)

func (r NodeRole) String() string {
	switch r {
	case NodeRoleMixed:
		return "mixed"
	case NodeRoleCompute:
		return "compute"
	case NodeRoleStorage:
		return "storage"
	default:
		return "unspecified"
	}
}

// ─── ServiceType ─────────────────────────────────────────────────────────────

// ServiceType определяет тип управляемого сервиса хранения данных.
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

func (t ServiceType) String() string {
	switch t {
	case ServiceTypePostgres:
		return "postgres"
	case ServiceTypeRedis:
		return "redis"
	case ServiceTypeMySQL:
		return "mysql"
	case ServiceTypeVolume:
		return "volume"
	case ServiceTypeAdminer:
		return "adminer"
	case ServiceTypePGAdmin:
		return "pgadmin"
	default:
		return "unknown"
	}
}

// ─── ServiceStatus ───────────────────────────────────────────────────────────

// ServiceStatus описывает состояние сервиса данных.
type ServiceStatus uint8

const (
	ServiceStatusUnspecified ServiceStatus = iota
	ServiceStatusStarting
	ServiceStatusRunning
	ServiceStatusStopped
	ServiceStatusError
)

func (s ServiceStatus) String() string {
	switch s {
	case ServiceStatusStarting:
		return "starting"
	case ServiceStatusRunning:
		return "running"
	case ServiceStatusStopped:
		return "stopped"
	case ServiceStatusError:
		return "error"
	default:
		return "unknown"
	}
}

