package database

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/tokens"
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
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	tm, err := tokens.NewWithKey(key, tokens.Config{})
	require.NoError(t, err)
	return tm
}
