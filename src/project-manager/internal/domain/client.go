package domain

import "context"

// ModerationClient определяет контракт взаимодействия с подсистемой модерации.
type ModerationClient interface {
	// SubmitDraft отправляет снимок черновика на проверку модератору.
	SubmitDraft(ctx context.Context, snapshot *ProjectSnapshot) (int64, error)
	// GetLatestRequest возвращает базовую информацию о последней заявке на модерацию.
	GetLatestRequest(ctx context.Context, projectID int64) (*ModerationRequestInfo, error)
}

// ProjectSnapshot представляет снимок состояния черновика игры на момент отправки на модерацию.
type ProjectSnapshot struct {
	ProjectID          int64
	OwnerID            string
	TitleRu            string
	TitleEn            string
	AboutRu            string
	AboutEn            string
	SeoRu              string
	SeoEn              string
	IconPath           string
	CoverPath          string
	VideoPath          string
	ActiveBuildVersion string
	DevURL             string
}

// ModerationRequestInfo содержит статус заявки на модерацию.
type ModerationRequestInfo struct {
	RequestID       int64
	ProjectID       int64
	Status          int16
	RejectionReason string
}
