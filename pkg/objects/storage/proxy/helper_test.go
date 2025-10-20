package proxy

import (
	"crypto/rand"
	"crypto/rsa"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	storage "github.com/theopenlane/core/pkg/objects/storage"
	storagetypes "github.com/theopenlane/core/pkg/objects/storage/types"
	"github.com/theopenlane/iam/tokens"
)

func TestGenerateDownloadURLWithSecret_BaseHandling(t *testing.T) {
	t.Parallel()

	tm := newTestTokenManager(t)

	secret := make([]byte, 128)
	_, err := rand.Read(secret)
	require.NoError(t, err)

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

			downloadURL, err := GenerateDownloadURLWithSecret(file, secret, time.Minute, storagetypes.DatabaseProvider, cfg)
			require.NoError(t, err)
			require.NotEmpty(t, downloadURL)

			expectedPrefix := tc.base + "/" + escapedID + "/download?token="
			require.Truef(t, strings.HasPrefix(downloadURL, expectedPrefix), "expected prefix %q, got %q", expectedPrefix, downloadURL)
			require.Contains(t, downloadURL, "token=")

			parsed, err := url.Parse(downloadURL)
			require.NoError(t, err)
			require.NotEmpty(t, parsed.Query().Get("token"))
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

	require.Equal(t, tm, cfg.TokenManager)
	require.Equal(t, "https://issuer.example.com/oauth2", cfg.TokenIssuer)
	require.Equal(t, "https://aud.example.com", cfg.TokenAudience)
	require.Equal(t, "https://api.example.com/v1/files", cfg.BaseURL)
}

func newTestTokenManager(t *testing.T) *tokens.TokenManager {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	tm, err := tokens.NewWithKey(key, tokens.Config{})
	require.NoError(t, err)

	return tm
}
