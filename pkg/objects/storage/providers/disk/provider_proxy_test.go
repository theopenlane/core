package disk

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/pkg/objects/storage"
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
