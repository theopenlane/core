package keystore

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write temp file %s: %v", path, err)
	}
	return path
}

func TestLoadProviderSpecsFromFile(t *testing.T) {
	content := `integrations:
  providers:
    sample:
      displayName: Sample Provider
      authType: oauth2
      oauth:
        clientId: sample-id
        clientSecret: sample-secret
        authUrl: https://example.com/auth
        tokenUrl: https://example.com/token
        scopes: ["profile"]
`
	tmp := t.TempDir()
	configPath := writeTempFile(t, tmp, "config.yaml", content)

	specs, err := LoadProviderSpecs(configPath)
	if err != nil {
		t.Fatalf("LoadProviderSpecs returned error: %v", err)
	}

	sample, ok := specs["sample"]
	if !ok {
		t.Fatalf("expected provider 'sample' to be loaded: %#v", specs)
	}

	if sample.Name != "sample" {
		t.Fatalf("expected Name to default to map key, got %q", sample.Name)
	}
	if sample.DisplayName != "Sample Provider" {
		t.Fatalf("unexpected display name: %q", sample.DisplayName)
	}
	if sample.AuthType != AuthTypeOAuth2 {
		t.Fatalf("unexpected auth type: %v", sample.AuthType)
	}
}

func TestLoadProviderSpecsFromDir(t *testing.T) {
	tmp := t.TempDir()
	// JSON spec without explicit name to ensure filename fallback works.
	jsonSpec := `{
  "displayName": "JSON Provider",
  "authType": "apikey",
  "apiKey": {
    "keyLabel": "API Key",
    "headerName": "X-API-KEY"
  }
}`
	writeTempFile(t, tmp, "JsonProvider.json", jsonSpec)

	yamlSpec := `integrations:
  providers:
    other:
      displayName: Other Provider
      authType: apikey
      apiKey:
        keyLabel: Key
        headerName: X-Key
`
	writeTempFile(t, tmp, "other.yaml", yamlSpec)

	specs, err := LoadProviderSpecs(tmp)
	if err != nil {
		t.Fatalf("LoadProviderSpecs returned error: %v", err)
	}

	if len(specs) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(specs))
	}

	jsonProvider, ok := specs["jsonprovider"]
	if !ok {
		t.Fatalf("expected jsonprovider entry in specs: %#v", specs)
	}
	if jsonProvider.Name != "JsonProvider" {
		t.Fatalf("expected Name to be derived from filename, got %q", jsonProvider.Name)
	}
	if jsonProvider.AuthType != AuthTypeAPIKey {
		t.Fatalf("unexpected auth type for json provider: %v", jsonProvider.AuthType)
	}

	if _, ok := specs["other"]; !ok {
		t.Fatalf("expected yaml provider 'other' to be loaded: %#v", specs)
	}
}
