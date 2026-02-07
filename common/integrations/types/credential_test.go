package types

import (
	"testing"

	"golang.org/x/oauth2"

	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/core/common/models"
)

func TestBuildCredentialPayloadValidation(t *testing.T) {
	if _, err := BuildCredentialPayload(ProviderUnknown); err != ErrProviderTypeRequired {
		t.Fatalf("expected ErrProviderTypeRequired, got %v", err)
	}

	if _, err := BuildCredentialPayload(ProviderType("test")); err != ErrCredentialSetRequired {
		t.Fatalf("expected ErrCredentialSetRequired, got %v", err)
	}
}

func TestBuildCredentialPayloadDefaults(t *testing.T) {
	payload, err := BuildCredentialPayload(ProviderType("test"), WithCredentialSet(models.CredentialSet{APIToken: "token"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Kind != CredentialKindOAuthToken {
		t.Fatalf("expected default kind oauth token")
	}
}

func TestCredentialBuilder(t *testing.T) {
	builder := NewCredentialBuilder(ProviderType("test")).With(WithCredentialSet(models.CredentialSet{APIToken: "token"}))
	payload, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Provider != ProviderType("test") {
		t.Fatalf("expected provider")
	}
}

func TestMergeScopes(t *testing.T) {
	out := MergeScopes([]string{"a"}, " b ", "a", "", "c")
	want := []string{"a", "b", "c"}
	if len(out) != len(want) {
		t.Fatalf("expected %d scopes, got %d", len(want), len(out))
	}
	for i := range want {
		if out[i] != want[i] {
			t.Fatalf("expected %q at %d, got %q", want[i], i, out[i])
		}
	}
}

func TestRedacted(t *testing.T) {
	payload := CredentialPayload{
		Provider: ProviderType("test"),
		Kind:     CredentialKindOAuthToken,
		Token: &oauth2.Token{
			AccessToken:  "access",
			RefreshToken: "refresh",
		},
		Data: models.CredentialSet{APIToken: "secret"},
	}

	redacted := payload.Redacted()
	if redacted.Token == nil || redacted.Token.AccessToken != "[redacted]" {
		t.Fatalf("expected access token redacted")
	}
	if redacted.Token.RefreshToken != "" {
		t.Fatalf("expected refresh token cleared")
	}
	if redacted.Data.APIToken != "" {
		t.Fatalf("expected credential data cleared")
	}
}

func TestCredentialOptions(t *testing.T) {
	payload := CredentialPayload{
		Token:  &oauth2.Token{AccessToken: "token"},
		Claims: &oidc.IDTokenClaims{TokenClaims: oidc.TokenClaims{Subject: "sub"}},
	}

	tokenOpt := payload.OAuthTokenOption()
	if !tokenOpt.IsPresent() {
		t.Fatalf("expected token option to be present")
	}
	if tokenOpt.MustGet().AccessToken != "token" {
		t.Fatalf("expected access token")
	}

	claimsOpt := payload.ClaimsOption()
	if !claimsOpt.IsPresent() {
		t.Fatalf("expected claims option to be present")
	}
	if claimsOpt.MustGet().Subject != "sub" {
		t.Fatalf("expected subject")
	}

	if optionFromPointer[*oauth2.Token](nil).IsPresent() {
		t.Fatalf("expected none option")
	}
}
