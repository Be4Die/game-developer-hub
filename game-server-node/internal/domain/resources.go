package domain

// ResourcesUsage описывает текущее потребление ресурсов контейнером.
type ResourcesUsage struct {
	CPU     float64 // процент использования CPU
	Memory  uint64  // использованная память в байтах
	Disk    uint64  // использованное дисковое пространство в байтах
	Network uint64  // суммарный сетевой трафик в байтах
}

// ResourcesMax описывает максимальные доступные ресурсы узла.
type ResourcesMax struct {
	CPUCores         uint32
	TotalMemorySize  uint64
	TotalDiskSpace   uint64
	NetworkBandwidth uint64
}
