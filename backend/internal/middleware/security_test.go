package middleware_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"warehouse-management-system/internal/middleware"
	"warehouse-management-system/internal/model"
)

func signedToken(t *testing.T, method jwt.SigningMethod, issuer string, expiresAt *jwt.NumericDate) string {
	t.Helper()
	claims := model.UserClaims{
		UserID: "user-id",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: expiresAt,
		},
	}
	token, err := jwt.NewWithClaims(method, claims).SignedString([]byte("test-secret"))
	assert.NoError(t, err)
	return token
}

func TestAuthMiddlewareJWTValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		token  string
		status int
	}{
		{name: "valid", token: signedToken(t, jwt.SigningMethodHS256, "warehouse-system", jwt.NewNumericDate(time.Now().Add(time.Minute))), status: http.StatusOK},
		{name: "missing expiration", token: signedToken(t, jwt.SigningMethodHS256, "warehouse-system", nil), status: http.StatusUnauthorized},
		{name: "wrong issuer", token: signedToken(t, jwt.SigningMethodHS256, "other-system", jwt.NewNumericDate(time.Now().Add(time.Minute))), status: http.StatusUnauthorized},
		{name: "wrong algorithm", token: signedToken(t, jwt.SigningMethodHS384, "warehouse-system", jwt.NewNumericDate(time.Now().Add(time.Minute))), status: http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			router.Use(middleware.AuthMiddleware(logger, []byte("test-secret")))
			router.GET("/secure", func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/secure", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tc.status, w.Code)
		})
	}
}

func TestCORSMiddlewareExactAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORSMiddleware([]string{"https://warehouse.example.com"}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	allowed := httptest.NewRequest(http.MethodGet, "/", nil)
	allowed.Header.Set("Origin", "https://warehouse.example.com")
	allowedRecorder := httptest.NewRecorder()
	router.ServeHTTP(allowedRecorder, allowed)
	assert.Equal(t, "https://warehouse.example.com", allowedRecorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", allowedRecorder.Header().Get("Access-Control-Allow-Credentials"))

	blocked := httptest.NewRequest(http.MethodGet, "/", nil)
	blocked.Header.Set("Origin", "https://evil.example.com")
	blockedRecorder := httptest.NewRecorder()
	router.ServeHTTP(blockedRecorder, blocked)
	assert.Empty(t, blockedRecorder.Header().Get("Access-Control-Allow-Origin"))
}
