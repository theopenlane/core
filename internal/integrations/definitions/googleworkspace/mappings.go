package googleworkspace

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Google Workspace user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'primaryEmail' in payload ? payload.primaryEmail : ""`},
	{Key: entityops.InputKeyDirectoryAccountPhoneNumber, Expr: `'recoveryPhone' in payload ? payload.recoveryPhone : ""`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'name' in payload && 'fullName' in payload.name ? payload.name.fullName : ""`},
	{Key: entityops.InputKeyDirectoryAccountGivenName, Expr: `'name' in payload && 'givenName' in payload.name ? payload.name.givenName : ""`},
	{Key: entityops.InputKeyDirectoryAccountFamilyName, Expr: `'name' in payload && 'familyName' in payload.name ? payload.name.familyName : ""`},
	{Key: entityops.InputKeyDirectoryAccountDirectoryInstanceID, Expr: `'customerId' in payload ? payload.customerId : ""`},

	{Key: entityops.InputKeyDirectoryAccountOrganizationUnit, Expr: `'orgUnitPath' in payload ? payload.orgUnitPath : ""`},
	{Key: entityops.InputKeyDirectoryAccountStatus, Expr: `dyn('deletionTime' in payload && payload.deletionTime != "" ? "DELETED" : ('suspended' in payload && payload.suspended ? "SUSPENDED" : ('archived' in payload && payload.archived ? "INACTIVE" : "ACTIVE")))`},
	{Key: entityops.InputKeyDirectoryAccountDepartment, Expr: `'organizations' in payload && size(payload.organizations.filter(o, ('primary' in o) && o.primary == true)) > 0 ? ('department' in payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0] ? payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0].department : "") : ""`},
	{Key: entityops.InputKeyDirectoryAccountJobTitle, Expr: `'organizations' in payload && size(payload.organizations.filter(o, ('primary' in o) && o.primary == true)) > 0 ? ('title' in payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0] ? payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0].title : "") : ""`},
	{Key: entityops.InputKeyDirectoryAccountAccountType, Expr: `'type' in payload ? payload.type : "USER"`},
	{Key: entityops.InputKeyDirectoryAccountMfaState, Expr: `dyn('isEnforcedIn2Sv' in payload && payload.isEnforcedIn2Sv ? "ENFORCED" : ('isEnrolledIn2Sv' in payload && payload.isEnrolledIn2Sv ? "ENABLED" : "DISABLED"))`},
	{Key: entityops.InputKeyDirectoryAccountLastLoginAt, Expr: `'lastLoginTime' in payload ? payload.lastLoginTime : ""`},
	{Key: entityops.InputKeyDirectoryAccountEmailAliases, Expr: `dyn('emails' in payload ? payload.emails.filter(e, !('primary' in e) || e.primary != true).map(e, e.address) : [])`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
	{Key: entityops.InputKeyDirectoryAccountAddedAt, Expr: `'creationTime' in payload ? payload.creationTime : ""`},
	{Key: entityops.InputKeyDirectoryAccountAvatarRemoteURL, Expr: `'thumbnailPhotoUrl' in payload ? payload.thumbnailPhotoUrl : null`},
})

// mapExprDirectoryGroup is the CEL mapping expression for Google Workspace group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryGroupEmail, Expr: `'email' in payload ? payload.email : ""`},
	{Key: entityops.InputKeyDirectoryGroupDisplayName, Expr: `'name' in payload ? payload.name : ""`},
	{Key: entityops.InputKeyDirectoryGroupClassification, Expr: `dyn('adminCreated' in payload && payload.adminCreated ? "TEAM" : "DISTRIBUTION")`},
	{Key: entityops.InputKeyDirectoryGroupStatus, Expr: `"ACTIVE"`},
	{Key: entityops.InputKeyDirectoryGroupSourceVersion, Expr: `'etag' in payload ? payload.etag : ""`},
	{Key: entityops.InputKeyDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Google Workspace membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryMembershipDirectoryAccountID, Expr: `'email' in payload ? payload.email : ""`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryGroupID, Expr: `resource`},
	{Key: entityops.InputKeyDirectoryMembershipRole, Expr: `dyn('role' in payload && payload.role != "" ? payload.role : "MEMBER")`},
	{Key: entityops.InputKeyDirectoryMembershipMetadata, Expr: "payload"},
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
