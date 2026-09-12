package domain

import "context"

// RequestFilter задает параметры фильтрации списка запросов на модерацию.
type RequestFilter struct {
	Status      *RequestStatus
	ModeratorID *string
	Limit       int
	Offset      int
}

// RequestRepo определяет контракт репозитория запросов на модерацию.
type RequestRepo interface {
	Create(ctx context.Context, req *ModerationRequest) (int64, error)
	Get(ctx context.Context, id int64) (*ModerationRequest, error)
	GetLatestByProject(ctx context.Context, projectID int64) (*ModerationRequest, error)
	List(ctx context.Context, filter RequestFilter) ([]*ModerationRequest, int, error)
	Claim(ctx context.Context, id int64, moderatorID string) error
	Resolve(ctx context.Context, id int64, status RequestStatus, reason, moderatorID string) error
	GetModeratorStats(ctx context.Context, moderatorID string) (*ModeratorStats, error)
	ListModeratorsStats(ctx context.Context) ([]*ModeratorStats, error)
	ListModeratorActivity(ctx context.Context, moderatorID string, actionType string, limit, offset int) ([]*ModeratorActivityItem, int, error)
}

// MessageRepo определяет контракт репозитория сообщений чата проекта.
type MessageRepo interface {
	Create(ctx context.Context, msg *ChatMessage) (int64, error)
	ListByProject(ctx context.Context, projectID int64, limit, offset int) ([]*ChatMessage, int, error)
	ListActiveChats(ctx context.Context, limit, offset int) ([]*ChatSummary, int, error)
}

// AttachmentRepo определяет контракт репозитория вложений чата.
type AttachmentRepo interface {
	Create(ctx context.Context, att *Attachment) error
	Get(ctx context.Context, id string, projectID int64) (*Attachment, error)
	GetByIDs(ctx context.Context, ids []string) ([]*Attachment, error)
	BindToMessage(ctx context.Context, attachmentIDs []string, messageID int64) error
	ListByMessageIDs(ctx context.Context, messageIDs []int64) (map[int64][]*Attachment, error)
	PurgeByProjectID(ctx context.Context, projectID int64) (int, []string, error)
}
