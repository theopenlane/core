package state

import (
	"encoding/json"
	"reflect"

	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

// IntegrationProviderState stores provider-specific integration state captured during auth/config
type IntegrationProviderState struct {
	// Providers contains provider-specific state by provider key
	Providers map[string]json.RawMessage `json:"providers,omitempty"`
}

// ProviderDataMap returns a cloned provider state map for a provider key
func (s IntegrationProviderState) ProviderDataMap(provider string) (map[string]any, error) {
	if provider == "" || len(s.Providers) == 0 {
		return nil, nil
	}

	raw := s.Providers[provider]
	if len(raw) == 0 {
		return nil, nil
	}

	var decoded map[string]any
	if err := jsonx.RoundTrip(raw, &decoded); err != nil {
		return nil, ErrProviderStateDecode
	}

	return mapx.DeepCloneMapAny(decoded), nil
}

// MergeProviderData deep-merges provider state and reports whether state changed
func (s *IntegrationProviderState) MergeProviderData(provider string, patch map[string]any) (bool, error) {
	if s == nil || provider == "" || patch == nil {
		return false, nil
	}

	if len(patch) == 0 {
		return false, nil
	}

	if s.Providers == nil {
		s.Providers = map[string]json.RawMessage{}
	}

	current, err := s.ProviderDataMap(provider)
	if err != nil {
		return false, err
	}

	next := mapx.DeepMergeMapAny(current, mapx.DeepCloneMapAny(patch))
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
