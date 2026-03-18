package config

import (
	"testing"

	"github.com/theopenlane/core/common/integrations/types"
)

func TestProviderSpecProviderType(t *testing.T) {
	spec := ProviderSpec{Name: "github"}
	if spec.ProviderType() != types.ProviderType("github") {
		t.Fatalf("expected provider type to match name")
	}
}

func TestProviderSpecToProviderConfig(t *testing.T) {
	spec := ProviderSpec{
		Name:        "github",
		DisplayName: "GitHub",
		Category:    "code",
		DocsURL:     "docs",
		LogoURL:     "logo",
		AuthType:    types.AuthKindOAuth2,
		CredentialsSchema: map[string]any{
			"type": "object",
		},
		Metadata: map[string]any{"foo": "bar"},
	}

	cfg := spec.ToProviderConfig()
	if cfg.Type != types.ProviderType("github") {
		t.Fatalf("expected provider type to match")
	}
	if cfg.DisplayName != "GitHub" {
		t.Fatalf("expected display name")
	}
	if cfg.Category != "code" {
		t.Fatalf("expected category")
	}
	if cfg.Auth != types.AuthKindOAuth2 {
		t.Fatalf("expected auth kind")
	}
	if cfg.Schema == nil {
		t.Fatalf("expected schema to be set")
	}
	if cfg.Metadata == nil {
		t.Fatalf("expected metadata to be set")
	}
}

func TestToProviderConfigsEmpty(t *testing.T) {
	if out := ToProviderConfigs(nil); out != nil {
		t.Fatalf("expected nil for empty map")
	}
}
