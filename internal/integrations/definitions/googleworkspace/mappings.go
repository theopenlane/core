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
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountPhoneNumber, Expr: `'phones' in payload && size(payload.phones) > 0
  ? (
      size(payload.phones.filter(p, p.type == "work")) > 0
        ? payload.phones.filter(p, p.type == "work")[0].value
        : payload.phones[0].value
    )
  : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'name' in payload && 'fullName' in payload.name ? payload.name.fullName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, Expr: `'name' in payload && 'givenName' in payload.name ? payload.name.givenName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, Expr: `'name' in payload && 'familyName' in payload.name ? payload.name.familyName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDirectoryInstanceID, Expr: `'customerId' in payload ? payload.customerId : ""`},

	{Key: integrationgenerated.IntegrationMappingDirectoryAccountOrganizationUnit, Expr: `'orgUnitPath' in payload ? payload.orgUnitPath : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('deletionTime' in payload && payload.deletionTime != "" ? "DELETED" : ('suspended' in payload && payload.suspended ? "SUSPENDED" : ('archived' in payload && payload.archived ? "INACTIVE" : "ACTIVE")))`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDepartment, Expr: `'organizations' in payload && size(payload.organizations.filter(o, ('primary' in o) && o.primary == true)) > 0 ? ('department' in payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0] ? payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0].department : "") : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountJobTitle, Expr: `'organizations' in payload && size(payload.organizations.filter(o, ('primary' in o) && o.primary == true)) > 0 ? ('title' in payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0] ? payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0].title : "") : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAccountType, Expr: `'type' in payload ? payload.type : "USER"`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountMfaState, Expr: `dyn('isEnforcedIn2Sv' in payload && payload.isEnforcedIn2Sv ? "ENFORCED" : ('isEnrolledIn2Sv' in payload && payload.isEnrolledIn2Sv ? "ENABLED" : "DISABLED"))`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountLastLoginAt, Expr: `'lastLoginTime' in payload ? payload.lastLoginTime : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountEmailAliases, Expr: `dyn('emails' in payload ? payload.emails.filter(e, !('primary' in e) || e.primary != true).map(e, e.address) : [])`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAddedAt, Expr: `'creationTime' in payload ? payload.creationTime : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAvatarRemoteURL, Expr: `'thumbnailPhotoUrl' in payload ? payload.thumbnailPhotoUrl : null`},
})

// mapExprDirectoryGroup is the CEL mapping expression for Google Workspace group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupEmail, Expr: `'email' in payload ? payload.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, Expr: `'name' in payload ? payload.name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupClassification, Expr: `dyn('adminCreated' in payload && payload.adminCreated ? "TEAM" : "DISTRIBUTION")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupStatus, Expr: `"ACTIVE"`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupSourceVersion, Expr: `'etag' in payload ? payload.etag : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Google Workspace membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, Expr: `'email' in payload ? payload.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, Expr: `resource`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, Expr: `'role' in payload ? payload.role : ""`},
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
