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

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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
	assert.NilError(t, err)

	assert.Check(t, is.Equal("https", parsed.Scheme))
	assert.Check(t, is.Equal("trustcenter.example.com", parsed.Host))
	assert.Check(t, is.Equal("/acme", parsed.Path))
	assert.Check(t, is.Equal("test-token-value", parsed.Query().Get("token")))
}

func TestTokenURL_PreservesSlugPath(t *testing.T) {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "trust.theopenlane.io",
		Path:   "/my-org",
	}

	result := TokenURL(baseURL, "jwt-abc123")

	assert.Check(t, is.Equal("https://trust.theopenlane.io/my-org?token=jwt-abc123", result))
}

func TestTokenURL_CustomDomainNoPath(t *testing.T) {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "trust.acme.com",
	}

	result := TokenURL(baseURL, "jwt-xyz")

	assert.Check(t, is.Equal("https://trust.acme.com?token=jwt-xyz", result))
}

func testTokenManager(t *testing.T) *tokens.TokenManager {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	assert.NilError(t, err)

	tm, err := tokens.NewWithKey(key, tokens.Config{
		Audience:        "https://api.example.com",
		Issuer:          "https://auth.example.com",
		AccessDuration:  time.Hour,
		RefreshDuration: 24 * time.Hour,
		RefreshOverlap:  -15 * time.Minute,
	})
	assert.NilError(t, err)

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
	assert.NilError(t, err)
	assert.Assert(t, result.AccessToken != "")

	parsed, err := url.Parse(result.URL)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("https", parsed.Scheme))
	assert.Check(t, is.Equal("trust.theopenlane.io", parsed.Host))
	assert.Check(t, is.Equal("/my-org", parsed.Path))
	assert.Check(t, parsed.Query().Get("token") != "")
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
	assert.NilError(t, err)

	parsed, err := url.Parse(result.URL)
	assert.NilError(t, err)

	assert.Check(t, is.Equal("trust.acme.com", parsed.Host))
	assert.Check(t, is.Len(parsed.Path, 0))
	assert.Check(t, parsed.Query().Get("token") != "")
}

func TestShorten_ReturnsOriginalWithoutClient(t *testing.T) {
	assert.Check(t, is.Equal("https://console.example.com/verify?token=abc", Shorten(context.Background(), nil, shortlinks.CreateRequest{URL: "https://console.example.com/verify?token=abc"})))
	assert.Check(t, is.Len(Shorten(context.Background(), nil, shortlinks.CreateRequest{}), 0))
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
	assert.NilError(t, err)

	req := shortlinks.CreateRequest{URL: "https://www.example.com/welcome"}
	assert.Check(t, is.Equal("https://s.example.com/r/abc", Shorten(context.Background(), client, req)))
	assert.Check(t, is.Equal("https://www.example.com/welcome", Shorten(context.Background(), client, req)), "service failure falls back to the original")
	assert.Check(t, is.Equal("https://www.example.com/welcome", Shorten(shortlinks.ContextWithoutShortlinks(context.Background()), client, req)), "suppressed context never calls the service")
	assert.Check(t, is.Equal(2, calls))
}
