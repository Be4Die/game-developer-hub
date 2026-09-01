package domain

import "time"

// Draft представляет рабочее состояние черновика игрового проекта.
// Содержит актуальные редактируемые метаданные и привязку к активной тестовой сборке.
type Draft struct {
	ProjectID          int64
	TitleRu            string
	TitleEn            string
	SeoRu              string
	SeoEn              string
	AboutRu            string
	AboutEn            string
	IconPath           string
	CoverPath          string
	VideoPath          string
	ActiveBuildVersion string
	DevURL             string
	UpdatedAt          time.Time
}

// DraftMeta содержит передаваемые пользователем метаданные для обновления черновика.
type DraftMeta struct {
	TitleRu            string
	TitleEn            string
	SeoRu              string
	SeoEn              string
	AboutRu            string
	AboutEn            string
	ActiveBuildVersion string
}

// IsReadyForModeration проверяет полноту заполнения черновика перед отправкой на модерацию.
// Возвращает nil, если все обязательные поля заполнены и выбрана активная сборка.
func (d *Draft) IsReadyForModeration() error {
	if d.TitleRu == "" && d.TitleEn == "" {
		return ErrDraftNotReady
	}
	if d.AboutRu == "" && d.AboutEn == "" {
		return ErrDraftNotReady
	}
	if d.ActiveBuildVersion == "" {
		return ErrDraftNotReady
	}
	return nil
}
