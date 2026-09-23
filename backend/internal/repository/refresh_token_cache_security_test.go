package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newRefreshTokenCacheSecurityTest(t *testing.T) (*refreshTokenCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return &refreshTokenCache{rdb: client}, mr
}

func testRefreshTokenData(userID int64, familyID string, generation int64) *service.RefreshTokenData {
	now := time.Now().UTC()
	return &service.RefreshTokenData{
		UserID:            userID,
		TokenVersion:      1,
		SessionGeneration: generation,
		FamilyID:          familyID,
		CreatedAt:         now,
		ExpiresAt:         now.Add(time.Hour),
	}
}

func TestRefreshTokenStoreAtomicallyCreatesPrimaryAndIndexes(t *testing.T) {
	cache, mr := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()

	require.NoError(t, cache.StoreRefreshToken(ctx, "hash-a", testRefreshTokenData(7, "family-a", 0), time.Hour))
	require.True(t, mr.Exists(refreshTokenKey("hash-a")))
	userMembers, err := cache.rdb.SMembers(ctx, userRefreshTokensKey(7)).Result()
	require.NoError(t, err)
	familyMembers, err := cache.rdb.SMembers(ctx, tokenFamilyKey("family-a")).Result()
	require.NoError(t, err)
	require.Contains(t, userMembers, "hash-a")
	require.Contains(t, familyMembers, "hash-a")
}

func TestRefreshTokenStoreRejectsIndexFailureWithoutPrimaryRecord(t *testing.T) {
	cache, mr := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()
	mr.Set(userRefreshTokensKey(7), "wrong-type")

	err := cache.StoreRefreshToken(ctx, "hash-b", testRefreshTokenData(7, "family-b", 0), time.Hour)
	require.Error(t, err)
	require.False(t, mr.Exists(refreshTokenKey("hash-b")))
	require.False(t, mr.Exists(tokenFamilyKey("family-b")))
}

func TestRefreshTokenStoreRejectsStaleGeneration(t *testing.T) {
	cache, mr := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()
	_, err := cache.IncrementSessionGeneration(ctx, 7)
	require.NoError(t, err)

	err = cache.StoreRefreshToken(ctx, "hash-c", testRefreshTokenData(7, "family-c", 0), time.Hour)
	require.ErrorIs(t, err, service.ErrTokenRevoked)
	require.False(t, mr.Exists(refreshTokenKey("hash-c")))
}

func TestConsumeRefreshTokenDistinguishesFirstUseFromReplay(t *testing.T) {
	cache, _ := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()
	require.NoError(t, cache.StoreRefreshToken(ctx, "single-parent", testRefreshTokenData(7, "single-family", 0), time.Hour))

	first, err := cache.ConsumeRefreshToken(ctx, "single-parent")
	require.NoError(t, err)
	require.False(t, first.Consumed)

	replay, err := cache.ConsumeRefreshToken(ctx, "single-parent")
	require.NoError(t, err)
	require.True(t, replay.Consumed)
}

func TestRevokeTokenFamilyClosesBothStoreOrderings(t *testing.T) {
	cache, mr := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()

	require.NoError(t, cache.StoreRefreshToken(ctx, "before", testRefreshTokenData(7, "family-r", 0), time.Hour))
	require.NoError(t, cache.StoreRefreshToken(ctx, "sibling", testRefreshTokenData(7, "family-r", 0), time.Hour))
	require.NoError(t, cache.StoreRefreshToken(ctx, "unrelated", testRefreshTokenData(7, "family-other", 0), time.Hour))
	require.NoError(t, cache.RevokeTokenFamily(ctx, "family-r", time.Hour))
	require.False(t, mr.Exists(refreshTokenKey("before")))
	require.False(t, mr.Exists(refreshTokenKey("sibling")))
	require.True(t, mr.Exists(refreshTokenKey("unrelated")))
	userMembers, err := cache.rdb.SMembers(ctx, userRefreshTokensKey(7)).Result()
	require.NoError(t, err)
	require.NotContains(t, userMembers, "before")
	require.NotContains(t, userMembers, "sibling")
	require.Contains(t, userMembers, "unrelated")

	err = cache.StoreRefreshToken(ctx, "after", testRefreshTokenData(7, "family-r", 0), time.Hour)
	require.ErrorIs(t, err, service.ErrTokenRevoked)
	require.False(t, mr.Exists(refreshTokenKey("after")))
}

func TestConcurrentRefreshReplayCannotLeaveDescendantAlive(t *testing.T) {
	cache, mr := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()
	require.NoError(t, cache.StoreRefreshToken(ctx, "parent", testRefreshTokenData(9, "race-family", 0), time.Hour))

	start := make(chan struct{})
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			data, err := cache.ConsumeRefreshToken(ctx, "parent")
			if err != nil {
				results <- err
				return
			}
			if data.Consumed {
				results <- cache.RevokeTokenFamily(ctx, data.FamilyID, time.Hour)
				return
			}
			err = cache.StoreRefreshToken(ctx, "descendant", testRefreshTokenData(9, data.FamilyID, data.SessionGeneration), time.Hour)
			if err != nil && !errors.Is(err, service.ErrTokenRevoked) {
				results <- err
				return
			}
			results <- nil
		}()
	}
	close(start)
	for i := 0; i < 2; i++ {
		require.NoError(t, <-results)
	}

	require.False(t, mr.Exists(refreshTokenKey("descendant")))
	require.True(t, mr.Exists(revokedTokenFamilyKey("race-family")))
}

func TestConsumeLoginSessionHasExactlyOneWinner(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	cache := NewTotpCache(client)
	ctx := context.Background()
	require.NoError(t, cache.SetLoginSession(ctx, "single-use", &service.TotpLoginSession{UserID: 11}, time.Minute))

	const workers = 8
	var wg sync.WaitGroup
	results := make(chan *service.TotpLoginSession, workers)
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			session, err := cache.ConsumeLoginSession(ctx, "single-use")
			results <- session
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)

	winners := 0
	for err := range errs {
		require.NoError(t, err)
	}
	for session := range results {
		if session != nil {
			winners++
			require.Equal(t, int64(11), session.UserID)
		}
	}
	require.Equal(t, 1, winners)

	_, err := cache.ConsumeLoginSession(ctx, "single-use")
	require.NoError(t, err)
}
