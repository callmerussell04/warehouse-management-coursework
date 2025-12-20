package handler_test

import (
	"bytes"
	"context"
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

type allowRateLimiter struct{}

func (allowRateLimiter) Allow(context.Context, string, int64, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}

type denyRateLimiter struct{}

func (denyRateLimiter) Allow(context.Context, string, int64, time.Duration) (bool, time.Duration, error) {
	return false, 30 * time.Second, nil
}

func setupUserRouter(svc handler.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewUserHandler(svc, allowRateLimiter{}, false, logger)

	r.POST("/users", h.Create)
	r.POST("/login", h.Login)
	r.POST("/logout", h.Logout)
	r.POST("/forgot-username", h.ForgotUsername)
	r.POST("/request-otp", h.RequestOTP)
	r.POST("/reset-password", h.ResetPassword)
	r.POST("/refresh", h.RefreshToken)
	r.GET("/users", h.GetList)
	r.DELETE("/users/:id", h.Delete)

	return r
}

func TestUserHandler_Create(t *testing.T) {
	type args struct {
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.UserService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
	}{
		{
			name: "Success",
			args: args{
				body: dto.CreateUserRequest{
					Username: "user1",
					Email:    "test@mail.com",
					FullName: "Test User",
					Role:     "worker",
				},
			},
			prepare: func(m *mocks.UserService) {
				u := &model.User{
					ID:       uuid.New(),
					Username: "user1",
					Email:    "test@mail.com",
					Role:     model.RoleWorker,
				}
				m.EXPECT().CreateUser(mock.Anything, "user1", "test@mail.com", "Test User", "worker").Return(u, nil)
			},
			expectedStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.UserResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "user1", resp.Username)
			},
		},
		{
			name: "Invalid Body",
			args: args{
				body: "invalid",
			},
			prepare: func(m *mocks.UserService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			args: args{
				body: dto.CreateUserRequest{
					Username: "dup",
					Email:    "dup@mail.com",
					FullName: "Dup",
					Role:     "admin",
				},
			},
			prepare: func(m *mocks.UserService) {
				m.EXPECT().CreateUser(mock.Anything, "dup", "dup@mail.com", "Dup", "admin").Return(nil, customErrors.ErrAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewUserService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupUserRouter(m)
			var body []byte
			if s, ok := tc.args.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.args.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
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

func TestUserHandler_Login(t *testing.T) {
	type args struct {
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.UserService)
		expectedStatus int
		checkCookie    bool
	}{
		{
			name: "Success",
			args: args{
				body: dto.LoginRequest{Username: "user", Password: "pass"},
			},
			prepare: func(m *mocks.UserService) {
				u := &model.User{ID: uuid.New(), Username: "user"}
				m.EXPECT().Login(mock.Anything, "user", "pass").Return("access", "refresh", u, nil)
			},
			expectedStatus: http.StatusOK,
			checkCookie:    true,
		},
		{
			name: "Invalid Body",
			args: args{
				body: "bad",
			},
			prepare: func(m *mocks.UserService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			args: args{
				body: dto.LoginRequest{Username: "user", Password: "wrong"},
			},
			prepare: func(m *mocks.UserService) {
				m.EXPECT().Login(mock.Anything, "user", "wrong").Return("", "", nil, customErrors.ErrUnauthorized)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewUserService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupUserRouter(m)
			var body []byte
			if s, ok := tc.args.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.args.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkCookie {
				cookies := w.Result().Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "refresh_token" && c.Value == "refresh" {
						found = true
						assert.True(t, c.HttpOnly)
						assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
						assert.Equal(t, "/auth", c.Path)
						assert.Empty(t, c.Domain)
						assert.False(t, c.Secure)
						break
					}
				}
				assert.True(t, found)
			}
		})
	}
}

func TestUserHandler_LoginRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewUserHandler(mocks.NewUserService(t), denyRateLimiter{}, false, logger)
	router.POST("/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"user","password":"password"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "30", w.Header().Get("Retry-After"))
}

func TestUserHandler_LoginSecureCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := mocks.NewUserService(t)
	svc.EXPECT().Login(mock.Anything, "user", "password").Return("access", "refresh", &model.User{ID: uuid.New()}, nil)
	h := handler.NewUserHandler(svc, allowRateLimiter{}, true, logger)
	router.POST("/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"user","password":"password"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, w.Result().Cookies()[0].Secure)
}

func TestUserHandler_Logout(t *testing.T) {
	m := mocks.NewUserService(t)
	r := setupUserRouter(m)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "refresh_token" && c.MaxAge < 0 {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestUserHandler_ForgotUsername(t *testing.T) {
	type args struct {
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.UserService)
		expectedStatus int
	}{
		{
			name: "Success",
			args: args{
				body: dto.ForgotUsernameRequest{Email: "test@mail.com"},
			},
			prepare: func(m *mocks.UserService) {
				m.EXPECT().RecoverUsername(mock.Anything, "test@mail.com").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid JSON",
			args: args{
				body: "bad",
			},
			prepare: func(m *mocks.UserService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			args: args{
				body: dto.ForgotUsernameRequest{Email: "err@mail.com"},
			},
			prepare: func(m *mocks.UserService) {
				m.EXPECT().RecoverUsername(mock.Anything, "err@mail.com").Return(errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewUserService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupUserRouter(m)
			var body []byte
			if s, ok := tc.args.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.args.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/forgot-username", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestUserHandler_RequestOTP(t *testing.T) {
	type args struct {
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.UserService)
		expectedStatus int
	}{
		{
			name: "Success",
			args: args{
				body: dto.SendOTPRequest{Email: "test@mail.com"},
			},
			prepare: func(m *mocks.UserService) {
				m.EXPECT().GenerateAndSendOTP(mock.Anything, "test@mail.com").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Service Error (Swallowed)",
			args: args{
				body: dto.SendOTPRequest{Email: "err@mail.com"},
			},
			prepare: func(m *mocks.UserService) {
				m.EXPECT().GenerateAndSendOTP(mock.Anything, "err@mail.com").Return(errors.New("fail"))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Body",
			args: args{
				body: "bad",
			},
			prepare: func(m *mocks.UserService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewUserService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupUserRouter(m)
			var body []byte
			if s, ok := tc.args.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.args.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/request-otp", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestUserHandler_ResetPassword(t *testing.T) {
	type args struct {
		body interface{}
	}

	tests := []struct {
		name           string
		args           args
		prepare        func(m *mocks.UserService)
		expectedStatus int
	}{
		{
			name: "Success",
			args: args{
				body: dto.ResetPasswordRequest{Email: "a@a.com", OTP: "123456", NewPassword: "new-password-123"},
			},
			prepare: func(m *mocks.UserService) {
				m.EXPECT().ResetPassword(mock.Anything, "a@a.com", "123456", "new-password-123").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Service Error",
			args: args{
				body: dto.ResetPasswordRequest{Email: "a@a.com", OTP: "123456", NewPassword: "valid-password"},
			},
			prepare: func(m *mocks.UserService) {
				m.EXPECT().ResetPassword(mock.Anything, "a@a.com", "123456", "valid-password").Return(customErrors.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Body (Validation Fail)",
			args: args{
				// Здесь тестируем именно валидацию (короткий пароль), сервис вызываться не должен
				body: dto.ResetPasswordRequest{Email: "a@a.com", OTP: "123456", NewPassword: "short"},
			},
			prepare: func(m *mocks.UserService) {
				// Ожиданий нет, так как ShouldBindJSON вернет ошибку раньше
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid JSON",
			args: args{
				body: "bad",
			},
			prepare: func(m *mocks.UserService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewUserService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupUserRouter(m)
			var body []byte
			if s, ok := tc.args.body.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tc.args.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestUserHandler_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		cookieValue    string
		prepare        func(m *mocks.UserService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
		checkClear     bool
	}{
		{
			name:        "Success",
			cookieValue: "valid-refresh",
			prepare: func(m *mocks.UserService) {
				m.EXPECT().RefreshToken(mock.Anything, "valid-refresh").Return("new-access", "new-refresh", nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp dto.RefreshTokenResponse
				assert.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "new-access", resp.AccessToken)
			},
		},
		{
			name:        "No Cookie",
			cookieValue: "",
			prepare: func(m *mocks.UserService) {
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "Service Error (Invalid)",
			cookieValue: "invalid",
			prepare: func(m *mocks.UserService) {
				m.EXPECT().RefreshToken(mock.Anything, "invalid").Return("", "", customErrors.ErrUnauthorized)
			},
			expectedStatus: http.StatusUnauthorized,
			checkClear:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewUserService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupUserRouter(m)
			req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
			if tc.cookieValue != "" {
				req.AddCookie(&http.Cookie{Name: "refresh_token", Value: tc.cookieValue})
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkBody != nil {
				tc.checkBody(t, w.Body.Bytes())
			}
			if tc.checkClear {
				cookies := w.Result().Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "refresh_token" && c.MaxAge < 0 {
						found = true
						break
					}
				}
				assert.True(t, found)
			}
		})
	}
}

func TestUserHandler_GetList(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		prepare        func(m *mocks.UserService)
		expectedStatus int
	}{
		{
			name:  "Success",
			query: "",
			prepare: func(m *mocks.UserService) {
				m.EXPECT().GetList(mock.Anything, 1, 10).Return([]*model.User{}, 0, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "Service Error",
			query: "",
			prepare: func(m *mocks.UserService) {
				m.EXPECT().GetList(mock.Anything, 1, 10).Return(nil, 0, errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewUserService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupUserRouter(m)
			req := httptest.NewRequest(http.MethodGet, "/users"+tc.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestUserHandler_Delete(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name           string
		url            string
		prepare        func(m *mocks.UserService)
		expectedStatus int
	}{
		{
			name: "Success",
			url:  "/users/" + id.String(),
			prepare: func(m *mocks.UserService) {
				m.EXPECT().Delete(mock.Anything, id).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Invalid ID",
			url:  "/users/bad",
			prepare: func(m *mocks.UserService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			url:  "/users/" + id.String(),
			prepare: func(m *mocks.UserService) {
				m.EXPECT().Delete(mock.Anything, id).Return(customErrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewUserService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupUserRouter(m)
			req := httptest.NewRequest(http.MethodDelete, tc.url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
