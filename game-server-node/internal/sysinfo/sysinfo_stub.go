//go:build !linux

package sysinfo

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/domain"
)

type StubProvider struct {
	mu      sync.Mutex
	ethName string
	rng     *rand.Rand
}

func NewProvider(ethName string) *StubProvider {
	if ethName == "" {
		ethName = "eth0_mock"
	}

	log.Printf("WARNING: sysinfo is running in STUB mode (non-Linux OS). Using mock data for interface: %s\n", ethName)

	return &StubProvider{
		ethName: ethName,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *StubProvider) GetMax() (domain.ResourcesMax, error) {
	return domain.ResourcesMax{
		CPUCores:         8,
		TotalMemorySize:  16 * 1024 * 1024 * 1024,  // 16 GB
		TotalDiskSpace:   500 * 1024 * 1024 * 1024, // 500 GB
		NetworkBandwidth: 1000 * 125000,            // 1 Gbps в байтах
	}, nil
}

func (p *StubProvider) GetUsage() (domain.ResourcesUsage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return domain.ResourcesUsage{
		CPU:     30.0 + (p.rng.Float64() * 20.0),
		Memory:  uint64(6+p.rng.Intn(2)) * 1024 * 1024 * 1024,
		Disk:    250 * 1024 * 1024 * 1024,
		Network: uint64(p.rng.Intn(5000) * 1024),
	}, nil
}
