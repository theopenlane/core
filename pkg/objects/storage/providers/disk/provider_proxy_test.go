package disk

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"strings"
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
		storage.WithLocalURL("http://localhost:8080"),
		storage.WithProxyPresignEnabled(true),
		storage.WithProxyPresignConfig(cfg),
	)

	provider, err := NewDiskProvider(opts)
	require.NoError(t, err)
	require.NotNil(t, provider)

	file := &storagetypes.File{
		ID: "01HZZTESTFILEDISK0000",
		FileMetadata: storagetypes.FileMetadata{
			Bucket: "test-bucket",
			Key:    "object.txt",
		},
	}

	url, err := provider.GetPresignedURL(context.Background(), file, &storagetypes.PresignedURLOptions{Duration: time.Minute})
	require.NoError(t, err, "with proxy presign enabled but no ent client, disk provider falls back to LocalURL-based presigning")
	require.NotEmpty(t, url)
	require.True(t, strings.HasPrefix(url, "http://localhost:8080/"), "expected LocalURL-based URL, got %q", url)
	require.Contains(t, url, "object.txt")
}

func newTestTokenManager(t *testing.T) *tokens.TokenManager {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	tm, err := tokens.NewWithKey(key, tokens.Config{})
	require.NoError(t, err)
	return tm
}
