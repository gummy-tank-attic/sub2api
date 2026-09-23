package service

import (
	"context"
	"errors"
	"time"
)

// ErrRefreshTokenNotFound is returned when a refresh token is not found in cache.
// This is used to abstract away the underlying cache implementation (e.g., redis.Nil).
var ErrRefreshTokenNotFound = errors.New("refresh token not found")

// RefreshTokenData 存储在Redis中的Refresh Token数据
type RefreshTokenData struct {
	UserID            int64     `json:"user_id"`
	TokenVersion      int64     `json:"token_version"` // 用于检测密码更改后的Token失效
	SessionGeneration int64     `json:"session_generation"`
	FamilyID          string    `json:"family_id"`              // Token家族ID，用于防重放攻击
	BindingHash       string    `json:"binding_hash,omitempty"` // 会话指纹哈希（IP+UA），会话绑定开启时校验
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	Consumed          bool      `json:"consumed,omitempty"`
}

// RefreshTokenCache 管理Refresh Token的Redis缓存
// 用于JWT Token刷新机制，支持Token轮转和防重放攻击
//
// Key 格式:
//   - refresh_token:{token_hash}     -> RefreshTokenData (JSON)
//   - user_refresh_tokens:{user_id}  -> Set<token_hash>
//   - token_family:{family_id}       -> Set<token_hash>
type RefreshTokenCache interface {
	// StoreRefreshToken 存储Refresh Token
	// tokenHash: Token的SHA256哈希值（不存储原始Token）
	// data: Token关联的数据
	// ttl: Token过期时间
	StoreRefreshToken(ctx context.Context, tokenHash string, data *RefreshTokenData, ttl time.Duration) error

	// GetRefreshToken 获取Refresh Token数据
	// 返回 (data, nil) 如果Token存在
	// 返回 (nil, ErrRefreshTokenNotFound) 如果Token不存在
	// 返回 (nil, err) 如果发生其他错误
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenData, error)

	// ConsumeRefreshToken atomically marks a refresh token as consumed and
	// returns its metadata. A consumed token must not be exchanged again.
	ConsumeRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenData, error)

	// DeleteRefreshToken 删除单个Refresh Token
	// 用于Token轮转时使旧Token失效
	DeleteRefreshToken(ctx context.Context, tokenHash string) error

	// DeleteUserRefreshTokens 删除用户的所有Refresh Token
	// 用于密码更改或用户主动登出所有设备
	DeleteUserRefreshTokens(ctx context.Context, userID int64) error

	// DeleteTokenFamily 删除整个Token家族
	// 用于检测到Token重放攻击时，撤销整个会话链
	DeleteTokenFamily(ctx context.Context, familyID string) error

	// AddToUserTokenSet 将Token添加到用户的Token集合
	// 用于跟踪用户的所有活跃Refresh Token
	AddToUserTokenSet(ctx context.Context, userID int64, tokenHash string, ttl time.Duration) error

	// AddToFamilyTokenSet 将Token添加到家族Token集合
	// 用于跟踪同一登录会话的所有Token
	AddToFamilyTokenSet(ctx context.Context, familyID string, tokenHash string, ttl time.Duration) error

	// GetUserTokenHashes 获取用户的所有Token哈希
	// 用于批量删除用户Token
	GetUserTokenHashes(ctx context.Context, userID int64) ([]string, error)

	// GetFamilyTokenHashes 获取家族的所有Token哈希
	// 用于批量删除家族Token
	GetFamilyTokenHashes(ctx context.Context, familyID string) ([]string, error)

	// IsTokenInFamily 检查Token是否属于指定家族
	// 用于验证Token家族关系
	IsTokenInFamily(ctx context.Context, familyID string, tokenHash string) (bool, error)
}

// AccessTokenRevocationStore persists a per-user revocation watermark for
// stateless access tokens. It is separate from refresh-token storage because
// deleting refresh sessions cannot invalidate already-issued JWTs.
type AccessTokenRevocationStore interface {
	RevokeAccessTokens(ctx context.Context, userID int64, revokedAt time.Time, ttl time.Duration) error
	GetAccessTokensRevokedAt(ctx context.Context, userID int64) (time.Time, error)
}

// SessionGenerationStore provides a monotonic per-user session epoch. Access
// and refresh credentials issued under an older epoch remain invalid even when
// they race with revoke-all or escape an index-deletion snapshot.
type SessionGenerationStore interface {
	GetSessionGeneration(ctx context.Context, userID int64) (int64, error)
	IncrementSessionGeneration(ctx context.Context, userID int64) (int64, error)
}

// RefreshTokenFamilyRevocationStore records a family-level revocation marker.
// The marker is checked atomically while storing replacement refresh tokens.
type RefreshTokenFamilyRevocationStore interface {
	RevokeTokenFamily(ctx context.Context, familyID string, ttl time.Duration) error
}
