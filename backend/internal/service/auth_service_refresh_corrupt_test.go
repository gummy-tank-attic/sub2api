//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type scriptedRefreshCache struct {
	refreshTokenCacheStub
	getErr     error
	live       *RefreshTokenData
	consumeErr error
	deleted    int
}

func (c *scriptedRefreshCache) GetRefreshToken(context.Context, string) (*RefreshTokenData, error) {
	if c.getErr != nil {
		return nil, c.getErr
	}
	return c.live, nil
}

func (c *scriptedRefreshCache) ConsumeRefreshToken(context.Context, string) (*RefreshTokenData, error) {
	if c.consumeErr != nil {
		return nil, c.consumeErr
	}
	return nil, ErrRefreshTokenNotFound
}

func (c *scriptedRefreshCache) DeleteRefreshToken(context.Context, string) error {
	c.deleted++
	return nil
}

func TestRefreshTokenPairCorruptRecordIsInvalidAndDeleted(t *testing.T) {
	cache := &scriptedRefreshCache{getErr: ErrRefreshTokenCorrupt}
	svc := &AuthService{refreshTokenCache: cache}

	_, err := svc.RefreshTokenPair(context.Background(), "rt_corrupt")
	require.ErrorIs(t, err, ErrRefreshTokenInvalid)
	require.NotErrorIs(t, err, ErrServiceUnavailable)
	require.Equal(t, 1, cache.deleted)
}

func TestRefreshTokenPairRedisOutageStaysUnavailable(t *testing.T) {
	cache := &scriptedRefreshCache{getErr: errors.New("redis down")}
	svc := &AuthService{refreshTokenCache: cache}

	_, err := svc.RefreshTokenPair(context.Background(), "rt_down")
	require.ErrorIs(t, err, ErrServiceUnavailable)
	require.Zero(t, cache.deleted)
}

func TestRefreshTokenPairCorruptConsumeIsInvalidAndDeleted(t *testing.T) {
	cache := &scriptedRefreshCache{
		live: &RefreshTokenData{
			UserID:       1,
			TokenVersion: 0,
			FamilyID:     "family",
			ExpiresAt:    time.Now().Add(time.Hour),
		},
		consumeErr: ErrRefreshTokenCorrupt,
	}
	svc := &AuthService{
		refreshTokenCache: cache,
		userRepo:          &userRepoStub{user: &User{ID: 1, Status: StatusActive, TokenVersionResolved: true}},
	}

	_, err := svc.RefreshTokenPair(context.Background(), "rt_consume")
	require.ErrorIs(t, err, ErrRefreshTokenInvalid)
	require.NotErrorIs(t, err, ErrServiceUnavailable)
	require.Equal(t, 1, cache.deleted)
}
