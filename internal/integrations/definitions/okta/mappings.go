package okta

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Okta user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'profile' in payload && payload.profile != null && 'email' in payload.profile ? payload.profile.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'profile' in payload && payload.profile != null && 'displayName' in payload.profile && payload.profile.displayName != null && payload.profile.displayName != "" ? payload.profile.displayName : ('profile' in payload && payload.profile != null && 'login' in payload.profile ? payload.profile.login : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, Expr: `'profile' in payload && payload.profile != null && 'firstName' in payload.profile && payload.profile.firstName != null ? payload.profile.firstName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, Expr: `'profile' in payload && payload.profile != null && 'lastName' in payload.profile && payload.profile.lastName != null ? payload.profile.lastName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDepartment, Expr: `'profile' in payload && payload.profile != null && 'department' in payload.profile ? payload.profile.department : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountJobTitle, Expr: `'profile' in payload && payload.profile != null && 'title' in payload.profile && payload.profile.title != null ? payload.profile.title : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('status' in payload ? (payload.status == "DEPROVISIONED" ? "DELETED" : (payload.status == "SUSPENDED" ? "SUSPENDED" : (payload.status == "STAGED" || payload.status == "PROVISIONED" ? "INACTIVE" : "ACTIVE"))) : "ACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountLastLoginAt, Expr: `'lastLogin' in payload && payload.lastLogin != null ? payload.lastLogin : null`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup is the CEL mapping expression for Okta group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupEmail, Expr: `""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, Expr: `'profile' in payload && payload.profile != null && 'name' in payload.profile ? payload.profile.name : ('id' in payload ? payload.id : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupClassification, Expr: `dyn('type' in payload && payload.type == "OKTA_GROUP" ? "TEAM" : "DISTRIBUTION")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Okta membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, Expr: `'member' in payload && payload.member != null && 'id' in payload.member && payload.member.id != "" ? payload.member.id : ('member' in payload && payload.member != null && 'profile' in payload.member && payload.member.profile != null && 'login' in payload.member.profile ? payload.member.profile.login : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, Expr: `'group' in payload && payload.group != null && 'id' in payload.group ? payload.group.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, Expr: "payload"},
})

// oktaMappings returns the built-in Okta ingest mappings
func oktaMappings() []types.MappingRegistration {
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
