package oidcgeneric

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0OIDCGEN00000000000000001")
	OIDCClient             = types.NewClientRef[*providerkit.AuthenticatedClient]()
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	ClaimsInspectOperation = types.NewOperationRef[ClaimsInspect]("claims.inspect")
)

const Slug = "oidc_generic"
