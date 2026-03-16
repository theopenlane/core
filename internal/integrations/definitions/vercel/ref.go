package vercel

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID            = types.NewDefinitionRef("def_01K0VERCEL00000000000000001")
	VercelClient            = types.NewClientRef[*providerkit.AuthenticatedClient]()
	HealthDefaultOperation  = types.NewOperationRef[HealthCheck]("health.default")
	ProjectsSampleOperation = types.NewOperationRef[ProjectsSample]("projects.sample")
)

const Slug = "vercel"
