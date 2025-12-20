package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"warehouse-management-system/internal/dto"
	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/handler"
	"warehouse-management-system/internal/model"
	"warehouse-management-system/mocks"
)

func setupProductRouter(svc handler.ProductService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewProductHandler(svc, logger)

	r.POST("/products", h.Create)
	r.GET("/products", h.GetList)
	r.GET("/products/:id", h.Get)
	r.PUT("/products/:id", h.Update)
	r.DELETE("/products/:id", h.Delete)
	r.POST("/products/:id/stock", h.UpdateStock)
	r.GET("/products/:id/history", h.GetHistory)

	return r
}

func TestProductHandler_Create(t *testing.T) {
	type args struct {
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.ProductService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name: "Success",
			args: args{
				body: dto.CreateProductRequest{SKU: "SKU-1", Name: "Phone"},
			},
			prepare: func(m *mocks.ProductService) {
				p := &model.Product{
					ID:        uuid.New(),
					SKU:       "SKU-1",
					Name:      "Phone",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				m.EXPECT().Create(mock.Anything, "SKU-1", "Phone").Return(p, nil)
			},
			expectedStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.ProductResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "SKU-1", resp.SKU)
				assert.NotEmpty(t, resp.ID)
			},
		},
		{
			name: "Invalid JSON",
			args: args{
				body: "invalid",
			},
			prepare: func(m *mocks.ProductService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			args: args{
				body: dto.CreateProductRequest{SKU: "DUP", Name: "Dup"},
			},
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().Create(mock.Anything, "DUP", "Dup").Return(nil, customErrors.ErrAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewProductService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupProductRouter(m)

			var body []byte
			if s, ok := tc.args.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.args.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestProductHandler_GetList(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		prepare        func(m *mocks.ProductService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name:  "Success Default",
			query: "",
			prepare: func(m *mocks.ProductService) {
				list := []*model.Product{{ID: uuid.New(), SKU: "A"}}
				m.EXPECT().GetList(mock.Anything, 1, 10).Return(list, 1, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.PagedProducts
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, 1, resp.Paging.Total)
				assert.Len(t, resp.Items, 1)
			},
		},
		{
			name:  "Invalid Params",
			query: "?page=-1",
			prepare: func(m *mocks.ProductService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Service Error",
			query: "",
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().GetList(mock.Anything, 1, 10).Return(nil, 0, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewProductService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupProductRouter(m)
			req := httptest.NewRequest(http.MethodGet, "/products"+tc.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestProductHandler_Get(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		prepare        func(m *mocks.ProductService)
		expectedStatus int
	}{
		{
			name: "Success",
			url:  "/products/" + id.String(),
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().GetByID(mock.Anything, id).Return(&model.Product{ID: id}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid ID",
			url:  "/products/invalid-uuid",
			prepare: func(m *mocks.ProductService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Not Found",
			url:  "/products/" + id.String(),
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().GetByID(mock.Anything, id).Return(nil, customErrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewProductService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupProductRouter(m)
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_Update(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		body           interface{}
		prepare        func(m *mocks.ProductService)
		expectedStatus int
	}{
		{
			name: "Success",
			url:  "/products/" + id.String(),
			body: dto.UpdateProductRequest{SKU: "NEW", Name: "New"},
			prepare: func(m *mocks.ProductService) {
				p := &model.Product{ID: id, SKU: "NEW", Name: "New"}
				m.EXPECT().Update(mock.Anything, id, "NEW", "New").Return(p, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid ID",
			url:  "/products/abc",
			body: dto.UpdateProductRequest{SKU: "N", Name: "N"},
			prepare: func(m *mocks.ProductService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			url:  "/products/" + id.String(),
			body: dto.UpdateProductRequest{SKU: "N", Name: "N"},
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().Update(mock.Anything, id, "N", "N").Return(nil, customErrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewProductService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupProductRouter(m)
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPut, tc.url, bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_Delete(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		prepare        func(m *mocks.ProductService)
		expectedStatus int
	}{
		{
			name: "Success",
			url:  "/products/" + id.String(),
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().Delete(mock.Anything, id).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Not Found",
			url:  "/products/" + id.String(),
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().Delete(mock.Anything, id).Return(customErrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewProductService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupProductRouter(m)
			req := httptest.NewRequest(http.MethodDelete, tc.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_UpdateStock(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		body           interface{}
		prepare        func(m *mocks.ProductService)
		expectedStatus int
	}{
		{
			name: "Success",
			url:  "/products/" + id.String() + "/stock",
			body: dto.UpdateStockRequest{Quantity: 10, Type: "income"},
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().ChangeStock(mock.Anything, id, int64(10), "income").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Body",
			url:  "/products/" + id.String() + "/stock",
			body: "bad",
			prepare: func(m *mocks.ProductService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			url:  "/products/" + id.String() + "/stock",
			body: dto.UpdateStockRequest{Quantity: 10, Type: "expense"},
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().ChangeStock(mock.Anything, id, int64(10), "expense").Return(customErrors.ErrInsufficientStock)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewProductService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupProductRouter(m)
			var body []byte
			if s, ok := tc.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.body)
			}
			req := httptest.NewRequest(http.MethodPost, tc.url, bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_GetHistory(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		prepare        func(m *mocks.ProductService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name: "Success",
			url:  "/products/" + id.String() + "/history?from=2023-01-01",
			prepare: func(m *mocks.ProductService) {
				hist := []*model.InventoryTransaction{
					{ID: uuid.New(), ProductID: id, Quantity: 5},
				}
				m.EXPECT().GetProductHistory(mock.Anything, id, 1, 10, "2023-01-01", "").Return(hist, 1, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.PagedTransactions
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Len(t, resp.Items, 1)
			},
		},
		{
			name: "Invalid UUID",
			url:  "/products/bad-uuid/history",
			prepare: func(m *mocks.ProductService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			url:  "/products/" + id.String() + "/history",
			prepare: func(m *mocks.ProductService) {
				m.EXPECT().GetProductHistory(mock.Anything, id, 1, 10, "", "").Return(nil, 0, errors.New("err"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewProductService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupProductRouter(m)
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, w.Body.Bytes())
			}
		})
	}
}
