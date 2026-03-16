package oidcgeneric

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Generic OIDC integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0OIDCGEN00000000000000001")
	// OIDCClient is the client ref for the OIDC userinfo HTTP client used by this definition
	OIDCClient = types.NewClientRef[*providerkit.AuthenticatedClient]()
	// HealthDefaultOperation is the operation ref for the Generic OIDC health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	// ClaimsInspectOperation is the operation ref for the OIDC claims inspect operation
	ClaimsInspectOperation = types.NewOperationRef[ClaimsInspect]("claims.inspect")
)

// Slug is the unique identifier for the Generic OIDC integration
const Slug = "oidc_generic"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}
