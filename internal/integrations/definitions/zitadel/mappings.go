package zitadel

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Zitadel user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'user_id' in payload ? payload.user_id : ""`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'human' in payload && 'email' in payload.human && 'email' in payload.human.email ? payload.human.email.email : ""`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'human' in payload ? ('profile' in payload.human && 'display_name' in payload.human.profile && payload.human.profile.display_name != "" ? payload.human.profile.display_name : ('username' in payload ? payload.username : "")) : ('machine' in payload && 'name' in payload.machine ? payload.machine.name : ('username' in payload ? payload.username : ""))`),
	entityops.DirectoryAccountFields.GivenName.Expr(`'human' in payload && 'profile' in payload.human && 'given_name' in payload.human.profile ? payload.human.profile.given_name : ""`),
	entityops.DirectoryAccountFields.FamilyName.Expr(`'human' in payload && 'profile' in payload.human && 'family_name' in payload.human.profile ? payload.human.profile.family_name : ""`),
	entityops.DirectoryAccountFields.PhoneNumber.Expr(`'human' in payload && 'phone' in payload.human && 'phone' in payload.human.phone ? payload.human.phone.phone : ""`),
	entityops.DirectoryAccountFields.AvatarRemoteURL.Expr(`'human' in payload && 'profile' in payload.human && 'avatar_url' in payload.human.profile ? payload.human.profile.avatar_url : ""`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn('state' in payload ? (payload.state == 1 ? "ACTIVE" : (payload.state == 2 ? "INACTIVE" : (payload.state == 3 ? "DELETED" : (payload.state == 4 ? "SUSPENDED" : "INACTIVE")))) : "INACTIVE")`),
	entityops.DirectoryAccountFields.AccountType.Expr(`dyn('human' in payload ? "USER" : "SERVICE")`),
	entityops.DirectoryAccountFields.AddedAt.Expr(`'details' in payload && 'creation_date' in payload.details ? payload.details.creation_date : null`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.DirectoryName.Expr("installation.name"),
)

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
