package domain

import "time"

type InstanceStatus uint8

const (
	InstanceStatusStarting InstanceStatus = iota + 1
	InstanceStatusRunning
	InstanceStatusStopping
	InstanceStatusStopped
	InstanceStatusCrashed
)

type Protocol uint8

const (
	ProtocolTCP Protocol = iota + 1
	ProtocolUDP
	ProtocolWebSocket
	ProtocolWebRTC
)

type Instance struct {
	ID               int64
	ContainerID      string // docker id
	ImageTag         string // docker image
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
