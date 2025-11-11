package config

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestFSLoader_LoadInterpolatesEnv(t *testing.T) {
	fsys := fstest.MapFS{
		"providers/github.json": {
			Data: []byte(`{
				"name": "github",
				"displayName": "GitHub",
				"category": "code",
				"authType": "oauth2",
				"active": true,
				"logoUrl": "logo",
				"docsUrl": "docs",
				"oauth": {
					"clientId": "${GITHUB_CLIENT_ID}",
					"clientSecret": "${GITHUB_CLIENT_SECRET}",
					"authUrl": "https://github.com/login/oauth/authorize",
					"tokenUrl": "https://github.com/login/oauth/access_token",
					"scopes": ["repo", "read:${GITHUB_SCOPE_SUFFIX}"],
					"authParams": {"audience": "${GITHUB_AUDIENCE}"}
				},
				"metadata": {
					"apiBase": "${GITHUB_API_BASE}",
					"nested": {"value": "${GITHUB_NESTED}"}
				},
				"defaults": {
					"labels": ["${GITHUB_LABEL}"]
				},
				"labels": {
					"tier": "${GITHUB_TIER}"
				},
				"credentialsSchema": {
					"type": "object",
					"properties": {
						"workspace": {
							"type": "string",
							"title": "${GITHUB_SCHEMA_TITLE}"
						}
					}
				}
			}`),
		},
	}

	env := map[string]string{
		"GITHUB_CLIENT_ID":     "client-id",
		"GITHUB_CLIENT_SECRET": "client-secret",
		"GITHUB_SCOPE_SUFFIX":  "org",
		"GITHUB_AUDIENCE":      "https://github.example.com",
		"GITHUB_API_BASE":      "https://api.github.com",
		"GITHUB_NESTED":        "nested",
		"GITHUB_LABEL":         "oss",
		"GITHUB_TIER":          "gold",
		"GITHUB_SCHEMA_TITLE":  "Org workspace",
	}

	loader := NewFSLoader(fsys, "providers")
	loader.EnvLookup = func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	}

	specs, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	spec, ok := specs[types.ProviderType("github")]
	if !ok {
		t.Fatalf("expected github spec to be loaded")
	}

	if spec.OAuth == nil {
		t.Fatalf("expected OAuth spec to be populated")
	}
	if spec.OAuth.ClientID != "client-id" || spec.OAuth.ClientSecret != "client-secret" {
		t.Fatalf("env variables not interpolated, got %+v", spec.OAuth)
	}
	if got := spec.OAuth.Scopes[1]; got != "read:org" {
		t.Fatalf("expected scope interpolation, got %q", got)
	}
	if got := spec.OAuth.AuthParams["audience"]; got != "https://github.example.com" {
		t.Fatalf("expected auth params interpolation, got %q", got)
	}

	value, ok := spec.Metadata["apiBase"].(string)
	if !ok || value != "https://api.github.com" {
		t.Fatalf("expected metadata interpolation, got %#v", spec.Metadata["apiBase"])
	}

	nested, ok := spec.Metadata["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested metadata map, got %#v", spec.Metadata["nested"])
	}
	if nested["value"] != "nested" {
		t.Fatalf("expected nested metadata interpolation, got %#v", nested["value"])
	}

	if spec.Labels["tier"] != "gold" {
		t.Fatalf("expected label interpolation, got %q", spec.Labels["tier"])
	}

	defaultLabels, ok := spec.Defaults["labels"].([]interface{})
	if !ok || len(defaultLabels) != 1 || defaultLabels[0] != "oss" {
		t.Fatalf("expected defaults interpolation, got %#v", spec.Defaults["labels"])
	}

	props, ok := spec.CredentialsSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected credentials schema properties map, got %#v", spec.CredentialsSchema["properties"])
	}
	workspace, ok := props["workspace"].(map[string]any)
	if !ok {
		t.Fatalf("expected workspace schema map, got %#v", props["workspace"])
	}
	if workspace["title"] != "Org workspace" {
		t.Fatalf("expected credentials schema interpolation, got %#v", workspace["title"])
	}
}

func TestFSLoader_LoadMissingEnv(t *testing.T) {
	fsys := fstest.MapFS{
		"providers/slack.json": {
			Data: []byte(`{
				"name": "slack",
				"displayName": "Slack",
				"category": "collab",
				"authType": "oauth2",
				"active": true,
				"oauth": {
					"clientId": "${SLACK_CLIENT_ID}"
				}
			}`),
		},
	}

	loader := NewFSLoader(fsys, "providers")
	loader.EnvLookup = func(string) (string, bool) {
		return "", false
	}

	_, err := loader.Load(context.Background())
	if err == nil {
		t.Fatalf("expected error when env variable missing")
	}
	if !errors.Is(err, integrations.ErrEnvVarNotDefined) {
		t.Fatalf("expected ErrEnvVarNotDefined, got %v", err)
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
	_, err := loader.Load(context.Background())
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
