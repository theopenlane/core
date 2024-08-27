package httpsling

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createTestServerForProxy creates a simple HTTP server for testing purposes
func createTestServerForProxy() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

// TestSetProxyValidProxy tests setting a valid proxy and making a request through it
func TestSetProxyValidProxy(t *testing.T) {
	server := createTestServerForProxy()

	defer server.Close()

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Indicate the request passed through the proxy
		w.Header().Set("X-Test-Proxy", "true")
		w.WriteHeader(http.StatusOK)
	}))

	defer proxyServer.Close()

	client := URL(server.URL)

	err := client.SetProxy(proxyServer.URL)
	assert.Nil(t, err, "Setting a valid proxy should not result in an error")

	resp, err := client.Get("/").Send(context.Background())
	assert.Nil(t, err, "Request through a valid proxy should succeed")
	assert.NotNil(t, resp, "Response should not be nil")
	assert.Equal(t, "true", resp.Header().Get("X-Test-Proxy"), "Request should have passed through the proxy")
}

// TestSetProxyInvalidProxy tests handling of invalid proxy URLs
func TestSetProxyInvalidProxy(t *testing.T) {
	server := createTestServerForProxy()

	defer server.Close()

	client := URL(server.URL)

	invalidProxyURL := "://invalid_url"
	err := client.SetProxy(invalidProxyURL)
	assert.NotNil(t, err, "Setting an invalid proxy URL should result in an error")
}

// TestSetProxyRemoveProxy tests removing proxy settings
func TestSetProxyRemoveProxy(t *testing.T) {
	server := createTestServerForProxy()

	defer server.Close()

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Proxy server response
		w.WriteHeader(http.StatusOK)
	}))

	defer proxyServer.Close()

	client := URL(server.URL)

	// Set then remove the proxy
	err := client.SetProxy(proxyServer.URL)
	assert.Nil(t, err, "Setting a proxy should not result in an error")

	client.RemoveProxy()

	// Make a request and check it doesn't go through the proxy
	resp, err := client.Get("/").Send(context.Background())
	assert.Nil(t, err, "Request after removing proxy should succeed")
	assert.NotNil(t, resp, "Response should not be ni.")
	assert.NotEqual(t, "true", resp.Header().Get("X-Test-Proxy"), "Request should not have passed through the proxy")
}
