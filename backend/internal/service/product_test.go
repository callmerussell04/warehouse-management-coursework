package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
	"warehouse-management-system/internal/service"
	"warehouse-management-system/mocks"
)

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestProductService_Create(t *testing.T) {
	type args struct {
		sku  string
		name string
	}
	tests := []struct {
		name      string
		args      args
		prepare   func(m *mocks.ProductRepository)
		wantError error
		checkRes  func(*testing.T, *model.Product)
	}{
		{
			name: "Success",
			args: args{sku: "SKU-1", name: "Product 1"},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(p *model.Product) bool {
						return p.SKU == "SKU-1" && p.Name == "Product 1" && p.Quantity == 0
					})).
					Return(nil)
			},
			wantError: nil,
			checkRes: func(t *testing.T, p *model.Product) {
				assert.NotNil(t, p)
				assert.Equal(t, "SKU-1", p.SKU)
				assert.NotEmpty(t, p.ID)
			},
		},
		{
			name: "Invalid Input - Empty SKU",
			args: args{sku: "", name: "Product 1"},
			prepare: func(m *mocks.ProductRepository) {
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "Repo Error",
			args: args{sku: "SKU-1", name: "Product 1"},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(errors.New("db error"))
			},
			wantError: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewProductRepository(t)
			if tc.prepare != nil {
				tc.prepare(repo)
			}

			svc := service.NewProductService(repo, newDiscardLogger())
			got, err := svc.Create(context.Background(), tc.args.sku, tc.args.name)

			if tc.wantError != nil {
				assert.Error(t, err)
				if errors.Is(tc.wantError, customErrors.ErrInvalidInput) {
					assert.ErrorIs(t, err, tc.wantError)
				} else {
					assert.Contains(t, err.Error(), tc.wantError.Error())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tc.checkRes != nil {
					tc.checkRes(t, got)
				}
			}
		})
	}
}

func TestProductService_GetList(t *testing.T) {
	type args struct {
		page     int
		pageSize int
	}
	expectedList := []*model.Product{{ID: uuid.New(), Name: "P1"}}

	tests := []struct {
		name      string
		args      args
		prepare   func(m *mocks.ProductRepository)
		wantList  []*model.Product
		wantCount int
		wantErr   bool
	}{
		{
			name: "Success - Default Pagination",
			args: args{page: 0, pageSize: 0},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().GetList(mock.Anything, 10, 0).Return(expectedList, 1, nil)
			},
			wantList:  expectedList,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "Success - Custom Pagination",
			args: args{page: 2, pageSize: 5},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().GetList(mock.Anything, 5, 5).Return(expectedList, 10, nil)
			},
			wantList:  expectedList,
			wantCount: 10,
			wantErr:   false,
		},
		{
			name: "Repo Error",
			args: args{page: 1, pageSize: 10},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().GetList(mock.Anything, 10, 0).Return(nil, 0, errors.New("fail"))
			},
			wantList:  nil,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewProductRepository(t)
			tc.prepare(repo)

			svc := service.NewProductService(repo, newDiscardLogger())
			got, count, err := svc.GetList(context.Background(), tc.args.page, tc.args.pageSize)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantList, got)
				assert.Equal(t, tc.wantCount, count)
			}
		})
	}
}

func TestProductService_Update(t *testing.T) {
	id := uuid.New()
	type args struct {
		id   uuid.UUID
		sku  string
		name string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(m *mocks.ProductRepository)
		wantErr error
	}{
		{
			name: "Success",
			args: args{id: id, sku: "NEW-SKU", name: "New Name"},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().Update(mock.Anything, mock.MatchedBy(func(p *model.Product) bool {
					return p.ID == id && p.SKU == "NEW-SKU" && p.Name == "New Name"
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Invalid Input",
			args: args{id: id, sku: "", name: "Name"},
			prepare: func(m *mocks.ProductRepository) {
			},
			wantErr: customErrors.ErrInvalidInput,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewProductRepository(t)
			if tc.prepare != nil {
				tc.prepare(repo)
			}

			svc := service.NewProductService(repo, newDiscardLogger())
			got, err := svc.Update(context.Background(), tc.args.id, tc.args.sku, tc.args.name)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.args.sku, got.SKU)
			}
		})
	}
}

func TestProductService_ChangeStock(t *testing.T) {
	id := uuid.New()
	type args struct {
		productID       uuid.UUID
		amount          int64
		transactionType string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(m *mocks.ProductRepository)
		wantErr bool
	}{
		{
			name: "Success Income",
			args: args{productID: id, amount: 10, transactionType: "income"},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().UpdateStock(mock.Anything, id, int64(10), model.TransactionIncome).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Success Expense",
			args: args{productID: id, amount: 5, transactionType: "expense"},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().UpdateStock(mock.Anything, id, int64(5), model.TransactionExpense).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Invalid Amount",
			args: args{productID: id, amount: -1, transactionType: "income"},
			prepare: func(m *mocks.ProductRepository) {
			},
			wantErr: true,
		},
		{
			name: "Invalid Type",
			args: args{productID: id, amount: 10, transactionType: "unknown"},
			prepare: func(m *mocks.ProductRepository) {
			},
			wantErr: true,
		},
		{
			name: "Repo Error",
			args: args{productID: id, amount: 10, transactionType: "income"},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().UpdateStock(mock.Anything, id, int64(10), model.TransactionIncome).Return(errors.New("db fail"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewProductRepository(t)
			if tc.prepare != nil {
				tc.prepare(repo)
			}
			svc := service.NewProductService(repo, newDiscardLogger())

			err := svc.ChangeStock(context.Background(), tc.args.productID, tc.args.amount, tc.args.transactionType)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProductService_GetProductHistory(t *testing.T) {
	id := uuid.New()
	type args struct {
		productID uuid.UUID
		page      int
		pageSize  int
		fromStr   string
		toStr     string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(m *mocks.ProductRepository)
		wantErr bool
	}{
		{
			name: "Success - Basic Pagination",
			args: args{productID: id, page: 1, pageSize: 10, fromStr: "", toStr: ""},
			prepare: func(m *mocks.ProductRepository) {
				m.EXPECT().GetProductHistory(mock.Anything, id, 10, 0, time.Time{}, time.Time{}).
					Return([]*model.InventoryTransaction{}, 0, nil)
			},
			wantErr: false,
		},
		{
			name: "Success - With Date Parsing (DateOnly)",
			args: args{productID: id, page: 1, pageSize: 10, fromStr: "2025-01-01", toStr: "2025-01-05"},
			prepare: func(m *mocks.ProductRepository) {
				from, _ := time.Parse(time.DateOnly, "2025-01-01")
				to, _ := time.Parse(time.DateOnly, "2025-01-05")
				toAdjusted := to.AddDate(0, 0, 1)

				m.EXPECT().GetProductHistory(mock.Anything, id, 10, 0, from, toAdjusted).
					Return(nil, 0, nil)
			},
			wantErr: false,
		},
		{
			name: "Success - RFC3339",
			args: args{productID: id, page: 1, pageSize: 10, fromStr: "2025-01-01T10:00:00Z", toStr: ""},
			prepare: func(m *mocks.ProductRepository) {
				from, _ := time.Parse(time.RFC3339, "2025-01-01T10:00:00Z")
				m.EXPECT().GetProductHistory(mock.Anything, id, 10, 0, from, time.Time{}).
					Return(nil, 0, nil)
			},
			wantErr: false,
		},
		{
			name: "Invalid From Format",
			args: args{productID: id, page: 1, pageSize: 10, fromStr: "invalid-date", toStr: ""},
			prepare: func(m *mocks.ProductRepository) {
			},
			wantErr: true,
		},
		{
			name: "Invalid To Format",
			args: args{productID: id, page: 1, pageSize: 10, fromStr: "", toStr: "bad-date"},
			prepare: func(m *mocks.ProductRepository) {
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewProductRepository(t)
			if tc.prepare != nil {
				tc.prepare(repo)
			}
			svc := service.NewProductService(repo, newDiscardLogger())

			_, _, err := svc.GetProductHistory(context.Background(), tc.args.productID, tc.args.page, tc.args.pageSize, tc.args.fromStr, tc.args.toStr)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProductService_GetByID(t *testing.T) {
	id := uuid.New()
	expectedProd := &model.Product{ID: id, Name: "Test"}

	repo := mocks.NewProductRepository(t)
	repo.EXPECT().GetByID(mock.Anything, id).Return(expectedProd, nil)

	svc := service.NewProductService(repo, newDiscardLogger())
	got, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, expectedProd, got)
}

func TestProductService_Delete(t *testing.T) {
	id := uuid.New()

	repo := mocks.NewProductRepository(t)
	repo.EXPECT().Delete(mock.Anything, id).Return(nil)

	svc := service.NewProductService(repo, newDiscardLogger())
	err := svc.Delete(context.Background(), id)

	assert.NoError(t, err)
}
