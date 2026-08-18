package domain

import "time"

// Release представляет опубликованную версию игры в продуктивном окружении.
// Содержит неизменяемый снимок метаданных и привязку к сборке на момент релиза.
type Release struct {
	ID          int64
	ProjectID   int64
	Version     string
	TitleRu     string
	TitleEn     string
	SeoRu       string
	SeoEn       string
	About       string
	IconPath    string
	CoverPath   string
	VideoPath   string
	ProdURL     string
	IsActive    bool
	PublishedBy string
	PublishedAt time.Time
	UnpublishAt *time.Time
}
