package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPaymentWebhookRejectsOversizedPayloadBeforeProviderLookup(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		for _, size := range []int{maxWebhookBodySize + 1, maxWebhookBodySize + 100} {
			t.Run(fmt.Sprintf("%s/%d", method, size), func(t *testing.T) {
				r := gin.New()
				h := &PaymentWebhookHandler{}
				r.Any("/webhook", h.EasyPayNotify)
				var req *http.Request
				if method == http.MethodGet {
					req = httptest.NewRequest(method, "/webhook?payload="+strings.Repeat("x", size), nil)
				} else {
					req = httptest.NewRequest(method, "/webhook", strings.NewReader(strings.Repeat("x", size)))
				}
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
			})
		}
	}
}
