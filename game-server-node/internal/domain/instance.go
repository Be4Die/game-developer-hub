// Package domain определяет бизнес-модели и интерфейсы ядра системы.
package domain

import "time"

// InstanceStatus описывает состояние игрового инстанса.
type InstanceStatus uint8

// Состояния игрового инстанса.
const (
	InstanceStatusStarting InstanceStatus = iota + 1
	InstanceStatusRunning
	InstanceStatusStopping
	InstanceStatusStopped
	InstanceStatusCrashed
)

// Protocol определяет сетевой протокол инстанса.
type Protocol uint8

// Сетевые протоколы инстансов.
const (
	ProtocolTCP Protocol = iota + 1
	ProtocolUDP
	ProtocolWebSocket
	ProtocolWebRTC
)

// Instance представляет запущенный игровой сервер.
type Instance struct {
	ID               int64
	ContainerID      string
	ImageTag         string
	Name             string
	GameID           int64
	BuildVersion     string
	Port             uint32
	Protocol         Protocol
	Status           InstanceStatus
	PlayerCount      *uint32
	MaxPlayers       uint32
	DeveloperPayload map[string]string
	StartedAt        time.Time
}
