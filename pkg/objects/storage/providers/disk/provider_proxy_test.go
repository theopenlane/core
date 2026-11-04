package disk

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
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
	assert.NilError(t, err)
	assert.Assert(t, provider != nil)

	file := &storagetypes.File{
		ID: "01HZZTESTFILEDISK0000",
		FileMetadata: storagetypes.FileMetadata{
			Bucket: "test-bucket",
			Key:    "object.txt",
		},
	}

	url, err := provider.GetPresignedURL(context.Background(), file, &storagetypes.PresignedURLOptions{Duration: time.Minute})
	assert.NilError(t, err, "with proxy presign enabled but no ent client, disk provider falls back to LocalURL-based presigning")
	assert.Assert(t, len(url) > 0)
	assert.Assert(t, strings.HasPrefix(url, "http://localhost:8080/"), "expected LocalURL-based URL, got %q", url)
	assert.Assert(t, is.Contains(url, "object.txt"))
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
