package middleware

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/requestmodel"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LunaSubagentOnly runs after authentication and before model rewrites/billing.
// Responses WebSocket payloads are checked in the handler on every turn.
func LunaSubagentOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request == nil || isResponsesWebSocketRoute(c) {
			c.Next()
			return
		}
		var body []byte
		var models []string
		if model := groupModelAllowlistModelFromParams(c); model != "" {
			models = append(models, model)
		}
		if model := c.Query("model"); model != "" {
			models = append(models, model)
		}
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			candidates, done := groupModelAllowlistModelsFromBody(c)
			if !done {
				return
			}
			models = append(models, candidates...)
			// The previous read populated PrereadBody; this does not consume it.
			body, _ = httputil.ReadRequestBodyWithPrealloc(c.Request)
			requestmodel.ResetRequestBody(c.Request, body)
		}
		blocked := service.BlockedDirectLunaModel(c, body, models)
		for _, model := range models {
			if service.IsLunaModel(model) {
				logger.L().Info("gateway.luna_subagent_policy", zap.Bool("allowed", blocked == ""))
				break
			}
		}
		if blocked != "" {
			service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonLocalPolicyDenied)
			MarkIngressRejected(c, IngressRejectModelNotAllowed)
			groupModelAllowlistErrorWriter(c)(c, http.StatusForbidden, service.LunaSubagentOnlyMessage)
			c.Abort()
			return
		}
		c.Next()
	}
}
