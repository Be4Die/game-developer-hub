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
// После создания поле node.ID заполняется сгенерированным BIGSERIAL идентификатором.
func (r *NodeRepo) Create(ctx context.Context, node *domain.Node) error {
	role := node.Role
	if role == 0 {
		role = domain.NodeRoleMixed
	}
	ingressMode := node.IngressMode
	if ingressMode == 0 {
		ingressMode = domain.IngressModePlatformProxy
	}

	const q = `
		INSERT INTO nodes (owner_id, address, token_hash, api_token, region, status, role,
		                   cpu_cores, total_memory, total_disk, agent_version,
		                   last_ping_at, created_at, updated_at, backups_enabled,
		                   ingress_mode, custom_domain)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING id
	`

	err := r.pool.QueryRow(ctx, q,
		node.OwnerID, node.Address, node.TokenHash, node.APIToken, node.Region, node.Status, role,
		node.CPUCores, node.TotalMemory, node.TotalDisk, node.AgentVersion,
		node.LastPingAt, node.CreatedAt, node.UpdatedAt, node.BackupsEnabled,
		ingressMode, node.CustomDomain,
	).Scan(&node.ID)
	if err != nil {
		if isPgUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("postgres.NodeRepo.Create: %w", err)
	}

	node.Role = role
	node.IngressMode = ingressMode
	return nil
}

// Update обновляет данные ноды. Возвращает ErrNotFound если нода не существует.
func (r *NodeRepo) Update(ctx context.Context, node *domain.Node) error {
	role := node.Role
	if role == 0 {
		role = domain.NodeRoleMixed
	}
	ingressMode := node.IngressMode
	if ingressMode == 0 {
		ingressMode = domain.IngressModePlatformProxy
	}

	const q = `
		UPDATE nodes SET owner_id=$1, address=$2, token_hash=$3, api_token=$4, region=$5, status=$6, role=$7,
		                 cpu_cores=$8, total_memory=$9, total_disk=$10,
		                 agent_version=$11, last_ping_at=$12, updated_at=$13, backups_enabled=$14,
		                 ingress_mode=$15, custom_domain=$16
		WHERE id=$17
	`

	tag, err := r.pool.Exec(ctx, q,
		node.OwnerID, node.Address, node.TokenHash, node.APIToken, node.Region, node.Status, role,
		node.CPUCores, node.TotalMemory, node.TotalDisk,
		node.AgentVersion, node.LastPingAt, node.UpdatedAt, node.BackupsEnabled,
		ingressMode, node.CustomDomain, node.ID,
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
		SELECT id, owner_id, address, token_hash, api_token, region, status, role,
		       cpu_cores, total_memory, total_disk, agent_version,
		       last_ping_at, created_at, updated_at, backups_enabled,
		       ingress_mode, custom_domain
		FROM nodes WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, q, id)
	return scanNode(row)
}

// GetByAddress возвращает ноду по gRPC-адресу. Возвращает ErrNotFound при отсутствии.
func (r *NodeRepo) GetByAddress(ctx context.Context, address string) (*domain.Node, error) {
	const q = `
		SELECT id, owner_id, address, token_hash, api_token, region, status, role,
		       cpu_cores, total_memory, total_disk, agent_version,
		       last_ping_at, created_at, updated_at, backups_enabled,
		       ingress_mode, custom_domain
		FROM nodes WHERE address = $1
	`

	row := r.pool.QueryRow(ctx, q, address)
	return scanNode(row)
}

// List возвращает все ноды. Опционально фильтрует по статусу (nil — без фильтра).
func (r *NodeRepo) List(ctx context.Context, status *domain.NodeStatus) ([]*domain.Node, error) {
	q := `
		SELECT id, owner_id, address, token_hash, api_token, region, status, role,
		       cpu_cores, total_memory, total_disk, agent_version,
		       last_ping_at, created_at, updated_at, backups_enabled,
		       ingress_mode, custom_domain
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

// UpdateRole обновляет роль вычислительной ноды (mixed, compute, storage).
func (r *NodeRepo) UpdateRole(ctx context.Context, id int64, role domain.NodeRole) error {
	const q = `UPDATE nodes SET role = $1, updated_at = NOW() WHERE id = $2`

	tag, err := r.pool.Exec(ctx, q, role, id)
	if err != nil {
		return fmt.Errorf("postgres.NodeRepo.UpdateRole: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// UpdateIngress обновляет сетевой режим ноды (platform_proxy / direct) и кастомный домен.
func (r *NodeRepo) UpdateIngress(ctx context.Context, id int64, mode domain.IngressMode, customDomain string) error {
	const q = `UPDATE nodes SET ingress_mode = $1, custom_domain = $2, updated_at = NOW() WHERE id = $3`

	tag, err := r.pool.Exec(ctx, q, mode, customDomain, id)
	if err != nil {
		return fmt.Errorf("postgres.NodeRepo.UpdateIngress: %w", err)
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
		&n.ID, &n.OwnerID, &n.Address, &n.TokenHash, &n.APIToken, &n.Region, &n.Status, &n.Role,
		&n.CPUCores, &n.TotalMemory, &n.TotalDisk, &n.AgentVersion,
		&n.LastPingAt, &n.CreatedAt, &n.UpdatedAt, &n.BackupsEnabled,
		&n.IngressMode, &n.CustomDomain,
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
