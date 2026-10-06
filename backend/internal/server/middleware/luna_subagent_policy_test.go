package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestLunaSubagentOnlyAdmission(t *testing.T) {
	const parent = "01a10ef5-ea6b-7bf3-8743-5613f860ec0e"
	tests := []struct {
		name, path, body string
		child            bool
		status           int
	}{
		{"direct responses", "/v1/responses", `{"model":"gpt-5.6-luna"}`, false, 403},
		{"direct chat", "/v1/chat/completions", `{"model":"gpt-6-luna"}`, false, 403},
		{"direct messages", "/v1/messages", `{"model":"gpt-5.6-luna"}`, false, 403},
		{"root alias", "/responses", `{"model":"gpt-5.6-luna"}`, false, 403},
		{"codex alias", "/backend-api/codex/responses", `{"model":"gpt-5.6-luna"}`, false, 403},
		{"duplicate model", "/v1/responses", `{"model":"gpt-6.1-sol","model":"gpt-5.6-luna"}`, false, 403},
		{"mixed case model key", "/v1/chat/completions", `{"Model":"gpt-5.6-luna"}`, false, 403},
		{"header child", "/v1/responses", `{"model":"gpt-5.6-luna","input":"probe"}`, true, 200},
		{"body child", "/v1/responses", `{"model":"gpt-5.6-luna","client_metadata":{"x-openai-subagent":"collab_spawn","x-codex-parent-thread-id":"` + parent + `"}}`, false, 200},
		{"Sol unaffected", "/v1/responses", `{"model":"gpt-6.1-sol","input":"probe"}`, false, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(LunaSubagentOnly())
			forwarded := false
			r.POST(tt.path, func(c *gin.Context) {
				forwarded = true
				body, err := io.ReadAll(c.Request.Body)
				require.NoError(t, err)
				require.Equal(t, tt.body, string(body))
				c.Status(http.StatusOK)
			})
			req := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.child {
				req.Header.Set("x-openai-subagent", "collab_spawn")
				req.Header.Set("x-codex-parent-thread-id", parent)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			require.Equal(t, tt.status, w.Code)
			require.Equal(t, tt.status == 200, forwarded, "denied requests must not reach billing/upstream")
		})
	}
}
