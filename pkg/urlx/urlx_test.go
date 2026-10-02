package urlx

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/v2/pkg/shortlinks"
)

func TestTokenURL_AppendsToken(t *testing.T) {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "trustcenter.example.com",
		Path:   "/acme",
	}

	result := TokenURL(baseURL, "test-token-value")

	parsed, err := url.Parse(result)
	require.NoError(t, err)

	assert.Equal(t, "https", parsed.Scheme)
	assert.Equal(t, "trustcenter.example.com", parsed.Host)
	assert.Equal(t, "/acme", parsed.Path)
	assert.Equal(t, "test-token-value", parsed.Query().Get("token"))
}

func TestTokenURL_PreservesSlugPath(t *testing.T) {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "trust.theopenlane.io",
		Path:   "/my-org",
	}

	result := TokenURL(baseURL, "jwt-abc123")

	assert.Equal(t, "https://trust.theopenlane.io/my-org?token=jwt-abc123", result)
}

func TestTokenURL_CustomDomainNoPath(t *testing.T) {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "trust.acme.com",
	}

	result := TokenURL(baseURL, "jwt-xyz")

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

func TestShorten_ReturnsOriginalWithoutClient(t *testing.T) {
	assert.Equal(t, "https://console.example.com/verify?token=abc", Shorten(context.Background(), nil, shortlinks.CreateRequest{URL: "https://console.example.com/verify?token=abc"}))
	assert.Empty(t, Shorten(context.Background(), nil, shortlinks.CreateRequest{}))
}

func TestShorten_UsesServiceAndFallsBack(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			_, _ = w.Write([]byte(`{"shortUrl":"https://s.example.com/r/abc"}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, err := shortlinks.NewClient("id", "secret", shortlinks.WithEndpointURL(server.URL))
	require.NoError(t, err)

	req := shortlinks.CreateRequest{URL: "https://www.example.com/welcome"}
	assert.Equal(t, "https://s.example.com/r/abc", Shorten(context.Background(), client, req))
	assert.Equal(t, "https://www.example.com/welcome", Shorten(context.Background(), client, req), "service failure falls back to the original")
	assert.Equal(t, "https://www.example.com/welcome", Shorten(shortlinks.ContextWithoutShortlinks(context.Background()), client, req), "suppressed context never calls the service")
	assert.Equal(t, 2, calls)
}
