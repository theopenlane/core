package state

import (
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
)

// IntegrationProviderState stores provider-specific integration state captured during auth/config
type IntegrationProviderState struct {
	// Providers contains provider-specific state by provider key
	Providers map[string]json.RawMessage `json:"providers,omitempty"`
}

// ProviderData returns the raw provider state for a provider key.
func (s IntegrationProviderState) ProviderData(provider string) json.RawMessage {
	if provider == "" || len(s.Providers) == 0 {
		return nil
	}

	return s.Providers[provider]
}

// MergeProviderData deep-merges provider state and reports whether state changed
func (s *IntegrationProviderState) MergeProviderData(provider string, patch json.RawMessage) (bool, error) {
	if s == nil || provider == "" || len(patch) == 0 {
		return false, nil
	}

	if s.Providers == nil {
		s.Providers = map[string]json.RawMessage{}
	}

	merged, changed, err := jsonx.DeepMerge(s.Providers[provider], patch)
	if err != nil {
		return false, ErrProviderStateDecode
	}

	if !changed {
		return false, nil
	}

	s.Providers[provider] = merged

	return true, nil
}
