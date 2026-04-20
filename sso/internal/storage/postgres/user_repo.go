// Package postgres предоставляет PostgreSQL-реализации репозиториев SSO.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Be4Die/game-developer-hub/sso/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository — PostgreSQL-реализация domain.UserRepository.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository создаёт репозиторий пользователей.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error { //nolint:revive
	const op = "postgres.UserRepository.Create"

	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (email, password_hash, display_name, role, status, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.Email, user.PasswordHash, user.DisplayName, user.Role, user.Status, user.EmailVerified, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%s: %w", op, domain.ErrAlreadyExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) { //nolint:revive
	const op = "postgres.UserRepository.GetByID"

	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, role, status, email_verified, created_at, updated_at
		FROM users WHERE id = $1
	`, id)

	user, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) { //nolint:revive
	const op = "postgres.UserRepository.GetByEmail"

	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, role, status, email_verified, created_at, updated_at
		FROM users WHERE email = $1
	`, email)

	user, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) error { //nolint:revive
	const op = "postgres.UserRepository.Update"

	tag, err := r.pool.Exec(ctx, `
		UPDATE users SET
			email = $2,
			password_hash = $3,
			display_name = $4,
			role = $5,
			status = $6,
			email_verified = $7,
			updated_at = $8
		WHERE id = $1
	`, user.ID, user.Email, user.PasswordHash, user.DisplayName, user.Role, user.Status, user.EmailVerified, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	}

	return nil
}

func (r *UserRepository) Search(ctx context.Context, query string, limit, offset int) ([]domain.User, int64, error) { //nolint:revive
	const op = "postgres.UserRepository.Search"

	searchPattern := "%" + query + "%"

	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE email ILIKE $1 OR display_name ILIKE $1`, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: count: %w", op, err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, email, password_hash, display_name, role, status, email_verified, created_at, updated_at
		FROM users WHERE email ILIKE $1 OR display_name ILIKE $1
		ORDER BY display_name
		LIMIT $2 OFFSET $3
	`, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: query: %w", op, err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: scan: %w", op, err)
		}
		users = append(users, *u)
	}

	return users, total, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(s userScanner) (*domain.User, error) {
	u := &domain.User{}
	err := s.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role, &u.Status, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("postgres.scanUser: %w", err)
	}
	return u, nil
}

func isUniqueViolation(err error) bool {
	// PostgreSQL error code 23505 = unique_violation
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
