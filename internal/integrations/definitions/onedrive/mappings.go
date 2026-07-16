package onedrive

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprInternalPolicy is the CEL mapping expression for OneDrive file payloads mapped to InternalPolicy
var mapExprInternalPolicy = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyInternalPolicyName, Expr: `'name' in payload && payload.name != "" ? payload.name : "Untitled Policy"`},
	{Key: entityops.InputKeyInternalPolicyExternalFileID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyInternalPolicyURL, Expr: `'webUrl' in payload ? payload.webUrl : null`},
	{Key: entityops.InputKeyInternalPolicyManagementMode, Expr: `"INTEGRATION"`},
	{Key: entityops.InputKeyInternalPolicyStatus, Expr: `"DRAFT"`},
})

// oneDriveMappings returns the built-in OneDrive ingest mappings
func oneDriveMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: entityops.SchemaInternalPolicy.Name,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprInternalPolicy,
			},
		},
	}
}
