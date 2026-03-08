package state

import (
	"encoding/json"
	"reflect"

	"github.com/theopenlane/core/pkg/mapx"
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

	var current map[string]any
	if raw := s.Providers[provider]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &current); err != nil {
			return false, ErrProviderStateDecode
		}
	}

	var patchMap map[string]any
	if err := json.Unmarshal(patch, &patchMap); err != nil {
		return false, ErrProviderStatePatchEncode
	}

	next := mapx.DeepMergeMapAny(current, mapx.DeepCloneMapAny(patchMap))
	if reflect.DeepEqual(current, next) {
		return false, nil
	}

	encoded, err := json.Marshal(next)
	if err != nil {
		return false, ErrProviderStatePatchEncode
	}

	s.Providers[provider] = encoded

	return true, nil
}
