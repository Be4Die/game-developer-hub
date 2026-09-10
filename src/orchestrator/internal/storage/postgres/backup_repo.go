package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BackupRepo реализует domain.BackupRepo поверх PostgreSQL.
type BackupRepo struct {
	pool *pgxpool.Pool
}

// NewBackupRepo создает репозиторий резервных копий.
func NewBackupRepo(pool *pgxpool.Pool) *BackupRepo {
	return &BackupRepo{pool: pool}
}

// Create сохраняет запись о бэкапе в базе данных.
func (r *BackupRepo) Create(ctx context.Context, b *domain.ServiceBackup) error {
	const q = `
		INSERT INTO node_service_backups (
			backup_id, node_id, service_id, service_name, service_type,
			file_name, size_bytes, checksum, backup_type, status, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (backup_id) DO UPDATE SET
			size_bytes = EXCLUDED.size_bytes,
			checksum = EXCLUDED.checksum,
			status = EXCLUDED.status
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, q,
		b.BackupID, b.NodeID, b.ServiceID, b.ServiceName, b.ServiceType,
		b.FileName, b.SizeBytes, b.Checksum, b.BackupType, b.Status, b.CreatedAt,
	).Scan(&b.ID)
	if err != nil {
		return fmt.Errorf("postgres.BackupRepo.Create: %w", err)
	}

	return nil
}

// GetByID возвращает запись о бэкапе по backup_id.
func (r *BackupRepo) GetByID(ctx context.Context, backupID string) (*domain.ServiceBackup, error) {
	const q = `
		SELECT id, backup_id, node_id, service_id, service_name, service_type,
		       file_name, size_bytes, checksum, backup_type, status, created_at
		FROM node_service_backups
		WHERE backup_id = $1
	`

	b := &domain.ServiceBackup{}
	err := r.pool.QueryRow(ctx, q, backupID).Scan(
		&b.ID, &b.BackupID, &b.NodeID, &b.ServiceID, &b.ServiceName, &b.ServiceType,
		&b.FileName, &b.SizeBytes, &b.Checksum, &b.BackupType, &b.Status, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.BackupRepo.GetByID: %w", err)
	}

	return b, nil
}

// ListByService возвращает список всех бэкапов для конкретного сервиса на ноде.
func (r *BackupRepo) ListByService(ctx context.Context, nodeID int64, serviceName string) ([]*domain.ServiceBackup, error) {
	const q = `
		SELECT id, backup_id, node_id, service_id, service_name, service_type,
		       file_name, size_bytes, checksum, backup_type, status, created_at
		FROM node_service_backups
		WHERE node_id = $1 AND service_name = $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, q, nodeID, serviceName)
	if err != nil {
		return nil, fmt.Errorf("postgres.BackupRepo.ListByService: %w", err)
	}
	defer rows.Close()

	var backups []*domain.ServiceBackup
	for rows.Next() {
		b := &domain.ServiceBackup{}
		err := rows.Scan(
			&b.ID, &b.BackupID, &b.NodeID, &b.ServiceID, &b.ServiceName, &b.ServiceType,
			&b.FileName, &b.SizeBytes, &b.Checksum, &b.BackupType, &b.Status, &b.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres.BackupRepo.ListByService scan: %w", err)
		}
		backups = append(backups, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.BackupRepo.ListByService rows: %w", err)
	}

	return backups, nil
}

// Update обновляет статус и параметры бэкапа.
func (r *BackupRepo) Update(ctx context.Context, b *domain.ServiceBackup) error {
	const q = `
		UPDATE node_service_backups
		SET size_bytes = $2, checksum = $3, status = $4
		WHERE backup_id = $1
	`

	tag, err := r.pool.Exec(ctx, q, b.BackupID, b.SizeBytes, b.Checksum, b.Status)
	if err != nil {
		return fmt.Errorf("postgres.BackupRepo.Update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete удаляет запись о бэкапе по его backup_id.
func (r *BackupRepo) Delete(ctx context.Context, backupID string) error {
	const q = `DELETE FROM node_service_backups WHERE backup_id = $1`

	tag, err := r.pool.Exec(ctx, q, backupID)
	if err != nil {
		return fmt.Errorf("postgres.BackupRepo.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
