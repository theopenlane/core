package cloudflare

import (
	cf "github.com/cloudflare/cloudflare-go/v6"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0CFLARE00000000000000001")
	CloudflareClient       = types.NewClientRef[*cf.Client]()
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
)

const Slug = "cloudflare"
