package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"warehouse-management-system/internal/dto"
	customErrors "warehouse-management-system/internal/errors"
)

type ReportRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewReportRepository(db *sql.DB, logger *slog.Logger) *ReportRepository {
	return &ReportRepository{db: db, logger: logger}
}

func (r *ReportRepository) GetTurnoverData(ctx context.Context, from, to time.Time) ([]dto.TurnoverReportItem, error) {
	query := `
		SELECT
			p.name, 
			p.sku,
			COALESCE(start_bal.balance_after, 0) as start_balance,
			COALESCE(moves.inc, 0) as income,
			COALESCE(moves.exp, 0) as expense
		FROM products p
		LEFT JOIN LATERAL (
			SELECT balance_after 
			FROM inventory_transactions it
			WHERE it.product_id = p.id AND it.created_at < $1
			ORDER BY it.created_at DESC 
			LIMIT 1
		) start_bal ON true
		LEFT JOIN (
			SELECT 
				product_id,
				SUM(CASE WHEN type = 'income' THEN quantity ELSE 0 END) as inc,
				SUM(CASE WHEN type = 'expense' THEN quantity ELSE 0 END) as exp
			FROM inventory_transactions
			WHERE created_at >= $1 AND created_at < $2
			GROUP BY product_id
		) moves ON moves.product_id = p.id
		ORDER BY p.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		r.logger.Error("failed to query turnover report", "error", err)
		return nil, customErrors.ErrInternal
	}
	defer rows.Close()

	var result []dto.TurnoverReportItem

	for rows.Next() {
		var item dto.TurnoverReportItem
		if err := rows.Scan(&item.ProductName, &item.SKU, &item.StartBalance, &item.Income, &item.Expense); err != nil {
			r.logger.Error("failed to scan report row", "error", err)
			return nil, customErrors.ErrInternal
		}

		item.EndBalance = item.StartBalance + item.Income - item.Expense

		if item.StartBalance != 0 || item.Income != 0 || item.Expense != 0 {
			result = append(result, item)
		}
	}

	return result, nil
}
