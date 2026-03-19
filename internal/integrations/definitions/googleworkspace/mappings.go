package googleworkspace

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Google Workspace user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'primaryEmail' in payload ? payload.primaryEmail : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'name' in payload && payload.name != null && 'fullName' in payload.name && payload.name.fullName != "" ? payload.name.fullName : ('primaryEmail' in payload ? payload.primaryEmail : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, Expr: `'name' in payload && payload.name != null && 'givenName' in payload.name ? payload.name.givenName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, Expr: `'name' in payload && payload.name != null && 'familyName' in payload.name ? payload.name.familyName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDirectoryName, Expr: `'customerId' in payload ? payload.customerId : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountOrganizationUnit, Expr: `'orgUnitPath' in payload ? payload.orgUnitPath : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('deletionTime' in payload && payload.deletionTime != "" ? "DELETED" : ('suspended' in payload && payload.suspended ? "SUSPENDED" : ('archived' in payload && payload.archived ? "INACTIVE" : "ACTIVE")))`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountMfaState, Expr: `dyn('isEnforcedIn2Sv' in payload && payload.isEnforcedIn2Sv ? "ENFORCED" : ('isEnrolledIn2Sv' in payload && payload.isEnrolledIn2Sv ? "ENABLED" : "DISABLED"))`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountLastLoginAt, Expr: `'lastLoginTime' in payload && payload.lastLoginTime != "" ? payload.lastLoginTime : null`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup is the CEL mapping expression for Google Workspace group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupEmail, Expr: `'email' in payload ? payload.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, Expr: `'name' in payload ? payload.name : ('email' in payload ? payload.email : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupClassification, Expr: `dyn('adminCreated' in payload && payload.adminCreated ? "TEAM" : "DISTRIBUTION")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupSourceVersion, Expr: `'etag' in payload ? payload.etag : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Google Workspace membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, Expr: `'member' in payload && payload.member != null && 'id' in payload.member && payload.member.id != "" ? payload.member.id : ('member' in payload && payload.member != null && 'email' in payload.member ? payload.member.email : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, Expr: `'group' in payload && payload.group != null && 'id' in payload.group && payload.group.id != "" ? payload.group.id : ('group' in payload && payload.group != null && 'email' in payload.group ? payload.group.email : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, Expr: `dyn('member' in payload && payload.member != null && 'role' in payload.member && payload.member.role != "" ? (payload.member.role == "OWNER" ? "OWNER" : (payload.member.role == "MANAGER" ? "MANAGER" : "MEMBER")) : "MEMBER")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipSource, Expr: `dyn("google_workspace")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, Expr: "payload"},
})

// googleWorkspaceMappings returns the built-in Google Workspace ingest mappings
func googleWorkspaceMappings() []types.MappingRegistration {
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
