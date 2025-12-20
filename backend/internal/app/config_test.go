package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setValidConfig(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("ADMIN_USERNAME", "admin")
	t.Setenv("ADMIN_EMAIL", "admin@example.com")
	t.Setenv("ADMIN_PASSWORD", "strong-password-123")
	t.Setenv("ALLOWED_ORIGINS", "http://localhost:5000,https://warehouse.example.com")
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("TRUSTED_PROXIES", "")
}

func TestLoadConfig(t *testing.T) {
	setValidConfig(t)

	cfg, err := loadConfig()
	require.NoError(t, err)
	assert.Len(t, cfg.jwtSecret, 32)
	assert.Equal(t, []string{"http://localhost:5000", "https://warehouse.example.com"}, cfg.allowedOrigins)
	assert.True(t, cfg.cookieSecure)
	assert.Empty(t, cfg.trustedProxies)
}

func TestLoadConfigParsesTrustedProxies(t *testing.T) {
	setValidConfig(t)
	t.Setenv("TRUSTED_PROXIES", "172.30.50.254,10.0.0.0/24")

	cfg, err := loadConfig()
	require.NoError(t, err)
	assert.Equal(t, []string{"172.30.50.254", "10.0.0.0/24"}, cfg.trustedProxies)
}

func TestClientIPOnlyUsesHeadersFromTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies([]string{"172.30.50.254"}))
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, c.ClientIP())
	})

	tests := []struct {
		name       string
		remoteAddr string
		forwarded  string
		want       string
	}{
		{
			name:       "trusted gateway forwards client address",
			remoteAddr: "172.30.50.254:12345",
			forwarded:  "198.51.100.10",
			want:       "198.51.100.10",
		},
		{
			name:       "untrusted client cannot spoof address",
			remoteAddr: "203.0.113.20:12345",
			forwarded:  "198.51.100.10",
			want:       "203.0.113.20",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remoteAddr
			req.Header.Set("X-Forwarded-For", tc.forwarded)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, req)

			assert.Equal(t, http.StatusOK, response.Code)
			assert.Equal(t, tc.want, response.Body.String())
		})
	}
}

func TestLoadConfigRejectsUnsafeValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "short JWT secret", key: "JWT_SECRET", value: "short"},
		{name: "short admin password", key: "ADMIN_PASSWORD", value: "short"},
		{name: "wildcard origin", key: "ALLOWED_ORIGINS", value: "*"},
		{name: "invalid cookie flag", key: "COOKIE_SECURE", value: "sometimes"},
		{name: "invalid trusted proxy", key: "TRUSTED_PROXIES", value: "reverse-proxy"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setValidConfig(t)
			t.Setenv(tc.key, tc.value)
			_, err := loadConfig()
			assert.Error(t, err)
		})
	}
}
