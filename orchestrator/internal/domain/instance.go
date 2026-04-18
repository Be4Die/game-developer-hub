package domain

import "time"

// Instance описывает экземпляр игрового сервера.
type Instance struct {
	ID               int64
	OwnerID          string // ID пользователя (из JWT).
	NodeID           int64
	ServerBuildID    int64
	GameID           int64
	Name             string
	BuildVersion     string
	Protocol         Protocol
	HostPort         uint32
	InternalPort     uint32
	Status           InstanceStatus
	PlayerCount      *uint32
	MaxPlayers       uint32
	DeveloperPayload map[string]string
	ServerAddress    string // IP-адрес ноды для клиентов
	StartedAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
