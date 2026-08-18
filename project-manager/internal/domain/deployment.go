package domain

import "time"

// DeploymentEnvironment определяет целевое окружение для развертывания веб-игры.
type DeploymentEnvironment int16

const (
	// DeploymentEnvDev обозначает тестовое рабочее окружение для разработчика и модератора.
	DeploymentEnvDev DeploymentEnvironment = 1
	// DeploymentEnvProd обозначает продуктивное окружение, доступное игрокам.
	DeploymentEnvProd DeploymentEnvironment = 2
)

// DeploymentStatus определяет результат операции развертывания.
type DeploymentStatus int16

const (
	// DeploymentStatusPending обозначает процесс подготовки или распаковки файлов.
	DeploymentStatusPending DeploymentStatus = 1
	// DeploymentStatusSuccess обозначает успешное завершение развертывания.
	DeploymentStatusSuccess DeploymentStatus = 2
	// DeploymentStatusFailed обозначает ошибку в процессе распаковки или публикации.
	DeploymentStatusFailed DeploymentStatus = 3
)

// DeploymentRecord представляет запись аудита истории развертываний проекта.
type DeploymentRecord struct {
	ID           int64
	ProjectID    int64
	Environment  DeploymentEnvironment
	Version      string
	Status       DeploymentStatus
	ErrorMessage string
	DeployedAt   time.Time
}

// DeploymentResult содержит результат выполнения операции развертывания драйвером.
type DeploymentResult struct {
	URL          string
	UnpackedPath string
	Success      bool
}
