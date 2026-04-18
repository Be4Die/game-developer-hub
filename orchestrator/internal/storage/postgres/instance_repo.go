package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InstanceRepo реализует domain.InstanceRepo поверх PostgreSQL.
// Безопасен для конкурентного использования.
type InstanceRepo struct {
	pool *pgxpool.Pool
}

// NewInstanceRepo создаёт хранилище инстансов. Требует инициализированный pool.
func NewInstanceRepo(pool *pgxpool.Pool) *InstanceRepo {
	return &InstanceRepo{pool: pool}
}

// Create регистрирует новый инстанс. Возвращает ErrAlreadyExists при дубликате.
func (r *InstanceRepo) Create(ctx context.Context, inst *domain.Instance) error {
	const q = `
		INSERT INTO instances (id, owner_id, node_id, server_build_id, game_id, name,
		                       build_version, protocol, host_port, internal_port,
		                       status, max_players, developer_payload,
		                       server_address, started_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
	`

	_, err := r.pool.Exec(ctx, q,
		inst.ID, inst.OwnerID, inst.NodeID, inst.ServerBuildID, inst.GameID, inst.Name,
		inst.BuildVersion, inst.Protocol, inst.HostPort, inst.InternalPort,
		inst.Status, inst.MaxPlayers, inst.DeveloperPayload,
		inst.ServerAddress, inst.StartedAt, inst.CreatedAt, inst.UpdatedAt,
	)
	if err != nil {
		if isPgUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("postgres.InstanceRepo.Create: %w", err)
	}

	return nil
}

// GetByID возвращает инстанс по идентификатору. Возвращает ErrNotFound при отсутствии.
func (r *InstanceRepo) GetByID(ctx context.Context, id int64) (*domain.Instance, error) {
	const q = `
		SELECT id, owner_id, node_id, server_build_id, game_id, name,
		       build_version, protocol, host_port, internal_port,
		       status, max_players, developer_payload,
		       server_address, started_at, created_at, updated_at
		FROM instances WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, q, id)
	return scanInstance(row)
}

// ListByGame возвращает все инстансы указанной игры.
// Опционально фильтрует по статусу (nil — без фильтра).
func (r *InstanceRepo) ListByGame(ctx context.Context, gameID int64, status *domain.InstanceStatus) ([]*domain.Instance, error) {
	q := `
		SELECT id, owner_id, node_id, server_build_id, game_id, name,
		       build_version, protocol, host_port, internal_port,
		       status, max_players, developer_payload,
		       server_address, started_at, created_at, updated_at
		FROM instances WHERE game_id = $1
	`

	args := []any{gameID}
	if status != nil {
		q += " AND status = $2"
		args = append(args, *status)
	}
	q += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres.InstanceRepo.ListByGame: %w", err)
	}
	defer rows.Close()

	var instances []*domain.Instance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.InstanceRepo.ListByGame: %w", err)
	}

	return instances, nil
}

// ListByNode возвращает все инстансы, запущенные на указанной ноде.
func (r *InstanceRepo) ListByNode(ctx context.Context, nodeID int64) ([]*domain.Instance, error) {
	const q = `
		SELECT id, owner_id, node_id, server_build_id, game_id, name,
		       build_version, protocol, host_port, internal_port,
		       status, max_players, developer_payload,
		       server_address, started_at, created_at, updated_at
		FROM instances WHERE node_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, q, nodeID)
	if err != nil {
		return nil, fmt.Errorf("postgres.InstanceRepo.ListByNode: %w", err)
	}
	defer rows.Close()

	var instances []*domain.Instance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.InstanceRepo.ListByNode: %w", err)
	}

	return instances, nil
}

// Update обновляет персистентные поля инстанса. Возвращает ErrNotFound при отсутствии.
func (r *InstanceRepo) Update(ctx context.Context, inst *domain.Instance) error {
	const q = `
		UPDATE instances SET owner_id=$1, node_id=$2, server_build_id=$3, game_id=$4,
		                     name=$5, build_version=$6, protocol=$7,
		                     host_port=$8, internal_port=$9, status=$10,
		                     max_players=$11, developer_payload=$12,
		                     server_address=$13, started_at=$14, updated_at=$15
		WHERE id=$16
	`

	tag, err := r.pool.Exec(ctx, q,
		inst.OwnerID, inst.NodeID, inst.ServerBuildID, inst.GameID, inst.Name,
		inst.BuildVersion, inst.Protocol, inst.HostPort, inst.InternalPort,
		inst.Status, inst.MaxPlayers, inst.DeveloperPayload,
		inst.ServerAddress, inst.StartedAt, inst.UpdatedAt, inst.ID,
	)
	if err != nil {
		return fmt.Errorf("postgres.InstanceRepo.Update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete удаляет запись инстанса из БД. Возвращает ErrNotFound при отсутствии.
func (r *InstanceRepo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM instances WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("postgres.InstanceRepo.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CountByGame возвращает количество записей инстансов для игры.
func (r *InstanceRepo) CountByGame(ctx context.Context, gameID int64) (int, error) {
	const q = `SELECT COUNT(*) FROM instances WHERE game_id = $1`

	var count int
	err := r.pool.QueryRow(ctx, q, gameID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("postgres.InstanceRepo.CountByGame: %w", err)
	}

	return count, nil
}

type instanceScanner interface {
	Scan(dest ...any) error
}

func scanInstance(s instanceScanner) (*domain.Instance, error) {
	inst := &domain.Instance{}
	err := s.Scan(
		&inst.ID, &inst.OwnerID, &inst.NodeID, &inst.ServerBuildID, &inst.GameID, &inst.Name,
		&inst.BuildVersion, &inst.Protocol, &inst.HostPort, &inst.InternalPort,
		&inst.Status, &inst.MaxPlayers, &inst.DeveloperPayload,
		&inst.ServerAddress, &inst.StartedAt, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.scanInstance: %w", err)
	}

	return inst, nil
}
