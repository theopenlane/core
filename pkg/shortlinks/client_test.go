package shortlinks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate_SendsMetadataAndOmitsEmptyFields(t *testing.T) {
	var received map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "client-id", r.Header.Get(headerAccessClientID))
		assert.Equal(t, "client-secret", r.Header.Get(headerAccessClientSecret))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"shortUrl":"https://s.example.com/r/abc1234"}`))
	}))
	defer server.Close()

	client, err := NewClient("client-id", "client-secret", WithEndpointURL(server.URL))
	require.NoError(t, err)

	shortURL, err := client.Create(context.Background(), CreateRequest{
		URL:  "https://console.example.com/invite?token=abc",
		Slug: "  custom ",
		Metadata: Metadata{
			OrganizationID: "org-123",
			Purpose:        PurposeWelcome,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "https://s.example.com/r/abc1234", shortURL)

	assert.Equal(t, "https://console.example.com/invite?token=abc", received["url"])
	assert.Equal(t, "custom", received["slug"])
	assert.Equal(t, map[string]any{
		"organization_id": "org-123",
		"purpose":         "welcome",
	}, received["metadata"], "empty metadata fields are omitted from the wire shape")
	assert.NotContains(t, received, "comment")
	assert.NotContains(t, received, "expiration", "no default TTL means no expiration")
}

func TestCreate_AppliesDefaultLinkTTL(t *testing.T) {
	var received map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		_, _ = w.Write([]byte(`{"shortUrl":"https://s.example.com/r/ttl"}`))
	}))
	defer server.Close()

	client, err := NewClient("client-id", "client-secret", WithEndpointURL(server.URL), WithLinkTTL(time.Hour))
	require.NoError(t, err)

	before := time.Now().Add(time.Hour).Unix()
	_, err = client.Create(context.Background(), CreateRequest{URL: "https://console.example.com"})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, int64(received["expiration"].(float64)), before)

	_, err = client.Create(context.Background(), CreateRequest{URL: "https://console.example.com", Expiration: 42})
	require.NoError(t, err)
	assert.Equal(t, float64(42), received["expiration"], "an explicit expiration wins over the default TTL")
}

func TestDeterministicSlug(t *testing.T) {
	first := DeterministicSlug("welcome", "user@example.com")
	assert.Equal(t, first, DeterministicSlug("welcome", "user@example.com"))
	assert.NotEqual(t, first, DeterministicSlug("welcome", "other@example.com"))
	assert.NotEqual(t, first, DeterministicSlug("welcomeuser@example.com"), "parts are delimited, not concatenated")
	assert.Len(t, first, deterministicSlugLength)
	assert.Regexp(t, `^[a-z2-7]+$`, first)
}

func TestCreate_OmitsMetadataWhenEmpty(t *testing.T) {
	var received map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		_, _ = w.Write([]byte(`{"shortUrl":"https://s.example.com/r/zzz"}`))
	}))
	defer server.Close()

	client, err := NewClient("client-id", "client-secret", WithEndpointURL(server.URL))
	require.NoError(t, err)

	_, err = client.Create(context.Background(), CreateRequest{URL: "https://console.example.com"})
	require.NoError(t, err)
	assert.NotContains(t, received, "metadata")
}

func TestCreate_RejectsMissingURL(t *testing.T) {
	client, err := NewClient("client-id", "client-secret")
	require.NoError(t, err)

	_, err = client.Create(context.Background(), CreateRequest{})
	assert.ErrorIs(t, err, ErrMissingURL)
}
