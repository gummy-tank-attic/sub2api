package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	refreshTokenKeyPrefix      = "refresh_token:"
	refreshTokenConsumedPrefix = "refresh_token_consumed:"
	userRefreshTokensPrefix    = "user_refresh_tokens:"
	tokenFamilyPrefix          = "token_family:"
	revokedTokenFamilyPrefix   = "revoked_token_family:"
	accessTokenRevokedPrefix   = "access_tokens_revoked:"
	sessionGenerationPrefix    = "session_generation:"
	refreshJSONFamilyID        = "family_id"
	refreshJSONUserID          = "user_id"
)

// refreshTokenFieldReader pulls one string or integer field out of a JSON
// object. Field names are passed in, and whitespace around the colon is allowed.
const refreshTokenFieldReader = `
local function json_string(payload, field)
  local _, _, found = string.find(payload, '"' .. field .. '"%s*:%s*"([^"]*)"')
  return found
end
local function json_integer(payload, field)
  local _, _, found = string.find(payload, '"' .. field .. '"%s*:%s*(%-?%d+)')
  return found
end
`

// refreshTokenKey generates the Redis key for a refresh token.
func refreshTokenKey(tokenHash string) string {
	return refreshTokenKeyPrefix + tokenHash
}

func refreshTokenConsumedKey(tokenHash string) string {
	return refreshTokenConsumedPrefix + tokenHash
}

// userRefreshTokensKey generates the Redis key for user's token set.
func userRefreshTokensKey(userID int64) string {
	return fmt.Sprintf("%s%d", userRefreshTokensPrefix, userID)
}

// tokenFamilyKey generates the Redis key for token family set.
func tokenFamilyKey(familyID string) string {
	return tokenFamilyPrefix + familyID
}

func revokedTokenFamilyKey(familyID string) string {
	return revokedTokenFamilyPrefix + familyID
}

func accessTokenRevokedKey(userID int64) string {
	return fmt.Sprintf("%s%d", accessTokenRevokedPrefix, userID)
}

func sessionGenerationKey(userID int64) string {
	return fmt.Sprintf("%s%d", sessionGenerationPrefix, userID)
}

type refreshTokenCache struct {
	rdb *redis.Client
}

// NewRefreshTokenCache creates a new RefreshTokenCache implementation.
func NewRefreshTokenCache(rdb *redis.Client) service.RefreshTokenCache {
	return &refreshTokenCache{rdb: rdb}
}

func (c *refreshTokenCache) StoreRefreshToken(ctx context.Context, tokenHash string, data *service.RefreshTokenData, ttl time.Duration) error {
	val, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal refresh token data: %w", err)
	}
	if ttl <= 0 {
		return fmt.Errorf("store refresh token: invalid ttl")
	}
	const script = `
if redis.call('EXISTS', KEYS[4]) == 1 then
  return 0
end
local current_generation = tonumber(redis.call('GET', KEYS[5]) or '0')
if current_generation ~= tonumber(ARGV[4]) then
  return -1
end
local user_type = redis.call('TYPE', KEYS[2]).ok
local family_type = redis.call('TYPE', KEYS[3]).ok
if (user_type ~= 'none' and user_type ~= 'set') or (family_type ~= 'none' and family_type ~= 'set') then
  return redis.error_reply('refresh token index has unexpected type')
end
redis.call('PSETEX', KEYS[1], ARGV[2], ARGV[1])
redis.call('SADD', KEYS[2], ARGV[3])
redis.call('PEXPIRE', KEYS[2], ARGV[2])
redis.call('SADD', KEYS[3], ARGV[3])
redis.call('PEXPIRE', KEYS[3], ARGV[2])
return 1`
	stored, err := c.rdb.Eval(ctx, script, []string{
		refreshTokenKey(tokenHash),
		userRefreshTokensKey(data.UserID),
		tokenFamilyKey(data.FamilyID),
		revokedTokenFamilyKey(data.FamilyID),
		sessionGenerationKey(data.UserID),
	}, string(val), ttl.Milliseconds(), tokenHash, data.SessionGeneration).Int()
	if err != nil {
		return err
	}
	if stored != 1 {
		return service.ErrTokenRevoked
	}
	return nil
}

func (c *refreshTokenCache) GetRefreshToken(ctx context.Context, tokenHash string) (*service.RefreshTokenData, error) {
	key := refreshTokenKey(tokenHash)
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, service.ErrRefreshTokenNotFound
		}
		return nil, err
	}
	var data service.RefreshTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("%w: unmarshal refresh token data: %v", service.ErrRefreshTokenCorrupt, err)
	}
	return &data, nil
}

const (
	refreshConsumeMissing int64 = 0
	refreshConsumeWon     int64 = 1
	refreshConsumeReplay  int64 = 2
)

// ConsumeRefreshToken atomically marks a live refresh token as used.
// The JSON value is left unchanged. Redis Lua numbers are IEEE-754 doubles,
// so rewriting the body with cjson would corrupt int64 fields above 2^53.
// Replay state lives in a sibling key that expires with the original token.
func (c *refreshTokenCache) ConsumeRefreshToken(ctx context.Context, tokenHash string) (*service.RefreshTokenData, error) {
	const script = `
local value = redis.call('GET', KEYS[1])
if not value then
  return {0, ''}
end
local ttl = redis.call('PTTL', KEYS[1])
local claimed
if ttl > 0 then
  claimed = redis.call('SET', KEYS[2], '1', 'PX', ttl, 'NX')
else
  claimed = redis.call('SET', KEYS[2], '1', 'NX')
end
if claimed then
  return {1, value}
end
return {2, value}`

	result, err := c.rdb.Eval(ctx, script, []string{
		refreshTokenKey(tokenHash),
		refreshTokenConsumedKey(tokenHash),
	}).Result()
	if err != nil {
		return nil, err
	}
	values, ok := result.([]any)
	if !ok || len(values) != 2 {
		return nil, fmt.Errorf("unexpected refresh token consume result")
	}
	status, ok := values[0].(int64)
	if !ok {
		return nil, fmt.Errorf("unexpected refresh token consume status")
	}
	encoded, ok := values[1].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected refresh token consume payload")
	}
	if status == refreshConsumeMissing {
		return nil, service.ErrRefreshTokenNotFound
	}
	if status != refreshConsumeWon && status != refreshConsumeReplay {
		return nil, fmt.Errorf("unexpected refresh token consume status %d", status)
	}
	var data service.RefreshTokenData
	if err := json.Unmarshal([]byte(encoded), &data); err != nil {
		return nil, fmt.Errorf("%w: unmarshal consumed refresh token data: %v", service.ErrRefreshTokenCorrupt, err)
	}
	// A legacy record may already carry consumed=true inside the JSON.
	// Newer records keep that flag false and use the sibling key instead.
	data.Consumed = status == refreshConsumeReplay || data.Consumed
	return &data, nil
}

func (c *refreshTokenCache) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	return c.rdb.Del(ctx, refreshTokenKey(tokenHash), refreshTokenConsumedKey(tokenHash)).Err()
}

func (c *refreshTokenCache) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	script := refreshTokenFieldReader + `
local hashes = redis.call('SMEMBERS', KEYS[1])
for _, hash in ipairs(hashes) do
  local token_key = ARGV[1] .. hash
  local value = redis.call('GET', token_key)
  if value then
    local family_id = json_string(value, ARGV[4])
    if family_id then
      redis.call('SREM', ARGV[2] .. family_id, hash)
    end
  end
  redis.call('DEL', token_key, ARGV[3] .. hash)
end
redis.call('DEL', KEYS[1])
return #hashes`
	return c.rdb.Eval(ctx, script, []string{userRefreshTokensKey(userID)}, refreshTokenKeyPrefix, tokenFamilyPrefix, refreshTokenConsumedPrefix, refreshJSONFamilyID).Err()
}

func (c *refreshTokenCache) DeleteTokenFamily(ctx context.Context, familyID string) error {
	return c.RevokeTokenFamily(ctx, familyID, 366*24*time.Hour)
}

// RevokeTokenFamily atomically marks a family revoked and deletes every token
// currently indexed in it. StoreRefreshToken checks the marker in the same
// Redis script, so a concurrent replacement is either deleted or rejected.
func (c *refreshTokenCache) RevokeTokenFamily(ctx context.Context, familyID string, ttl time.Duration) error {
	if strings.TrimSpace(familyID) == "" {
		return nil
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	script := refreshTokenFieldReader + `
local hashes = redis.call('SMEMBERS', KEYS[1])
redis.call('PSETEX', KEYS[2], ARGV[1], '1')
for _, hash in ipairs(hashes) do
  local token_key = ARGV[2] .. hash
  local value = redis.call('GET', token_key)
  if value then
    local user_id = json_integer(value, ARGV[5])
    if user_id then
      redis.call('SREM', ARGV[3] .. user_id, hash)
    end
  end
  redis.call('DEL', token_key, ARGV[4] .. hash)
end
redis.call('DEL', KEYS[1])
return #hashes`
	return c.rdb.Eval(ctx, script, []string{
		tokenFamilyKey(familyID),
		revokedTokenFamilyKey(familyID),
	}, ttl.Milliseconds(), refreshTokenKeyPrefix, userRefreshTokensPrefix, refreshTokenConsumedPrefix, refreshJSONUserID).Err()
}

func (c *refreshTokenCache) AddToUserTokenSet(ctx context.Context, userID int64, tokenHash string, ttl time.Duration) error {
	key := userRefreshTokensKey(userID)
	pipe := c.rdb.Pipeline()
	pipe.SAdd(ctx, key, tokenHash)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *refreshTokenCache) AddToFamilyTokenSet(ctx context.Context, familyID string, tokenHash string, ttl time.Duration) error {
	key := tokenFamilyKey(familyID)
	pipe := c.rdb.Pipeline()
	pipe.SAdd(ctx, key, tokenHash)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *refreshTokenCache) GetUserTokenHashes(ctx context.Context, userID int64) ([]string, error) {
	key := userRefreshTokensKey(userID)
	return c.rdb.SMembers(ctx, key).Result()
}

func (c *refreshTokenCache) GetFamilyTokenHashes(ctx context.Context, familyID string) ([]string, error) {
	key := tokenFamilyKey(familyID)
	return c.rdb.SMembers(ctx, key).Result()
}

func (c *refreshTokenCache) IsTokenInFamily(ctx context.Context, familyID string, tokenHash string) (bool, error) {
	key := tokenFamilyKey(familyID)
	return c.rdb.SIsMember(ctx, key, tokenHash).Result()
}

func (c *refreshTokenCache) RevokeAccessTokens(ctx context.Context, userID int64, revokedAt time.Time, ttl time.Duration) error {
	return c.rdb.Set(ctx, accessTokenRevokedKey(userID), strconv.FormatInt(revokedAt.UnixNano(), 10), ttl).Err()
}

func (c *refreshTokenCache) GetAccessTokensRevokedAt(ctx context.Context, userID int64) (time.Time, error) {
	value, err := c.rdb.Get(ctx, accessTokenRevokedKey(userID)).Result()
	if err == redis.Nil {
		return time.Time{}, service.ErrRefreshTokenNotFound
	}
	if err != nil {
		return time.Time{}, err
	}
	nanos, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse access-token revocation watermark: %w", err)
	}
	return time.Unix(0, nanos), nil
}

func (c *refreshTokenCache) GetSessionGeneration(ctx context.Context, userID int64) (int64, error) {
	value, err := c.rdb.Get(ctx, sessionGenerationKey(userID)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (c *refreshTokenCache) IncrementSessionGeneration(ctx context.Context, userID int64) (int64, error) {
	return c.rdb.Incr(ctx, sessionGenerationKey(userID)).Result()
}
