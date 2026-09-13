package domain

import (
	"fmt"
	"net"
)

// ServerEndpoint описывает данные для подключения к игровому серверу.
// Возвращается клиентам игр при discovery.
type ServerEndpoint struct {
	InstanceID  int64
	Address     string
	Port        uint32
	Protocol    Protocol
	PlayerCount *uint32
	MaxPlayers  uint32
	Path        string
}

// BuildServerEndpoint формирует эндпоинт подключения к инстансу с учетом сетевого режима ноды.
func BuildServerEndpoint(inst *Instance, node *Node, proxyHost string, proxyPort uint32, playerCount *uint32) ServerEndpoint {
	var (
		addr string
		port uint32
		path string
	)

	if node != nil && node.IngressMode == IngressModeDirect {
		if node.CustomDomain != "" {
			addr = node.CustomDomain
		} else {
			host, _, err := net.SplitHostPort(node.Address)
			if err != nil {
				host = node.Address
			}
			addr = host
		}
		port = inst.HostPort
		path = ""
	} else {
		// Platform Proxy (default)
		addr = proxyHost
		if addr == "" {
			addr = "localhost"
		}
		port = proxyPort
		if port == 0 {
			port = 80
		}
		path = fmt.Sprintf("/game-proxy/%d/%d", inst.GameID, inst.ID)
	}

	return ServerEndpoint{
		InstanceID:  inst.ID,
		Address:     addr,
		Port:        port,
		Protocol:    inst.Protocol,
		PlayerCount: playerCount,
		MaxPlayers:  inst.MaxPlayers,
		Path:        path,
	}
}
