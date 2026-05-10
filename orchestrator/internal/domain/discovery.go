package domain

// DiscoveryStatus описывает состояние ответа discovery для клиента игры.
type DiscoveryStatus int32

// Возможные статусы discovery.
const (
	DiscoveryStatusUnspecified DiscoveryStatus = iota
	DiscoveryStatusReady
	DiscoveryStatusStarting
	DiscoveryStatusCapacityReached
	DiscoveryStatusUnavailable
)

func (s DiscoveryStatus) String() string {
	switch s {
	case DiscoveryStatusReady:
		return "ready"
	case DiscoveryStatusStarting:
		return "starting"
	case DiscoveryStatusCapacityReached:
		return "capacity_reached"
	case DiscoveryStatusUnavailable:
		return "unavailable"
	default:
		return "unspecified"
	}
}

// DiscoveryResult — результат вызова DiscoverServers.
// Возвращается клиентам игр для принятия решения о подключении.
type DiscoveryResult struct {
	Status  DiscoveryStatus
	Servers []ServerEndpoint
	Message string // human-readable explanation
}
