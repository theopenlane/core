package cloudflare

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Cloudflare account member payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'user_id' in payload && payload.user_id != "" ? payload.user_id : ('email' in payload ? payload.email : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'email' in payload ? payload.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'first_name' in payload && payload.first_name != "" ? (payload.first_name + ('last_name' in payload && payload.last_name != "" ? " " + payload.last_name : "")) : ('email' in payload ? payload.email : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, Expr: `'first_name' in payload ? payload.first_name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, Expr: `'last_name' in payload ? payload.last_name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountMfaState, Expr: `dyn('two_factor_enabled' in payload && payload.two_factor_enabled ? "ENABLED" : "DISABLED")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('status' in payload && payload.status == "accepted" ? "ACTIVE" : "INACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
})

// cloudflareMappings returns the built-in Cloudflare ingest mappings
func cloudflareMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryAccount,
			},
		},
	}
}
