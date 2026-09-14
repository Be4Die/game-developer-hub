package domain

import "context"

// OrchestratorClient определяет контракт взаимодействия с сервисом оркестратора (orchestrator).
type OrchestratorClient interface {
	// GrantPlatformAccess выдает квоту серверов платформы проекту.
	GrantPlatformAccess(
		ctx context.Context,
		projectID int64,
		maxInstances int32,
		maxTotalCPU uint32,
		maxTotalMemoryMB uint64,
		maxInstanceCPU uint32,
		maxInstanceMemoryMB uint64,
	) error
	// RevokePlatformAccess отзывает доступ к платформенным мощностям проекта.
	RevokePlatformAccess(ctx context.Context, projectID int64) error
}
