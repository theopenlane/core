package onedrive

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprInternalPolicy is the CEL mapping expression for OneDrive file payloads mapped to InternalPolicy
var mapExprInternalPolicy = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingInternalPolicyName, Expr: `'name' in payload && payload.name != "" ? payload.name : "Untitled Policy"`},
	{Key: integrationgenerated.IntegrationMappingInternalPolicyExternalFileID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingInternalPolicyURL, Expr: `'webUrl' in payload ? payload.webUrl : null`},
	{Key: integrationgenerated.IntegrationMappingInternalPolicyManagementMode, Expr: `"INTEGRATION"`},
	{Key: integrationgenerated.IntegrationMappingInternalPolicyStatus, Expr: `"DRAFT"`},
})

// oneDriveMappings returns the built-in OneDrive ingest mappings
func oneDriveMappings() []types.MappingRegistration {
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
