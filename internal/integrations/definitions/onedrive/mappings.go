package onedrive

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprInternalPolicy is the CEL mapping expression for OneDrive file payloads mapped to InternalPolicy
var mapExprInternalPolicy = providerkit.CelMapExpr(
	entityops.InternalPolicyFields.Name.Expr(`'name' in payload && payload.name != "" ? payload.name : "Untitled Policy"`),
	entityops.InternalPolicyFields.ExternalFileID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.InternalPolicyFields.URL.Expr(`'webUrl' in payload ? payload.webUrl : null`),
	entityops.InternalPolicyFields.ManagementMode.Expr(`"INTEGRATION"`),
)
