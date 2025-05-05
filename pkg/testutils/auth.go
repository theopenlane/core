package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"io"
	"log"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
)

// NewRedisClient creates a new redis client for testing using miniredis
func NewRedisClient() *redis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:             mr.Addr(),
		DisableIndentity: true, // # spellcheck:off
	})

	return client
}

// CreateSessionManager creates a new session manager for testing
func CreateSessionManager() sessions.Store[map[string]any] {
	hashKey := randomString(32)  //nolint:mnd
	blockKey := randomString(32) //nolint:mnd

	sm := sessions.NewCookieStore[map[string]any](sessions.DebugCookieConfig,
		hashKey, blockKey,
	)

	return sm
}

// randomString generates a random string of n bytes
func randomString(n int) []byte {
	id := make([]byte, n)

	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		panic(err) // This shouldn't happen
	}

	return id
}

// CreateTokenManager creates a new token manager for testing
func CreateTokenManager(refreshOverlap time.Duration) (*tokens.TokenManager, error) {
	conf := tokens.Config{
		Audience:        "http://localhost:17608",
		Issuer:          "http://localhost:17608",
		AccessDuration:  1 * time.Hour, //nolint:mnd
		RefreshDuration: 2 * time.Hour, //nolint:mnd
		RefreshOverlap:  refreshOverlap,
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048) //nolint:mnd
	if err != nil {
		return nil, err
	}

	return tokens.NewWithKey(key, conf)
}
