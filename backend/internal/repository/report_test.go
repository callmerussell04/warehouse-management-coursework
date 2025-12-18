package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"warehouse-management-system/internal/dto"
	"warehouse-management-system/internal/repository"
)

func TestReportRepository_GetTurnoverData(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewReportRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM inventory_transactions")
		_, _ = db.Exec("DELETE FROM products")
	}()

	prodID := uuid.New()
	_, err := db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, $2, $3, $4)",
		prodID, "REP-SKU", "Report Item", 100)
	require.NoError(t, err)

	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		from      time.Time
		to        time.Time
		prepare   func()
		wantLen   int
		checkItem func(*testing.T, dto.TurnoverReportItem)
	}{
		{
			name: "Complex Movement",
			from: baseTime,
			to:   baseTime.AddDate(0, 1, 0),
			prepare: func() {
				_, _ = db.Exec("DELETE FROM inventory_transactions")

				_, err = db.Exec(`INSERT INTO inventory_transactions (product_id, type, quantity, balance_after, created_at) 
					VALUES ($1, 'income', 10, 10, $2)`,
					prodID, baseTime.Add(-24*time.Hour))
				require.NoError(t, err)

				_, err = db.Exec(`INSERT INTO inventory_transactions (product_id, type, quantity, balance_after, created_at) 
					VALUES ($1, 'income', 5, 15, $2)`,
					prodID, baseTime.Add(10*24*time.Hour))
				require.NoError(t, err)

				_, err = db.Exec(`INSERT INTO inventory_transactions (product_id, type, quantity, balance_after, created_at) 
					VALUES ($1, 'expense', 3, 12, $2)`,
					prodID, baseTime.Add(20*24*time.Hour))
				require.NoError(t, err)

				_, err = db.Exec(`INSERT INTO inventory_transactions (product_id, type, quantity, balance_after, created_at) 
					VALUES ($1, 'income', 100, 112, $2)`,
					prodID, baseTime.AddDate(0, 1, 5))
				require.NoError(t, err)
			},
			wantLen: 1,
			checkItem: func(t *testing.T, item dto.TurnoverReportItem) {
				assert.Equal(t, "Report Item", item.ProductName)
				assert.Equal(t, "REP-SKU", item.SKU)
				assert.Equal(t, 10, item.StartBalance)
				assert.Equal(t, 5, item.Income)
				assert.Equal(t, 3, item.Expense)
				assert.Equal(t, 12, item.EndBalance)
			},
		},
		{
			name: "Only Start Balance (No movement in period)",
			from: baseTime,
			to:   baseTime.AddDate(0, 1, 0),
			prepare: func() {
				_, _ = db.Exec("DELETE FROM inventory_transactions")
				_, err = db.Exec(`INSERT INTO inventory_transactions (product_id, type, quantity, balance_after, created_at) 
					VALUES ($1, 'income', 50, 50, $2)`,
					prodID, baseTime.Add(-24*time.Hour))
				require.NoError(t, err)
			},
			wantLen: 1,
			checkItem: func(t *testing.T, item dto.TurnoverReportItem) {
				assert.Equal(t, 50, item.StartBalance)
				assert.Equal(t, 0, item.Income)
				assert.Equal(t, 0, item.Expense)
				assert.Equal(t, 50, item.EndBalance)
			},
		},
		{
			name: "No Data Before or Inside (Empty Report)",
			from: baseTime,
			to:   baseTime.AddDate(0, 1, 0),
			prepare: func() {
				_, _ = db.Exec("DELETE FROM inventory_transactions")
				_, err = db.Exec(`INSERT INTO inventory_transactions (product_id, type, quantity, balance_after, created_at) 
					VALUES ($1, 'income', 50, 50, $2)`,
					prodID, baseTime.AddDate(0, 2, 0))
				require.NoError(t, err)
			},
			wantLen:   0,
			checkItem: nil,
		},
		{
			name: "Only Movement Inside (Start Balance 0)",
			from: baseTime,
			to:   baseTime.AddDate(0, 1, 0),
			prepare: func() {
				_, _ = db.Exec("DELETE FROM inventory_transactions")
				_, err = db.Exec(`INSERT INTO inventory_transactions (product_id, type, quantity, balance_after, created_at) 
					VALUES ($1, 'income', 20, 20, $2)`,
					prodID, baseTime.Add(1*time.Hour))
				require.NoError(t, err)
			},
			wantLen: 1,
			checkItem: func(t *testing.T, item dto.TurnoverReportItem) {
				assert.Equal(t, 0, item.StartBalance)
				assert.Equal(t, 20, item.Income)
				assert.Equal(t, 0, item.Expense)
				assert.Equal(t, 20, item.EndBalance)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare()
			}

			list, err := repo.GetTurnoverData(context.Background(), tc.from, tc.to)
			assert.NoError(t, err)
			assert.Len(t, list, tc.wantLen)

			if tc.wantLen > 0 && tc.checkItem != nil {
				tc.checkItem(t, list[0])
			}
		})
	}
}
