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
