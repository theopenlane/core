package types

import "encoding/json"

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
	// Tags are optional catalog labels for the definition
	Tags []string `json:"tags,omitempty"`
	// Labels stores arbitrary metadata for the definition
	Labels map[string]string `json:"labels,omitempty"`
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
}

// OperatorConfigRegistration describes operator-owned configuration for a definition
type OperatorConfigRegistration struct {
	// Schema is the JSON schema used to collect operator-owned configuration
	Schema json.RawMessage `json:"schema,omitempty"`
}

// UserInputRegistration describes installation-scoped user input
type UserInputRegistration struct {
	// Schema is the JSON schema used to collect installation-scoped user input
	Schema json.RawMessage `json:"schema,omitempty"`
}
