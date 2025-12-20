package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

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
			switch pqErr.Code {
			case "23505":
				return customErrors.ErrAlreadyExists
			case "23514":
				return customErrors.ErrInvalidInput
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
			switch pqErr.Code {
			case "23505":
				return customErrors.ErrAlreadyExists
			case "23514":
				return customErrors.ErrInvalidInput
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
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return customErrors.NewAppError(customErrors.ErrConflict, "product is referenced by inventory history or an order")
		}
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

func (r *ProductRepository) UpdateStock(ctx context.Context, productID uuid.UUID, amount int64, transType model.TransactionType) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return customErrors.ErrInternal
	}
	defer tx.Rollback()

	var currentQty int64
	err = tx.QueryRowContext(ctx, "SELECT quantity FROM products WHERE id = $1 FOR UPDATE", productID).Scan(&currentQty)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customErrors.ErrNotFound
		}
		return customErrors.ErrInternal
	}

	newQty := currentQty
	switch transType {
	case model.TransactionIncome:
		if currentQty > (1<<63-1)-amount {
			return customErrors.NewAppError(customErrors.ErrInvalidInput, "stock quantity overflow")
		}
		newQty += amount
	case model.TransactionExpense:
		if currentQty < amount {
			return customErrors.ErrInsufficientStock
		}
		newQty -= amount
	default:
		return customErrors.NewAppError(customErrors.ErrInvalidInput, "invalid transaction type")
	}

	_, err = tx.ExecContext(ctx, "UPDATE products SET quantity = $1, updated_at = NOW() WHERE id = $2", newQty, productID)
	if err != nil {
		return customErrors.ErrInternal
	}

	historyQuery := `
		INSERT INTO inventory_transactions (id, product_id, type, quantity, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`
	_, err = tx.ExecContext(ctx, historyQuery, uuid.New(), productID, transType, amount, newQty)
	if err != nil {
		r.logger.Error("failed to insert inventory transaction", "error", err)
		return customErrors.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		return customErrors.ErrInternal
	}

	r.logger.Info("stock updated", "product_id", productID, "type", transType, "new_balance", newQty)
	return nil
}

func (r *ProductRepository) GetProductHistory(ctx context.Context, productID uuid.UUID, limit, offset int, from, to time.Time) ([]*model.InventoryTransaction, int, error) {
	baseQuery := `
		SELECT it.id, it.product_id, p.name, it.type, it.quantity, it.balance_after, it.created_at
		FROM inventory_transactions it
		JOIN products p ON it.product_id = p.id
		WHERE it.product_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM inventory_transactions it WHERE it.product_id = $1`

	args := []interface{}{productID}
	argID := 2

	if !from.IsZero() {
		filter := fmt.Sprintf(" AND it.created_at >= $%d", argID)
		baseQuery += filter
		countQuery += filter
		args = append(args, from)
		argID++
	}
	if !to.IsZero() {
		filter := fmt.Sprintf(" AND it.created_at < $%d", argID)
		baseQuery += filter
		countQuery += filter
		args = append(args, to)
		argID++
	}

	baseQuery += fmt.Sprintf(" ORDER BY it.created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)

	queryArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, baseQuery, queryArgs...)
	if err != nil {
		r.logger.Error("failed to list product history", "error", err)
		return nil, 0, customErrors.ErrInternal
	}
	defer rows.Close()

	var list []*model.InventoryTransaction
	for rows.Next() {
		t := &model.InventoryTransaction{}
		var typeStr string
		if err := rows.Scan(&t.ID, &t.ProductID, &t.ProductName, &typeStr, &t.Quantity, &t.BalanceAfter, &t.CreatedAt); err != nil {
			r.logger.Error("scan error", "error", err)
			return nil, 0, customErrors.ErrInternal
		}
		t.Type = model.TransactionType(typeStr)
		list = append(list, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, customErrors.ErrInternal
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, customErrors.ErrInternal
	}

	return list, total, nil
}
