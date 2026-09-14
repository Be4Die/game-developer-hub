package domain

import "time"

// ModerationSnapshot представляет сущность неизменяемого аудит-снимка решения модерации.
type ModerationSnapshot struct {
	ID            int64         `json:"id"`
	RequestID     int64         `json:"request_id"`
	ProjectID     int64         `json:"project_id"`
	Status        RequestStatus `json:"status"`
	FormatVersion int           `json:"format_version"`
	SnapshotJSON  string        `json:"snapshot_json"`
	SnapshotBlob  []byte        `json:"snapshot_blob"`
	CreatedAt     time.Time     `json:"created_at"`
}

// SnapshotPayload содержит полную структурированную декомпозицию снимка.
type SnapshotPayload struct {
	Project        SnapshotProjectData    `json:"project"`
	Media          SnapshotMediaData      `json:"media"`
	Verdict        SnapshotVerdictData    `json:"verdict"`
	ChatTranscript []*SnapshotMessageItem `json:"chat_transcript"`
}

// SnapshotProjectData метаданные игрового проекта на момент вердикта.
type SnapshotProjectData struct {
	ID            int64  `json:"id"`
	OwnerID       string `json:"owner_id"`
	DeveloperName string `json:"developer_name"`
	TitleRu       string `json:"title_ru"`
	TitleEn       string `json:"title_en"`
	SeoRu         string `json:"seo_ru"`
	SeoEn         string `json:"seo_en"`
	AboutRu       string `json:"about_ru"`
	AboutEn       string `json:"about_en"`
	BuildVersion  string `json:"build_version"`
	DevURL        string `json:"dev_url"`
	ProdURL       string `json:"prod_url,omitempty"`
	IsOnline      bool   `json:"is_online"`
}

// SnapshotMediaData медиа-материалы проекта со слепками и контрольными суммами.
type SnapshotMediaData struct {
	Icon  SnapshotMediaItem `json:"icon"`
	Cover SnapshotMediaItem `json:"cover"`
	Video SnapshotMediaItem `json:"video"`
}

// SnapshotMediaItem описание медиафайла с метаданными, превью и sha256.
type SnapshotMediaItem struct {
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	MimeType      string `json:"mime_type"`
	Width         int    `json:"width,omitempty"`
	Height        int    `json:"height,omitempty"`
	Sha256        string `json:"sha256"`
	ThumbnailData string `json:"thumbnail_data,omitempty"` // data:image/webp;base64,...
	OriginalURL   string `json:"original_url,omitempty"`
}

// SnapshotVerdictData зафиксированное решение модератора.
type SnapshotVerdictData struct {
	Status               RequestStatus            `json:"status"`
	ModeratorID          string                   `json:"moderator_id"`
	ModeratorName        string                   `json:"moderator_name"`
	SubmittedAt          string                   `json:"submitted_at"`
	StartedReviewAt      string                   `json:"started_review_at"`
	ResolvedAt           string                   `json:"resolved_at"`
	ReviewDuration       string                   `json:"review_duration"`
	RejectionReason      string                   `json:"rejection_reason,omitempty"`
	Comment              string                   `json:"comment,omitempty"`
	ProdURL              string                   `json:"prod_url,omitempty"`
	Violations           []*SnapshotViolationItem `json:"violations,omitempty"`
	RequestType          RequestType              `json:"request_type,omitempty"`
	Reason               string                   `json:"reason,omitempty"`
	MaxInstances         int32                    `json:"max_instances,omitempty"`
	MaxTotalCPUMillis    uint32                   `json:"max_total_cpu_millis,omitempty"`
	MaxTotalMemoryMB     uint64                   `json:"max_total_memory_mb,omitempty"`
	MaxInstanceCPUMillis uint32                   `json:"max_instance_cpu_millis,omitempty"`
	MaxInstanceMemoryMB  uint64                   `json:"max_instance_memory_mb,omitempty"`
}

// SnapshotViolationItem пункт выявленного нарушения регламента в снимке.
type SnapshotViolationItem struct {
	RuleCode    string                    `json:"rule_code"`
	RuleTitle   string                    `json:"rule_title"`
	Description string                    `json:"description"`
	Attachments []*SnapshotAttachmentItem `json:"attachments,omitempty"`
}

// SnapshotMessageItem сообщение в замороженном срезе переписки тикета.
type SnapshotMessageItem struct {
	ID          int64                     `json:"id"`
	SenderID    string                    `json:"sender_id"`
	SenderRole  SenderRole                `json:"sender_role"`
	SenderName  string                    `json:"sender_name"`
	MessageType MessageType               `json:"message_type"`
	Content     string                    `json:"content"`
	CreatedAt   string                    `json:"created_at"`
	Payload     map[string]any            `json:"payload,omitempty"`
	Attachments []*SnapshotAttachmentItem `json:"attachments,omitempty"`
}

// SnapshotAttachmentItem вложение к сообщению в срезе переписки.
type SnapshotAttachmentItem struct {
	ID            string `json:"id"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
	MimeType      string `json:"mime_type"`
	Sha256        string `json:"sha256,omitempty"`
	ThumbnailData string `json:"thumbnail_data,omitempty"`
	URL           string `json:"url,omitempty"`
	DownloadURL   string `json:"download_url,omitempty"`
}
