// Package postgres реализует адаптеры постоянного хранения данных в PostgreSQL для сервиса moderation.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Be4Die/game-developer-hub/moderation/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RequestRepo реализует domain.RequestRepo поверх PostgreSQL.
type RequestRepo struct {
	pool *pgxpool.Pool
}

// NewRequestRepo создаёт новый экземпляр RequestRepo.
func NewRequestRepo(pool *pgxpool.Pool) *RequestRepo {
	return &RequestRepo{pool: pool}
}

// Create создает новый запрос на модерацию.
func (r *RequestRepo) Create(ctx context.Context, req *domain.ModerationRequest) (int64, error) {
	snapshotJSON, err := json.Marshal(req.Snapshot)
	if err != nil {
		return 0, fmt.Errorf("marshal snapshot: %w", err)
	}

	query := `
		INSERT INTO moderation_requests (
			project_id, owner_id, moderator_id, status, snapshot_meta,
			rejection_reason, submitted_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW(), NOW()
		)
		RETURNING id, submitted_at, created_at, updated_at
	`

	var id int64
	var submittedAt, createdAt, updatedAt time.Time
	err = r.pool.QueryRow(
		ctx, query,
		req.ProjectID,
		req.OwnerID,
		req.ModeratorID,
		int16(req.Status),
		snapshotJSON,
		req.RejectionReason,
	).Scan(&id, &submittedAt, &createdAt, &updatedAt)
	if err != nil {
		return 0, fmt.Errorf("RequestRepo.Create: %w", err)
	}

	req.ID = id
	req.SubmittedAt = submittedAt
	req.CreatedAt = createdAt
	req.UpdatedAt = updatedAt

	return id, nil
}

// Get возвращает запрос на модерацию по его первичному ключу.
func (r *RequestRepo) Get(ctx context.Context, id int64) (*domain.ModerationRequest, error) {
	query := `
		SELECT
			id, project_id, owner_id, moderator_id, status, snapshot_meta,
			rejection_reason, submitted_at, started_review_at, resolved_at,
			created_at, updated_at
		FROM moderation_requests
		WHERE id = $1
	`

	req := &domain.ModerationRequest{}
	var snapshotRaw []byte
	var statusInt int16

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&req.ID,
		&req.ProjectID,
		&req.OwnerID,
		&req.ModeratorID,
		&statusInt,
		&snapshotRaw,
		&req.RejectionReason,
		&req.SubmittedAt,
		&req.StartedReviewAt,
		&req.ResolvedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("RequestRepo.Get: %w", err)
	}

	req.Status = domain.RequestStatus(statusInt)
	if len(snapshotRaw) > 0 {
		_ = json.Unmarshal(snapshotRaw, &req.Snapshot)
	}

	return req, nil
}

// GetLatestByProject возвращает последнюю отправленную заявку по ID проекта.
func (r *RequestRepo) GetLatestByProject(ctx context.Context, projectID int64) (*domain.ModerationRequest, error) {
	query := `
		SELECT
			id, project_id, owner_id, moderator_id, status, snapshot_meta,
			rejection_reason, submitted_at, started_review_at, resolved_at,
			created_at, updated_at
		FROM moderation_requests
		WHERE project_id = $1
		ORDER BY submitted_at DESC
		LIMIT 1
	`

	req := &domain.ModerationRequest{}
	var snapshotRaw []byte
	var statusInt int16

	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&req.ID,
		&req.ProjectID,
		&req.OwnerID,
		&req.ModeratorID,
		&statusInt,
		&snapshotRaw,
		&req.RejectionReason,
		&req.SubmittedAt,
		&req.StartedReviewAt,
		&req.ResolvedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("RequestRepo.GetLatestByProject: %w", err)
	}

	req.Status = domain.RequestStatus(statusInt)
	if len(snapshotRaw) > 0 {
		_ = json.Unmarshal(snapshotRaw, &req.Snapshot)
	}

	return req, nil
}

// List возвращает постраничный список запросов согласно переданным фильтрам.
func (r *RequestRepo) List(ctx context.Context, filter domain.RequestFilter) ([]*domain.ModerationRequest, int, error) {
	var whereConditions []string
	var args []any
	argIdx := 1

	if filter.Status != nil && *filter.Status != domain.RequestStatusUnspecified {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, int16(*filter.Status))
		argIdx++
	}

	if filter.ModeratorID != nil && *filter.ModeratorID != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("moderator_id = $%d", argIdx))
		args = append(args, *filter.ModeratorID)
		argIdx++
	}

	whereSQL := ""
	if len(whereConditions) > 0 {
		whereSQL = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Подсчёт общего числа записей
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM moderation_requests %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("RequestRepo.List count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT
			id, project_id, owner_id, moderator_id, status, snapshot_meta,
			rejection_reason, submitted_at, started_review_at, resolved_at,
			created_at, updated_at
		FROM moderation_requests
		%s
		ORDER BY submitted_at ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("RequestRepo.List select: %w", err)
	}
	defer rows.Close()

	var requests []*domain.ModerationRequest
	for rows.Next() {
		req := &domain.ModerationRequest{}
		var snapshotRaw []byte
		var statusInt int16

		err := rows.Scan(
			&req.ID,
			&req.ProjectID,
			&req.OwnerID,
			&req.ModeratorID,
			&statusInt,
			&snapshotRaw,
			&req.RejectionReason,
			&req.SubmittedAt,
			&req.StartedReviewAt,
			&req.ResolvedAt,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("RequestRepo.List scan: %w", err)
		}

		req.Status = domain.RequestStatus(statusInt)
		if len(snapshotRaw) > 0 {
			_ = json.Unmarshal(snapshotRaw, &req.Snapshot)
		}

		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("RequestRepo.List rows: %w", err)
	}

	return requests, total, nil
}

// Claim закрепляет заявку за модератором и переводит в статус IN_REVIEW.
func (r *RequestRepo) Claim(ctx context.Context, id int64, moderatorID string) error {
	query := `
		UPDATE moderation_requests
		SET
			moderator_id = $2,
			status = $3,
			started_review_at = COALESCE(started_review_at, NOW()),
			updated_at = NOW()
		WHERE id = $1 AND (moderator_id = '' OR moderator_id = $2) AND status IN (1, 2)
	`

	tag, err := r.pool.Exec(ctx, query, id, moderatorID, int16(domain.RequestStatusInReview))
	if err != nil {
		return fmt.Errorf("RequestRepo.Claim: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAlreadyClaimed
	}

	return nil
}

// Resolve переводит заявку в финальный статус (APPROVED или REJECTED).
func (r *RequestRepo) Resolve(ctx context.Context, id int64, status domain.RequestStatus, reason, moderatorID string) error {
	query := `
		UPDATE moderation_requests
		SET
			status = $2,
			rejection_reason = $3,
			moderator_id = CASE WHEN moderator_id = '' THEN $4 ELSE moderator_id END,
			resolved_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`

	tag, err := r.pool.Exec(ctx, query, id, int16(status), reason, moderatorID)
	if err != nil {
		return fmt.Errorf("RequestRepo.Resolve: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetModeratorStats собирает агрегированную статистику по конкретному модератору.
func (r *RequestRepo) GetModeratorStats(ctx context.Context, moderatorID string) (*domain.ModeratorStats, error) {
	stats := &domain.ModeratorStats{
		ModeratorID: moderatorID,
	}

	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 3) AS approved_count,
			COUNT(*) FILTER (WHERE status = 4) AS rejected_count,
			COUNT(*) FILTER (WHERE status = 2) AS in_review_count,
			COUNT(*) AS total_assigned,
			COALESCE(
				AVG(EXTRACT(EPOCH FROM (resolved_at - started_review_at)))
				FILTER (WHERE status IN (3, 4) AND resolved_at IS NOT NULL AND started_review_at IS NOT NULL),
				0
			) AS avg_duration_sec,
			COUNT(*) FILTER (WHERE status IN (3, 4) AND resolved_at >= CURRENT_DATE) AS today_resolved,
			COUNT(*) FILTER (WHERE status IN (3, 4) AND resolved_at >= DATE_TRUNC('week', NOW())) AS week_resolved,
			COUNT(*) FILTER (WHERE status IN (3, 4) AND resolved_at >= DATE_TRUNC('month', NOW())) AS month_resolved
		FROM moderation_requests
		WHERE moderator_id = $1
	`

	var avgDuration float64
	err := r.pool.QueryRow(ctx, query, moderatorID).Scan(
		&stats.ApprovedCount,
		&stats.RejectedCount,
		&stats.InReviewCount,
		&stats.TotalAssigned,
		&avgDuration,
		&stats.TodayResolved,
		&stats.WeekResolved,
		&stats.MonthResolved,
	)
	if err != nil {
		return nil, fmt.Errorf("RequestRepo.GetModeratorStats query: %w", err)
	}

	stats.AvgReviewDurationSeconds = int64(avgDuration)
	stats.TotalResolved = stats.ApprovedCount + stats.RejectedCount
	if stats.TotalResolved > 0 {
		stats.ApprovalRate = math.Round((float64(stats.ApprovedCount)/float64(stats.TotalResolved)*100)*10) / 10
		stats.RejectionRate = math.Round((float64(stats.RejectedCount)/float64(stats.TotalResolved)*100)*10) / 10
	}

	// Количество отправленных сообщений модератором в чатах проектов
	msgQuery := `
		SELECT COUNT(*)
		FROM moderation_messages
		WHERE sender_id = $1 AND sender_role = 2
	`
	_ = r.pool.QueryRow(ctx, msgQuery, moderatorID).Scan(&stats.MessagesSent)

	return stats, nil
}

// ListModeratorsStats собирает сводную статистику по всем модераторам системы.
func (r *RequestRepo) ListModeratorsStats(ctx context.Context) ([]*domain.ModeratorStats, error) {
	query := `
		SELECT
			moderator_id,
			COUNT(*) FILTER (WHERE status = 3) AS approved_count,
			COUNT(*) FILTER (WHERE status = 4) AS rejected_count,
			COUNT(*) FILTER (WHERE status = 2) AS in_review_count,
			COUNT(*) AS total_assigned,
			COALESCE(
				AVG(EXTRACT(EPOCH FROM (resolved_at - started_review_at)))
				FILTER (WHERE status IN (3, 4) AND resolved_at IS NOT NULL AND started_review_at IS NOT NULL),
				0
			) AS avg_duration_sec,
			COUNT(*) FILTER (WHERE status IN (3, 4) AND resolved_at >= CURRENT_DATE) AS today_resolved,
			COUNT(*) FILTER (WHERE status IN (3, 4) AND resolved_at >= DATE_TRUNC('week', NOW())) AS week_resolved,
			COUNT(*) FILTER (WHERE status IN (3, 4) AND resolved_at >= DATE_TRUNC('month', NOW())) AS month_resolved
		FROM moderation_requests
		WHERE moderator_id != ''
		GROUP BY moderator_id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("RequestRepo.ListModeratorsStats query: %w", err)
	}
	defer rows.Close()

	statsMap := make(map[string]*domain.ModeratorStats)
	var result []*domain.ModeratorStats

	for rows.Next() {
		stats := &domain.ModeratorStats{}
		var avgDuration float64
		if err := rows.Scan(
			&stats.ModeratorID,
			&stats.ApprovedCount,
			&stats.RejectedCount,
			&stats.InReviewCount,
			&stats.TotalAssigned,
			&avgDuration,
			&stats.TodayResolved,
			&stats.WeekResolved,
			&stats.MonthResolved,
		); err != nil {
			return nil, fmt.Errorf("RequestRepo.ListModeratorsStats scan: %w", err)
		}

		stats.AvgReviewDurationSeconds = int64(avgDuration)
		stats.TotalResolved = stats.ApprovedCount + stats.RejectedCount
		if stats.TotalResolved > 0 {
			stats.ApprovalRate = math.Round((float64(stats.ApprovedCount)/float64(stats.TotalResolved)*100)*10) / 10
			stats.RejectionRate = math.Round((float64(stats.RejectedCount)/float64(stats.TotalResolved)*100)*10) / 10
		}

		statsMap[stats.ModeratorID] = stats
		result = append(result, stats)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("RequestRepo.ListModeratorsStats rows: %w", err)
	}

	// Подсчёт сообщений по каждому модератору
	msgQuery := `
		SELECT sender_id, COUNT(*)
		FROM moderation_messages
		WHERE sender_role = 2 AND sender_id != ''
		GROUP BY sender_id
	`
	msgRows, err := r.pool.Query(ctx, msgQuery)
	if err == nil {
		defer msgRows.Close()
		for msgRows.Next() {
			var senderID string
			var count int32
			if err := msgRows.Scan(&senderID, &count); err == nil {
				if s, ok := statsMap[senderID]; ok {
					s.MessagesSent = count
				}
			}
		}
	}

	return result, nil
}

// ListModeratorActivity возвращает постраничный журнал всех действий конкретного модератора (сообщения, решения, закрытия диалогов).
func (r *RequestRepo) ListModeratorActivity(ctx context.Context, moderatorID string, actionType string, limit, offset int) ([]*domain.ModeratorActivityItem, int, error) {
	var whereConditions []string
	var args []any
	argIdx := 1

	whereConditions = append(whereConditions, fmt.Sprintf("m.sender_id = $%d", argIdx))
	args = append(args, moderatorID)
	argIdx++

	switch actionType {
	case "chat_message":
		whereConditions = append(whereConditions, "m.message_type = 1")
	case "dialog_closed":
		whereConditions = append(whereConditions, "m.message_type = 3 AND m.payload->>'dialog_status' = 'closed'")
	case "claimed":
		whereConditions = append(whereConditions, "m.message_type = 3 AND (m.payload->>'dialog_status' IS NULL OR m.payload->>'dialog_status' != 'closed')")
	case "approved":
		whereConditions = append(whereConditions, "m.message_type = 4")
	case "rejected":
		whereConditions = append(whereConditions, "m.message_type = 5")
	}

	whereSQL := "WHERE " + strings.Join(whereConditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM moderation_messages m %s", whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("RequestRepo.ListModeratorActivity count: %w", err)
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT
			m.id,
			m.project_id,
			m.message_type,
			m.content,
			m.payload,
			m.created_at,
			COALESCE(r.snapshot_meta->>'title_ru', r.snapshot_meta->>'title_en', '') AS project_title,
			COALESCE(r.snapshot_meta->>'icon_path', '') AS project_icon,
			COALESCE(r.snapshot_meta->>'active_build_version', '') AS build_version
		FROM moderation_messages m
		LEFT JOIN LATERAL (
			SELECT snapshot_meta
			FROM moderation_requests
			WHERE (m.request_id IS NOT NULL AND id = m.request_id)
			   OR (m.request_id IS NULL AND project_id = m.project_id)
			ORDER BY id DESC
			LIMIT 1
		) r ON true
		%s
		ORDER BY m.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("RequestRepo.ListModeratorActivity select: %w", err)
	}
	defer rows.Close()

	var items []*domain.ModeratorActivityItem
	for rows.Next() {
		var id, projectID int64
		var msgType int16
		var content string
		var payloadRaw []byte
		var createdAt time.Time
		var projectTitle, projectIcon, buildVersion string

		err := rows.Scan(
			&id,
			&projectID,
			&msgType,
			&content,
			&payloadRaw,
			&createdAt,
			&projectTitle,
			&projectIcon,
			&buildVersion,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("RequestRepo.ListModeratorActivity scan: %w", err)
		}

		item := &domain.ModeratorActivityItem{
			ID:           id,
			ProjectID:    projectID,
			ProjectTitle: projectTitle,
			ProjectIcon:  projectIcon,
			BuildVersion: buildVersion,
			CreatedAt:    createdAt,
		}

		switch msgType {
		case 1:
			item.ActionType = "chat_message"
			item.ActionTitle = "Сообщение в чате"
			item.Details = content
		case 3:
			var payload map[string]any
			if len(payloadRaw) > 0 {
				_ = json.Unmarshal(payloadRaw, &payload)
			}
			if payload != nil && payload["dialog_status"] == "closed" {
				item.ActionType = "dialog_closed"
				item.ActionTitle = "Вопрос решён"
				item.Details = content
			} else {
				item.ActionType = "claimed"
				item.ActionTitle = "Взял на проверку"
				item.Details = content
			}
		case 4:
			item.ActionType = "approved"
			item.ActionTitle = "Одобрил публикацию"
			var payload map[string]any
			if len(payloadRaw) > 0 {
				_ = json.Unmarshal(payloadRaw, &payload)
			}
			if payload != nil && payload["comment"] != nil && payload["comment"] != "" {
				item.Details = fmt.Sprintf("%v", payload["comment"])
			} else {
				item.Details = content
			}
		case 5:
			item.ActionType = "rejected"
			item.ActionTitle = "Отклонил заявку"
			var payload map[string]any
			if len(payloadRaw) > 0 {
				_ = json.Unmarshal(payloadRaw, &payload)
			}
			if payload != nil && payload["reason"] != nil && payload["reason"] != "" {
				item.Details = fmt.Sprintf("%v", payload["reason"])
			} else {
				item.Details = content
			}
		default:
			item.ActionType = "event"
			item.ActionTitle = "Действие"
			item.Details = content
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("RequestRepo.ListModeratorActivity rows: %w", err)
	}

	return items, total, nil
}
