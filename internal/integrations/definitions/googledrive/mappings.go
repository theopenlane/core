package googledrive

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprInternalPolicy is the CEL mapping expression for Google Drive file payloads mapped to InternalPolicy
var mapExprInternalPolicy = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyInternalPolicyName, Expr: `'name' in payload && payload.name != "" ? payload.name : "Untitled Policy"`},
	{Key: entityops.InputKeyInternalPolicyExternalFileID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyInternalPolicyManagementMode, Expr: `"INTEGRATION"`},
	{Key: entityops.InputKeyInternalPolicyStatus, Expr: `"DRAFT"`},
})

// googleDriveMappings returns the built-in Google Drive ingest mappings
func googleDriveMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaInternalPolicy,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprInternalPolicy,
			},
		},
	}
}
