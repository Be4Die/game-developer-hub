package domain

// ChatSummary содержит краткую информацию о последнем сообщении в чате проекта.
type ChatSummary struct {
	ProjectID   int64
	LastMessage *ChatMessage
}
