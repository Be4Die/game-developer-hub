package domain

type ResourcesUsage struct {
	CPU     float64
	Memory  uint64
	Disk    uint64
	Network uint64
}

type ResourcesMax struct {
	CPUCores         uint32
	TotalMemorySize  uint64
	TotalDiskSpace   uint64
	NetworkBandwidth uint64
}
