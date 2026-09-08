package postgres

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BlockRepo реализует domain.BlockRepo для таблицы user_access_blocks.
type BlockRepo struct {
	pool *pgxpool.Pool
}

// NewBlockRepo создаёт репозиторий блокировок пользователей.
func NewBlockRepo(pool *pgxpool.Pool) *BlockRepo {
	return &BlockRepo{pool: pool}
}

// Block добавляет пользователя в черный список (игнорирует повторы благодаря ON CONFLICT).
func (r *BlockRepo) Block(ctx context.Context, b *domain.UserBlock) error {
	const query = `
		INSERT INTO user_access_blocks (user_id, blocked_user_id, blocked_user_email, blocked_user_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, blocked_user_id) DO UPDATE
		SET blocked_user_email = CASE WHEN EXCLUDED.blocked_user_email <> '' THEN EXCLUDED.blocked_user_email ELSE user_access_blocks.blocked_user_email END,
		    blocked_user_name = CASE WHEN EXCLUDED.blocked_user_name <> '' THEN EXCLUDED.blocked_user_name ELSE user_access_blocks.blocked_user_name END
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, b.UserID, b.BlockedUserID, b.BlockedUserEmail, b.BlockedUserName).
		Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return fmt.Errorf("postgres.BlockRepo.Block: %w", err)
	}
	return nil
}

// Unblock удаляет пользователя из черного списка.
func (r *BlockRepo) Unblock(ctx context.Context, userID, blockedUserID string) error {
	const query = `DELETE FROM user_access_blocks WHERE user_id = $1 AND blocked_user_id = $2`
	_, err := r.pool.Exec(ctx, query, userID, blockedUserID)
	if err != nil {
		return fmt.Errorf("postgres.BlockRepo.Unblock: %w", err)
	}
	return nil
}

// IsBlocked проверяет, заблокировал ли blockerID пользователя targetID.
func (r *BlockRepo) IsBlocked(ctx context.Context, blockerID, targetID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM user_access_blocks WHERE user_id = $1 AND blocked_user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, blockerID, targetID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("postgres.BlockRepo.IsBlocked: %w", err)
	}
	return exists, nil
}

// ListBlocked возвращает всех заблокированных текущим пользователем.
func (r *BlockRepo) ListBlocked(ctx context.Context, userID string) ([]*domain.UserBlock, error) {
	const query = `
		SELECT id, user_id, blocked_user_id, blocked_user_email, blocked_user_name, created_at
		FROM user_access_blocks
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres.BlockRepo.ListBlocked: %w", err)
	}
	defer rows.Close()

	var blocks []*domain.UserBlock
	for rows.Next() {
		var b domain.UserBlock
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.BlockedUserID, &b.BlockedUserEmail, &b.BlockedUserName, &b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.BlockRepo.ListBlocked scan: %w", err)
		}
		blocks = append(blocks, &b)
	}
	return blocks, nil
}
