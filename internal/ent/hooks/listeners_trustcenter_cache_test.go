package hooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestBuildTrustCenterURL(t *testing.T) {
	tests := []struct {
		name          string
		customDomain  string
		slug          string
		defaultDomain string
		scheme        string
		expected      string
	}{
		{
			name:          "custom domain takes precedence",
			customDomain:  "trust.example.com",
			slug:          "my-org",
			defaultDomain: "trust.theopenlane.io",
			expected:      "https://trust.example.com",
		},
		{
			name:          "slug with default domain",
			customDomain:  "",
			slug:          "my-org",
			defaultDomain: "trust.theopenlane.io",
			expected:      "https://trust.theopenlane.io/my-org",
		},
		{
			name:          "empty when no custom domain and no default",
			customDomain:  "",
			slug:          "my-org",
			defaultDomain: "",
			expected:      "",
		},
		{
			name:          "empty when no custom domain and no slug",
			customDomain:  "",
			slug:          "",
			defaultDomain: "trust.theopenlane.io",
			expected:      "",
		},
		{
			name:          "http scheme uses default domain for all requests",
			customDomain:  "custom.example.com",
			slug:          "my-org",
			defaultDomain: "127.0.0.1:12345",
			scheme:        "http",
			expected:      "http://127.0.0.1:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldConfig := trustCenterConfig
			t.Cleanup(func() { SetTrustCenterConfig(oldConfig) })

			SetTrustCenterConfig(TrustCenterConfig{
				DefaultTrustCenterDomain: tt.defaultDomain,
				CacheRefreshScheme:       tt.scheme,
			})

			result := buildTrustCenterURL(tt.customDomain, tt.slug)
			assert.Check(t, is.Equal(tt.expected, result))
		})
	}
}

func TestTriggerCacheRefresh(t *testing.T) {
	initialBackoff, maxBackoff := cacheRefreshInitialBackoff, cacheRefreshMaxBackoff
	cacheRefreshInitialBackoff, cacheRefreshMaxBackoff = time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() { cacheRefreshInitialBackoff, cacheRefreshMaxBackoff = initialBackoff, maxBackoff })

	t.Run("successful request", func(t *testing.T) {
		var receivedFresh string
		var receivedUserAgent string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedFresh = r.URL.Query().Get(cacheRefreshParam)
			receivedUserAgent = r.Header.Get("User-Agent")
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(server.Close)

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.Check(t, err)
		assert.Check(t, is.Equal(cacheRefreshValue, receivedFresh))
		assert.Check(t, is.Equal(cacheRefreshUserAgent, receivedUserAgent))
	})

	t.Run("sends shared secret header when configured", func(t *testing.T) {
		oldConfig := trustCenterConfig
		t.Cleanup(func() { SetTrustCenterConfig(oldConfig) })

		SetTrustCenterConfig(TrustCenterConfig{CacheRefreshSecret: "refresh-secret"})

		var receivedSecret string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedSecret = r.Header.Get(cacheRefreshAuthHeader)
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(server.Close)

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.Check(t, err)
		assert.Check(t, is.Equal("refresh-secret", receivedSecret))
	})

	t.Run("omits shared secret header when not configured", func(t *testing.T) {
		oldConfig := trustCenterConfig
		t.Cleanup(func() { SetTrustCenterConfig(oldConfig) })

		SetTrustCenterConfig(TrustCenterConfig{})

		var secretHeaderPresent bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, secretHeaderPresent = r.Header[cacheRefreshAuthHeader]
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(server.Close)

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.Check(t, err)
		assert.Check(t, !secretHeaderPresent)
	})

	t.Run("4xx error returns immediately without retry", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			w.WriteHeader(http.StatusNotFound)
		}))
		t.Cleanup(server.Close)

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.Check(t, err != nil)
		assert.Check(t, is.ErrorIs(err, ErrCacheRefreshFailed))
		assert.Check(t, is.Equal(int32(1), requestCount.Load()))
	})

	t.Run("5xx error retries", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.Check(t, err != nil)
		assert.Check(t, is.ErrorIs(err, ErrCacheRefreshFailed))
		assert.Check(t, is.Equal(int32(cacheRefreshMaxRetries), requestCount.Load()))
	})

	t.Run("retry succeeds on second attempt", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := requestCount.Add(1)
			if count == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(server.Close)

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.Check(t, err)
		assert.Check(t, is.Equal(int32(2), requestCount.Load()))
	})
}
