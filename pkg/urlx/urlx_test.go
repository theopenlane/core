package urlx

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/tokens"
)

func TestBuildTokenURL_AppendsToken(t *testing.T) {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "trustcenter.example.com",
		Path:   "/acme",
	}

	result, err := BuildTokenURL(context.Background(), nil, baseURL, "test-token-value")
	require.NoError(t, err)

	parsed, err := url.Parse(result)
	require.NoError(t, err)

	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "trustcenter.example.com", parsed.Host)
	assert.Equal(t, "/acme", parsed.Path)
	assert.Equal(t, "test-token-value", parsed.Query().Get("token"))
}

func TestBuildTokenURL_PreservesSlugPath(t *testing.T) {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "trust.theopenlane.io",
		Path:   "/my-org",
	}

	result, err := BuildTokenURL(context.Background(), nil, baseURL, "jwt-abc123")
	require.NoError(t, err)

	assert.Equal(t, "https://trust.theopenlane.io/my-org?token=jwt-abc123", result)
}

func TestBuildTokenURL_CustomDomainNoPath(t *testing.T) {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "trust.acme.com",
	}

	result, err := BuildTokenURL(context.Background(), nil, baseURL, "jwt-xyz")
	require.NoError(t, err)

	assert.Equal(t, "https://trust.acme.com?token=jwt-xyz", result)
}

func testTokenManager(t *testing.T) *tokens.TokenManager {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	tm, err := tokens.NewWithKey(key, tokens.Config{
		Audience:        "https://api.example.com",
		Issuer:          "https://auth.example.com",
		AccessDuration:  time.Hour,
		RefreshDuration: 24 * time.Hour,
		RefreshOverlap:  -15 * time.Minute,
	})
	require.NoError(t, err)

	return tm
}

func TestGenerateAnonTokenURL_DefaultDomainWithSlug(t *testing.T) {
	tm := testTokenManager(t)

	baseURL := url.URL{
		Scheme: "https",
		Host:   "trust.theopenlane.io",
		Path:   "/my-org",
	}

	result, err := GenerateAnonTokenURL(context.Background(), tm, nil, baseURL, AnonTokenRequest{
		Prefix:    "anon-tc-",
		SubjectID: "req-123",
		OrgID:     "org-456",
		Email:     "user@example.com",
		Duration:  time.Hour,
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.AccessToken)

	parsed, err := url.Parse(result.URL)
	require.NoError(t, err)

	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "trust.theopenlane.io", parsed.Host)
	assert.Equal(t, "/my-org", parsed.Path)
	assert.NotEmpty(t, parsed.Query().Get("token"))
}

func TestGenerateAnonTokenURL_CustomDomainNoSlug(t *testing.T) {
	tm := testTokenManager(t)

	baseURL := url.URL{
		Scheme: "https",
		Host:   "trust.acme.com",
	}

	result, err := GenerateAnonTokenURL(context.Background(), tm, nil, baseURL, AnonTokenRequest{
		Prefix:    "anon-tc-",
		SubjectID: "req-789",
		OrgID:     "org-456",
		Email:     "user@example.com",
		Duration:  time.Hour,
	})
	require.NoError(t, err)

	parsed, err := url.Parse(result.URL)
	require.NoError(t, err)

	assert.Equal(t, "trust.acme.com", parsed.Host)
	assert.Empty(t, parsed.Path)
	assert.NotEmpty(t, parsed.Query().Get("token"))
}
