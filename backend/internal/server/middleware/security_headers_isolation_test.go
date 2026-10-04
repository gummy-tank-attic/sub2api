package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSecurityHeadersAuthNoStore(t *testing.T) {
	for _, path := range []string{"/api/v1/auth/login", "/api/v1/auth/refresh", "/api/v1/auth/me", "/api/v1/auth/oauth/oidc/callback"} {
		for _, status := range []int{http.StatusOK, http.StatusUnauthorized, http.StatusServiceUnavailable} {
			t.Run(path+http.StatusText(status), func(t *testing.T) {
				r := gin.New()
				r.Use(SecurityHeaders(config.CSPConfig{}, nil))
				r.Any(path, func(c *gin.Context) { c.Status(status) })
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
				require.Equal(t, status, w.Code)
				require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
			})
		}
	}
}

func TestSecurityHeadersModelRoutesSkipFramePolicy(t *testing.T) {
	for _, path := range []string{"/v1/chat/completions", "/v1/messages", "/v1/responses", "/v1beta/models", "/antigravity/v1/messages", "/responses", "/images/generations"} {
		t.Run(path, func(t *testing.T) {
			called := false
			r := gin.New()
			r.Use(SecurityHeaders(config.CSPConfig{Enabled: true}, func() []string {
				called = true
				return []string{"https://example.com"}
			}))
			r.POST(path, func(c *gin.Context) {
				require.Empty(t, GetNonceFromContext(c))
				c.Header("Cache-Control", "no-cache")
				c.SSEvent("message", "ok")
				c.Writer.Flush()
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, nil))
			require.False(t, called)
			require.True(t, w.Flushed)
			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
			require.Empty(t, w.Header().Get("Content-Security-Policy"))
			require.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			require.Contains(t, w.Body.String(), "data:ok")
		})
	}
}

func TestSecurityHeadersPreservesPublicCachePolicy(t *testing.T) {
	for _, path := range []string{"/", "/assets/app.js", "/api/v1/settings/public"} {
		t.Run(path, func(t *testing.T) {
			r := gin.New()
			r.Use(SecurityHeaders(config.CSPConfig{}, nil))
			r.GET(path, func(c *gin.Context) {
				require.Empty(t, c.Writer.Header().Get("Cache-Control"))
				c.Header("Cache-Control", "public, max-age=3600")
				c.Status(http.StatusOK)
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, "public, max-age=3600", w.Header().Get("Cache-Control"))
		})
	}
}
