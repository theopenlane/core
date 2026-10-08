package types //nolint:revive

import (
	"context"
	"encoding/json"

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
	UserInput *InputRegistration `json:"userInput,omitempty"`
	// Connections lists the connection methods exposed by the definition
	Connections []Connector `json:"-"`
	// Installation describes the installation metadata layout declared once for the definition
	Installation *InstallationRegistration `json:"installation,omitempty"`
	// Operations lists the operations the definition exposes
	Operations []OperationRegistration `json:"operations,omitempty"`
	// Mappings lists the default mappings shipped with the definition
	Mappings []MappingRegistration `json:"mappings,omitempty"`
	// Webhooks lists the webhook contracts exposed by the definition
	Webhooks []WebhookRegistration `json:"webhooks,omitempty"`
	// GalaListeners declares standalone gala listeners registered on the integration runtime
	GalaListeners []GalaListenerRegistration `json:"-"`
	// RuntimeIntegration declares this definition provisioned from a single runtime config struct
	RuntimeIntegration *RuntimeIntegrationRegistration `json:"runtimeIntegration,omitempty"`
}

// GalaListenerRegistration declares a gala listener registered at runtime startup
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

// UpgradeFunc reshapes a stored document from the layout, connection, or operation name it was persisted under into the current layout
type UpgradeFunc func(ctx context.Context, req InstallationRequest, from string, stored json.RawMessage) (json.RawMessage, error)

// ValidateFunc checks a payload that already satisfies its schema for constraints the schema cannot express
type ValidateFunc func(ctx context.Context, req InstallationRequest, payload json.RawMessage) error

// InputRegistration describes one stored installation-scoped input layout and how older documents move onto it
type InputRegistration struct {
	// Name is the stable layout name the stored document is keyed by
	Name string `json:"name"`
	// Schema is the JSON schema the stored document conforms to
	Schema json.RawMessage `json:"schema,omitempty"`
	// Upgrade reshapes a stored document from the layout it was persisted under, nil when none is declared
	Upgrade UpgradeFunc `json:"-"`
	// Validate checks a schema-valid payload for semantic constraints, nil when none is declared
	Validate ValidateFunc `json:"-"`
}

// Clone returns a copy of the input registration with its own schema bytes
func (r InputRegistration) Clone() InputRegistration {
	r.Schema = jsonx.CloneRawMessage(r.Schema)

	return r
}

// MetaInfo is data shown to the user during credential setup of an integration
type MetaInfo struct {
	// Value is the Value to show to the user
	Value string
	// AllowCopy displays a copy to clipboard button
	AllowCopy bool
}

// ConnectionList returns the connections declared by the definition
func (d Definition) ConnectionList() []Connection {
	return lo.Map(d.Connections, func(c Connector, _ int) Connection {
		return c.Connection()
	})
}

// Connection returns the connection with the given name
func (d Definition) Connection(name string) (Connection, bool) {
	return lo.Find(d.ConnectionList(), func(c Connection) bool {
		return c.Credential.Name == name
	})
}

// Operation returns the operation registration for the given name
func (d Definition) Operation(name string) (OperationRegistration, bool) {
	return lo.Find(d.Operations, func(r OperationRegistration) bool {
		return r.Name == name
	})
}

// Webhook returns the webhook registration for the given contract name
func (d Definition) Webhook(name string) (WebhookRegistration, bool) {
	return lo.Find(d.Webhooks, func(r WebhookRegistration) bool {
		return r.Name == name
	})
}

// resolveReplaced returns the registration identified by key, or the one whose replaces lists key; replaced reports the fallback
func resolveReplaced[T any, K comparable](registrations []T, key K, id func(T) K, replaces func(T) []K) (registration T, replaced bool, ok bool) {
	if registration, ok := lo.Find(registrations, func(r T) bool { return id(r) == key }); ok {
		return registration, false, true
	}

	registration, ok = lo.Find(registrations, func(r T) bool { return lo.Contains(replaces(r), key) })

	return registration, ok, ok
}

// ResolveConnection returns the connection for the name, or the one whose connection replaces it; replaced reports the fallback
func (d Definition) ResolveConnection(name string) (registration Connection, replaced bool, ok bool) {
	return resolveReplaced(d.ConnectionList(), name,
		func(c Connection) string { return c.Credential.Name },
		func(c Connection) []string { return c.Replaces })
}

// ResolveOperation returns the registration for the name, or the one whose operation replaces it; replaced reports the fallback
func (d Definition) ResolveOperation(name string) (registration OperationRegistration, replaced bool, ok bool) {
	return resolveReplaced(d.Operations, name,
		func(r OperationRegistration) string { return r.Name },
		func(r OperationRegistration) []string { return r.Replaces })
}

// ResolveWebhook returns the registration for the name, or the one whose contract replaces it; replaced reports the fallback
func (d Definition) ResolveWebhook(name string) (registration WebhookRegistration, replaced bool, ok bool) {
	return resolveReplaced(d.Webhooks, name,
		func(r WebhookRegistration) string { return r.Name },
		func(r WebhookRegistration) []string { return r.Replaces })
}

// DefinitionProviderState stores installation-scoped state for one definition
type DefinitionProviderState struct {
	// CredentialRef identifies the connection active for the installation
	CredentialRef string `json:"credentialRef"`
}

// ProviderState returns the persisted provider state for this definition
func (d Definition) ProviderState(state IntegrationProviderState) (DefinitionProviderState, error) {
	var out DefinitionProviderState
	if err := jsonx.UnmarshalIfPresent(state.Providers[d.ID], &out); err != nil {
		return DefinitionProviderState{}, err
	}

	return out, nil
}

// WithProviderState returns a copy of the provider state with this definition's state updated
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
