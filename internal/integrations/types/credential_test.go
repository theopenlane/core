package types

import (
	"testing"
	"time"

	"github.com/theopenlane/core/common/models"
)

func TestCloneCredentialSet(t *testing.T) {
	expiry := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	input := models.CredentialSet{
		APIToken:     "token",
		ProviderData: []byte(`{"region":"us-east-1"}`),
		Claims: map[string]any{
			"sub": "abc",
			"nested": map[string]any{
				"role": "admin",
			},
		},
		OAuthExpiry: &expiry,
	}

	cloned := CloneCredentialSet(input)
	if cloned.APIToken != "token" {
		t.Fatalf("expected APIToken to be copied")
	}
	if string(cloned.ProviderData) != string(input.ProviderData) {
		t.Fatalf("expected ProviderData to match")
	}
	if cloned.OAuthExpiry == nil || !cloned.OAuthExpiry.Equal(expiry) {
		t.Fatalf("expected OAuthExpiry to be copied")
	}

	cloned.ProviderData[0] = 'x'
	if string(input.ProviderData) == string(cloned.ProviderData) {
		t.Fatalf("expected ProviderData deep clone")
	}

	cloned.Claims["sub"] = "changed"
	nested := cloned.Claims["nested"].(map[string]any) //nolint:forcetypeassert
	nested["role"] = "viewer"

	if input.Claims["sub"] != "abc" {
		t.Fatalf("expected Claims top-level deep clone")
	}
	originalNested := input.Claims["nested"].(map[string]any) //nolint:forcetypeassert
	if originalNested["role"] != "admin" {
		t.Fatalf("expected Claims nested deep clone")
	}
}

func TestInferAuthKind(t *testing.T) {
	tests := []struct {
		name string
		set  models.CredentialSet
		want AuthKind
	}{
		{
			name: "oidc claims",
			set:  models.CredentialSet{Claims: map[string]any{"sub": "abc"}},
			want: AuthKindOIDC,
		},
		{
			name: "oauth token",
			set:  models.CredentialSet{OAuthAccessToken: "access"},
			want: AuthKindOAuth2,
		},
		{
			name: "api key",
			set:  models.CredentialSet{APIToken: "token"},
			want: AuthKindAPIKey,
		},
		{
			name: "client credentials",
			set:  models.CredentialSet{ClientID: "id", ClientSecret: "secret"},
			want: AuthKindOAuth2ClientCredentials,
		},
		{
			name: "aws federation",
			set:  models.CredentialSet{AccessKeyID: "key"},
			want: AuthKindAWSFederation,
		},
		{
			name: "workload identity",
			set:  models.CredentialSet{ServiceAccountKey: "key"},
			want: AuthKindWorkloadIdentity,
		},
		{
			name: "unknown",
			set:  models.CredentialSet{},
			want: AuthKindUnknown,
		},
	}

	for _, tt := range tests {
		got := InferAuthKind(tt.set)
		if got != tt.want {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.want, got)
		}
	}
}

func TestIsCredentialSetEmpty(t *testing.T) {
	if !IsCredentialSetEmpty(models.CredentialSet{}) {
		t.Fatalf("expected empty credential set to be empty")
	}

	if IsCredentialSetEmpty(models.CredentialSet{APIToken: "token"}) {
		t.Fatalf("expected API token credential set to be non-empty")
	}

	if IsCredentialSetEmpty(models.CredentialSet{ProviderData: []byte(`{"region":"us-east-1"}`)}) {
		t.Fatalf("expected provider data credential set to be non-empty")
	}

	expiry := time.Now().UTC().Add(time.Hour)
	if IsCredentialSetEmpty(models.CredentialSet{OAuthExpiry: &expiry}) {
		t.Fatalf("expected expiry-only credential set to be non-empty")
	}

	if IsCredentialSetEmpty(models.CredentialSet{Claims: map[string]any{"sub": "abc"}}) {
		t.Fatalf("expected claims credential set to be non-empty")
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
