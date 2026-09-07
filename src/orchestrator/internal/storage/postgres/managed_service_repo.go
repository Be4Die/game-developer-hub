package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/orchestrator/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ManagedServiceRepo реализует domain.ManagedServiceRepo поверх PostgreSQL.
type ManagedServiceRepo struct {
	pool *pgxpool.Pool
}

// NewManagedServiceRepo создает репозиторий управляемых сервисов.
func NewManagedServiceRepo(pool *pgxpool.Pool) *ManagedServiceRepo {
	return &ManagedServiceRepo{pool: pool}
}

// Create добавляет новую запись об управляемом сервисе.
func (r *ManagedServiceRepo) Create(ctx context.Context, s *domain.ManagedService) error {
	const q = `
		INSERT INTO node_services (node_id, owner_id, allowed_game_ids, service_type, name,
		                           container_id, host_port, connection_uri, credentials, status,
		                           volume_path, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	credsJSON, err := json.Marshal(s.Credentials)
	if err != nil {
		credsJSON = []byte("{}")
	}

	allowedGames := s.AllowedGameIDs
	if allowedGames == nil {
		allowedGames = []int64{}
	}

	err = r.pool.QueryRow(ctx, q,
		s.NodeID, s.OwnerID, allowedGames, s.ServiceType, s.Name,
		s.ContainerID, s.HostPort, s.ConnectionURI, credsJSON, s.Status,
		s.VolumePath, s.CreatedAt, s.UpdatedAt,
	).Scan(&s.ID)
	if err != nil {
		if isPgUniqueViolation(err) {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("postgres.ManagedServiceRepo.Create: %w", err)
	}

	return nil
}

// GetByID возвращает сервис по идентификатору.
func (r *ManagedServiceRepo) GetByID(ctx context.Context, id int64) (*domain.ManagedService, error) {
	const q = `
		SELECT id, node_id, owner_id, allowed_game_ids, service_type, name,
		       container_id, host_port, connection_uri, credentials, status,
		       volume_path, created_at, updated_at
		FROM node_services WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, q, id)
	return scanManagedService(row)
}

// GetByName возвращает сервис по имени на ноде.
func (r *ManagedServiceRepo) GetByName(ctx context.Context, nodeID int64, name string) (*domain.ManagedService, error) {
	const q = `
		SELECT id, node_id, owner_id, allowed_game_ids, service_type, name,
		       container_id, host_port, connection_uri, credentials, status,
		       volume_path, created_at, updated_at
		FROM node_services WHERE node_id = $1 AND name = $2
	`
	row := r.pool.QueryRow(ctx, q, nodeID, name)
	return scanManagedService(row)
}

// ListByNode возвращает все сервисы на ноде.
func (r *ManagedServiceRepo) ListByNode(ctx context.Context, nodeID int64) ([]*domain.ManagedService, error) {
	const q = `
		SELECT id, node_id, owner_id, allowed_game_ids, service_type, name,
		       container_id, host_port, connection_uri, credentials, status,
		       volume_path, created_at, updated_at
		FROM node_services WHERE node_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, nodeID)
	if err != nil {
		return nil, fmt.Errorf("postgres.ManagedServiceRepo.ListByNode: %w", err)
	}
	defer rows.Close()

	return scanManagedServices(rows)
}

// ListByOwner возвращает все сервисы владельца.
func (r *ManagedServiceRepo) ListByOwner(ctx context.Context, ownerID string) ([]*domain.ManagedService, error) {
	const q = `
		SELECT id, node_id, owner_id, allowed_game_ids, service_type, name,
		       container_id, host_port, connection_uri, credentials, status,
		       volume_path, created_at, updated_at
		FROM node_services WHERE owner_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, ownerID)
	if err != nil {
		return nil, fmt.Errorf("postgres.ManagedServiceRepo.ListByOwner: %w", err)
	}
	defer rows.Close()

	return scanManagedServices(rows)
}

// ListByGame возвращает сервисы, к которым разрешен доступ данной игре (или всем играм, если allowed_game_ids пуст).
func (r *ManagedServiceRepo) ListByGame(ctx context.Context, gameID int64) ([]*domain.ManagedService, error) {
	const q = `
		SELECT id, node_id, owner_id, allowed_game_ids, service_type, name,
		       container_id, host_port, connection_uri, credentials, status,
		       volume_path, created_at, updated_at
		FROM node_services
		WHERE cardinality(allowed_game_ids) = 0 OR $1 = ANY(allowed_game_ids)
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, gameID)
	if err != nil {
		return nil, fmt.Errorf("postgres.ManagedServiceRepo.ListByGame: %w", err)
	}
	defer rows.Close()

	return scanManagedServices(rows)
}

// Update обновляет метаданные сервиса.
func (r *ManagedServiceRepo) Update(ctx context.Context, s *domain.ManagedService) error {
	const q = `
		UPDATE node_services
		SET allowed_game_ids = $1, container_id = $2, host_port = $3,
		    connection_uri = $4, status = $5, volume_path = $6, updated_at = NOW()
		WHERE id = $7
	`
	tag, err := r.pool.Exec(ctx, q,
		s.AllowedGameIDs, s.ContainerID, s.HostPort, s.ConnectionURI, s.Status, s.VolumePath, s.ID,
	)
	if err != nil {
		return fmt.Errorf("postgres.ManagedServiceRepo.Update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete удаляет запись сервиса.
func (r *ManagedServiceRepo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM node_services WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("postgres.ManagedServiceRepo.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanManagedService(row pgx.Row) (*domain.ManagedService, error) {
	s := &domain.ManagedService{}
	var credsRaw []byte

	err := row.Scan(
		&s.ID, &s.NodeID, &s.OwnerID, &s.AllowedGameIDs, &s.ServiceType, &s.Name,
		&s.ContainerID, &s.HostPort, &s.ConnectionURI, &credsRaw, &s.Status,
		&s.VolumePath, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scanManagedService: %w", err)
	}

	if len(credsRaw) > 0 {
		_ = json.Unmarshal(credsRaw, &s.Credentials)
	}
	return s, nil
}

func scanManagedServices(rows pgx.Rows) ([]*domain.ManagedService, error) {
	var list []*domain.ManagedService
	for rows.Next() {
		s, err := scanManagedService(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scanManagedServices: %w", err)
	}
	return list, nil
}

var _ domain.ManagedServiceRepo = (*ManagedServiceRepo)(nil)
