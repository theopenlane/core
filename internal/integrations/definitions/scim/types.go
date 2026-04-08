package scim

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable reference for the SCIM integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SCIM000000000000000001")
	// SCIMAuthWebhook is the stable identity handle for the SCIM authentication webhook
	SCIMAuthWebhook = types.NewWebhookRef("scim.auth")
	// directorySyncSchema is the operation ref for the SCIM directory sync operation
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
	// healthCheckSchema is the operation ref for the SCIM health check operation
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
)

// UserInput captures optional user-provided configuration for the SCIM integration
type UserInput struct {
	// Name is the human-readable label for this SCIM directory (e.g. "Okta Production")
	Name string `json:"name,omitempty" jsonschema:"required,title=Directory Name,description=Human-readable label for this SCIM directory."`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting (allows inclusion, exclusion, etc.)"`
	// PrimaryDirectory marks this installation as the authoritative directory source for identity holder enrichment and lifecycle derivation
	PrimaryDirectory bool `json:"primaryDirectory,omitempty" jsonschema:"title=Primary Directory"`
}
