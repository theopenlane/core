package azureentraid

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID              = types.NewDefinitionRef("def_01K0AZENTRAID000000000000001")
	EntraClient               = types.NewClientRef[*providerkit.AuthenticatedClient]()
	HealthDefaultOperation    = types.NewOperationRef[HealthCheck]("health.default")
	DirectoryInspectOperation = types.NewOperationRef[DirectoryInspect]("directory.inspect")
)

const Slug = "azure_entra_id"
