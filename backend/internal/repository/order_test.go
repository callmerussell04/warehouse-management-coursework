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
				assert.Equal(t, 5, o.Items[0].Quantity)
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

func TestOrderRepository_Update(t *testing.T) {
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
		order     *model.Order
		wantError error
		checkRes  func(*testing.T)
	}{
		{
			name: "Success Update Status",
			order: &model.Order{
				ID:        orderID,
				Status:    model.StatusCompleted,
				UpdatedAt: time.Now(),
			},
			wantError: nil,
			checkRes: func(t *testing.T) {
				var status string
				err := db.QueryRow("SELECT status FROM orders WHERE id = $1", orderID).Scan(&status)
				assert.NoError(t, err)
				assert.Equal(t, "completed", status)
			},
		},
		{
			name: "Not Found",
			order: &model.Order{
				ID:        uuid.New(),
				Status:    model.StatusCanceled,
				UpdatedAt: time.Now(),
			},
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Update(context.Background(), tc.order)

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
