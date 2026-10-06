package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHomepageURL(t *testing.T) {
	for _, tc := range []struct {
		method, path        string
		status              int
		location, canonical string
	}{
		{"GET", "/home?ref=a%2Fb&ref=c", 308, "/?ref=a%2Fb&ref=c", ""},
		{"HEAD", "/home/", 308, "/", ""},
		{"GET", "/", 200, "", `<https://www.fxvia.com/>; rel="canonical"`},
		{"GET", "/login", 200, "", ""},
		{"POST", "/home", 200, "", ""},
		{"POST", "/v1/messages", 200, "", ""},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if !handleHomepageURL(c) {
					c.Next()
				}
			})
			router.NoRoute(func(c *gin.Context) { c.Status(http.StatusOK) })
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			require.Equal(t, tc.status, w.Code)
			require.Equal(t, tc.location, w.Header().Get("Location"))
			require.Equal(t, tc.canonical, w.Header().Get("Link"))
		})
	}
}
