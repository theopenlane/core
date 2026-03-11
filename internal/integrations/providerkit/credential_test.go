package providerkit

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestOAuthTokenFromCredential_Missing(t *testing.T) {
	credential := types.CredentialSet{}
	if _, err := OAuthTokenFromCredential(credential); !errors.Is(err, ErrOAuthTokenMissing) {
		t.Fatalf("expected ErrOAuthTokenMissing, got %v", err)
	}
}

func TestOAuthTokenFromCredential_EmptyAccess(t *testing.T) {
	credential := types.CredentialSet{
		OAuthRefreshToken: "refresh-token",
	}
	if _, err := OAuthTokenFromCredential(credential); !errors.Is(err, ErrAccessTokenEmpty) {
		t.Fatalf("expected ErrAccessTokenEmpty, got %v", err)
	}
}

func TestOAuthTokenFromCredential_WithExpiry(t *testing.T) {
	now := time.Now()
	credential := types.CredentialSet{
		OAuthExpiry: &now,
	}
	if _, err := OAuthTokenFromCredential(credential); !errors.Is(err, ErrAccessTokenEmpty) {
		t.Fatalf("expected ErrAccessTokenEmpty when only expiry is set, got %v", err)
	}
}

func TestOAuthTokenFromCredential_Success(t *testing.T) {
	credential := types.CredentialSet{
		OAuthAccessToken: "access-token",
	}

	token, err := OAuthTokenFromCredential(credential)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "access-token" {
		t.Fatalf("expected access-token, got %q", token)
	}
}

func TestAPITokenFromCredential_Missing(t *testing.T) {
	credential := types.CredentialSet{}
	if _, err := APITokenFromCredential(credential); !errors.Is(err, ErrAPITokenMissing) {
		t.Fatalf("expected ErrAPITokenMissing, got %v", err)
	}
}

func TestAPITokenFromCredential_MissingKey(t *testing.T) {
	credential := types.CredentialSet{
		ProviderData: json.RawMessage(`{"orgUrl":"https://example.okta.com"}`),
	}
	if _, err := APITokenFromCredential(credential); !errors.Is(err, ErrAPITokenMissing) {
		t.Fatalf("expected ErrAPITokenMissing, got %v", err)
	}
}

func TestAPITokenFromCredential_EmptyToken(t *testing.T) {
	credential := types.CredentialSet{
		ProviderData: json.RawMessage(`{"apiToken":""}`),
	}
	if _, err := APITokenFromCredential(credential); !errors.Is(err, ErrAPITokenMissing) {
		t.Fatalf("expected ErrAPITokenMissing for empty token, got %v", err)
	}
}

func TestAPITokenFromCredential_WhitespaceToken(t *testing.T) {
	credential := types.CredentialSet{
		ProviderData: json.RawMessage(`{"apiToken":"   "}`),
	}
	if _, err := APITokenFromCredential(credential); !errors.Is(err, ErrAPITokenMissing) {
		t.Fatalf("expected ErrAPITokenMissing for whitespace token, got %v", err)
	}
}

func TestAPITokenFromCredential_Success(t *testing.T) {
	credential := types.CredentialSet{
		ProviderData: json.RawMessage(`{"apiToken":"my-secret-token","orgUrl":"https://example.okta.com"}`),
	}

	token, err := APITokenFromCredential(credential)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "my-secret-token" {
		t.Fatalf("expected my-secret-token, got %q", token)
	}
}

func TestAPITokenFromCredential_InvalidJSON(t *testing.T) {
	credential := types.CredentialSet{
		ProviderData: json.RawMessage(`not-json`),
	}
	if _, err := APITokenFromCredential(credential); !errors.Is(err, ErrAPITokenMissing) {
		t.Fatalf("expected ErrAPITokenMissing for invalid JSON, got %v", err)
	}
}
