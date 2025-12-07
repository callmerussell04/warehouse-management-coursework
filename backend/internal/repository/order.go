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

type OrderRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewOrderRepository(db *sql.DB, logger *slog.Logger) *OrderRepository {
	return &OrderRepository{db: db, logger: logger}
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("repo: failed to begin transaction", "error", err)
		return customErrors.ErrInternal
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	orderQuery := `
		INSERT INTO orders (id, counterparty_id, status, order_date, created_at, updated_at, order_type, destination) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.ExecContext(ctx, orderQuery,
		order.ID, order.CounterpartyID, order.Status, order.OrderDate,
		order.CreatedAt, order.UpdatedAt,
		order.OrderType, order.Destination)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return customErrors.ErrNotFound
		}
		r.logger.Error("repo: failed to create order", "error", err)
		return customErrors.ErrInternal
	}

	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity) 
		VALUES ($1, $2, $3)`

	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery,
			order.ID, item.ProductID, item.Quantity)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && (pqErr.Code == "23503" || pqErr.Code == "23505") {
				return customErrors.ErrNotFound
			}
			r.logger.Error("repo: failed to create order item", "error", err)
			return customErrors.ErrInternal
		}
	}

	return nil
}

func (r *OrderRepository) getItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error) {
	query := `
		SELECT product_id, quantity 
		FROM order_items 
		WHERE order_id = $1`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		r.logger.Error("repo: failed to query order items", "order_id", orderID, "error", err)
		return nil, customErrors.ErrInternal
	}
	defer rows.Close()

	items := []model.OrderItem{}
	for rows.Next() {
		item := model.OrderItem{OrderID: orderID}
		err := rows.Scan(
			&item.ProductID, &item.Quantity,
		)
		if err != nil {
			r.logger.Error("repo: failed to scan order item row", "order_id", orderID, "error", err)
			return nil, customErrors.ErrInternal
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("repo: rows iteration error for order items", "order_id", orderID, "error", err)
		return nil, customErrors.ErrInternal
	}

	return items, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	orderQuery := `
		SELECT id, counterparty_id, status, order_date, created_at, updated_at, order_type, destination 
		FROM orders 
		WHERE id = $1`

	order := &model.Order{}
	var statusStr, typeStr string
	var destination sql.NullString

	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.ID, &order.CounterpartyID, &statusStr, &order.OrderDate,
		&order.CreatedAt, &order.UpdatedAt,
		&typeStr, &destination,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, customErrors.ErrNotFound
	}
	if err != nil {
		r.logger.Error("repo: failed to get order", "id", id, "error", err)
		return nil, customErrors.ErrInternal
	}

	order.Status = model.OrderStatus(statusStr)
	order.OrderType = model.OrderType(typeStr)
	order.Destination = destination.String

	items, err := r.getItemsByOrderID(ctx, id)
	if err != nil {
		return nil, err
	}

	order.Items = items
	return order, nil
}

func (r *OrderRepository) Update(ctx context.Context, order *model.Order) error {
	orderQuery := `
		UPDATE orders 
		SET status = $1, updated_at = $2
		WHERE id = $3`

	res, err := r.db.ExecContext(ctx, orderQuery, order.Status, order.UpdatedAt, order.ID)

	if err != nil {
		r.logger.Error("repo: failed to update order status", "id", order.ID, "error", err)
		return customErrors.ErrInternal
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return customErrors.ErrNotFound
	}

	return nil
}

func (r *OrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM orders WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("repo: failed to delete order", "id", id, "error", err)
		return customErrors.ErrInternal
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return customErrors.ErrNotFound
	}
	return nil
}

func (r *OrderRepository) GetList(ctx context.Context, limit, offset int) ([]*model.Order, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, counterparty_id, status, order_date, created_at, updated_at, order_type, destination
		FROM orders 
		ORDER BY order_date DESC 
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("repo: failed to list orders", "error", err)
		return nil, 0, customErrors.ErrInternal
	}
	defer rows.Close()

	list := make([]*model.Order, 0, limit)
	for rows.Next() {
		order := &model.Order{}
		var statusStr, typeStr string
		var destination sql.NullString

		err := rows.Scan(
			&order.ID, &order.CounterpartyID, &statusStr, &order.OrderDate,
			&order.CreatedAt, &order.UpdatedAt,
			&typeStr, &destination,
		)

		if err != nil {
			r.logger.Error("repo: failed to scan order row", "error", err)
			return nil, 0, customErrors.ErrInternal
		}
		order.Status = model.OrderStatus(statusStr)
		order.OrderType = model.OrderType(typeStr)
		order.Destination = destination.String
		list = append(list, order)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("repo: rows iteration error", "error", err)
		return nil, 0, customErrors.ErrInternal
	}

	var totalCount int
	countQuery := "SELECT COUNT(id) FROM orders"
	err = r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		r.logger.Error("repo: failed to count orders", "error", err)
		return nil, 0, customErrors.ErrInternal
	}

	return list, totalCount, nil
}
