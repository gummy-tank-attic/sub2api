package service

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBlockedDirectLunaModel(t *testing.T) {
	const parent = "01a10ef5-ea6b-7bf3-8743-5613f860ec0e"
	tests := []struct {
		name, model, body string
		headers           map[string]string
		allow             bool
	}{
		{name: "direct Luna", model: "gpt-5.6-luna"},
		{name: "unrelated Sol", model: "gpt-6.1-sol", allow: true},
		{name: "provider alias", model: "openai/gpt-6-luna-2026-09-22"},
		{name: "header child", model: "gpt-5.6-luna", headers: map[string]string{"x-openai-subagent": "collab_spawn", "x-codex-parent-thread-id": parent}, allow: true},
		{name: "body child", model: "gpt-6-luna", body: `{"client_metadata":{"x-openai-subagent":"collab_spawn","x-codex-parent-thread-id":"` + parent + `"}}`, allow: true},
		{name: "turn metadata child", model: "gpt-5.6-luna", body: `{"client_metadata":{"x-codex-turn-metadata":"{\"subagent_kind\":\"thread_spawn\",\"parent_thread_id\":\"` + parent + `\"}"}}`, allow: true},
		{name: "missing parent", model: "gpt-6-luna", headers: map[string]string{"x-openai-subagent": "collab_spawn"}},
		{name: "invalid parent", model: "gpt-6-luna", headers: map[string]string{"x-openai-subagent": "collab_spawn", "x-codex-parent-thread-id": "invalid"}},
		{name: "title helper", model: "gpt-6-luna", headers: map[string]string{"x-openai-subagent": "thread_title", "x-codex-parent-thread-id": parent}},
		{name: "review helper", model: "gpt-6-luna", headers: map[string]string{"x-openai-subagent": "guardian", "x-codex-parent-thread-id": parent}},
		{name: "conflicting kinds", model: "gpt-6-luna", headers: map[string]string{"x-openai-subagent": "collab_spawn", "x-codex-parent-thread-id": parent}, body: `{"client_metadata":{"subagent_kind":"review"}}`},
		{name: "conflicting parents", model: "gpt-6-luna", headers: map[string]string{"x-openai-subagent": "collab_spawn", "x-codex-parent-thread-id": parent}, body: `{"client_metadata":{"parent_thread_id":"01a10ef6-3f00-7741-b805-ac38d0537745"}}`},
		{name: "invalid metadata", model: "gpt-6-luna", headers: map[string]string{"x-openai-subagent": "collab_spawn", "x-codex-parent-thread-id": parent, "x-codex-turn-metadata": "{"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
			for k, v := range tt.headers {
				c.Request.Header.Set(k, v)
			}
			blocked := BlockedDirectLunaModel(c, []byte(tt.body), []string{tt.model})
			require.Equal(t, tt.allow, blocked == "")
		})
	}
}

func TestBlockedDirectLunaModelChecksEveryCandidate(t *testing.T) {
	require.Equal(t, "gpt-5.6-luna", BlockedDirectLunaModel(nil, nil, []string{"gpt-6.1-sol", "gpt-5.6-luna"}))
}
