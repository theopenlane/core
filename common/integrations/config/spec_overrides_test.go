package config

import (
	"context"
	"testing"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/types"
)

// TestMergeProviderSpecs verifies provider overrides are merged onto loaded defaults.
func TestMergeProviderSpecs(t *testing.T) {
	t.Parallel()

	base := map[types.ProviderType]ProviderSpec{
		types.ProviderType("github"): {
			Name:        "github",
			DisplayName: "GitHub",
			Category:    "code",
			Active:      lo.ToPtr(true),
			OAuth: &OAuthSpec{
				ClientID:     "base-id",
				ClientSecret: "base-secret",
				AuthURL:      "https://github.com/login/oauth/authorize",
				TokenURL:     "https://github.com/login/oauth/access_token",
				Scopes:       []string{"repo"},
				AuthParams: map[string]string{
					"allow_signup": "true",
				},
				TokenParams: map[string]string{
					"audience": "base",
				},
			},
			Labels: map[string]string{
				"vendor": "github",
			},
			Metadata: map[string]any{
				"ui": map[string]any{
					"enabled": true,
				},
			},
		},
	}

	overrides := map[string]ProviderSpec{
		"github": {
			DisplayName: "GitHub Enterprise",
			OAuth: &OAuthSpec{
				ClientID: "override-id",
				Scopes:   []string{"repo", "read:org"},
				AuthParams: map[string]string{
					"prompt": "consent",
				},
				TokenParams: map[string]string{
					"resource": "override",
				},
			},
			Labels: map[string]string{
				"product": "scm",
			},
			Metadata: map[string]any{
				"ui": map[string]any{
					"beta": true,
				},
			},
		},
	}

	merged := MergeProviderSpecs(context.Background(), base, overrides)
	spec := merged[types.ProviderType("github")]

	if spec.DisplayName != "GitHub Enterprise" {
		t.Fatalf("expected display name override, got %q", spec.DisplayName)
	}
	if spec.OAuth == nil {
		t.Fatal("expected oauth block")
	}
	if spec.OAuth.ClientID != "override-id" {
		t.Fatalf("expected client id override, got %q", spec.OAuth.ClientID)
	}
	if spec.OAuth.ClientSecret != "base-secret" {
		t.Fatalf("expected base client secret to remain, got %q", spec.OAuth.ClientSecret)
	}
	if len(spec.OAuth.Scopes) != 2 {
		t.Fatalf("expected scope override, got %v", spec.OAuth.Scopes)
	}
	if spec.OAuth.AuthParams["allow_signup"] != "true" {
		t.Fatalf("expected existing auth param to remain, got %v", spec.OAuth.AuthParams)
	}
	if spec.OAuth.AuthParams["prompt"] != "consent" {
		t.Fatalf("expected auth param override, got %v", spec.OAuth.AuthParams)
	}
	if spec.OAuth.TokenParams["audience"] != "base" || spec.OAuth.TokenParams["resource"] != "override" {
		t.Fatalf("expected token params merge, got %v", spec.OAuth.TokenParams)
	}
	if spec.Labels["vendor"] != "github" || spec.Labels["product"] != "scm" {
		t.Fatalf("expected labels merge, got %v", spec.Labels)
	}

	ui, ok := spec.Metadata["ui"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested metadata map, got %#v", spec.Metadata["ui"])
	}
	if enabled, ok := ui["enabled"].(bool); !ok || !enabled {
		t.Fatalf("expected existing nested metadata key to remain, got %v", ui)
	}
	if beta, ok := ui["beta"].(bool); !ok || !beta {
		t.Fatalf("expected nested metadata override key, got %v", ui)
	}
}

// TestMergeProviderSpecsAllowsBooleanOverrideToFalse verifies that Active and Visible can be
// explicitly overridden to false, which was not possible before *bool was used.
func TestMergeProviderSpecsAllowsBooleanOverrideToFalse(t *testing.T) {
	t.Parallel()

	base := map[types.ProviderType]ProviderSpec{
		types.ProviderType("github"): {Name: "github", Active: lo.ToPtr(true), Visible: lo.ToPtr(true)},
	}

	overrides := map[string]ProviderSpec{
		"github": {Active: lo.ToPtr(false)},
	}

	merged := MergeProviderSpecs(context.Background(), base, overrides)
	spec := merged[types.ProviderType("github")]

	if spec.Active == nil || *spec.Active {
		t.Fatalf("expected Active to be overridden to false, got %v", spec.Active)
	}
	if spec.Visible == nil || !*spec.Visible {
		t.Fatalf("expected Visible to remain true, got %v", spec.Visible)
	}
}

// TestMergeProviderSpecsIgnoresUnknown verifies unknown provider keys are ignored.
func TestMergeProviderSpecsIgnoresUnknown(t *testing.T) {
	t.Parallel()

	base := map[types.ProviderType]ProviderSpec{
		types.ProviderType("github"): {Name: "github", Active: lo.ToPtr(true)},
	}

	overrides := map[string]ProviderSpec{
		"doesnotexist": {Name: "doesnotexist"},
	}

	merged := MergeProviderSpecs(context.Background(), base, overrides)
	if len(merged) != 1 {
		t.Fatalf("expected merged map size to remain unchanged, got %d", len(merged))
	}
	if _, ok := merged[types.ProviderType("github")]; !ok {
		t.Fatal("expected existing provider to remain")
	}
}
