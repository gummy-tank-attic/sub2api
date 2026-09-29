package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// wideTokenVersion is the first integer float64 cannot represent exactly.
func wideTokenVersion() int64 {
	return (int64(1) << 53) + 1
}

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

func TestConsumeRefreshTokenPreservesWideTokenVersion(t *testing.T) {
	cache, mr := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()
	version := wideTokenVersion()
	generation, err := cache.IncrementSessionGeneration(ctx, 7)
	require.NoError(t, err)
	data := testRefreshTokenData(7, "family-big", generation)
	data.TokenVersion = version
	require.NoError(t, cache.StoreRefreshToken(ctx, "big-parent", data, time.Hour))

	before, err := cache.rdb.Get(ctx, refreshTokenKey("big-parent")).Result()
	require.NoError(t, err)

	first, err := cache.ConsumeRefreshToken(ctx, "big-parent")
	require.NoError(t, err)
	require.False(t, first.Consumed)
	require.Equal(t, version, first.TokenVersion)
	require.Equal(t, int64(7), first.UserID)
	require.Equal(t, generation, first.SessionGeneration)

	after, err := cache.rdb.Get(ctx, refreshTokenKey("big-parent")).Result()
	require.NoError(t, err)
	require.Equal(t, before, after)
	require.Contains(t, after, strconv.FormatInt(version, 10))
	require.NotContains(t, after, "e+")
	require.True(t, mr.Exists(refreshTokenConsumedKey("big-parent")))

	replay, err := cache.ConsumeRefreshToken(ctx, "big-parent")
	require.NoError(t, err)
	require.True(t, replay.Consumed)
	require.Equal(t, version, replay.TokenVersion)

	require.NoError(t, cache.RevokeTokenFamily(ctx, "family-big", time.Hour))
	require.False(t, mr.Exists(refreshTokenKey("big-parent")))
	require.False(t, mr.Exists(refreshTokenConsumedKey("big-parent")))
}

func TestRefreshTokenReadsScientificNotationWithoutDroppingIdentity(t *testing.T) {
	cache, mr := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()
	version := wideTokenVersion()
	payload := fmt.Sprintf("{\n  \"consumed\": true,\n  \"user_id\": 7,\n  \"token_version\": %s,\n  \"session_generation\": 4,\n  \"family_id\": \"family-sci\"\n}", strconv.FormatFloat(float64(version), 'e', -1, 64))
	require.NoError(t, mr.Set(refreshTokenKey("sci"), payload))

	got, err := cache.GetRefreshToken(ctx, "sci")
	require.NoError(t, err)
	require.Equal(t, int64(7), got.UserID)
	require.Equal(t, "family-sci", got.FamilyID)
	require.Equal(t, int64(4), got.SessionGeneration)
	require.NotEqual(t, version, got.TokenVersion)

	consumed, err := cache.ConsumeRefreshToken(ctx, "sci")
	require.NoError(t, err)
	require.True(t, consumed.Consumed)
	require.Equal(t, "family-sci", consumed.FamilyID)
	require.Equal(t, got.TokenVersion, consumed.TokenVersion)
	raw, err := cache.rdb.Get(ctx, refreshTokenKey("sci")).Result()
	require.NoError(t, err)
	require.Equal(t, payload, raw)
}

func TestRefreshTokenCorruptPayloadIsMarkedCorrupt(t *testing.T) {
	cache, mr := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()
	require.NoError(t, mr.Set(refreshTokenKey("bad"), "not-json"))

	_, err := cache.GetRefreshToken(ctx, "bad")
	require.ErrorIs(t, err, service.ErrRefreshTokenCorrupt)

	_, err = cache.ConsumeRefreshToken(ctx, "bad")
	require.ErrorIs(t, err, service.ErrRefreshTokenCorrupt)
	raw, getErr := cache.rdb.Get(ctx, refreshTokenKey("bad")).Result()
	require.NoError(t, getErr)
	require.Equal(t, "not-json", raw)
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

func TestRefreshTokenIndexCleanupReadsSpacedJSON(t *testing.T) {
	cache, mr := newRefreshTokenCacheSecurityTest(t)
	ctx := context.Background()
	version := wideTokenVersion()
	hash := "spaced-hash"
	payload := fmt.Sprintf("{\n  \"user_id\" : %d,\n  \"token_version\" : %d,\n  \"family_id\" : \"family-spaced\"\n}", int64(7), version)
	require.NoError(t, mr.Set(refreshTokenKey(hash), payload))
	require.NoError(t, mr.Set(refreshTokenConsumedKey(hash), "1"))
	require.NoError(t, cache.rdb.SAdd(ctx, userRefreshTokensKey(7), hash).Err())
	require.NoError(t, cache.rdb.SAdd(ctx, tokenFamilyKey("family-spaced"), hash).Err())

	require.NoError(t, cache.RevokeTokenFamily(ctx, "family-spaced", time.Hour))
	require.False(t, mr.Exists(refreshTokenKey(hash)))
	require.False(t, mr.Exists(refreshTokenConsumedKey(hash)))
	userMembers, err := cache.rdb.SMembers(ctx, userRefreshTokensKey(7)).Result()
	require.NoError(t, err)
	require.NotContains(t, userMembers, hash)

	otherHash := "spaced-user"
	require.NoError(t, mr.Set(refreshTokenKey(otherHash), payload))
	require.NoError(t, mr.Set(refreshTokenConsumedKey(otherHash), "1"))
	require.NoError(t, cache.rdb.SAdd(ctx, userRefreshTokensKey(7), otherHash).Err())
	require.NoError(t, cache.rdb.SAdd(ctx, tokenFamilyKey("family-spaced"), otherHash).Err())
	require.NoError(t, cache.DeleteUserRefreshTokens(ctx, 7))
	require.False(t, mr.Exists(refreshTokenKey(otherHash)))
	require.False(t, mr.Exists(refreshTokenConsumedKey(otherHash)))
	familyMembers, err := cache.rdb.SMembers(ctx, tokenFamilyKey("family-spaced")).Result()
	require.NoError(t, err)
	require.NotContains(t, familyMembers, otherHash)
}

func TestRefreshTokenJSONFieldsMatchStructTags(t *testing.T) {
	typ := reflect.TypeOf(service.RefreshTokenData{})
	family, ok := typ.FieldByName("FamilyID")
	require.True(t, ok)
	user, ok := typ.FieldByName("UserID")
	require.True(t, ok)
	require.Equal(t, refreshJSONFamilyID, strings.Split(family.Tag.Get("json"), ",")[0])
	require.Equal(t, refreshJSONUserID, strings.Split(user.Tag.Get("json"), ",")[0])
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
