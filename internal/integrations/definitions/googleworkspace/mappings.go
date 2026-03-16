package googleworkspace

import (
	"strconv"
	"strings"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/types"
)

type celMapEntry struct {
	key  string
	expr string
}

// celMapExpr renders CEL map entries into a CEL object literal string
func celMapExpr(entries []celMapEntry) string {
	if len(entries) == 0 {
		return "{}"
	}

	var b strings.Builder

	b.WriteString("{\n")

	for i, entry := range entries {
		b.WriteString("  ")
		b.WriteString(strconv.Quote(entry.key))
		b.WriteString(": ")
		b.WriteString(entry.expr)

		if i < len(entries)-1 {
			b.WriteString(",")
		}

		b.WriteString("\n")
	}

	b.WriteString("}")

	return b.String()
}

var mapExprDirectoryAccount = celMapExpr([]celMapEntry{
	{key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, expr: `'id' in payload ? payload.id : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, expr: `'primaryEmail' in payload ? payload.primaryEmail : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, expr: `'name' in payload && payload.name != null && 'fullName' in payload.name && payload.name.fullName != "" ? payload.name.fullName : ('primaryEmail' in payload ? payload.primaryEmail : "")`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, expr: `'name' in payload && payload.name != null && 'givenName' in payload.name ? payload.name.givenName : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, expr: `'name' in payload && payload.name != null && 'familyName' in payload.name ? payload.name.familyName : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountDirectoryName, expr: `'customerId' in payload ? payload.customerId : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountOrganizationUnit, expr: `'orgUnitPath' in payload ? payload.orgUnitPath : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, expr: `dyn('deletionTime' in payload && payload.deletionTime != "" ? "DELETED" : ('suspended' in payload && payload.suspended ? "SUSPENDED" : ('archived' in payload && payload.archived ? "INACTIVE" : "ACTIVE")))`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountMfaState, expr: `dyn('isEnforcedIn2Sv' in payload && payload.isEnforcedIn2Sv ? "ENFORCED" : ('isEnrolledIn2Sv' in payload && payload.isEnrolledIn2Sv ? "ENABLED" : "DISABLED"))`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountLastLoginAt, expr: `'lastLoginTime' in payload && payload.lastLoginTime != "" ? payload.lastLoginTime : null`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, expr: "payload"},
})

var mapExprDirectoryGroup = celMapExpr([]celMapEntry{
	{key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, expr: `'id' in payload ? payload.id : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupEmail, expr: `'email' in payload ? payload.email : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, expr: `'name' in payload ? payload.name : ('email' in payload ? payload.email : "")`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupClassification, expr: `dyn('adminCreated' in payload && payload.adminCreated ? "TEAM" : "DISTRIBUTION")`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupStatus, expr: `dyn("ACTIVE")`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupSourceVersion, expr: `'etag' in payload ? payload.etag : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, expr: "payload"},
})

var mapExprDirectoryMembership = celMapExpr([]celMapEntry{
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, expr: `'member' in payload && payload.member != null && 'id' in payload.member && payload.member.id != "" ? payload.member.id : ('member' in payload && payload.member != null && 'email' in payload.member ? payload.member.email : "")`},
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, expr: `'group' in payload && payload.group != null && 'id' in payload.group && payload.group.id != "" ? payload.group.id : ('group' in payload && payload.group != null && 'email' in payload.group ? payload.group.email : "")`},
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, expr: `dyn('member' in payload && payload.member != null && 'role' in payload.member && payload.member.role != "" ? (payload.member.role == "OWNER" ? "OWNER" : (payload.member.role == "MANAGER" ? "MANAGER" : "MEMBER")) : "MEMBER")`},
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipSource, expr: `dyn("google_workspace")`},
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, expr: "payload"},
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
