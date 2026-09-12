package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AttachmentRepo реализует domain.AttachmentRepo поверх PostgreSQL.
type AttachmentRepo struct {
	pool *pgxpool.Pool
}

// NewAttachmentRepo создаёт новый экземпляр AttachmentRepo.
func NewAttachmentRepo(pool *pgxpool.Pool) *AttachmentRepo {
	return &AttachmentRepo{pool: pool}
}

// Create сохраняет запись о новом вложении.
func (r *AttachmentRepo) Create(ctx context.Context, att *domain.Attachment) error {
	query := `
		INSERT INTO moderation_attachments (
			id, project_id, message_id, uploader_id, uploader_role,
			file_name, file_size, mime_type, storage_path, is_purged, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW()
		)
		RETURNING created_at
	`

	var createdAt time.Time
	err := r.pool.QueryRow(
		ctx, query,
		att.ID,
		att.ProjectID,
		att.MessageID,
		att.UploaderID,
		int16(att.UploaderRole),
		att.FileName,
		att.FileSize,
		att.MimeType,
		att.StoragePath,
		att.IsPurged,
	).Scan(&createdAt)
	if err != nil {
		return fmt.Errorf("AttachmentRepo.Create: %w", err)
	}

	att.CreatedAt = createdAt
	return nil
}

// Get возвращает вложение по ID и ProjectID.
func (r *AttachmentRepo) Get(ctx context.Context, id string, projectID int64) (*domain.Attachment, error) {
	query := `
		SELECT id, project_id, message_id, uploader_id, uploader_role,
		       file_name, file_size, mime_type, storage_path, is_purged, created_at
		FROM moderation_attachments
		WHERE id = $1 AND project_id = $2
	`

	att := &domain.Attachment{}
	var roleInt int16
	err := r.pool.QueryRow(ctx, query, id, projectID).Scan(
		&att.ID,
		&att.ProjectID,
		&att.MessageID,
		&att.UploaderID,
		&roleInt,
		&att.FileName,
		&att.FileSize,
		&att.MimeType,
		&att.StoragePath,
		&att.IsPurged,
		&att.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("attachment not found: %s", id)
		}
		return nil, fmt.Errorf("AttachmentRepo.Get: %w", err)
	}

	att.UploaderRole = domain.SenderRole(roleInt)
	return att, nil
}

// GetByIDs возвращает список вложений по их идентификаторам.
func (r *AttachmentRepo) GetByIDs(ctx context.Context, ids []string) ([]*domain.Attachment, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, project_id, message_id, uploader_id, uploader_role,
		       file_name, file_size, mime_type, storage_path, is_purged, created_at
		FROM moderation_attachments
		WHERE id = ANY($1)
	`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("AttachmentRepo.GetByIDs: %w", err)
	}
	defer rows.Close()

	var attachments []*domain.Attachment
	for rows.Next() {
		att := &domain.Attachment{}
		var roleInt int16
		if err := rows.Scan(
			&att.ID,
			&att.ProjectID,
			&att.MessageID,
			&att.UploaderID,
			&roleInt,
			&att.FileName,
			&att.FileSize,
			&att.MimeType,
			&att.StoragePath,
			&att.IsPurged,
			&att.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("AttachmentRepo.GetByIDs scan: %w", err)
		}
		att.UploaderRole = domain.SenderRole(roleInt)
		attachments = append(attachments, att)
	}

	return attachments, nil
}

// BindToMessage привязывает список ранее загруженных вложений к отправленному сообщению.
func (r *AttachmentRepo) BindToMessage(ctx context.Context, attachmentIDs []string, messageID int64) error {
	if len(attachmentIDs) == 0 {
		return nil
	}

	query := `
		UPDATE moderation_attachments
		SET message_id = $1
		WHERE id = ANY($2) AND (message_id IS NULL OR message_id = $1)
	`

	_, err := r.pool.Exec(ctx, query, messageID, attachmentIDs)
	if err != nil {
		return fmt.Errorf("AttachmentRepo.BindToMessage: %w", err)
	}

	return nil
}

// ListByMessageIDs возвращает мапу вложений, сгруппированных по ID сообщений.
func (r *AttachmentRepo) ListByMessageIDs(ctx context.Context, messageIDs []int64) (map[int64][]*domain.Attachment, error) {
	result := make(map[int64][]*domain.Attachment)
	if len(messageIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT id, project_id, message_id, uploader_id, uploader_role,
		       file_name, file_size, mime_type, storage_path, is_purged, created_at
		FROM moderation_attachments
		WHERE message_id = ANY($1)
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("AttachmentRepo.ListByMessageIDs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		att := &domain.Attachment{}
		var roleInt int16
		if err := rows.Scan(
			&att.ID,
			&att.ProjectID,
			&att.MessageID,
			&att.UploaderID,
			&roleInt,
			&att.FileName,
			&att.FileSize,
			&att.MimeType,
			&att.StoragePath,
			&att.IsPurged,
			&att.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("AttachmentRepo.ListByMessageIDs scan: %w", err)
		}
		att.UploaderRole = domain.SenderRole(roleInt)

		if att.MessageID != nil {
			result[*att.MessageID] = append(result[*att.MessageID], att)
		}
	}

	return result, nil
}

// PurgeByProjectID помечает вложения проекта как очищенные и возвращает список путей файлов для удаления из S3.
func (r *AttachmentRepo) PurgeByProjectID(ctx context.Context, projectID int64) (int, []string, error) {
	query := `
		UPDATE moderation_attachments
		SET is_purged = TRUE
		WHERE project_id = $1 AND is_purged = FALSE
		RETURNING storage_path
	`

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return 0, nil, fmt.Errorf("AttachmentRepo.PurgeByProjectID: %w", err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return 0, nil, fmt.Errorf("AttachmentRepo.PurgeByProjectID scan: %w", err)
		}
		if path != "" {
			paths = append(paths, path)
		}
	}

	return len(paths), paths, nil
}
