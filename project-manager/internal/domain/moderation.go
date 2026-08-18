package domain

import "time"

// ModerationStatus определяет текущее состояние рассмотрения заявки модератором.
type ModerationStatus int16

const (
	// ModerationStatusPending обозначает ожидание проверки модератором.
	ModerationStatusPending ModerationStatus = 1
	// ModerationStatusApproved обозначает одобрение проекта модератором.
	ModerationStatusApproved ModerationStatus = 2
	// ModerationStatusRejected обозначает отклонение проекта модератором с указанием причин.
	ModerationStatusRejected ModerationStatus = 3
)

// ModerationTicket представляет тикет проверки проекта модератором.
// Фиксирует снимок состояния черновика на момент отправки на модерацию.
type ModerationTicket struct {
	ID                 int64
	ProjectID          int64
	OwnerID            string
	GameTitle          string
	GameDescription    string
	Status             ModerationStatus
	SnapshotMeta       string // JSON снимок метаданных черновика
	RejectionReason    string
	ModeratorID        string
	DevURL             string
	ActiveBuildVersion string
	SubmittedAt        time.Time
	ResolvedAt         *time.Time
}
