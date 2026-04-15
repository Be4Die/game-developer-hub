package domain

import "time"

// Node описывает вычислительный узел в реестре оркестратора.
type Node struct {
	ID           int64
	Address      string // gRPC-адрес ноды (host:port)
	TokenHash    []byte // хеш авторизационного токена (для проверки при авторизации)
	APIToken     string // plaintext токен (для gRPC-запросов к ноде)
	Region       string // опционально
	Status       NodeStatus
	CPUCores     uint32    // получено из NodeInfo
	TotalMemory  uint64    // получено из NodeInfo
	TotalDisk    uint64    // получено из NodeInfo
	AgentVersion string    // версия агента ноды
	LastPingAt   time.Time // время последнего heartbeat
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
