package hooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
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
			defer func() { trustCenterConfig = oldConfig }()

			trustCenterConfig.DefaultTrustCenterDomain = tt.defaultDomain
			trustCenterConfig.CacheRefreshScheme = tt.scheme

			result := buildTrustCenterURL(tt.customDomain, tt.slug)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTriggerCacheRefresh(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var receivedFresh string
		var receivedUserAgent string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedFresh = r.URL.Query().Get(cacheRefreshParam)
			receivedUserAgent = r.Header.Get("User-Agent")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.NoError(t, err)
		assert.Equal(t, cacheRefreshValue, receivedFresh)
		assert.Equal(t, cacheRefreshUserAgent, receivedUserAgent)
	})

	t.Run("4xx error returns immediately without retry", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrCacheRefreshFailed)
		assert.Equal(t, int32(1), requestCount.Load())
	})

	t.Run("5xx error retries", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrCacheRefreshFailed)
		assert.Equal(t, int32(cacheRefreshMaxRetries), requestCount.Load())
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
		defer server.Close()

		err := triggerCacheRefresh(context.Background(), server.URL)
		assert.NoError(t, err)
		assert.Equal(t, int32(2), requestCount.Load())
	})
}
