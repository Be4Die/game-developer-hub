package sysinfo

import (
	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

type Provider interface {
	GetMax() (domain.ResourcesMax, error)
	GetUsage() (domain.ResourcesUsage, error)
}
