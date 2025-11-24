package config

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestFSLoader_LoadSupportsYAML(t *testing.T) {
	fsys := fstest.MapFS{
		"providers/github.yaml": {
			Data: []byte(`
name: github
displayName: GitHub
category: code
authType: oauth2
active: true
oauth:
  clientId: abc
  clientSecret: def
  authUrl: https://example.com/auth
  tokenUrl: https://example.com/token
`),
		},
	}

	loader := NewFSLoader(fsys, "providers")
	specs, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, ok := specs[types.ProviderType("github")]; !ok {
		t.Fatalf("expected github spec to be loaded from YAML")
	}
}

func TestFSLoader_LoadUnsupportedSchemaVersion(t *testing.T) {
	fsys := fstest.MapFS{
		"providers/github.json": {
			Data: []byte(`{
				"name": "github",
				"displayName": "GitHub",
				"category": "code",
				"authType": "oauth2",
				"active": true,
				"schemaVersion": "v9"
			}`),
		},
	}

	loader := NewFSLoader(fsys, "providers")
	_, err := loader.Load()
	if err == nil {
		t.Fatalf("expected schema version error")
	}
	if !errors.Is(err, integrations.ErrSchemaVersionUnsupported) {
		t.Fatalf("expected ErrSchemaVersionUnsupported, got %v", err)
	}
}

func TestToProviderConfigs(t *testing.T) {
	specs := map[types.ProviderType]ProviderSpec{
		types.ProviderType("github"): {
			Name:        "github",
			DisplayName: "GitHub",
			AuthType:    types.AuthKindOAuth2,
			Active:      true,
		},
	}

	configs := ToProviderConfigs(specs)
	if len(configs) != 1 {
		t.Fatalf("expected one config, got %d", len(configs))
	}

	cfg, ok := configs[types.ProviderType("github")]
	if !ok {
		t.Fatalf("expected github config")
	}
	if cfg.DisplayName != "GitHub" {
		t.Fatalf("expected display name to propagate, got %q", cfg.DisplayName)
	}
}
