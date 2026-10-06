package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleHomepageURL canonicalizes only browser homepage requests, never API POSTs.
func handleHomepageURL(c *gin.Context) bool {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		return false
	}
	if c.Request.URL.Path == "/home" || c.Request.URL.Path == "/home/" {
		target := "/"
		if c.Request.URL.RawQuery != "" {
			target += "?" + c.Request.URL.RawQuery
		}
		c.Redirect(http.StatusPermanentRedirect, target)
		c.Abort()
		return true
	}
	if c.Request.URL.Path == "/" {
		c.Header("Link", `<https://www.fxvia.com/>; rel="canonical"`)
	}
	return false
}
