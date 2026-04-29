package slack

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Slack workspace user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'email' in payload ? payload.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'display_name' in payload && payload.display_name != "" ? payload.display_name : ('real_name' in payload && payload.real_name != "" ? payload.real_name : ('name' in payload ? payload.name : ""))`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, Expr: `'first_name' in payload ? payload.first_name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, Expr: `'last_name' in payload ? payload.last_name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountJobTitle, Expr: `'title' in payload ? payload.title : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAvatarRemoteURL, Expr: `'avatar_url' in payload ? payload.avatar_url : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountMfaState, Expr: `dyn(payload.has_2fa ? "ENABLED" : "DISABLED")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn(payload.deleted ? "INACTIVE" : "ACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDirectoryInstanceID, Expr: `payload.team_id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAccountType, Expr: `payload.is_bot ? "SERVICE" : payload.is_external ? "GUEST" : "USER"`},
})

// slackMappings returns the built-in Slack ingest mappings
func slackMappings() []types.MappingRegistration {
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
