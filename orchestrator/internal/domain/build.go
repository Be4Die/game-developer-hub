package domain

import "time"

// ServerBuild описывает загруженный серверный билд игры.
type ServerBuild struct {
	ID           int64
	OwnerID      string
	GameID       int64
	UploadedBy   int64 // ID пользователя (0 — неизвестно).
	Version      string
	ImageTag     string
	Protocol     Protocol
	InternalPort uint32 // порт внутри контейнера
	MaxPlayers   uint32
	FileURL      string // путь к файлу в хранилище
	FileSize     int64
	CreatedAt    time.Time
}
