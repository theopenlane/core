package scim

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the SCIM Directory Sync integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SCIM000000000000000001")

	healthCheckSchema, healthCheckOperation     = providerkit.OperationSchema[HealthCheck]()
	directorySyncSchema, directorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}
