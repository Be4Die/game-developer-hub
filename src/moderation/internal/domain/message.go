package domain

import "time"

// SenderRole роль автора сообщения в чате проекта.
type SenderRole int16

// Константы ролей отправителей.
const (
	SenderRoleUnspecified SenderRole = 0
	SenderRoleDeveloper   SenderRole = 1
	SenderRoleModerator   SenderRole = 2
	SenderRoleSystem      SenderRole = 3
)

// MessageType тип события в чате проекта.
type MessageType int16

// Константы типов сообщений.
const (
	MessageTypeUnspecified   MessageType = 0
	MessageTypeText          MessageType = 1 // Обычное текстовое сообщение
	MessageTypeSubmitted     MessageType = 2 // Системное: проект отправлен на проверку
	MessageTypeStatusChanged MessageType = 3 // Системное: модератор взял заявку в работу
	MessageTypeApproved      MessageType = 4 // Системное: проект успешно одобрен
	MessageTypeRejected      MessageType = 5 // Системное: проект отклонен
)

// Attachment представляет метаданные медиа-файла (фото или видео), прикрепленного к сообщению.
type Attachment struct {
	ID           string
	ProjectID    int64
	MessageID    *int64
	UploaderID   string
	UploaderRole SenderRole
	FileName     string
	FileSize     int64
	MimeType     string
	StoragePath  string
	IsPurged     bool
	CreatedAt    time.Time
}

// ViolationItem представляет конкретный пункт нарушения правил модерации со своими доказательствами.
type ViolationItem struct {
	RuleCode      string        `json:"rule_code"`
	RuleTitle     string        `json:"rule_title"`
	Description   string        `json:"description"`
	AttachmentIDs []string      `json:"attachment_ids,omitempty"`
	Attachments   []*Attachment `json:"attachments,omitempty"`
}

// ChatMessage представляет реплику или системное событие в рамках чата проекта.
type ChatMessage struct {
	ID          int64
	ProjectID   int64
	RequestID   *int64
	SenderID    string
	SenderRole  SenderRole
	MessageType MessageType
	Content     string
	Attachments []*Attachment
	Payload     map[string]any
	CreatedAt   time.Time
}
