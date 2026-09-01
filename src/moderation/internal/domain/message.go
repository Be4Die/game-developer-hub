package domain

import "time"

// SenderRole роль автора сообщения в чате проекта.
type SenderRole int16

const (
	SenderRoleUnspecified SenderRole = 0
	SenderRoleDeveloper   SenderRole = 1
	SenderRoleModerator   SenderRole = 2
	SenderRoleSystem      SenderRole = 3
)

// MessageType тип события в чате проекта.
type MessageType int16

const (
	MessageTypeUnspecified   MessageType = 0
	MessageTypeText          MessageType = 1 // Обычное текстовое сообщение
	MessageTypeSubmitted     MessageType = 2 // Системное: проект отправлен на проверку
	MessageTypeStatusChanged MessageType = 3 // Системное: модератор взял заявку в работу
	MessageTypeApproved      MessageType = 4 // Системное: проект успешно одобрен
	MessageTypeRejected      MessageType = 5 // Системное: проект отклонен
)

// ChatMessage представляет реплику или системное событие в рамках чата проекта.
type ChatMessage struct {
	ID          int64
	ProjectID   int64
	RequestID   *int64
	SenderID    string
	SenderRole  SenderRole
	MessageType MessageType
	Content     string
	Payload     map[string]any
	CreatedAt   time.Time
}
