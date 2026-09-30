package googledrive

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprInternalPolicy maps Google Drive file payloads to InternalPolicy
var mapExprInternalPolicy = providerkit.CelMapExpr(
	entityops.InternalPolicyFields.Name.Expr(`'name' in payload && payload.name != "" ? payload.name : "Untitled Policy"`),
	entityops.InternalPolicyFields.ExternalFileID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.InternalPolicyFields.ManagementMode.Expr(`"INTEGRATION"`),
)
