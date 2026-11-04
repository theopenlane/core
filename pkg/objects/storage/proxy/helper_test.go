package proxy

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/url"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/iam/tokens"
)

func TestGenerateDownloadURLWithSecret_BaseHandling(t *testing.T) {
	t.Parallel()

	tm := newTestTokenManager(t)

	secret := make([]byte, 128)
	_, err := rand.Read(secret)
	assert.NilError(t, err)

	file := &storagetypes.File{
		ID: "01HZZTESTFILE00000000000000",
		FileMetadata: storagetypes.FileMetadata{
			Bucket:  "test-bucket",
			Key:     "path/to/object.txt",
			FullURI: "database://test-bucket/path/to/object.txt",
		},
	}

	escapedID := url.PathEscape(file.ID)

	testCases := []struct {
		name string
		base string
	}{
		{
			name: "Localhost",
			base: "http://localhost:17608/v1/files",
		},
		{
			name: "Production",
			base: "https://api.example.com/v1/files",
		},
		{
			name: "TrailingPath",
			base: "https://api.example.com/root/v1/files",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()

			cfg := storage.NewProxyPresignConfig(
				storage.WithProxyPresignTokenManager(tm),
				storage.WithProxyPresignBaseURL(tc.base),
			)

			downloadURL, err := GenerateDownloadURLWithSecret(file, secret, time.Minute, cfg)
			assert.NilError(t, err)
			assert.Assert(t, len(downloadURL) > 0)

			expectedPrefix := tc.base + "/" + escapedID + "/download?token="
			assert.Assert(t, strings.HasPrefix(downloadURL, expectedPrefix), "expected prefix %q, got %q", expectedPrefix, downloadURL)
			assert.Assert(t, is.Contains(downloadURL, "token="))

			parsed, err := url.Parse(downloadURL)
			assert.NilError(t, err)
			assert.Assert(t, len(parsed.Query().Get("token")) > 0)
		})
	}
}

func TestNewProxyPresignConfigOptions(t *testing.T) {
	t.Parallel()

	tm := newTestTokenManager(t)

	cfg := storage.NewProxyPresignConfig(
		storage.WithProxyPresignTokenManager(tm),
		storage.WithProxyPresignTokenIssuer("https://issuer.example.com/oauth2"),
		storage.WithProxyPresignTokenAudience("https://aud.example.com"),
		storage.WithProxyPresignBaseURL("https://api.example.com/v1/files"),
	)

	assert.Assert(t, is.Equal(tm, cfg.TokenManager))
	assert.Assert(t, is.Equal("https://issuer.example.com/oauth2", cfg.TokenIssuer))
	assert.Assert(t, is.Equal("https://aud.example.com", cfg.TokenAudience))
	assert.Assert(t, is.Equal("https://api.example.com/v1/files", cfg.BaseURL))
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
