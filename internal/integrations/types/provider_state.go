package types

import (
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
)

// DefinitionProviderState stores installation-scoped state for one definition
type DefinitionProviderState struct {
	// CredentialRef identifies which credential-schema-selected connection mode is active for the installation
	CredentialRef CredentialSlotID `json:"credentialRef"`
}

// ProviderState returns the persisted provider state for this definition
func (d Definition) ProviderState(state IntegrationProviderState) (DefinitionProviderState, error) {
	if state.Providers == nil {
		return DefinitionProviderState{}, nil
	}

	raw, ok := state.Providers[d.ID]
	if !ok || len(raw) == 0 {
		return DefinitionProviderState{}, nil
	}

	var out DefinitionProviderState
	if err := jsonx.UnmarshalIfPresent(raw, &out); err != nil {
		return DefinitionProviderState{}, err
	}

	return out, nil
}

// WithProviderState returns a copy of the installation provider state with this definition's state updated
func (d Definition) WithProviderState(state IntegrationProviderState, next DefinitionProviderState) (IntegrationProviderState, error) {
	raw, err := jsonx.ToRawMessage(next)
	if err != nil {
		return IntegrationProviderState{}, err
	}

	out := IntegrationProviderState{
		Providers: map[string]json.RawMessage{},
	}

	for key, value := range state.Providers {
		out.Providers[key] = jsonx.CloneRawMessage(value)
	}

	out.Providers[d.ID] = raw

	return out, nil
}
