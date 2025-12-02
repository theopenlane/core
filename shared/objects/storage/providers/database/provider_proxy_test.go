package database

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/tokens"
	storage "github.com/theopenlane/shared/objects/storage"
	storagetypes "github.com/theopenlane/shared/objects/storage/types"
)

func TestProvider_GetPresignedURL_Proxy(t *testing.T) {
	tm := newTestTokenManager(t)
	cfg := &storage.ProxyPresignConfig{
		TokenManager: tm,
		BaseURL:      "http://localhost:17608",
	}
	opts := storage.NewProviderOptions(
		storage.WithBucket("test-bucket"),
		storage.WithProxyPresignEnabled(true),
		storage.WithProxyPresignConfig(cfg),
	)

	provider := &Provider{
		options:     opts,
		proxyConfig: cfg,
	}

	file := &storagetypes.File{
		ID: "01HZZTESTFILEDB000000",
		FileMetadata: storagetypes.FileMetadata{
			Bucket: "test-bucket",
			Key:    "object.txt",
		},
	}

	url, err := provider.GetPresignedURL(context.Background(), file, &storagetypes.PresignedURLOptions{Duration: time.Minute})
	require.Error(t, err, "database provider with proxy presign enabled requires ent client in context for presigned URLs")
	require.Empty(t, url)
}

func newTestTokenManager(t *testing.T) *tokens.TokenManager {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	conf := tokens.Config{
		Audience:        "http://localhost:17608",
		Issuer:          "http://localhost:17608",
		AccessDuration:  time.Hour,
		RefreshDuration: 2 * time.Hour,
		RefreshOverlap:  -15 * time.Minute,
	}

	tm, err := tokens.NewWithKey(key, conf)
	require.NoError(t, err)
	return tm
}
