package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

//go:embed providers/*.json
var providerFS embed.FS

type embeddedProvider struct {
	Name              string         `json:"name"`
	CredentialsSchema map[string]any `json:"credentialsSchema"`
}

// LoadEmbeddedSchemas returns credential schemas decoded from the embedded provider specs.
func LoadEmbeddedSchemas() (map[string]map[string]any, error) {
	entries, err := providerFS.ReadDir("providers")
	if err != nil {
		return nil, fmt.Errorf("read providers directory: %w", err)
	}
	result := make(map[string]map[string]any, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		data, err := providerFS.ReadFile(filepath.Join("providers", name))
		if err != nil {
			return nil, fmt.Errorf("read provider spec %s: %w", name, err)
		}
		var embedded embeddedProvider
		if err := json.Unmarshal(data, &embedded); err != nil {
			return nil, fmt.Errorf("unmarshal provider spec %s: %w", name, err)
		}
		if embedded.Name == "" {
			embedded.Name = strings.TrimSuffix(name, filepath.Ext(name))
		}
		if len(embedded.CredentialsSchema) == 0 {
			continue
		}
		result[strings.ToLower(embedded.Name)] = embedded.CredentialsSchema
	}
	return result, nil
}
