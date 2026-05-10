package domain

import "time"

// OrchestrationMode определяет режим автоматической оркестрации серверов проекта.
type OrchestrationMode uint8

// Режимы оркестрации.
const (
	OrchestrationModeDisabled OrchestrationMode = iota + 1
	OrchestrationModeKeepAlive
	OrchestrationModeScaleToZero
)

func (m OrchestrationMode) String() string {
	switch m {
	case OrchestrationModeDisabled:
		return "disabled"
	case OrchestrationModeKeepAlive:
		return "keep_alive"
	case OrchestrationModeScaleToZero:
		return "scale_to_zero"
	default:
		return "unknown"
	}
}

// ScaleBehavior определяет поведение при переполнении инстанса.
type ScaleBehavior uint8

// Варианты поведения при переполнении.
const (
	ScaleBehaviorSpawn ScaleBehavior = iota + 1
	ScaleBehaviorQueue
)

func (b ScaleBehavior) String() string {
	switch b {
	case ScaleBehaviorSpawn:
		return "spawn"
	case ScaleBehaviorQueue:
		return "queue"
	default:
		return "unknown"
	}
}

// GamePolicy описывает правила автоматической оркестрации серверов для конкретной игры.
type GamePolicy struct {
	GameID                int64
	OwnerID               string // ID владельца проекта (из JWT)
	Mode                  OrchestrationMode
	TargetInstances       int32
	AutoRestart           bool
	ScaleToZeroTimeout    int32 // минут
	DefaultBuildVersion   string
	MaxPlayersPerInstance int32
	MaxInstancesPerGame   int32
	ScaleBehavior         ScaleBehavior
	NodePreference        string // "auto" или "node-<id>"
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// IsAuto возвращает true, если режим требует автоматического вмешательства системы.
func (p *GamePolicy) IsAuto() bool {
	return p.Mode == OrchestrationModeKeepAlive || p.Mode == OrchestrationModeScaleToZero
}
