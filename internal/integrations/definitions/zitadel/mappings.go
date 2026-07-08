package zitadel

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Zitadel user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'user_id' in payload ? payload.user_id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'human' in payload && 'email' in payload.human && 'email' in payload.human.email ? payload.human.email.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'human' in payload && 'profile' in payload.human && 'display_name' in payload.human.profile && payload.human.profile.display_name != "" ? payload.human.profile.display_name : ('human' in payload && 'profile' in payload.human && 'given_name' in payload.human.profile && payload.human.profile.given_name != "" ? payload.human.profile.given_name + ('family_name' in payload.human.profile ? " " + payload.human.profile.family_name : "") : ('username' in payload ? payload.username : ""))`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, Expr: `'human' in payload && 'profile' in payload.human && 'given_name' in payload.human.profile ? payload.human.profile.given_name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, Expr: `'human' in payload && 'profile' in payload.human && 'family_name' in payload.human.profile ? payload.human.profile.family_name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('state' in payload ? (payload.state == 1 ? "ACTIVE" : (payload.state == 2 ? "INACTIVE" : (payload.state == 3 ? "DELETED" : (payload.state == 4 ? "SUSPENDED" : "INACTIVE")))) : "INACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAccountType, Expr: `dyn('human' in payload ? "USER" : "SERVICE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAddedAt, Expr: `'details' in payload && 'creation_date' in payload.details ? payload.details.creation_date : null`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
})

// zitadelMappings returns the built-in Zitadel ingest mappings
func zitadelMappings() []types.MappingRegistration {
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