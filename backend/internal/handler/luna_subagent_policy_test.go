package handler

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOpenAIResponsesWebSocket_LunaPolicyEveryTurn(t *testing.T) {
	const child = `"client_metadata":{"x-openai-subagent":"collab_spawn","x-codex-parent-thread-id":"01a10ef5-ea6b-7bf3-8743-5613f860ec0e"}`
	t.Run("child Luna is forwarded and billed", func(t *testing.T) {
		got := runOpenAIResponsesWebSocketUsageLogCase(t, openAIResponsesWSUsageLogCase{
			firstPayload: `{"type":"response.create","model":"gpt-5.6-luna",` + child + `}`,
		})
		require.Len(t, got.upstreamPayloads, 1)
		require.Len(t, got.logs, 1)
		require.Equal(t, "gpt-5.6-luna", got.logs[0].RequestedModel)
	})
	t.Run("Sol connection cannot switch to direct Luna", func(t *testing.T) {
		runOpenAIResponsesWebSocketUsageLogCase(t, openAIResponsesWSUsageLogCase{
			firstPayload:            `{"type":"response.create","model":"gpt-5.6-sol"}`,
			secondPayload:           `{"type":"response.create","model":"gpt-5.6-luna"}`,
			secondTurnCloseExpected: true,
			closeReason:             "Luna",
		})
	})
	t.Run("inherited Luna also requires a child marker", func(t *testing.T) {
		runOpenAIResponsesWebSocketUsageLogCase(t, openAIResponsesWSUsageLogCase{
			firstPayload:            `{"type":"response.create","model":"gpt-5.6-luna",` + child + `}`,
			secondPayload:           `{"type":"response.create"}`,
			secondTurnCloseExpected: true,
			closeReason:             "Luna",
		})
	})
}

func TestLunaPolicyDenialDoesNotPenalizeUpstream(t *testing.T) {
	err := service.NewOpenAIWSClientCloseError(coderws.StatusPolicyViolation, service.LunaSubagentOnlyMessage, errLunaSubagentPolicyDenied)
	require.False(t, shouldReportOpenAIWSProxyAccountFailure(fmt.Errorf("wrapped: %w", err)))
}

func TestOpenAIResponsesWebSocket_RejectsDirectLunaBeforeScheduling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newOpenAIHandlerForPreviousResponseIDValidation(t, nil)
	s := newOpenAIWSHandlerTestServer(t, h, middleware.AuthSubject{UserID: 1, Concurrency: 1})
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := coderws.Dial(ctx, "ws"+strings.TrimPrefix(s.URL, "http")+"/openai/v1/responses", nil)
	require.NoError(t, err)
	defer func() { _ = conn.CloseNow() }()
	require.NoError(t, conn.Write(ctx, coderws.MessageText, []byte(`{"type":"response.create","model":"gpt-5.6-luna","input":"probe"}`)))
	_, _, err = conn.Read(ctx)
	require.Equal(t, coderws.StatusPolicyViolation, coderws.CloseStatus(err))
	var closeErr coderws.CloseError
	require.ErrorAs(t, err, &closeErr)
	require.Contains(t, closeErr.Reason, "Luna")
}
