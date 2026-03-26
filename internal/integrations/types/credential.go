package types

import (
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
)

// CredentialSet is the persisted credential bundle used by integrations
type CredentialSet = models.CredentialSet

// CredentialBinding links one durable credential slot to its persisted credential bundle
type CredentialBinding struct {
	// Ref is the durable credential slot identifier
	Ref CredentialSlotID `json:"ref"`
	// Credential is the persisted credential bundle for the slot
	Credential CredentialSet `json:"credential"`
}

// CredentialBindings lists persisted credential bundles by durable credential slot
type CredentialBindings []CredentialBinding

// Resolve returns the credential bound to the supplied slot ref when present
func (bindings CredentialBindings) Resolve(ref CredentialSlotID) (CredentialSet, bool) {
	for _, binding := range bindings {
		if binding.Ref == ref {
			return binding.Credential, true
		}
	}

	return CredentialSet{}, false
}

// With returns a copy of the bindings with the given ref set to credential,
// replacing an existing binding for the same ref or appending a new one
func (bindings CredentialBindings) With(ref CredentialSlotID, credential CredentialSet) CredentialBindings {
	out := make(CredentialBindings, len(bindings))
	copy(out, bindings)

	for i := range out {
		if out[i].Ref == ref {
			out[i].Credential = credential
			return out
		}
	}

	return append(out, CredentialBinding{Ref: ref, Credential: credential})
}

// CredentialRegistration declares how a definition accepts credentials
type CredentialRegistration struct {
	// Ref is the durable credential slot identifier
	Ref CredentialSlotID `json:"ref"`
	// Name is the user-facing credential slot name
	Name string `json:"name,omitempty"`
	// Description describes when this credential slot should be used
	Description string `json:"description,omitempty"`
	// Schema is the JSON schema used to collect credentials
	Schema json.RawMessage `json:"schema,omitempty"`
}

// CredentialRegistration returns the credential registration for the given ref
func (d Definition) CredentialRegistration(ref CredentialSlotID) (CredentialRegistration, error) {
	reg, found := lo.Find(d.CredentialRegistrations, func(r CredentialRegistration) bool {
		return r.Ref == ref
	})
	if !found {
		return CredentialRegistration{}, ErrCredentialRefNotFound
	}

	return reg, nil
}
