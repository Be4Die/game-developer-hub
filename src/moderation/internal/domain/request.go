// Package domain содержит бизнес-сущности и правила предметной области модерации.
package domain

import "time"

// RequestStatus статус рассмотрения запроса на модерацию.
type RequestStatus int16

// Константы статусов рассмотрения запроса.
const (
	RequestStatusUnspecified RequestStatus = 0
	RequestStatusPending     RequestStatus = 1 // Ожидает назначения модератора
	RequestStatusInReview    RequestStatus = 2 // В процессе проверки модератором
	RequestStatusApproved    RequestStatus = 3 // Одобрено и опубликовано
	RequestStatusRejected    RequestStatus = 4 // Отклонено с указанием причины
	RequestStatusCancelled   RequestStatus = 5 // Отозвано разработчиком
)

// RequestType тип заявки на модерацию.
type RequestType int16

const (
	RequestTypeUnspecified        RequestType = 0
	RequestTypeProjectPublication RequestType = 1 // Публикация проекта
	RequestTypeServerAccess       RequestType = 2 // Доступ к серверам платформы
)

// ProjectSnapshot фиксирует неизменяемый снимок данных черновика на момент отправки на модерацию.
type ProjectSnapshot struct {
	ProjectID          int64  `json:"project_id"`
	TitleRu            string `json:"title_ru"`
	TitleEn            string `json:"title_en"`
	SeoRu              string `json:"seo_ru"`
	SeoEn              string `json:"seo_en"`
	AboutRu            string `json:"about_ru"`
	AboutEn            string `json:"about_en"`
	IconPath           string `json:"icon_path"`
	CoverPath          string `json:"cover_path"`
	VideoPath          string `json:"video_path"`
	ActiveBuildVersion string `json:"active_build_version"`
	DevURL             string `json:"dev_url"`
	IsOnline           bool   `json:"is_online"`
}

// ModerationRequest представляет заявку на модерацию игрового проекта.
type ModerationRequest struct {
	ID              int64
	ProjectID       int64
	OwnerID         string
	ModeratorID     string
	Status          RequestStatus
	Type            RequestType
	Snapshot        ProjectSnapshot
	Reason          string
	MaxInstances         int32
	MaxTotalCPUMillis    uint32
	MaxTotalMemoryMB     uint64
	MaxInstanceCPUMillis uint32
	MaxInstanceMemoryMB  uint64
	ModeratorComment     string
	RejectionReason      string
	SubmittedAt          time.Time
	StartedReviewAt      *time.Time
	ResolvedAt           *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
