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

func setupCounterpartyRouter(svc handler.CounterpartyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewCounterpartyHandler(svc, logger)

	r.POST("/counterparties", h.Create)
	r.GET("/counterparties", h.GetList)
	r.GET("/counterparties/:id", h.Get)
	r.PUT("/counterparties/:id", h.Update)
	r.DELETE("/counterparties/:id", h.Delete)

	return r
}

func TestCounterpartyHandler_Create(t *testing.T) {
	type args struct {
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.CounterpartyService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name: "Success",
			args: args{
				body: dto.CreateCounterpartyRequest{
					Name:        "Client A",
					Type:        "client",
					PhoneNumber: "123",
					Email:       "a@a.com",
				},
			},
			prepare: func(m *mocks.CounterpartyService) {
				cp := &model.Counterparty{
					ID:          uuid.New(),
					Name:        "Client A",
					Type:        model.CounterpartyClient,
					PhoneNumber: "123",
					Email:       "a@a.com",
				}
				m.EXPECT().Create(mock.Anything, "Client A", "client", "123", "a@a.com").Return(cp, nil)
			},
			expectedStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.CounterpartyResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.NotEmpty(t, resp.ID)
				assert.Equal(t, "Client A", resp.Name)
			},
		},
		{
			name: "Invalid JSON",
			args: args{
				body: "invalid",
			},
			prepare: func(m *mocks.CounterpartyService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			args: args{
				body: dto.CreateCounterpartyRequest{
					Name: "Dup",
					Type: "client",
				},
			},
			prepare: func(m *mocks.CounterpartyService) {
				m.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, customErrors.ErrAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewCounterpartyService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupCounterpartyRouter(m)

			var body []byte
			if s, ok := tc.args.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.args.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/counterparties", bytes.NewBuffer(body))
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

func TestCounterpartyHandler_Get(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		prepare        func(m *mocks.CounterpartyService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name: "Success",
			url:  "/counterparties/" + id.String(),
			prepare: func(m *mocks.CounterpartyService) {
				cp := &model.Counterparty{ID: id, Name: "Test"}
				m.EXPECT().GetByID(mock.Anything, id).Return(cp, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.CounterpartyResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, id.String(), resp.ID)
			},
		},
		{
			name: "Invalid ID",
			url:  "/counterparties/invalid-uuid",
			prepare: func(m *mocks.CounterpartyService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Not Found",
			url:  "/counterparties/" + id.String(),
			prepare: func(m *mocks.CounterpartyService) {
				m.EXPECT().GetByID(mock.Anything, id).Return(nil, customErrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewCounterpartyService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupCounterpartyRouter(m)
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

func TestCounterpartyHandler_Update(t *testing.T) {
	id := uuid.New()

	type args struct {
		url  string
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.CounterpartyService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name: "Success",
			args: args{
				url:  "/counterparties/" + id.String(),
				body: dto.UpdateCounterpartyRequest{Name: "New Name", PhoneNumber: "999", Email: "new@mail.com"},
			},
			prepare: func(m *mocks.CounterpartyService) {
				cp := &model.Counterparty{ID: id, Name: "New Name", PhoneNumber: "999", Email: "new@mail.com"}
				m.EXPECT().Update(mock.Anything, id, "New Name", "999", "new@mail.com").Return(cp, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.CounterpartyResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "New Name", resp.Name)
			},
		},
		{
			name: "Invalid ID",
			args: args{
				url:  "/counterparties/bad",
				body: dto.UpdateCounterpartyRequest{Name: "N"},
			},
			prepare: func(m *mocks.CounterpartyService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Body",
			args: args{
				url:  "/counterparties/" + id.String(),
				body: "bad json",
			},
			prepare: func(m *mocks.CounterpartyService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			args: args{
				url:  "/counterparties/" + id.String(),
				body: dto.UpdateCounterpartyRequest{Name: "N"},
			},
			prepare: func(m *mocks.CounterpartyService) {
				m.EXPECT().Update(mock.Anything, id, "N", "", "").Return(nil, customErrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewCounterpartyService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupCounterpartyRouter(m)

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

func TestCounterpartyHandler_Delete(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		prepare        func(m *mocks.CounterpartyService)
		expectedStatus int
	}{
		{
			name: "Success",
			url:  "/counterparties/" + id.String(),
			prepare: func(m *mocks.CounterpartyService) {
				m.EXPECT().Delete(mock.Anything, id).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Invalid ID",
			url:  "/counterparties/bad",
			prepare: func(m *mocks.CounterpartyService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Not Found",
			url:  "/counterparties/" + id.String(),
			prepare: func(m *mocks.CounterpartyService) {
				m.EXPECT().Delete(mock.Anything, id).Return(customErrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewCounterpartyService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupCounterpartyRouter(m)
			req := httptest.NewRequest(http.MethodDelete, tc.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestCounterpartyHandler_GetList(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		prepare        func(m *mocks.CounterpartyService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name:  "Success Default",
			query: "",
			prepare: func(m *mocks.CounterpartyService) {
				list := []*model.Counterparty{{ID: uuid.New(), Name: "A"}}
				m.EXPECT().GetList(mock.Anything, 1, 10, "").Return(list, 1, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.PagedCounterparties
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, 1, resp.Paging.Total)
				assert.Len(t, resp.Items, 1)
			},
		},
		{
			name:  "Success With Type",
			query: "?type=client",
			prepare: func(m *mocks.CounterpartyService) {
				m.EXPECT().GetList(mock.Anything, 1, 10, "client").Return([]*model.Counterparty{}, 0, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Invalid Pagination Params",
			query: "?page=-1",
			prepare: func(m *mocks.CounterpartyService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Invalid Type Param",
			query: "?type=unknown",
			prepare: func(m *mocks.CounterpartyService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Service Error",
			query: "",
			prepare: func(m *mocks.CounterpartyService) {
				m.EXPECT().GetList(mock.Anything, 1, 10, "").Return(nil, 0, errors.New("err"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewCounterpartyService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupCounterpartyRouter(m)
			req := httptest.NewRequest(http.MethodGet, "/counterparties"+tc.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, w.Body.Bytes())
			}
		})
	}
}
