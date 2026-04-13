package domain

// ResourceLimits описывает ограничения ресурсов для контейнера.
type ResourceLimits struct {
	CPUMillis   *uint32
	MemoryBytes *uint64
}
