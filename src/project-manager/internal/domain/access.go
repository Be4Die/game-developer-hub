package domain

import (
	"context"
	"time"
)

// Константы гранулярных разрешений разработчиков на проект.
const (
	// PermEditInfo даёт право редактировать метаданные черновика (названия, описания, SEO).
	PermEditInfo = "PERM_EDIT_INFO"
	// PermUploadBuild даёт право загружать и удалять билды веб-игры.
	PermUploadBuild = "PERM_UPLOAD_BUILD"
	// PermUploadMedia даёт право загружать иконку, обложку и промо-видео.
	PermUploadMedia = "PERM_UPLOAD_MEDIA"
	// PermViewStats даёт право просматривать статистику проекта.
	PermViewStats = "PERM_VIEW_STATS"
	// PermManageServers даёт право управлять выделенными игровыми серверами.
	PermManageServers = "PERM_MANAGE_SERVERS"
	// PermSubmitModeration даёт право отправлять черновик на модерацию и снимать с публикации.
	PermSubmitModeration = "PERM_SUBMIT_MODERATION"
)

// AllPermissions возвращает список всех допустимых разрешений.
func AllPermissions() []string {
	return []string{
		PermEditInfo,
		PermUploadBuild,
		PermUploadMedia,
		PermViewStats,
		PermManageServers,
		PermSubmitModeration,
	}
}

// IsValidPermission проверяет, является ли разрешение валидным.
func IsValidPermission(p string) bool {
	switch p {
	case PermEditInfo, PermUploadBuild, PermUploadMedia, PermViewStats, PermManageServers, PermSubmitModeration:
		return true
	default:
		return false
	}
}

// InvitationStatus статус жизненного цикла приглашения в проект.
type InvitationStatus int16

const (
	InvitationStatusPending  InvitationStatus = 1
	InvitationStatusAccepted InvitationStatus = 2
	InvitationStatusDeclined InvitationStatus = 3
	InvitationStatusCanceled InvitationStatus = 4
)

// Member сущность участника проекта с набором выданных прав.
type Member struct {
	ID          int64
	ProjectID   int64
	UserID      string
	UserEmail   string
	UserName    string
	Permissions []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// HasPermission проверяет наличие конкретного права у участника.
func (m *Member) HasPermission(perm string) bool {
	for _, p := range m.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// Invitation сущность приглашения к совместной разработке проекта.
type Invitation struct {
	ID           int64
	ProjectID    int64
	ProjectTitle string
	ProjectIcon  string
	InviterID    string
	InviterEmail string
	InviterName  string
	InviteeID    string
	InviteeEmail string
	Permissions  []string
	Status       InvitationStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserBlock запись о блокировке пользователя в черном списке.
type UserBlock struct {
	ID               int64
	UserID           string
	BlockedUserID    string
	BlockedUserEmail string
	BlockedUserName  string
	CreatedAt        time.Time
}

// SharedProject агрегат совместного проекта с правами участника.
type SharedProject struct {
	Project     *Project
	Permissions []string
	OwnerEmail  string
	OwnerName   string
	JoinedAt    time.Time
}

// MemberRepo определяет контракт репозитория участников проектов.
type MemberRepo interface {
	Add(ctx context.Context, m *Member) error
	Get(ctx context.Context, projectID int64, userID string) (*Member, error)
	ListByProject(ctx context.Context, projectID int64) ([]*Member, error)
	ListByUser(ctx context.Context, userID string) ([]*Member, error)
	UpdatePermissions(ctx context.Context, projectID int64, userID string, permissions []string) error
	Delete(ctx context.Context, projectID int64, userID string) error
	IsMember(ctx context.Context, projectID int64, userID string) (bool, error)
}

// InvitationRepo определяет контракт репозитория приглашений.
type InvitationRepo interface {
	Create(ctx context.Context, inv *Invitation) (int64, error)
	Get(ctx context.Context, id int64) (*Invitation, error)
	GetPending(ctx context.Context, projectID int64, inviteeID string) (*Invitation, error)
	ListIncoming(ctx context.Context, inviteeID string) ([]*Invitation, error)
	ListOutgoing(ctx context.Context, inviterID string, projectID int64) ([]*Invitation, error)
	UpdateStatus(ctx context.Context, id int64, status InvitationStatus) error
	CancelAllPendingBetween(ctx context.Context, inviterID, inviteeID string) error
}

// BlockRepo определяет контракт репозитория черного списка.
type BlockRepo interface {
	Block(ctx context.Context, b *UserBlock) error
	Unblock(ctx context.Context, userID, blockedUserID string) error
	IsBlocked(ctx context.Context, blockerID, targetID string) (bool, error)
	ListBlocked(ctx context.Context, userID string) ([]*UserBlock, error)
}
