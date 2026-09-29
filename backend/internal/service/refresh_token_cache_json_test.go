package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func wideTokenVersion() int64 {
	return (int64(1) << 53) + 1
}

func TestRefreshTokenDataUnmarshalKeepsWideTokenVersion(t *testing.T) {
	version := wideTokenVersion()
	raw := []byte(fmt.Sprintf(`{"user_id":7,"token_version":%d,"session_generation":4,"family_id":"family-big","created_at":"2026-09-28T00:00:00Z","expires_at":"2026-09-29T00:00:00Z"}`, version))

	var data RefreshTokenData
	require.NoError(t, json.Unmarshal(raw, &data))
	require.Equal(t, version, data.TokenVersion)
	require.Equal(t, int64(7), data.UserID)
	require.Equal(t, int64(4), data.SessionGeneration)
	require.False(t, data.Consumed)
}

func TestRefreshTokenDataUnmarshalAcceptsScientificNotation(t *testing.T) {
	version := wideTokenVersion()
	raw := []byte(fmt.Sprintf("{\n  \"consumed\": true,\n  \"user_id\": 7,\n  \"token_version\": %s,\n  \"session_generation\": 4,\n  \"family_id\": \"family-sci\"\n}", strconv.FormatFloat(float64(version), 'e', -1, 64)))

	var data RefreshTokenData
	require.NoError(t, json.Unmarshal(raw, &data))
	require.Equal(t, int64(7), data.UserID)
	require.Equal(t, "family-sci", data.FamilyID)
	require.Equal(t, int64(4), data.SessionGeneration)
	require.True(t, data.Consumed)
	require.NotEqual(t, version, data.TokenVersion)
}
