// Package sysinfo предоставляет информацию о ресурсах системы.
package sysinfo

import (
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

// Provider предоставляет информацию о ресурсах системы.
type Provider interface {
	// GetMax возвращает максимальные доступные ресурсы узла.
	GetMax() (domain.ResourcesMax, error)
	// GetUsage возвращает текущее потребление ресурсов.
	GetUsage() (domain.ResourcesUsage, error)
}
