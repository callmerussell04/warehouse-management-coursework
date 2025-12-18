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

func setupOrderRouter(svc handler.OrderService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewOrderHandler(svc, logger)

	r.POST("/orders", h.Create)
	r.GET("/orders", h.GetList)
	r.GET("/orders/:id", h.Get)
	r.PUT("/orders/:id", h.Update)
	r.DELETE("/orders/:id", h.Delete)

	return r
}

func TestOrderHandler_Create(t *testing.T) {
	cpID := uuid.New()
	prodID := uuid.New()
	dest := "Warehouse A"

	type args struct {
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.OrderService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name: "Success",
			args: args{
				body: dto.CreateOrderRequest{
					CounterpartyID: cpID.String(),
					OrderType:      "inbound",
					OrderDate:      time.Now(),
					Destination:    &dest,
					Items: []dto.OrderItemRequest{
						{ProductID: prodID.String(), Quantity: 10},
					},
				},
			},
			prepare: func(m *mocks.OrderService) {
				created := &model.Order{
					ID:             uuid.New(),
					CounterpartyID: cpID,
					OrderType:      model.OrderInbound,
					Status:         model.StatusPending,
					Destination:    dest,
					Items: []model.OrderItem{
						{ProductID: prodID, Quantity: 10},
					},
				}
				m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
					return o.CounterpartyID == cpID && o.OrderType == model.OrderInbound && len(o.Items) == 1
				})).Return(created, nil)
			},
			expectedStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.OrderResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, cpID.String(), resp.CounterpartyID)
				assert.Equal(t, dest, resp.Destination)
			},
		},
		{
			name: "Invalid JSON",
			args: args{
				body: "bad-json",
			},
			prepare: func(m *mocks.OrderService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Validation Error",
			args: args{
				body: dto.CreateOrderRequest{
					CounterpartyID: "",
					OrderType:      "bad-type",
				},
			},
			prepare: func(m *mocks.OrderService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			args: args{
				body: dto.CreateOrderRequest{
					CounterpartyID: cpID.String(),
					OrderType:      "outbound",
					OrderDate:      time.Now(),
					Items:          []dto.OrderItemRequest{{ProductID: prodID.String(), Quantity: 1}},
				},
			},
			prepare: func(m *mocks.OrderService) {
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, customErrors.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewOrderService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupOrderRouter(m)

			var body []byte
			if s, ok := tc.args.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.args.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestOrderHandler_Get(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		prepare        func(m *mocks.OrderService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name: "Success",
			url:  "/orders/" + id.String(),
			prepare: func(m *mocks.OrderService) {
				order := &model.Order{
					ID:        id,
					OrderType: model.OrderInbound,
					Status:    model.StatusCompleted,
				}
				m.EXPECT().GetByID(mock.Anything, id).Return(order, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.OrderResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, id.String(), resp.ID)
			},
		},
		{
			name: "Invalid ID",
			url:  "/orders/invalid-uuid",
			prepare: func(m *mocks.OrderService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Not Found",
			url:  "/orders/" + id.String(),
			prepare: func(m *mocks.OrderService) {
				m.EXPECT().GetByID(mock.Anything, id).Return(nil, customErrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewOrderService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupOrderRouter(m)
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

func TestOrderHandler_Update(t *testing.T) {
	id := uuid.New()
	status := "completed"

	type args struct {
		url  string
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.OrderService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name: "Success",
			args: args{
				url:  "/orders/" + id.String(),
				body: dto.UpdateOrderRequest{Status: &status},
			},
			prepare: func(m *mocks.OrderService) {
				order := &model.Order{
					ID:     id,
					Status: model.StatusCompleted,
				}
				m.EXPECT().Update(mock.Anything, id, status).Return(order, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.OrderResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, status, resp.Status)
			},
		},
		{
			name: "Invalid ID",
			args: args{
				url:  "/orders/bad-id",
				body: dto.UpdateOrderRequest{Status: &status},
			},
			prepare: func(m *mocks.OrderService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Body",
			args: args{
				url:  "/orders/" + id.String(),
				body: "bad-json",
			},
			prepare: func(m *mocks.OrderService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			args: args{
				url:  "/orders/" + id.String(),
				body: dto.UpdateOrderRequest{Status: &status},
			},
			prepare: func(m *mocks.OrderService) {
				m.EXPECT().Update(mock.Anything, id, status).Return(nil, customErrors.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewOrderService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupOrderRouter(m)

			var body []byte
			if s, ok := tc.args.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.args.body)
			}

			req := httptest.NewRequest(http.MethodPut, tc.args.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestOrderHandler_Delete(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		prepare        func(m *mocks.OrderService)
		expectedStatus int
	}{
		{
			name: "Success",
			url:  "/orders/" + id.String(),
			prepare: func(m *mocks.OrderService) {
				m.EXPECT().Delete(mock.Anything, id).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Invalid ID",
			url:  "/orders/invalid",
			prepare: func(m *mocks.OrderService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			url:  "/orders/" + id.String(),
			prepare: func(m *mocks.OrderService) {
				m.EXPECT().Delete(mock.Anything, id).Return(customErrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewOrderService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupOrderRouter(m)
			req := httptest.NewRequest(http.MethodDelete, tc.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestOrderHandler_GetList(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		prepare        func(m *mocks.OrderService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name:  "Success Default",
			query: "",
			prepare: func(m *mocks.OrderService) {
				list := []*model.Order{{ID: uuid.New()}}
				m.EXPECT().GetList(mock.Anything, 1, 10).Return(list, 1, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.PagedOrders
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, 1, resp.TotalCount)
				assert.Len(t, resp.Data, 1)
			},
		},
		{
			name:  "Invalid Params",
			query: "?page=-1",
			prepare: func(m *mocks.OrderService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Service Error",
			query: "",
			prepare: func(m *mocks.OrderService) {
				m.EXPECT().GetList(mock.Anything, 1, 10).Return(nil, 0, errors.New("db fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewOrderService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupOrderRouter(m)
			req := httptest.NewRequest(http.MethodGet, "/orders"+tc.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, w.Body.Bytes())
			}
		})
	}
}
