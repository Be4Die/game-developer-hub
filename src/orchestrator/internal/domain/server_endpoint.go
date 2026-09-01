package domain

// ServerEndpoint описывает данные для подключения к игровому серверу.
// Возвращается клиентам игр при discovery.
type ServerEndpoint struct {
	InstanceID  int64
	Address     string
	Port        uint32
	Protocol    Protocol
	PlayerCount *uint32
	MaxPlayers  uint32
}
