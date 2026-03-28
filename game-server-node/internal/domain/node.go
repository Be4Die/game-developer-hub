package domain

import "time"

type Node struct {
	Version          string
	Region           string
	CPUCores         uint32
	TotalMemorySize  uint64
	TotalDiskSpace   uint64
	NetworkBandwidth uint64
	StartedAt        time.Time
}
