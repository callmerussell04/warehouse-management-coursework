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

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("repo: failed to begin transaction", "error", err)
		return customErrors.ErrInternal
	}
	defer tx.Rollback()

	orderQuery := `
		INSERT INTO orders (id, counterparty_id, status, order_date, created_at, updated_at, order_type, destination) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.ExecContext(ctx, orderQuery,
		order.ID, order.CounterpartyID, order.Status, order.OrderDate,
		order.CreatedAt, order.UpdatedAt,
		order.OrderType, order.Destination)

	if err != nil {
		if mapped := mapOrderWriteError(err); mapped != nil {
			return mapped
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
			if mapped := mapOrderWriteError(err); mapped != nil {
				return mapped
			}
			r.logger.Error("repo: failed to create order item", "error", err)
			return customErrors.ErrInternal
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("repo: failed to commit order", "error", err)
		return customErrors.ErrInternal
	}
	return nil
}

type queryContexter interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (r *OrderRepository) getItemsByOrderID(ctx context.Context, queryer queryContexter, orderID uuid.UUID) ([]model.OrderItem, error) {
	query := `
		SELECT product_id, quantity 
		FROM order_items 
		WHERE order_id = $1
		ORDER BY product_id`

	rows, err := queryer.QueryContext(ctx, query, orderID)
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

	items, err := r.getItemsByOrderID(ctx, r.db, id)
	if err != nil {
		return nil, err
	}

	order.Items = items
	return order, nil
}

func (r *OrderRepository) Transition(ctx context.Context, id uuid.UUID, targetStatus model.OrderStatus) (*model.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, customErrors.ErrInternal
	}
	defer tx.Rollback()

	order := &model.Order{}
	var statusStr, typeStr string
	var destination sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT id, counterparty_id, status, order_date, created_at, updated_at, order_type, destination
		FROM orders WHERE id = $1 FOR UPDATE`, id).Scan(
		&order.ID, &order.CounterpartyID, &statusStr, &order.OrderDate,
		&order.CreatedAt, &order.UpdatedAt, &typeStr, &destination,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, customErrors.ErrNotFound
	}
	if err != nil {
		r.logger.Error("repo: failed to lock order", "id", id, "error", err)
		return nil, customErrors.ErrInternal
	}
	order.Status = model.OrderStatus(statusStr)
	order.OrderType = model.OrderType(typeStr)
	order.Destination = destination.String

	if !isAllowedOrderTransition(order.Status, targetStatus) {
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "invalid order status transition")
	}

	items, err := r.getItemsByOrderID(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items

	if targetStatus == model.StatusCompleted {
		transactionType := model.TransactionIncome
		if order.OrderType == model.OrderOutbound {
			transactionType = model.TransactionExpense
		} else if order.OrderType != model.OrderInbound {
			return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "invalid order type")
		}

		for _, item := range items {
			var currentQuantity int64
			err := tx.QueryRowContext(ctx, "SELECT quantity FROM products WHERE id = $1 FOR UPDATE", item.ProductID).Scan(&currentQuantity)
			if errors.Is(err, sql.ErrNoRows) {
				return nil, customErrors.ErrNotFound
			}
			if err != nil {
				return nil, customErrors.ErrInternal
			}

			newQuantity := currentQuantity
			if transactionType == model.TransactionIncome {
				if currentQuantity > (1<<63-1)-item.Quantity {
					return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "stock quantity overflow")
				}
				newQuantity += item.Quantity
			} else {
				if currentQuantity < item.Quantity {
					return nil, customErrors.ErrInsufficientStock
				}
				newQuantity -= item.Quantity
			}

			if _, err := tx.ExecContext(ctx, "UPDATE products SET quantity = $1, updated_at = NOW() WHERE id = $2", newQuantity, item.ProductID); err != nil {
				return nil, customErrors.ErrInternal
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO inventory_transactions (id, product_id, type, quantity, balance_after, created_at)
				VALUES ($1, $2, $3, $4, $5, NOW())`, uuid.New(), item.ProductID, transactionType, item.Quantity, newQuantity); err != nil {
				return nil, customErrors.ErrInternal
			}
		}
	}

	order.Status = targetStatus
	if err := tx.QueryRowContext(ctx, `
		UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2
		RETURNING updated_at`, targetStatus, id).Scan(&order.UpdatedAt); err != nil {
		return nil, customErrors.ErrInternal
	}
	if err := tx.Commit(); err != nil {
		return nil, customErrors.ErrInternal
	}
	return order, nil
}

func isAllowedOrderTransition(current, target model.OrderStatus) bool {
	switch current {
	case model.StatusPending:
		return target == model.StatusProcessing || target == model.StatusCompleted || target == model.StatusCanceled
	case model.StatusProcessing:
		return target == model.StatusCompleted || target == model.StatusCanceled
	default:
		return false
	}
}

func mapOrderWriteError(err error) error {
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return nil
	}
	switch pqErr.Code {
	case "23503":
		return customErrors.ErrNotFound
	case "23505", "23514":
		return customErrors.ErrInvalidInput
	default:
		return nil
	}
}

func (r *OrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM orders WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("repo: failed to delete order", "id", id, "error", err)
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
