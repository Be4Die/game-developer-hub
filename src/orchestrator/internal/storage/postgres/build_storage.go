package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BuildStorage реализует domain.BuildStorage поверх PostgreSQL.
// Хранит только метаданные билдов. Файлы хранятся отдельно в BuildStorageFS.
// Безопасен для конкурентного использования.
type BuildStorage struct {
	pool *pgxpool.Pool
}

// NewBuildStorage создаёт хранилище метаданных билдов. Требует инициализированный pool.
func NewBuildStorage(pool *pgxpool.Pool) *BuildStorage {
	return &BuildStorage{pool: pool}
}

// Create регистрирует новый билд. Возвращает ErrAlreadyExists при дубликате версии.
func (s *BuildStorage) Create(ctx context.Context, build *domain.ServerBuild) error {
	const q = `
		INSERT INTO server_builds (id, owner_id, game_id, uploaded_by, version, image_tag,
		                           protocol, internal_port, max_players,
		                           file_url, file_size, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`

	_, err := s.pool.Exec(ctx, q,
		build.ID, build.OwnerID, build.GameID, build.UploadedBy, build.Version, build.ImageTag,
		build.Protocol, build.InternalPort, build.MaxPlayers,
		build.FileURL, build.FileSize, build.CreatedAt,
	)
	if err != nil {
		if isPgUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("postgres.BuildStorage.Create: %w", err)
	}

	return nil
}

// GetByID возвращает билд по идентификатору. Возвращает ErrNotFound при отсутствии.
func (s *BuildStorage) GetByID(ctx context.Context, id int64) (*domain.ServerBuild, error) {
	const q = `
		SELECT id, owner_id, game_id, uploaded_by, version, image_tag,
		       protocol, internal_port, max_players,
		       file_url, file_size, created_at
		FROM server_builds WHERE id = $1
	`

	row := s.pool.QueryRow(ctx, q, id)
	return scanBuild(row)
}

// GetByVersion возвращает билд по game_id и версии. Возвращает ErrNotFound при отсутствии.
func (s *BuildStorage) GetByVersion(ctx context.Context, gameID int64, version string) (*domain.ServerBuild, error) {
	const q = `
		SELECT id, owner_id, game_id, uploaded_by, version, image_tag,
		       protocol, internal_port, max_players,
		       file_url, file_size, created_at
		FROM server_builds WHERE game_id = $1 AND version = $2
	`

	row := s.pool.QueryRow(ctx, q, gameID, version)
	return scanBuild(row)
}

// ListByGame возвращает билды игры, отсортированные по дате (новые первыми).
// Limit ограничивает количество (0 — без ограничения).
func (s *BuildStorage) ListByGame(ctx context.Context, gameID int64, limit int) ([]*domain.ServerBuild, error) {
	q := `
		SELECT id, owner_id, game_id, uploaded_by, version, image_tag,
		       protocol, internal_port, max_players,
		       file_url, file_size, created_at
		FROM server_builds WHERE game_id = $1
		ORDER BY created_at DESC
	`

	args := []any{gameID}
	if limit > 0 {
		q += " LIMIT $2"
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres.BuildStorage.ListByGame: %w", err)
	}
	defer rows.Close()

	var builds []*domain.ServerBuild
	for rows.Next() {
		b, err := scanBuild(rows)
		if err != nil {
			return nil, err
		}
		builds = append(builds, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.BuildStorage.ListByGame: %w", err)
	}

	return builds, nil
}

// CountByGame возвращает количество билдов для игры.
func (s *BuildStorage) CountByGame(ctx context.Context, gameID int64) (int, error) {
	const q = `SELECT COUNT(*) FROM server_builds WHERE game_id = $1`

	var count int
	err := s.pool.QueryRow(ctx, q, gameID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("postgres.BuildStorage.CountByGame: %w", err)
	}

	return count, nil
}

// Delete удаляет билд из хранилища. Возвращает ErrNotFound при отсутствии.
func (s *BuildStorage) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM server_builds WHERE id = $1`

	tag, err := s.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("postgres.BuildStorage.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CountActiveInstancesByBuild возвращает количество активных инстансов, использующих билд.
// Активные — это инстансы в статусах starting, running, stopping.
func (s *BuildStorage) CountActiveInstancesByBuild(ctx context.Context, buildID int64) (int, error) {
	const q = `
		SELECT COUNT(*) FROM instances
		WHERE server_build_id = $1
		  AND status IN ($2,$3,$4)
	`

	var count int
	err := s.pool.QueryRow(ctx, q,
		buildID,
		domain.InstanceStatusStarting,
		domain.InstanceStatusRunning,
		domain.InstanceStatusStopping,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("postgres.BuildStorage.CountActiveInstancesByBuild: %w", err)
	}

	return count, nil
}

type buildScanner interface {
	Scan(dest ...any) error
}

func scanBuild(s buildScanner) (*domain.ServerBuild, error) {
	b := &domain.ServerBuild{}
	err := s.Scan(
		&b.ID, &b.OwnerID, &b.GameID, &b.UploadedBy, &b.Version, &b.ImageTag,
		&b.Protocol, &b.InternalPort, &b.MaxPlayers,
		&b.FileURL, &b.FileSize, &b.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.scanBuild: %w", err)
	}

	return b, nil
}
