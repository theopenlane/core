package testutils

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"

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
		MaintNotificationsConfig: &maintnotifications.Config{ // compatibility with go-redis v9.16.1
			Mode: maintnotifications.ModeDisabled,
		},
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
	if refreshOverlap >= 0 {
		refreshOverlap = -15 * time.Minute
	}

	conf := tokens.Config{
		Audience:        "http://localhost:17608",
		Issuer:          "http://localhost:17608",
		AccessDuration:  1 * time.Hour, //nolint:mnd
		RefreshDuration: 2 * time.Hour, //nolint:mnd
		RefreshOverlap:  refreshOverlap,
	}

	if -refreshOverlap >= conf.AccessDuration {
		refreshOverlap = -(conf.AccessDuration - time.Second)
		conf.RefreshOverlap = refreshOverlap
	}

	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	tm, err := tokens.NewWithKey(key, conf)
	if err != nil {
		return nil, fmt.Errorf("new token manager: %w", err)
	}

	return tm, nil
}
