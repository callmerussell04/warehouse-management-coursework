package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepository(db *sql.DB, logger *slog.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (r *UserRepository) CreateUser(ctx context.Context, u *model.User) error {
	query := `
		INSERT INTO users (id, username, email, full_name, role, is_active, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		u.ID, u.Username, u.Email, u.FullName, u.Role, u.IsActive, u.PasswordHash,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return customErrors.NewAppError(customErrors.ErrAlreadyExists, "Email or Username already exists")
		}
		r.logger.Error("repo: failed to create user", "error", err)
		return customErrors.ErrInternal
	}
	return nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT id, username, email, password_hash, full_name, role, is_active FROM users WHERE username = $1`
	return r.scanUser(ctx, query, username)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, username, email, password_hash, full_name, role, is_active FROM users WHERE email = $1`
	return r.scanUser(ctx, query, email)
}

func (r *UserRepository) GetList(ctx context.Context, limit, offset int) ([]*model.User, int, error) {
	var totalCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalCount)
	if err != nil {
		r.logger.Error("repo: failed to count users", "error", err)
		return nil, 0, customErrors.ErrInternal
	}

	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active
		FROM users 
		ORDER BY full_name DESC 
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("repo: failed to get user list", "error", err)
		return nil, 0, customErrors.ErrInternal
	}
	defer rows.Close()

	users := make([]*model.User, 0, limit)
	for rows.Next() {
		u := &model.User{}
		var roleStr string
		err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName,
			&roleStr, &u.IsActive,
		)
		if err != nil {
			r.logger.Error("repo: failed to scan user row", "error", err)
			return nil, 0, customErrors.ErrInternal
		}
		u.Role = model.Role(roleStr)
		users = append(users, u)
	}
	if rows.Err() != nil {
		r.logger.Error("repo: rows error in user list", "error", rows.Err())
		return nil, 0, customErrors.ErrInternal
	}

	return users, totalCount, nil
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	query := `
		UPDATE users 
		SET full_name = $1, email = $2, role = $3
		WHERE id = $4`

	res, err := r.db.ExecContext(ctx, query, u.FullName, u.Email, u.Role, u.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return customErrors.NewAppError(customErrors.ErrAlreadyExists, "Email or Username already exists")
		}
		r.logger.Error("repo: failed to update user", "error", err, "id", u.ID)
		return customErrors.ErrInternal
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return customErrors.ErrInternal
	}
	if rows == 0 {
		return customErrors.ErrNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		r.logger.Error("repo: failed to delete user", "error", err, "id", id)
		return customErrors.ErrInternal
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return customErrors.ErrInternal
	}
	if rows == 0 {
		return customErrors.ErrNotFound
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active
		FROM users 
		WHERE id = $1`

	u := &model.User{}
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName,
		&roleStr, &u.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, customErrors.ErrNotFound
	}
	if err != nil {
		r.logger.Error("repo: failed to get user by ID", "error", err)
		return nil, customErrors.ErrInternal
	}
	u.Role = model.Role(roleStr)
	return u, nil
}

func (r *UserRepository) UpdatePasswordAndActivate(ctx context.Context, email, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, is_active = TRUE WHERE email = $2`

	res, err := r.db.ExecContext(ctx, query, passwordHash, email)
	if err != nil {
		r.logger.Error("repo: failed to update user password", "error", err)
		return customErrors.ErrInternal
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return customErrors.ErrInternal
	}
	if rows == 0 {
		return customErrors.ErrNotFound
	}

	return nil
}

func (r *UserRepository) scanUser(ctx context.Context, query string, arg interface{}) (*model.User, error) {
	u := &model.User{}
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &roleStr, &u.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, customErrors.ErrNotFound
	}
	if err != nil {
		r.logger.Error("repo: failed to get user", "error", err)
		return nil, customErrors.ErrInternal
	}
	u.Role = model.Role(roleStr)
	return u, nil
}
