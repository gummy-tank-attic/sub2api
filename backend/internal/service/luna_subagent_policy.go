package service

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

const LunaSubagentOnlyMessage = "Luna is available only for spawned Codex subagents; choose a Sol model for direct requests."

// IsLunaModel includes provider-qualified and dated public model IDs.
func IsLunaModel(model string) bool {
	model = strings.ToLower(strings.TrimSpace(model))
	if i := strings.LastIndex(model, "/"); i >= 0 {
		model = model[i+1:]
	}
	for _, base := range []string{"gpt-5.6-luna", "gpt-6-luna"} {
		if model == base || strings.HasPrefix(model, base+"-") {
			return true
		}
	}
	return false
}

// BlockedDirectLunaModel is a client compatibility policy, not trusted identity
// verification: Codex supplies these unsigned markers. Only explicit spawned
// children qualify; title, memory, compact and review helpers do not.
// All supplied markers must agree, including HTTP/WS body metadata.
func BlockedDirectLunaModel(c *gin.Context, body []byte, models []string) string {
	blocked := ""
	for _, model := range models {
		if IsLunaModel(model) {
			blocked = model
			break
		}
	}
	if blocked == "" {
		return ""
	}
	spawn, valid, parent := false, true, ""
	addKind := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			return
		}
		if value != "collab_spawn" && value != "thread_spawn" {
			valid = false
			return
		}
		spawn = true
	}
	addParent := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil || (parent != "" && parent != id.String()) {
			valid = false
			return
		}
		parent = id.String()
	}
	addMetadata := func(raw string) {
		if strings.TrimSpace(raw) == "" {
			return
		}
		if !gjson.Valid(raw) || !gjson.Parse(raw).IsObject() {
			valid = false
			return
		}
		addKind(gjson.Get(raw, "subagent_kind").String())
		addParent(gjson.Get(raw, "parent_thread_id").String())
	}
	if c != nil && c.Request != nil {
		for _, value := range c.Request.Header.Values(openAISubagentHeader) {
			addKind(value)
		}
		for _, value := range c.Request.Header.Values(codexParentThreadIDHeader) {
			addParent(value)
		}
		for _, value := range c.Request.Header.Values(codexTurnMetadataHeader) {
			addMetadata(value)
		}
	}
	if len(body) > 0 {
		cm := gjson.GetBytes(body, "client_metadata")
		addKind(cm.Get(openAISubagentHeader).String())
		addParent(cm.Get(codexParentThreadIDHeader).String())
		addKind(cm.Get("subagent_kind").String())
		addParent(cm.Get("parent_thread_id").String())
		addMetadata(cm.Get(codexTurnMetadataHeader).String())
	}
	if valid && spawn && parent != "" {
		return ""
	}
	return blocked
}
