package scim

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the SCIM Directory Sync integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SCIM000000000000000001")

	// HealthDefaultOperation is the operation ref for the SCIM health check
	_, HealthDefaultOperation = providerkit.OperationSchema[HealthCheck]()
	// DirectorySyncOperation is the operation ref for the SCIM directory sync operation
	_, DirectorySyncOperation = providerkit.OperationSchema[DirectorySync]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}
