package types //nolint:revive

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// DefinitionSpec describes the catalog-visible metadata for one definition
type DefinitionSpec struct {
	// ID is the canonical opaque identifier for the definition
	ID string `json:"id"`
	// Family is the optional grouping label for related definitions
	Family string `json:"family,omitempty"`
	// DisplayName is the UI-facing name for the definition
	DisplayName string `json:"displayName"`
	// Description is the user-facing description for the definition
	Description string `json:"description,omitempty"`
	// Category is the catalog category for the definition
	Category string `json:"category,omitempty"`
	// DocsURL links to documentation for the definition
	DocsURL string `json:"docsUrl,omitempty"`
	// LogoURL links to a catalog logo asset
	LogoURL string `json:"logoUrl,omitempty"`
	// Tags are UI-facing labels that describe what the integration provides
	Tags []string `json:"tags,omitempty"`
	// Active indicates whether the definition is enabled
	Active bool `json:"active"`
	// Visible indicates whether the definition is visible in catalog surfaces
	Visible bool `json:"visible"`
}

// Definition is the installable and executable integration unit
type Definition struct {
	// DefinitionSpec is the base catalog metadata for the definition
	DefinitionSpec `json:"spec"`
	// OperatorConfig describes operator-owned configuration for the definition
	OperatorConfig *OperatorConfigRegistration `json:"operatorConfig,omitempty"`
	// UserInput describes installation-scoped user input for the definition
	UserInput *UserInputRegistration `json:"userInput,omitempty"`
	// CredentialRegistrations describes the credential slots exposed by the definition
	CredentialRegistrations []CredentialRegistration `json:"credentialRegistrations,omitempty"`
	// Connections describes the connection modes exposed by the definition
	Connections []ConnectionRegistration `json:"connections,omitempty"`
	// Clients lists the clients the definition can build
	Clients []ClientRegistration `json:"clients,omitempty"`
	// Operations lists the operations the definition exposes
	Operations []OperationRegistration `json:"operations,omitempty"`
	// Mappings lists the default mappings shipped with the definition
	Mappings []MappingRegistration `json:"mappings,omitempty"`
	// Webhooks lists the webhook contracts exposed by the definition
	Webhooks []WebhookRegistration `json:"webhooks,omitempty"`
	// GalaListeners declares standalone gala listeners registered on the integration runtime
	GalaListeners []GalaListenerRegistration `json:"-"`
	// RuntimeIntegration declares that this definition can be fully provisioned from a single runtime config struct
	RuntimeIntegration *RuntimeIntegrationRegistration `json:"runtimeIntegration,omitempty"`
}

// GalaListenerRegistration declares a gala listener that should be registered on the integration runtime at startup
type GalaListenerRegistration struct {
	// Name is a stable listener identifier for diagnostics
	Name string
	// Register registers the listener on the supplied gala runtime with access to runtime services
	Register func(g *gala.Gala, services RuntimeServices) ([]gala.ListenerID, error)
}

// OperatorConfigRegistration describes operator-owned configuration for a definition
type OperatorConfigRegistration struct {
	// Schema is the JSON schema used to collect operator-owned configuration
	Schema json.RawMessage `json:"schema,omitempty"`
}

// BackfillFunc completes a stored payload missing values from the live installation
type BackfillFunc func(ctx context.Context, req InstallationRequest, payload json.RawMessage) (json.RawMessage, error)

// UserInputRegistration describes installation-scoped user input
type UserInputRegistration struct {
	// Schema is the JSON schema used to collect installation-scoped user input
	Schema json.RawMessage `json:"schema,omitempty"`
	// Replaces lists the retired layout names whose stored user input converts into this schema
	Replaces []string `json:"-"`
	// Convert reshapes stored user input from a retired layout into this schema
	Convert func(old json.RawMessage) (json.RawMessage, error) `json:"-"`
	// Backfill completes stored user input missing values
	Backfill BackfillFunc `json:"-"`
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
	// Recommended indicates the method that is recommend if there are multiple options
	Recommended bool `json:"recommended,omitempty"`
	// Replaces lists the retired slots whose stored payloads convert into this slot
	Replaces []CredentialSlotID `json:"-"`
	// Convert reshapes a payload stored under a retired slot into this slot's schema
	Convert func(from CredentialSlotID, old json.RawMessage) (json.RawMessage, error) `json:"-"`
	// Backfill completes a stored payload missing values
	Backfill BackfillFunc `json:"-"`
}

// ConnectionRegistration describes one connection mode for a definition
type ConnectionRegistration struct {
	// CredentialRef is the user-facing credential schema that selects this connection mode
	CredentialRef CredentialSlotID `json:"credentialRef"`
	// Name is the user-facing connection mode name
	Name string `json:"name,omitempty"`
	// Description explains what the connection mode does
	Description string `json:"description,omitempty"`
	// Meta is additional data the user might need to setup the integration with key-value pairs
	Meta map[string]MetaInfo `json:"meta,omitempty"`
	// CredentialRefs lists the credential slots used by this connection mode
	CredentialRefs []CredentialSlotID `json:"credentialRefs,omitempty"`
	// ClientRefs lists the clients initialized by this connection mode
	ClientRefs []ClientID `json:"-"`
	// HealthCheck exercises the connection's credentials before persistence and during health assessments
	HealthCheck *HealthCheckRegistration `json:"-"`
	// Integration describes installation-scoped metadata derived by this connection mode
	Integration *InstallationRegistration `json:"installation,omitempty"`
	// Auth describes how this connection mode performs auth when supported
	Auth *AuthRegistration `json:"auth,omitempty"`
	// Disconnect describes how this connection mode tears down an installation
	Disconnect *DisconnectRegistration `json:"disconnect,omitempty"`
}

// HealthCheckRegistration declares the credential-exercising health check for one connection mode
type HealthCheckRegistration struct {
	// ClientRef identifies which registered client the check builds; when empty the handler
	// receives only the credential bindings
	ClientRef ClientID `json:"-"`
	// Handle executes the check
	Handle OperationHandler `json:"-"`
}

// MetaInfo is data to store for the UI to present to the user during credential setup of an integration
type MetaInfo struct {
	// Value is the Value to show to the user
	Value string
	// allow copy will display a opy to clipboard button
	AllowCopy bool
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

// CredentialSchema returns the stored schema of one credential slot, from its registration or from the auth flow that fills it
func (d Definition) CredentialSchema(slot CredentialSlotID) json.RawMessage {
	registration, err := d.CredentialRegistration(slot)
	if err == nil && len(registration.Schema) > 0 {
		return registration.Schema
	}

	connection, found := lo.Find(d.Connections, func(c ConnectionRegistration) bool {
		return c.Auth != nil && c.Auth.CredentialRef == slot
	})
	if !found {
		return nil
	}

	return connection.Auth.Schema
}

// CredentialReplacing returns the credential registration whose slot takes over payloads stored under the retired slot
func (d Definition) CredentialReplacing(retired CredentialSlotID) (CredentialRegistration, bool) {
	return lo.Find(d.CredentialRegistrations, func(r CredentialRegistration) bool {
		return lo.Contains(r.Replaces, retired)
	})
}

// WebhookReplacing returns the webhook registration whose contract takes over rows persisted under the retired name
func (d Definition) WebhookReplacing(retired string) (WebhookRegistration, bool) {
	return lo.Find(d.Webhooks, func(r WebhookRegistration) bool {
		return lo.Contains(r.Replaces, retired)
	})
}

// ConnectionRegistration returns the connection registration for the given credential slot
func (d Definition) ConnectionRegistration(ref CredentialSlotID) (ConnectionRegistration, error) {
	reg, found := lo.Find(d.Connections, func(r ConnectionRegistration) bool {
		return r.CredentialRef == ref
	})
	if !found {
		return ConnectionRegistration{}, fmt.Errorf("%w: %s not found", ErrConnectionRefNotFound, ref)
	}

	return reg, nil
}

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
