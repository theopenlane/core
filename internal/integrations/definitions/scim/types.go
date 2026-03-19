package scim

import "github.com/theopenlane/core/internal/integrations/types"

var (
	// DefinitionID is the stable identifier for the SCIM Directory Sync integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SCIM000000000000000001")
	// HealthDefaultOperation is the operation ref for the SCIM health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck](types.HealthDefaultOperation)
	// DirectorySyncOperation is the operation ref for the SCIM directory sync operation
	DirectorySyncOperation = types.NewOperationRef[DirectorySync]("directory.sync")
)

// Slug is the unique identifier for the SCIM Directory Sync integration
const Slug = "scim_directory_sync"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}
