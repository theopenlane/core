package shortlinks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestCreate_SendsMetadataAndOmitsEmptyFields(t *testing.T) {
	var received map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Check(t, is.Equal("client-id", r.Header.Get(headerAccessClientID)))
		assert.Check(t, is.Equal("client-secret", r.Header.Get(headerAccessClientSecret)))
		assert.NilError(t, json.NewDecoder(r.Body).Decode(&received))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"shortUrl":"https://s.example.com/r/abc1234"}`))
	}))
	defer server.Close()

	client, err := NewClient("client-id", "client-secret", WithEndpointURL(server.URL))
	assert.NilError(t, err)

	shortURL, err := client.Create(context.Background(), CreateRequest{
		URL:  "https://console.example.com/invite?token=abc",
		Slug: "  custom ",
		Metadata: Metadata{
			OrganizationID: "org-123",
			Purpose:        PurposeWelcome,
		},
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal("https://s.example.com/r/abc1234", shortURL))

	assert.Check(t, is.Equal("https://console.example.com/invite?token=abc", received["url"]))
	assert.Check(t, is.Equal("custom", received["slug"]))
	assert.Check(t, is.DeepEqual(map[string]any{
		"organization_id": "org-123",
		"purpose":         "welcome",
	}, received["metadata"]), "empty metadata fields are omitted from the wire shape")
	_, hasComment := received["comment"]
	assert.Check(t, !hasComment)
	_, hasExpiration := received["expiration"]
	assert.Check(t, !hasExpiration, "no default TTL means no expiration")
}

func TestCreate_AppliesDefaultLinkTTL(t *testing.T) {
	var received map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NilError(t, json.NewDecoder(r.Body).Decode(&received))
		_, _ = w.Write([]byte(`{"shortUrl":"https://s.example.com/r/ttl"}`))
	}))
	defer server.Close()

	client, err := NewClient("client-id", "client-secret", WithEndpointURL(server.URL), WithLinkTTL(time.Hour))
	assert.NilError(t, err)

	before := time.Now().Add(time.Hour).Unix()
	_, err = client.Create(context.Background(), CreateRequest{URL: "https://console.example.com"})
	assert.NilError(t, err)
	assert.Check(t, int64(received["expiration"].(float64)) >= before)

	_, err = client.Create(context.Background(), CreateRequest{URL: "https://console.example.com", Expiration: 42})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(float64(42), received["expiration"]), "an explicit expiration wins over the default TTL")
}

func TestDeterministicSlug(t *testing.T) {
	first := DeterministicSlug("welcome", "user@example.com")
	assert.Check(t, is.Equal(first, DeterministicSlug("welcome", "user@example.com")))
	assert.Check(t, first != DeterministicSlug("welcome", "other@example.com"))
	assert.Check(t, first != DeterministicSlug("welcomeuser@example.com"), "parts are delimited, not concatenated")
	assert.Check(t, is.Len(first, deterministicSlugLength))
	assert.Check(t, is.Regexp(`^[a-z2-7]+$`, first))
}

func TestCreate_OmitsMetadataWhenEmpty(t *testing.T) {
	var received map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NilError(t, json.NewDecoder(r.Body).Decode(&received))
		_, _ = w.Write([]byte(`{"shortUrl":"https://s.example.com/r/zzz"}`))
	}))
	defer server.Close()

	client, err := NewClient("client-id", "client-secret", WithEndpointURL(server.URL))
	assert.NilError(t, err)

	_, err = client.Create(context.Background(), CreateRequest{URL: "https://console.example.com"})
	assert.NilError(t, err)
	_, hasMetadata := received["metadata"]
	assert.Check(t, !hasMetadata)
}

func TestCreate_RejectsMissingURL(t *testing.T) {
	client, err := NewClient("client-id", "client-secret")
	assert.NilError(t, err)

	_, err = client.Create(context.Background(), CreateRequest{})
	assert.Check(t, is.ErrorIs(err, ErrMissingURL))
}
