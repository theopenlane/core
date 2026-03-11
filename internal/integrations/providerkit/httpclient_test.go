package providerkit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestNewAuthenticatedClient_ClonesHeaders(t *testing.T) {
	headers := map[string]string{"X-Test": "one"}

	client := NewAuthenticatedClient("https://api.example.com/", "token", headers)
	if client.BaseURL != "https://api.example.com/" {
		t.Fatalf("expected trimmed base URL, got %q", client.BaseURL)
	}

	headers["X-Test"] = "two"
	if got := client.Headers["X-Test"]; got != "one" {
		t.Fatalf("expected cloned headers, got %q", got)
	}
}

func TestAuthenticatedClientGetJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("expected Authorization header, got %q", got)
		}
		if got := r.URL.Path; got != "/resource" {
			t.Fatalf("expected path /resource, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	client := NewAuthenticatedClient(server.URL, "token", nil)
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

func TestAuthenticatedClientGetJSONWithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Fatalf("expected limit=5, got %q", got)
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

	if err := client.GetJSONWithParams(context.Background(), "/items", params, &out); err != nil {
		t.Fatalf("GetJSONWithParams error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok response")
	}
}

func TestAuthenticatedClientPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %q", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	client := NewAuthenticatedClient(server.URL, "token", nil)
	var out struct {
		OK bool `json:"ok"`
	}

	if err := client.PostJSON(context.Background(), "/resource", map[string]any{"foo": "bar"}, &out); err != nil {
		t.Fatalf("PostJSON error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok response")
	}
}

func TestAuthenticatedClientGetJSON_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	t.Cleanup(server.Close)

	client := NewAuthenticatedClient(server.URL, "bad-token", nil)
	var out map[string]any

	err := client.GetJSON(context.Background(), "/resource", &out)
	if err == nil {
		t.Fatalf("expected error for 401 response")
	}
	if !errors.Is(err, ErrHTTPRequestFailed) {
		t.Fatalf("expected ErrHTTPRequestFailed, got %v", err)
	}

	var httpErr *HTTPRequestError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPRequestError")
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", httpErr.StatusCode)
	}
}

func TestAuthenticatedClientFromClient(t *testing.T) {
	client := &AuthenticatedClient{BearerToken: "token"}

	if AuthenticatedClientFromClient(types.NewClientInstance(client)) == nil {
		t.Fatalf("expected client to be unwrapped")
	}

	if AuthenticatedClientFromClient(types.NewClientInstance("not-a-client")) != nil {
		t.Fatalf("expected nil for non-matching type")
	}
}

func TestResolveAuthenticatedClient_FromPooled(t *testing.T) {
	pooled := &AuthenticatedClient{
		BaseURL:     "https://api.example.com",
		BearerToken: "existing-token",
		Headers:     map[string]string{"X-Custom": "value"},
	}

	input := types.OperationInput{
		Client: types.NewClientInstance(pooled),
	}

	client, err := ResolveAuthenticatedClient(input, nil, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.BearerToken != "existing-token" {
		t.Fatalf("expected pooled token, got %q", client.BearerToken)
	}
}

func TestResolveAuthenticatedClient_FromPooled_AppliesBaseURL(t *testing.T) {
	pooled := &AuthenticatedClient{
		BearerToken: "token",
	}

	input := types.OperationInput{
		Client: types.NewClientInstance(pooled),
	}

	client, err := ResolveAuthenticatedClient(input, nil, "https://api.example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.BaseURL != "https://api.example.com" {
		t.Fatalf("expected base URL to be applied, got %q", client.BaseURL)
	}
}

func TestResolveAuthenticatedClient_FromCredential(t *testing.T) {
	extract := func(types.CredentialSet) (string, error) {
		return "extracted-token", nil
	}

	input := types.OperationInput{}

	client, err := ResolveAuthenticatedClient(input, extract, "https://api.example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.BearerToken != "extracted-token" {
		t.Fatalf("expected extracted token, got %q", client.BearerToken)
	}
	if client.BaseURL != "https://api.example.com" {
		t.Fatalf("expected base URL, got %q", client.BaseURL)
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
			params:  url.Values{"page": {"2"}},
			want:    "https://api.example.com/items?page=2",
		},
		{
			name:    "empty params not appended",
			baseURL: "https://api.example.com/",
			path:    "items",
			params:  url.Values{},
			want:    "https://api.example.com/items",
		},
		{
			name:    "empty base URL keeps relative path",
			baseURL: "",
			path:    "items",
			want:    "items",
		},
		{
			name:    "empty base URL with params",
			baseURL: "",
			path:    "items",
			params:  url.Values{"q": {"foo"}},
			want:    "items?q=foo",
		},
		{
			name:    "absolute path with params",
			baseURL: "https://api.example.com/",
			path:    "https://other.example.com/v1/users",
			params:  url.Values{"page": {"1"}},
			want:    "https://other.example.com/v1/users?page=1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := buildEndpointURL(tc.baseURL, tc.path, tc.params)
			if got != tc.want {
				t.Fatalf("buildEndpointURL(%q, %q, %v) = %q, want %q", tc.baseURL, tc.path, tc.params, got, tc.want)
			}
		})
	}
}
