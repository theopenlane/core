package types

import (
	"context"
	"encoding/json"
)

// DefinitionID is the immutable opaque identifier for one installable definition
type DefinitionID string

// DefinitionSpec describes the catalog-visible metadata for one definition
type DefinitionSpec struct {
	// ID is the canonical opaque identifier for the definition
	ID DefinitionID `json:"id"`
	// Slug is the human-readable alias for the definition
	Slug string `json:"slug"`
	// Version is the manifest or implementation version for the definition
	Version string `json:"version"`
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
	// Spec is the base catalog metadata for the definition
	Spec DefinitionSpec `json:"spec"`
	// OperatorConfig describes operator-owned configuration for the definition
	OperatorConfig *OperatorConfigRegistration `json:"operatorConfig,omitempty"`
	// UserInput describes installation-scoped user input for the definition
	UserInput *UserInputRegistration `json:"userInput,omitempty"`
	// Credentials describes how the definition accepts credentials
	Credentials *CredentialRegistration `json:"credentials,omitempty"`
	// Auth describes the definition's auth flow when it has one
	Auth *AuthRegistration `json:"auth,omitempty"`
	// Clients lists the clients the definition can build
	Clients []ClientRegistration `json:"clients,omitempty"`
	// Operations lists the operations the definition exposes
	Operations []OperationRegistration `json:"operations,omitempty"`
	// Mappings lists the default mappings shipped with the definition
	Mappings []MappingRegistration `json:"mappings,omitempty"`
	// Webhooks lists the webhook contracts exposed by the definition
	Webhooks []WebhookRegistration `json:"webhooks,omitempty"`
}

// OperatorConfigFunc normalizes or validates operator-owned configuration
type OperatorConfigFunc func(ctx context.Context, value json.RawMessage) (json.RawMessage, error)

// UserInputFunc normalizes or validates installation-scoped user input
type UserInputFunc func(ctx context.Context, value json.RawMessage) (json.RawMessage, error)

// OperatorConfigRegistration describes operator-owned configuration for a definition
type OperatorConfigRegistration struct {
	// Schema is the JSON schema used to collect operator-owned configuration
	Schema json.RawMessage `json:"schema,omitempty"`
	// Normalize canonicalizes raw operator-owned configuration
	Normalize OperatorConfigFunc `json:"-"`
	// Validate verifies that operator-owned configuration is usable
	Validate OperatorConfigFunc `json:"-"`
}

// UserInputRegistration describes installation-scoped user input
type UserInputRegistration struct {
	// Schema is the JSON schema used to collect installation-scoped user input
	Schema json.RawMessage `json:"schema,omitempty"`
	// Normalize canonicalizes raw user input
	Normalize UserInputFunc `json:"-"`
	// Validate verifies that user input is usable
	Validate UserInputFunc `json:"-"`
}

// AuthRegistration describes how one definition starts and completes auth
type AuthRegistration struct {
	// Start initializes an auth flow
	Start AuthStartFunc `json:"-"`
	// Complete finalizes an auth flow
	Complete AuthCompleteFunc `json:"-"`
	// Refresh exchanges an existing credential for a refreshed one
	Refresh AuthRefreshFunc `json:"-"`
}
