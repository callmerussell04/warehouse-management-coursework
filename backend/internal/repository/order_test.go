package repository_test

import (
	"context"
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

func TestOrderRepository_Create(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewOrderRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM orders")
		_, _ = db.Exec("DELETE FROM products")
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	cpID := uuid.New()
	_, err := db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
		cpID, "Supplier 1", "supplier", "123", "sup@mail.com")
	require.NoError(t, err)

	prodID := uuid.New()
	_, err = db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, $2, $3, $4)",
		prodID, "SKU-1", "Prod 1", 100)
	require.NoError(t, err)

	orderID := uuid.New()

	type args struct {
		order *model.Order
	}

	tests := []struct {
		name      string
		args      args
		wantError error
		checkRes  func(*testing.T)
	}{
		{
			name: "Success Inbound",
			args: args{
				order: &model.Order{
					ID:             orderID,
					CounterpartyID: cpID,
					Status:         model.StatusPending,
					OrderType:      model.OrderInbound,
					OrderDate:      time.Now(),
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
					Items: []model.OrderItem{
						{ProductID: prodID, Quantity: 10},
					},
				},
			},
			wantError: nil,
			checkRes: func(t *testing.T) {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM orders WHERE id = $1", orderID).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 1, count)

				var itemCount int
				err = db.QueryRow("SELECT COUNT(*) FROM order_items WHERE order_id = $1", orderID).Scan(&itemCount)
				assert.NoError(t, err)
				assert.Equal(t, 1, itemCount)
			},
		},
		{
			name: "Counterparty Not Found",
			args: args{
				order: &model.Order{
					ID:             uuid.New(),
					CounterpartyID: uuid.New(), // Random
					Status:         model.StatusPending,
					OrderType:      model.OrderInbound,
					OrderDate:      time.Now(),
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
					Items:          []model.OrderItem{},
				},
			},
			wantError: customErrors.ErrNotFound,
		},
		{
			name: "Product Not Found",
			args: args{
				order: &model.Order{
					ID:             uuid.New(),
					CounterpartyID: cpID,
					Status:         model.StatusPending,
					OrderType:      model.OrderInbound,
					OrderDate:      time.Now(),
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
					Items: []model.OrderItem{
						{ProductID: uuid.New(), Quantity: 5}, // Random product
					},
				},
			},
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tc.args.order)

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

func TestOrderRepository_GetByID(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewOrderRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM orders")
		_, _ = db.Exec("DELETE FROM products")
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	cpID := uuid.New()
	_, err := db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
		cpID, "Client 1", "client", "111", "cli@mail.com")
	require.NoError(t, err)

	prodID := uuid.New()
	_, err = db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, $2, $3, $4)",
		prodID, "SKU-TEST", "Prod Test", 50)
	require.NoError(t, err)

	orderID := uuid.New()
	_, err = db.Exec(`INSERT INTO orders (id, counterparty_id, status, order_date, created_at, updated_at, order_type, destination) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		orderID, cpID, "pending", time.Now(), time.Now(), time.Now(), "outbound", "Address")
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)",
		orderID, prodID, 5)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uuid.UUID
		wantError error
		checkRes  func(*testing.T, *model.Order)
	}{
		{
			name:      "Success",
			id:        orderID,
			wantError: nil,
			checkRes: func(t *testing.T, o *model.Order) {
				assert.Equal(t, model.OrderOutbound, o.OrderType)
				assert.Equal(t, "Address", o.Destination)
				assert.Len(t, o.Items, 1)
				assert.Equal(t, int64(5), o.Items[0].Quantity)
			},
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
				if tc.checkRes != nil {
					tc.checkRes(t, got)
				}
			}
		})
	}
}

func TestOrderRepository_Transition(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewOrderRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM orders")
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	cpID := uuid.New()
	_, err := db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
		cpID, "Sup", "supplier", "000", "s@m.com")
	require.NoError(t, err)

	orderID := uuid.New()
	_, err = db.Exec(`INSERT INTO orders (id, counterparty_id, status, order_date, created_at, updated_at, order_type) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		orderID, cpID, "pending", time.Now(), time.Now(), time.Now(), "inbound")
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uuid.UUID
		status    model.OrderStatus
		wantError error
		checkRes  func(*testing.T)
	}{
		{
			name:      "Success Update Status",
			id:        orderID,
			status:    model.StatusCompleted,
			wantError: nil,
			checkRes: func(t *testing.T) {
				var status string
				err := db.QueryRow("SELECT status FROM orders WHERE id = $1", orderID).Scan(&status)
				assert.NoError(t, err)
				assert.Equal(t, "completed", status)
			},
		},
		{
			name:      "Not Found",
			id:        uuid.New(),
			status:    model.StatusCanceled,
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updated, err := repo.Transition(context.Background(), tc.id, tc.status)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
				assert.Nil(t, updated)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.status, updated.Status)
				if tc.checkRes != nil {
					tc.checkRes(t)
				}
			}
		})
	}
}

func TestOrderRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewOrderRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM orders")
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	cpID := uuid.New()
	_, _ = db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
		cpID, "S", "supplier", "1", "e")

	orderID := uuid.New()
	_, _ = db.Exec(`INSERT INTO orders (id, counterparty_id, status, order_date, created_at, updated_at, order_type) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		orderID, cpID, "canceled", time.Now(), time.Now(), time.Now(), "inbound")

	tests := []struct {
		name      string
		id        uuid.UUID
		wantError error
	}{
		{
			name:      "Success",
			id:        orderID,
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
				_ = db.QueryRow("SELECT COUNT(*) FROM orders WHERE id = $1", tc.id).Scan(&count)
				assert.Equal(t, 0, count)
			}
		})
	}
}

func TestOrderRepository_GetList(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewOrderRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM orders")
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	cpID := uuid.New()
	_, _ = db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
		cpID, "S", "supplier", "1", "e")

	for i := 0; i < 5; i++ {
		_, err := db.Exec(`INSERT INTO orders (id, counterparty_id, status, order_date, created_at, updated_at, order_type) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), cpID, "pending", time.Now().Add(time.Duration(i)*time.Hour), time.Now(), time.Now(), "inbound")
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
			name:      "All",
			limit:     10,
			offset:    0,
			wantCount: 5,
			wantTotal: 5,
		},
		{
			name:      "Limit 2",
			limit:     2,
			offset:    0,
			wantCount: 2,
			wantTotal: 5,
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

func TestOrderRepository_CompletionIsAtomic(t *testing.T) {
	t.Run("completion changes stock and history once", func(t *testing.T) {
		db := newTestDB(t)
		defer db.Close()
		repo := repository.NewOrderRepository(db, newDiscardLogger())
		ctx := context.Background()
		cpID, productID, orderID := uuid.New(), uuid.New(), uuid.New()
		_, err := db.Exec("INSERT INTO counterparties (id, name, type) VALUES ($1, 'Client', 'client')", cpID)
		require.NoError(t, err)
		_, err = db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, 'ATOMIC-1', 'Atomic', 10)", productID)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO orders (id, counterparty_id, status, order_date, order_type)
			VALUES ($1, $2, 'processing', NOW(), 'outbound')`, orderID, cpID)
		require.NoError(t, err)
		_, err = db.Exec("INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, 6)", orderID, productID)
		require.NoError(t, err)

		order, err := repo.Transition(ctx, orderID, model.StatusCompleted)
		require.NoError(t, err)
		assert.Equal(t, model.StatusCompleted, order.Status)

		var quantity int64
		var historyCount int
		require.NoError(t, db.QueryRow("SELECT quantity FROM products WHERE id = $1", productID).Scan(&quantity))
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM inventory_transactions WHERE product_id = $1", productID).Scan(&historyCount))
		assert.Equal(t, int64(4), quantity)
		assert.Equal(t, 1, historyCount)
	})

	t.Run("insufficient stock rolls back every write", func(t *testing.T) {
		db := newTestDB(t)
		defer db.Close()
		repo := repository.NewOrderRepository(db, newDiscardLogger())
		ctx := context.Background()
		cpID := uuid.New()
		firstID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		secondID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
		orderID := uuid.New()
		_, err := db.Exec("INSERT INTO counterparties (id, name, type) VALUES ($1, 'Client', 'client')", cpID)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO products (id, sku, name, quantity) VALUES
			($1, 'ROLLBACK-1', 'First', 10), ($2, 'ROLLBACK-2', 'Second', 1)`, firstID, secondID)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO orders (id, counterparty_id, status, order_date, order_type)
			VALUES ($1, $2, 'processing', NOW(), 'outbound')`, orderID, cpID)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO order_items (order_id, product_id, quantity) VALUES
			($1, $2, 5), ($1, $3, 5)`, orderID, firstID, secondID)
		require.NoError(t, err)

		order, err := repo.Transition(ctx, orderID, model.StatusCompleted)
		assert.ErrorIs(t, err, customErrors.ErrInsufficientStock)
		assert.Nil(t, order)

		var firstQuantity, secondQuantity int64
		var historyCount int
		var status model.OrderStatus
		require.NoError(t, db.QueryRow("SELECT quantity FROM products WHERE id = $1", firstID).Scan(&firstQuantity))
		require.NoError(t, db.QueryRow("SELECT quantity FROM products WHERE id = $1", secondID).Scan(&secondQuantity))
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM inventory_transactions WHERE product_id IN ($1, $2)", firstID, secondID).Scan(&historyCount))
		require.NoError(t, db.QueryRow("SELECT status FROM orders WHERE id = $1", orderID).Scan(&status))
		assert.Equal(t, int64(10), firstQuantity)
		assert.Equal(t, int64(1), secondQuantity)
		assert.Zero(t, historyCount)
		assert.Equal(t, model.StatusProcessing, status)
	})

	t.Run("concurrent completion applies once", func(t *testing.T) {
		db := newTestDB(t)
		defer db.Close()
		repo := repository.NewOrderRepository(db, newDiscardLogger())
		cpID, productID, orderID := uuid.New(), uuid.New(), uuid.New()
		_, err := db.Exec("INSERT INTO counterparties (id, name, type) VALUES ($1, 'Client', 'client')", cpID)
		require.NoError(t, err)
		_, err = db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, 'CONCURRENT-1', 'Concurrent', 10)", productID)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO orders (id, counterparty_id, status, order_date, order_type)
			VALUES ($1, $2, 'processing', NOW(), 'outbound')`, orderID, cpID)
		require.NoError(t, err)
		_, err = db.Exec("INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, 3)", orderID, productID)
		require.NoError(t, err)

		errorsCh := make(chan error, 2)
		for i := 0; i < 2; i++ {
			go func() {
				_, transitionErr := repo.Transition(context.Background(), orderID, model.StatusCompleted)
				errorsCh <- transitionErr
			}()
		}
		var successCount int
		for i := 0; i < 2; i++ {
			if <-errorsCh == nil {
				successCount++
			}
		}

		var quantity int64
		var historyCount int
		require.NoError(t, db.QueryRow("SELECT quantity FROM products WHERE id = $1", productID).Scan(&quantity))
		require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM inventory_transactions WHERE product_id = $1", productID).Scan(&historyCount))
		assert.Equal(t, 1, successCount)
		assert.Equal(t, int64(7), quantity)
		assert.Equal(t, 1, historyCount)
	})
}
