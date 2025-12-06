package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lib/pq"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
)

type CounterpartyRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewCounterpartyRepository(db *sql.DB, logger *slog.Logger) *CounterpartyRepository {
	return &CounterpartyRepository{db: db, logger: logger}
}

func (r *CounterpartyRepository) Create(ctx context.Context, c *model.Counterparty) error {
	query := `
		INSERT INTO counterparties (id, name, type, phone_number, email) 
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.Name, c.Type, c.PhoneNumber, c.Email,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return customErrors.ErrAlreadyExists
		}
		r.logger.Error("repository: failed to create counterparty", "error", err)
		return customErrors.ErrInternal
	}
	return nil
}

func (r *CounterpartyRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Counterparty, error) {
	query := `
		SELECT id, name, type, phone_number, email
		FROM counterparties 
		WHERE id = $1`

	c := &model.Counterparty{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Type, &c.PhoneNumber, &c.Email,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, customErrors.ErrNotFound
	}
	if err != nil {
		r.logger.Error("repository: failed to get counterparty", "id", id, "error", err)
		return nil, customErrors.ErrInternal
	}
	return c, nil
}

func (r *CounterpartyRepository) Update(ctx context.Context, c *model.Counterparty) error {
	query := `
		UPDATE counterparties 
		SET name = $1, phone_number = $2, email = $3
		WHERE id = $4`

	res, err := r.db.ExecContext(ctx, query, c.Name, c.PhoneNumber, c.Email, c.ID)

	if err != nil {
		r.logger.Error("repository: failed to update counterparty", "id", c.ID, "error", err)
		return customErrors.ErrInternal
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return customErrors.ErrNotFound
	}

	return nil
}

func (r *CounterpartyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM counterparties WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("repository: failed to delete counterparty", "id", id, "error", err)
		return customErrors.ErrInternal
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return customErrors.ErrNotFound
	}
	return nil
}

func (r *CounterpartyRepository) GetList(ctx context.Context, limit, offset int) ([]*model.Counterparty, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, name, type, phone_number, email
		FROM counterparties 
		ORDER BY name ASC 
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("repository: failed to list counterparties", "error", err)
		return nil, 0, customErrors.ErrInternal
	}
	defer rows.Close()

	list := make([]*model.Counterparty, 0, limit)
	for rows.Next() {
		c := &model.Counterparty{}
		err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.PhoneNumber, &c.Email)
		if err != nil {
			r.logger.Error("repository: failed to scan counterparty row", "error", err)
			return nil, 0, customErrors.ErrInternal
		}
		list = append(list, c)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("repository: rows iteration error", "error", err)
		return nil, 0, customErrors.ErrInternal
	}

	var totalCount int
	countQuery := "SELECT COUNT(id) FROM counterparties"
	err = r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		r.logger.Error("repository: failed to count counterparties", "error", err)
		return nil, 0, customErrors.ErrInternal
	}

	return list, totalCount, nil
}
