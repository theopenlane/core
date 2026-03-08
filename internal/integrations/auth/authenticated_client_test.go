package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestNewAuthenticatedClient_ClonesHeadersAndKeepsToken(t *testing.T) {
	headers := map[string]string{
		"X-Test": "one",
	}

	client := NewAuthenticatedClient(" https://api.example.com/ ", "  token ", headers)
	if client.BaseURL != "https://api.example.com/" {
		t.Fatalf("expected base URL to be trimmed, got %q", client.BaseURL)
	}
	if client.BearerToken != "  token " {
		t.Fatalf("expected token to be preserved, got %q", client.BearerToken)
	}

	headers["X-Test"] = "two"
	if got := client.Headers["X-Test"]; got != "one" {
		t.Fatalf("expected cloned headers, got %q", got)
	}
}

func TestAuthenticatedClientGetJSON_UsesClientState(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/resource" {
			t.Fatalf("expected request path to include base URL prefix, got %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("expected Authorization header, got %q", got)
		}
		if got := r.Header.Get("X-Test"); got != "value" {
			t.Fatalf("expected custom header, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	client := NewAuthenticatedClient(server.URL, "token", map[string]string{"X-Test": "value"})
	var out struct {
		OK bool `json:"ok"`
	}

	if err := client.GetJSON(context.Background(), "/resource", &out); err != nil {
		t.Fatalf("GetJSON error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok response")
	}
}

func TestAuthenticatedClientGetJSON_AbsoluteURLWithoutBase(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("expected Authorization header, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	var out struct {
		OK bool `json:"ok"`
	}

	client := NewAuthenticatedClient("", "token", nil)
	if err := client.GetJSON(context.Background(), server.URL, &out); err != nil {
		t.Fatalf("GetJSON error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok response")
	}
}

func TestAuthenticatedClientGetJSONWithParams(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/v1/items" {
			t.Fatalf("expected path /v1/items, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Fatalf("expected query param limit=5, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	client := NewAuthenticatedClient(server.URL, "token", nil)
	params := url.Values{}
	params.Set("limit", "5")
	var out struct {
		OK bool `json:"ok"`
	}

	if err := client.GetJSONWithParams(context.Background(), "/v1/items", params, &out); err != nil {
		t.Fatalf("GetJSONWithParams error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok response")
	}
}

func TestBuildEndpointURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		path    string
		params  url.Values
		want    string
	}{
		{
			name:    "absolute path preserved",
			baseURL: "https://api.example.com/",
			path:    "https://other.example.com/v1/users",
			want:    "https://other.example.com/v1/users",
		},
		{
			name:    "base with trailing slash and path without leading slash",
			baseURL: "https://api.example.com/",
			path:    "v1/users",
			want:    "https://api.example.com/v1/users",
		},
		{
			name:    "base without trailing slash and path with leading slash",
			baseURL: "https://api.example.com",
			path:    "/v1/users",
			want:    "https://api.example.com/v1/users",
		},
		{
			name:    "query params appended",
			baseURL: "https://api.example.com/",
			path:    "items",
			params:  url.Values{"page": {"2"}, "per_page": {"10"}},
			want:    "https://api.example.com/items?page=2&per_page=10",
		},
		{
			name:    "empty params not appended",
			baseURL: "https://api.example.com/",
			path:    "items",
			params:  url.Values{},
			want:    "https://api.example.com/items",
		},
		{
			name:    "nil base URL keeps relative path",
			baseURL: "",
			path:    "items",
			want:    "items",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := buildEndpointURL(tc.baseURL, tc.path, tc.params)
			if got != tc.want {
				t.Fatalf("buildEndpointURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAuthenticatedClientFromClient(t *testing.T) {
	client := &AuthenticatedClient{BearerToken: "token"}
	if AuthenticatedClientFromClient(types.NewClientInstance(client)) == nil {
		t.Fatalf("expected client to be unwrapped")
	}
	if AuthenticatedClientFromClient(types.NewClientInstance("not a client")) != nil {
		t.Fatalf("expected nil for non-client")
	}
}
