package domain

// ResourceUsage описывает текущее потребление ресурсов нодой или инстансом.
type ResourceUsage struct {
	CPUUsagePercent    float64
	MemoryUsedBytes    uint64
	DiskUsedBytes      uint64
	NetworkBytesPerSec uint64
}
