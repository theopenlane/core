package azureentraid

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount maps Azure Entra ID user payloads to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'mail' in payload && payload.mail != "" ? payload.mail : ('userPrincipalName' in payload ? payload.userPrincipalName : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'displayName' in payload && payload.displayName != "" ? payload.displayName : ('userPrincipalName' in payload ? payload.userPrincipalName : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, Expr: `'givenName' in payload ? payload.givenName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, Expr: `'surname' in payload ? payload.surname : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDepartment, Expr: `'department' in payload ? payload.department : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountJobTitle, Expr: `'jobTitle' in payload ? payload.jobTitle : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('accountEnabled' in payload && payload.accountEnabled ? "ACTIVE" : "INACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup maps Azure Entra ID group payloads to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupEmail, Expr: `'mail' in payload ? payload.mail : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, Expr: `'displayName' in payload && payload.displayName != "" ? payload.displayName : ('mail' in payload ? payload.mail : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupClassification, Expr: `dyn('groupTypes' in payload && payload.groupTypes != null && payload.groupTypes.exists(t, t == "Unified") ? "TEAM" : ('securityEnabled' in payload && payload.securityEnabled ? "SECURITY" : "DISTRIBUTION"))`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership maps Azure Entra ID membership payloads to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, Expr: `'member' in payload && payload.member != null && 'id' in payload.member && payload.member.id != "" ? payload.member.id : ('member' in payload && payload.member != null && 'email' in payload.member ? payload.member.email : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, Expr: `'group' in payload && payload.group != null && 'id' in payload.group && payload.group.id != "" ? payload.group.id : ('group' in payload && payload.group != null && 'email' in payload.group ? payload.group.email : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, Expr: "payload"},
})

// entraIDMappings returns the built-in Azure Entra ID ingest mappings
func entraIDMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryAccount,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryGroup,
			},
		},
		{
			Schema: integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryMembership,
			},
		},
	}
}
