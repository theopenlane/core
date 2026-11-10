package s3

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/objects/storage/proxy"
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
	assert.NilError(t, err)

	file := &storagetypes.File{
		ID:      "01HZZTESTFILE00000000000000",
		Bucket:  "test-bucket",
		Key:     "path/to/object.txt",
		FullURI: "s3://test-bucket/path/to/object.txt",
	}

	url, err := proxy.GenerateDownloadURLWithSecret(file, secret, time.Minute, cfg)
	assert.NilError(t, err)
	assert.Assert(t, len(url) > 0)

	assert.Assert(t, strings.HasPrefix(url, baseURL), "URL should start with BaseURL, got %q", url)
	assert.Assert(t, is.Contains(url, file.ID), "URL should contain file ID")
	assert.Assert(t, is.Contains(url, "token="), "URL should contain token parameter")

	parts := strings.Split(url, "token=")
	assert.Assert(t, is.Len(parts, 2))
	tokenParam := parts[1]

	unescaped := strings.ReplaceAll(tokenParam, "%2B", "+")
	unescaped = strings.ReplaceAll(unescaped, "%2F", "/")
	unescaped = strings.ReplaceAll(unescaped, "%3D", "=")

	_, err = base64.RawURLEncoding.DecodeString(unescaped)
	assert.NilError(t, err, "token should be valid base64")
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
		storage.WithRegion("us-east-1"),
		storage.WithProxyPresignEnabled(true),
		storage.WithProxyPresignConfig(cfg),
		storage.WithCredentials(storage.ProviderCredentials{
			AccessKeyID:     "test-key",
			SecretAccessKey: "test-secret",
		}),
	)

	provider, err := NewS3Provider(opts)
	assert.NilError(t, err)
	assert.Assert(t, provider != nil)

	url, err := provider.GetPresignedURL(t.Context(), &storagetypes.File{
		ID:     "01HZZTESTFILE00000000000000",
		Bucket: "test-bucket",
		Key:    "path/to/object.txt",
	}, &storagetypes.PresignedURLOptions{Duration: time.Minute})

	assert.NilError(t, err, "with proxy presign enabled but no ent client, should fall back to native S3 presigning")
	assert.Assert(t, len(url) > 0)
	assert.Assert(t, is.Contains(url, "X-Amz-Algorithm=AWS4-HMAC-SHA256"))
}

func newTestTokenManager(t *testing.T) *tokens.TokenManager {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	assert.NilError(t, err)

	conf := tokens.Config{
		Audience:        "http://localhost:17608",
		Issuer:          "http://localhost:17608",
		AccessDuration:  time.Hour,
		RefreshDuration: 2 * time.Hour,
		RefreshOverlap:  -15 * time.Minute,
	}

	tm, err := tokens.NewWithKey(key, conf)
	assert.NilError(t, err)
	return tm
}
