package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/integrations/types"
)

func TestOAuthTokenFromPayload_Missing(t *testing.T) {
	payload := types.CredentialPayload{}
	if _, err := OAuthTokenFromPayload(payload); !errors.Is(err, ErrOAuthTokenMissing) {
		t.Fatalf("expected ErrOAuthTokenMissing, got %v", err)
	}
}

func TestOAuthTokenFromPayload_EmptyAccess(t *testing.T) {
	payload := types.CredentialPayload{
		Token: &oauth2.Token{AccessToken: ""},
	}
	if _, err := OAuthTokenFromPayload(payload); !errors.Is(err, ErrAccessTokenEmpty) {
		t.Fatalf("expected ErrAccessTokenEmpty, got %v", err)
	}
}

func TestOAuthTokenFromPayload_Success(t *testing.T) {
	payload := types.CredentialPayload{
		Token: &oauth2.Token{AccessToken: "token"},
	}
	token, err := OAuthTokenFromPayload(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "token" {
		t.Fatalf("expected token, got %q", token)
	}
}

func TestAPITokenFromPayload(t *testing.T) {
	payload := types.CredentialPayload{}
	if _, err := APITokenFromPayload(payload); !errors.Is(err, ErrAPITokenMissing) {
		t.Fatalf("expected ErrAPITokenMissing, got %v", err)
	}

	payload.Data.APIToken = "  token "
	token, err := APITokenFromPayload(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "  token " {
		t.Fatalf("expected token to be preserved, got %q", token)
	}
}

func TestRandomState(t *testing.T) {
	state1, err := RandomState(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	state2, err := RandomState(8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state1 == "" || state2 == "" {
		t.Fatalf("expected non-empty state values")
	}
	if state1 == state2 {
		t.Fatalf("expected different state values")
	}
}

func TestHTTPGetJSON(t *testing.T) {
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
	if err := HTTPGetJSON(context.Background(), nil, server.URL, "token", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok response")
	}
}

func TestHTTPGetJSON_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad"}`))
	}))
	t.Cleanup(server.Close)

	var out map[string]any
	err := HTTPGetJSON(context.Background(), nil, server.URL, "token", nil, &out)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrHTTPRequestFailed) {
		t.Fatalf("expected ErrHTTPRequestFailed, got %v", err)
	}
	var httpErr *HTTPRequestError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPRequestError")
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", httpErr.StatusCode)
	}
	if httpErr.Body == "" {
		t.Fatalf("expected body to be captured")
	}
}

func TestHTTPPostJSON(t *testing.T) {
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
	if err := HTTPPostJSON(context.Background(), nil, server.URL, "token", nil, map[string]any{"foo": "bar"}, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok response")
	}
}

func TestHTTPPostJSON_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"nope"}`))
	}))
	t.Cleanup(server.Close)

	var out map[string]any
	err := HTTPPostJSON(context.Background(), nil, server.URL, "token", nil, map[string]any{"foo": "bar"}, &out)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrHTTPRequestFailed) {
		t.Fatalf("expected ErrHTTPRequestFailed, got %v", err)
	}
}
