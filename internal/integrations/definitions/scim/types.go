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
	// DirectorySyncOperation is the stable identity handle for the SCIM directory sync operation
	directorySyncSchema, DirectorySyncOperation = providerkit.OperationSchema[DirectorySync]()

	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
)

// UserInput captures optional user-provided configuration for the SCIM integration
type UserInput struct {
	// Name is the human-readable label for this SCIM directory (e.g. "Okta Production")
	Name string `json:"name,omitempty" jsonschema:"title=Directory Name,description=Human-readable label for this SCIM directory."`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}
