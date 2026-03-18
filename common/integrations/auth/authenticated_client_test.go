package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAuthenticatedClient_ClonesHeadersAndKeepsToken(t *testing.T) {
	headers := map[string]string{
		"X-Test": "one",
	}

	client := NewAuthenticatedClient("  token ", headers)
	if client.BearerToken != "  token " {
		t.Fatalf("expected token to be preserved, got %q", client.BearerToken)
	}

	headers["X-Test"] = "two"
	if got := client.Headers["X-Test"]; got != "one" {
		t.Fatalf("expected cloned headers, got %q", got)
	}
}

func TestGetJSONWithClient_UsesProvidedClient(t *testing.T) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	client := NewAuthenticatedClient("token", map[string]string{"X-Test": "value"})
	var out struct {
		OK bool `json:"ok"`
	}

	if err := GetJSONWithClient(context.Background(), client, server.URL, "ignored", nil, &out); err != nil {
		t.Fatalf("GetJSONWithClient error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok response")
	}
}

func TestGetJSONWithClient_FallsBackToBearer(t *testing.T) {
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

	if err := GetJSONWithClient(context.Background(), nil, server.URL, "token", nil, &out); err != nil {
		t.Fatalf("GetJSONWithClient error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok response")
	}
}

func TestAuthenticatedClientFromAny(t *testing.T) {
	client := &AuthenticatedClient{BearerToken: "token"}
	if AuthenticatedClientFromAny(client) == nil {
		t.Fatalf("expected client to be unwrapped")
	}
	if AuthenticatedClientFromAny("not a client") != nil {
		t.Fatalf("expected nil for non-client")
	}
}
