// Package postgres реализует хранилище персистентных данных в PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ошибки пакета postgres.
var (
	ErrInvalidDSN = errors.New("invalid DSN")
)

// NodeRepo реализует domain.NodeRepo поверх PostgreSQL.
// Безопасен для конкурентного использования.
type NodeRepo struct {
	pool *pgxpool.Pool
}

// NewNodeRepo создаёт хранилище нод. Требует инициализированный pool подключений.
func NewNodeRepo(pool *pgxpool.Pool) *NodeRepo {
	return &NodeRepo{pool: pool}
}

// Create добавляет новую ноду в реестр. Возвращает ErrAlreadyExists при дубликате.
func (r *NodeRepo) Create(ctx context.Context, node *domain.Node) error {
	const q = `
		INSERT INTO nodes (id, address, token_hash, api_token, region, status,
		                   cpu_cores, total_memory, total_disk, agent_version,
		                   last_ping_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`

	_, err := r.pool.Exec(ctx, q,
		node.ID, node.Address, node.TokenHash, node.APIToken, node.Region, node.Status,
		node.CPUCores, node.TotalMemory, node.TotalDisk, node.AgentVersion,
		node.LastPingAt, node.CreatedAt, node.UpdatedAt,
	)
	if err != nil {
		if isPgUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("postgres.NodeRepo.Create: %w", err)
	}

	return nil
}

// Update обновляет данные ноды. Возвращает ErrNotFound если нода не существует.
func (r *NodeRepo) Update(ctx context.Context, node *domain.Node) error {
	const q = `
		UPDATE nodes SET address=$1, token_hash=$2, api_token=$3, region=$4, status=$5,
		                 cpu_cores=$6, total_memory=$7, total_disk=$8,
		                 agent_version=$9, last_ping_at=$10, updated_at=$11
		WHERE id=$12
	`

	tag, err := r.pool.Exec(ctx, q,
		node.Address, node.TokenHash, node.APIToken, node.Region, node.Status,
		node.CPUCores, node.TotalMemory, node.TotalDisk,
		node.AgentVersion, node.LastPingAt, node.UpdatedAt, node.ID,
	)
	if err != nil {
		return fmt.Errorf("postgres.NodeRepo.Update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetByID возвращает ноду по идентификатору. Возвращает ErrNotFound при отсутствии.
func (r *NodeRepo) GetByID(ctx context.Context, id int64) (*domain.Node, error) {
	const q = `
		SELECT id, address, token_hash, api_token, region, status,
		       cpu_cores, total_memory, total_disk, agent_version,
		       last_ping_at, created_at, updated_at
		FROM nodes WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, q, id)
	return scanNode(row)
}

// GetByAddress возвращает ноду по gRPC-адресу. Возвращает ErrNotFound при отсутствии.
func (r *NodeRepo) GetByAddress(ctx context.Context, address string) (*domain.Node, error) {
	const q = `
		SELECT id, address, token_hash, api_token, region, status,
		       cpu_cores, total_memory, total_disk, agent_version,
		       last_ping_at, created_at, updated_at
		FROM nodes WHERE address = $1
	`

	row := r.pool.QueryRow(ctx, q, address)
	return scanNode(row)
}

// List возвращает все ноды. Опционально фильтрует по статусу (nil — без фильтра).
func (r *NodeRepo) List(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
	q := `
		SELECT id, address, token_hash, api_token, region, status,
		       cpu_cores, total_memory, total_disk, agent_version,
		       last_ping_at, created_at, updated_at
		FROM nodes
	`

	args := []any{}
	if status != nil {
		q += " WHERE status = $1"
		args = append(args, *status)
	}
	q += " ORDER BY created_at DESC"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres.NodeRepo.List: %w", err)
	}
	defer rows.Close()

	var nodes []*domain.Node
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.NodeRepo.List: %w", err)
	}

	return nodes, nil
}

// Delete удаляет ноду из реестра. Возвращает ErrNotFound при отсутствии.
func (r *NodeRepo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM nodes WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("postgres.NodeRepo.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateLastPing обновляет время последнего heartbeat ноды.
// Возвращает ErrNotFound если нода не существует.
func (r *NodeRepo) UpdateLastPing(ctx context.Context, id int64) error {
	const q = `UPDATE nodes SET last_ping_at = NOW(), updated_at = NOW() WHERE id = $1`

	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("postgres.NodeRepo.UpdateLastPing: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

type nodeScanner interface {
	Scan(dest ...any) error
}

func scanNode(s nodeScanner) (*domain.Node, error) {
	n := &domain.Node{}
	err := s.Scan(
		&n.ID, &n.Address, &n.TokenHash, &n.APIToken, &n.Region, &n.Status,
		&n.CPUCores, &n.TotalMemory, &n.TotalDisk, &n.AgentVersion,
		&n.LastPingAt, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.scanNode: %w", err)
	}

	return n, nil
}

// isPgUniqueViolation проверяет, является ли ошибка нарушением UNIQUE-ограничения PostgreSQL.
func isPgUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
