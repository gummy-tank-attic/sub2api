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
	refreshTokenKeyPrefix    = "refresh_token:"
	userRefreshTokensPrefix  = "user_refresh_tokens:"
	tokenFamilyPrefix        = "token_family:"
	revokedTokenFamilyPrefix = "revoked_token_family:"
	accessTokenRevokedPrefix = "access_tokens_revoked:"
	sessionGenerationPrefix  = "session_generation:"
)

// refreshTokenKey generates the Redis key for a refresh token.
func refreshTokenKey(tokenHash string) string {
	return refreshTokenKeyPrefix + tokenHash
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
		return nil, fmt.Errorf("unmarshal refresh token data: %w", err)
	}
	return &data, nil
}

// ConsumeRefreshToken atomically changes a live refresh-token record into a
// consumed tombstone. Keeping the tombstone until the original TTL expires
// lets the service identify replay and revoke the whole token family.
func (c *refreshTokenCache) ConsumeRefreshToken(ctx context.Context, tokenHash string) (*service.RefreshTokenData, error) {
	const script = `
local value = redis.call('GET', KEYS[1])
if not value then
  return {0, ''}
end
local data = cjson.decode(value)
if data.consumed == true then
  return {2, value}
end
data.consumed = true
local encoded = cjson.encode(data)
local ttl = redis.call('PTTL', KEYS[1])
if ttl > 0 then
  redis.call('PSETEX', KEYS[1], ttl, encoded)
else
  redis.call('SET', KEYS[1], encoded)
end
return {1, encoded}`

	result, err := c.rdb.Eval(ctx, script, []string{refreshTokenKey(tokenHash)}).Result()
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
	if status == 0 {
		return nil, service.ErrRefreshTokenNotFound
	}
	var data service.RefreshTokenData
	if err := json.Unmarshal([]byte(encoded), &data); err != nil {
		return nil, fmt.Errorf("unmarshal consumed refresh token data: %w", err)
	}
	// The stored tombstone always contains consumed=true. Expose whether this
	// call won the transition (status=1) or observed an earlier consumer
	// (status=2), rather than copying the persisted tombstone flag verbatim.
	data.Consumed = status == 2
	return &data, nil
}

func (c *refreshTokenCache) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	key := refreshTokenKey(tokenHash)
	return c.rdb.Del(ctx, key).Err()
}

func (c *refreshTokenCache) DeleteUserRefreshTokens(ctx context.Context, userID int64) error {
	const script = `
local hashes = redis.call('SMEMBERS', KEYS[1])
for _, hash in ipairs(hashes) do
  local token_key = ARGV[1] .. hash
  local value = redis.call('GET', token_key)
  if value then
    local ok, data = pcall(cjson.decode, value)
    if ok and data.family_id then
      redis.call('SREM', ARGV[2] .. data.family_id, hash)
    end
  end
  redis.call('DEL', token_key)
end
redis.call('DEL', KEYS[1])
return #hashes`
	return c.rdb.Eval(ctx, script, []string{userRefreshTokensKey(userID)}, refreshTokenKeyPrefix, tokenFamilyPrefix).Err()
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
	const script = `
local hashes = redis.call('SMEMBERS', KEYS[1])
redis.call('PSETEX', KEYS[2], ARGV[1], '1')
for _, hash in ipairs(hashes) do
  local token_key = ARGV[2] .. hash
  local value = redis.call('GET', token_key)
  if value then
    local ok, data = pcall(cjson.decode, value)
    if ok and data.user_id then
      redis.call('SREM', ARGV[3] .. tostring(data.user_id), hash)
    end
  end
  redis.call('DEL', token_key)
end
redis.call('DEL', KEYS[1])
return #hashes`
	return c.rdb.Eval(ctx, script, []string{
		tokenFamilyKey(familyID),
		revokedTokenFamilyKey(familyID),
	}, ttl.Milliseconds(), refreshTokenKeyPrefix, userRefreshTokensPrefix).Err()
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
