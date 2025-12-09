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

type ProductRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewProductRepository(db *sql.DB, logger *slog.Logger) *ProductRepository {
	return &ProductRepository{
		db:     db,
		logger: logger,
	}
}

func (r *ProductRepository) Create(ctx context.Context, p *model.Product) error {
	query := `
		INSERT INTO products (id, sku, name, quantity, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		p.ID, p.SKU, p.Name, p.Quantity, p.CreatedAt, p.UpdatedAt,
	).Scan(&p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return customErrors.ErrAlreadyExists
			}
		}
		r.logger.Error("repository: failed to create product", "error", err)
		return customErrors.ErrInternal
	}

	return nil
}

func (r *ProductRepository) GetList(ctx context.Context, limit, offset int) ([]*model.Product, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, sku, name, quantity, created_at, updated_at 
		FROM products 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("repository: failed to list products", "error", err)
		return nil, 0, customErrors.ErrInternal
	}
	defer rows.Close()

	products := make([]*model.Product, 0, limit)
	for rows.Next() {
		p := &model.Product{}
		err := rows.Scan(
			&p.ID, &p.SKU, &p.Name, &p.Quantity, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("repository: failed to scan product row", "error", err)
			return nil, 0, customErrors.ErrInternal
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("repository: rows iteration error", "error", err)
		return nil, 0, customErrors.ErrInternal
	}

	var totalCount int
	countQuery := "SELECT COUNT(id) FROM products"
	err = r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		r.logger.Error("repository: failed to count products", "error", err)
		return nil, 0, customErrors.ErrInternal
	}

	return products, totalCount, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	query := `
		SELECT id, sku, name, quantity, created_at, updated_at 
		FROM products 
		WHERE id = $1`

	p := &model.Product{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Quantity, &p.CreatedAt, &p.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, customErrors.ErrNotFound
	}
	if err != nil {
		r.logger.Error("repository: failed to get product", "id", id, "error", err)
		return nil, customErrors.ErrInternal
	}

	return p, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *model.Product) error {
	query := `
		UPDATE products 
		SET sku = $1, name = $2, updated_at = NOW() 
		WHERE id = $3 
		RETURNING updated_at, quantity, created_at`

	err := r.db.QueryRowContext(ctx, query, p.SKU, p.Name, p.ID).
		Scan(&p.UpdatedAt, &p.Quantity, &p.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return customErrors.ErrNotFound
	}
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return customErrors.ErrAlreadyExists
			}
		}
		r.logger.Error("repository: failed to update product", "id", p.ID, "error", err)
		return customErrors.ErrInternal
	}

	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("repository: failed to delete product", "id", id, "error", err)
		return customErrors.ErrInternal
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return customErrors.ErrInternal
	}

	if rowsAffected == 0 {
		return customErrors.ErrNotFound
	}

	return nil
}

func (r *ProductRepository) UpdateQuantity(ctx context.Context, id uuid.UUID, delta int) (int, error) {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, customErrors.ErrInternal
	}
	defer tx.Rollback()

	var currentQty int
	err = tx.QueryRowContext(ctx, "SELECT quantity FROM products WHERE id = $1 FOR UPDATE", id).Scan(&currentQty)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, customErrors.ErrNotFound
	}
	if err != nil {
		return 0, customErrors.ErrInternal
	}

	newQty := currentQty + delta
	if newQty < 0 {
		return currentQty, customErrors.ErrInsufficientStock
	}

	_, err = tx.ExecContext(ctx, "UPDATE products SET quantity = $1, updated_at = NOW() WHERE id = $2", newQty, id)
	if err != nil {
		return 0, customErrors.ErrInternal
	}

	if err = tx.Commit(); err != nil {
		return 0, customErrors.ErrInternal
	}

	return newQty, nil
}

func (r *ProductRepository) AddTransaction(ctx context.Context, t *model.InventoryTransaction) error {
	query := `
		INSERT INTO inventory_transactions (id, product_id, type, quantity, created_at) 
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query, t.ID, t.ProductID, t.Type, t.Quantity, t.CreatedAt)
	if err != nil {
		r.logger.Error("repository: failed to add transaction", "error", err)
		return customErrors.ErrInternal
	}
	return nil
}
