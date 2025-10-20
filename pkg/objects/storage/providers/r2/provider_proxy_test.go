package r2

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	storage "github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/objects/storage/proxy"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/tokens"
)

func TestProvider_GetPresignedURL_Proxy(t *testing.T) {
	baseURL := "http://localhost:17608"
	tm := newTestTokenManager(t)
	cfg := &storage.ProxyPresignConfig{
		TokenManager: tm,
		BaseURL:      baseURL,
	}

	secret := make([]byte, 128)
	_, err := rand.Read(secret)
	require.NoError(t, err)

	file := &storagetypes.File{
		ID: "01HZZTESTFILER2AAAAAAA",
		FileMetadata: storagetypes.FileMetadata{
			Bucket:  "test-bucket",
			Key:     "object.txt",
			FullURI: "r2://test-bucket/object.txt",
		},
	}

	url, err := proxy.GenerateDownloadURLWithSecret(file, secret, time.Minute, cfg)
	require.NoError(t, err)
	require.NotEmpty(t, url)

	require.True(t, strings.HasPrefix(url, baseURL), "URL should start with BaseURL, got %q", url)
	require.Contains(t, url, file.ID, "URL should contain file ID")
	require.Contains(t, url, "token=", "URL should contain token parameter")

	parts := strings.Split(url, "token=")
	require.Len(t, parts, 2)
	tokenParam := parts[1]

	unescaped := strings.ReplaceAll(tokenParam, "%2B", "+")
	unescaped = strings.ReplaceAll(unescaped, "%2F", "/")
	unescaped = strings.ReplaceAll(unescaped, "%3D", "=")

	_, err = base64.RawURLEncoding.DecodeString(unescaped)
	require.NoError(t, err, "token should be valid base64")
}

func TestProvider_GetPresignedURL_FallbackToNative(t *testing.T) {
	baseURL := "http://localhost:17608"
	tm := newTestTokenManager(t)
	cfg := &storage.ProxyPresignConfig{
		TokenManager: tm,
		BaseURL:      baseURL,
	}
	opts := storage.NewProviderOptions(
		storage.WithBucket("test-bucket"),
		storage.WithProxyPresignEnabled(true),
		storage.WithProxyPresignConfig(cfg),
		storage.WithCredentials(storage.ProviderCredentials{
			AccountID:       "test-account",
			AccessKeyID:     "test-key",
			SecretAccessKey: "test-secret",
		}),
	)

	provider, err := NewR2Provider(opts)
	require.NoError(t, err)
	require.NotNil(t, provider)

	url, err := provider.GetPresignedURL(t.Context(), &storagetypes.File{
		ID: "01HZZTESTFILER2AAAAAAA",
		FileMetadata: storagetypes.FileMetadata{
			Bucket: "test-bucket",
			Key:    "object.txt",
		},
	}, &storagetypes.PresignedURLOptions{Duration: time.Minute})

	require.NoError(t, err, "with proxy presign enabled but no ent client, should fall back to native R2 presigning")
	require.NotEmpty(t, url)
	require.Contains(t, url, "X-Amz-Algorithm=AWS4-HMAC-SHA256")
}

func newTestTokenManager(t *testing.T) *tokens.TokenManager {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	tm, err := tokens.NewWithKey(key, tokens.Config{})
	require.NoError(t, err)
	return tm
}
