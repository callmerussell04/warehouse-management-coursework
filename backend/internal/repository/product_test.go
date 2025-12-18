package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
	"warehouse-management-system/internal/repository"
)

const testDBDSN = "postgres://postgres:postgres@localhost:5431/test_db?sslmode=disable"

func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", testDBDSN)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM products")
	require.NoError(t, err)

	return db
}

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestProductRepository_Create(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewProductRepository(db, newDiscardLogger())

	type args struct {
		product *model.Product
	}

	tests := []struct {
		name      string
		args      args
		prepare   func()
		wantError error
		checkRes  func(*testing.T, *model.Product)
	}{
		{
			name: "Success",
			args: args{
				product: &model.Product{
					ID:        uuid.New(),
					SKU:       "SKU-100",
					Name:      "Test Product",
					Quantity:  10,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			wantError: nil,
			checkRes: func(t *testing.T, p *model.Product) {
				assert.NotZero(t, p.CreatedAt)
				assert.NotZero(t, p.UpdatedAt)

				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM products WHERE id = $1", p.ID).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "Duplicate SKU",
			args: args{
				product: &model.Product{
					ID:        uuid.New(),
					SKU:       "DUPLICATE-SKU",
					Name:      "New Name",
					Quantity:  0,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			prepare: func() {
				_, err := db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, $2, $3, $4)",
					uuid.New(), "DUPLICATE-SKU", "Old Name", 5)
				require.NoError(t, err)
			},
			wantError: customErrors.ErrAlreadyExists,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare()
			}

			err := repo.Create(context.Background(), tc.args.product)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
				if tc.checkRes != nil {
					tc.checkRes(t, tc.args.product)
				}
			}
		})
	}
}

func TestProductRepository_GetByID(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewProductRepository(db, newDiscardLogger())

	existingID := uuid.New()
	_, err := db.Exec("INSERT INTO products (id, sku, name, quantity, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())",
		existingID, "FIND-ME", "Find Me", 50)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uuid.UUID
		wantError error
		wantName  string
	}{
		{
			name:      "Success",
			id:        existingID,
			wantError: nil,
			wantName:  "Find Me",
		},
		{
			name:      "Not Found",
			id:        uuid.New(),
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetByID(context.Background(), tc.id)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.wantName, got.Name)
				assert.Equal(t, tc.id, got.ID)
			}
		})
	}
}

func TestProductRepository_Update(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewProductRepository(db, newDiscardLogger())

	id := uuid.New()
	_, err := db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, $2, $3, $4)",
		id, "OLD-SKU", "Old Name", 10)
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, $2, $3, $4)",
		uuid.New(), "BUSY-SKU", "Busy", 5)
	require.NoError(t, err)

	tests := []struct {
		name      string
		product   *model.Product
		wantError error
		checkRes  func(*testing.T)
	}{
		{
			name: "Success",
			product: &model.Product{
				ID:   id,
				SKU:  "NEW-SKU",
				Name: "New Name",
			},
			wantError: nil,
			checkRes: func(t *testing.T) {
				var sku, name string
				err := db.QueryRow("SELECT sku, name FROM products WHERE id = $1", id).Scan(&sku, &name)
				assert.NoError(t, err)
				assert.Equal(t, "NEW-SKU", sku)
				assert.Equal(t, "New Name", name)
			},
		},
		{
			name: "Not Found",
			product: &model.Product{
				ID:   uuid.New(),
				SKU:  "ANY",
				Name: "ANY",
			},
			wantError: customErrors.ErrNotFound,
		},
		{
			name: "Conflict SKU",
			product: &model.Product{
				ID:   id,
				SKU:  "BUSY-SKU",
				Name: "Try Update",
			},
			wantError: customErrors.ErrAlreadyExists,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Update(context.Background(), tc.product)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
				if tc.checkRes != nil {
					tc.checkRes(t)
				}
			}
		})
	}
}

func TestProductRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewProductRepository(db, newDiscardLogger())

	id := uuid.New()
	_, err := db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, $2, $3, $4)",
		id, "DEL-SKU", "Delete Me", 10)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uuid.UUID
		wantError error
	}{
		{
			name:      "Success",
			id:        id,
			wantError: nil,
		},
		{
			name:      "Not Found",
			id:        uuid.New(),
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Delete(context.Background(), tc.id)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
				var count int
				_ = db.QueryRow("SELECT COUNT(*) FROM products WHERE id = $1", tc.id).Scan(&count)
				assert.Equal(t, 0, count)
			}
		})
	}
}

func TestProductRepository_UpdateStock(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewProductRepository(db, newDiscardLogger())

	id := uuid.New()
	initialQty := 10
	_, err := db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, $2, $3, $4)",
		id, "STOCK-SKU", "Stock Item", initialQty)
	require.NoError(t, err)

	type args struct {
		amount    int
		transType model.TransactionType
	}

	tests := []struct {
		name      string
		args      args
		wantError error
		wantQty   int
	}{
		{
			name: "Success Income",
			args: args{
				amount:    5,
				transType: model.TransactionIncome,
			},
			wantError: nil,
			wantQty:   15,
		},
		{
			name: "Success Expense",
			args: args{
				amount:    3,
				transType: model.TransactionExpense,
			},
			wantError: nil,
			wantQty:   12,
		},
		{
			name: "Insufficient Stock",
			args: args{
				amount:    100,
				transType: model.TransactionExpense,
			},
			wantError: customErrors.ErrInsufficientStock,
			wantQty:   12,
		},
		{
			name: "Invalid Type",
			args: args{
				amount:    1,
				transType: "unknown",
			},
			wantError: customErrors.ErrInvalidInput,
			wantQty:   12,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.UpdateStock(context.Background(), id, tc.args.amount, tc.args.transType)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)

				var historyCount int
				errHist := db.QueryRow("SELECT COUNT(*) FROM inventory_transactions WHERE product_id = $1 AND quantity = $2 AND type = $3",
					id, tc.args.amount, tc.args.transType).Scan(&historyCount)
				assert.NoError(t, errHist)
				assert.GreaterOrEqual(t, historyCount, 1)
			}

			var currentQty int
			errQty := db.QueryRow("SELECT quantity FROM products WHERE id = $1", id).Scan(&currentQty)
			assert.NoError(t, errQty)
			assert.Equal(t, tc.wantQty, currentQty)
		})
	}
}

func TestProductRepository_GetList(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewProductRepository(db, newDiscardLogger())

	for i := 0; i < 15; i++ {
		sku := fmt.Sprintf("SKU-%d", i)
		_, err := db.Exec("INSERT INTO products (id, sku, name, quantity, created_at) VALUES ($1, $2, $3, $4, NOW())",
			uuid.New(), sku, "Prod", 1)
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
		wantTotal int
	}{
		{
			name:      "First Page",
			limit:     10,
			offset:    0,
			wantCount: 10,
			wantTotal: 15,
		},
		{
			name:      "Second Page",
			limit:     10,
			offset:    10,
			wantCount: 5,
			wantTotal: 15,
		},
		{
			name:      "Zero Limit (Defaults)",
			limit:     0,
			offset:    0,
			wantCount: 10,
			wantTotal: 15,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list, total, err := repo.GetList(context.Background(), tc.limit, tc.offset)
			assert.NoError(t, err)
			assert.Len(t, list, tc.wantCount)
			assert.Equal(t, tc.wantTotal, total)
		})
	}
}

func TestProductRepository_GetProductHistory(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewProductRepository(db, newDiscardLogger())

	prodID := uuid.New()
	_, err := db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, $2, $3, $4)",
		prodID, "HIST-SKU", "History Item", 100)
	require.NoError(t, err)

	err = repo.UpdateStock(context.Background(), prodID, 10, model.TransactionIncome)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	err = repo.UpdateStock(context.Background(), prodID, 5, model.TransactionExpense)
	require.NoError(t, err)

	tests := []struct {
		name      string
		from      time.Time
		to        time.Time
		wantCount int
	}{
		{
			name:      "All History",
			from:      time.Time{},
			to:        time.Time{},
			wantCount: 2,
		},
		{
			name:      "Future Date (Empty)",
			from:      time.Now().Add(time.Hour),
			to:        time.Time{},
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list, _, err := repo.GetProductHistory(context.Background(), prodID, 10, 0, tc.from, tc.to)
			assert.NoError(t, err)
			assert.Len(t, list, tc.wantCount)

			if tc.wantCount > 0 {
				assert.Equal(t, "History Item", list[0].ProductName)
				assert.NotEmpty(t, list[0].BalanceAfter)
			}
		})
	}
}
