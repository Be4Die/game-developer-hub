package postgres

import (
	"context"
	"fmt"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeploymentRepo реализует domain.DeploymentRepo для аудита развертываний.
type DeploymentRepo struct {
	pool *pgxpool.Pool
}

// NewDeploymentRepo создаёт новый экземпляр репозитория деплоев.
func NewDeploymentRepo(pool *pgxpool.Pool) *DeploymentRepo {
	return &DeploymentRepo{pool: pool}
}

// Create сохраняет запись о развертывании в базе данных.
func (r *DeploymentRepo) Create(ctx context.Context, d *domain.DeploymentRecord) error {
	const query = `
		INSERT INTO deployments (project_id, environment, version, status, error_message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, deployed_at
	`
	err := r.pool.QueryRow(ctx, query,
		d.ProjectID, d.Environment, d.Version, d.Status, d.ErrorMessage,
	).Scan(&d.ID, &d.DeployedAt)
	if err != nil {
		return fmt.Errorf("postgres.DeploymentRepo.Create: %w", err)
	}
	return nil
}

// ListByProject возвращает историю развертываний проекта.
func (r *DeploymentRepo) ListByProject(ctx context.Context, projectID int64, limit int) ([]*domain.DeploymentRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	const query = `
		SELECT id, project_id, environment, version, status, error_message, deployed_at
		FROM deployments
		WHERE project_id = $1
		ORDER BY deployed_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("postgres.DeploymentRepo.ListByProject: %w", err)
	}
	defer rows.Close()

	var records []*domain.DeploymentRecord
	for rows.Next() {
		var d domain.DeploymentRecord
		if err := rows.Scan(
			&d.ID, &d.ProjectID, &d.Environment, &d.Version, &d.Status, &d.ErrorMessage, &d.DeployedAt,
		); err != nil {
			return nil, fmt.Errorf("postgres.DeploymentRepo.ListByProject scan: %w", err)
		}
		records = append(records, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.DeploymentRepo.ListByProject rows: %w", err)
	}
	return records, nil
}
