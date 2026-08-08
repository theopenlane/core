package zitadel

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Zitadel user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `'user_id' in payload ? payload.user_id : ""`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'human' in payload && 'email' in payload.human && 'email' in payload.human.email ? payload.human.email.email : ""`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'human' in payload && 'profile' in payload.human && 'display_name' in payload.human.profile && payload.human.profile.display_name != "" ? payload.human.profile.display_name : ('human' in payload && 'profile' in payload.human && 'given_name' in payload.human.profile && payload.human.profile.given_name != "" ? payload.human.profile.given_name + ('family_name' in payload.human.profile ? " " + payload.human.profile.family_name : "") : ('username' in payload ? payload.username : ""))`},
	{Key: entityops.InputKeyDirectoryAccountGivenName, Expr: `'human' in payload && 'profile' in payload.human && 'given_name' in payload.human.profile ? payload.human.profile.given_name : ""`},
	{Key: entityops.InputKeyDirectoryAccountFamilyName, Expr: `'human' in payload && 'profile' in payload.human && 'family_name' in payload.human.profile ? payload.human.profile.family_name : ""`},
	{Key: entityops.InputKeyDirectoryAccountStatus, Expr: `dyn('state' in payload ? (payload.state == 1 ? "ACTIVE" : (payload.state == 2 ? "INACTIVE" : (payload.state == 3 ? "DELETED" : (payload.state == 4 ? "SUSPENDED" : "INACTIVE")))) : "INACTIVE")`},
	{Key: entityops.InputKeyDirectoryAccountAccountType, Expr: `dyn('human' in payload ? "USER" : "SERVICE")`},
	{Key: entityops.InputKeyDirectoryAccountAddedAt, Expr: `'details' in payload && 'creation_date' in payload.details ? payload.details.creation_date : null`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
})

// zitadelMappings returns the built-in Zitadel ingest mappings
func zitadelMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: entityops.SchemaDirectoryAccount.Name,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryAccount,
			},
		},
	}
}
