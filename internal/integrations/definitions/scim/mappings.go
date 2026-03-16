package scim

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
	{key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, expr: `'externalId' in payload && payload.externalId != "" ? payload.externalId : ('userName' in payload && payload.userName != "" ? payload.userName : ('emails' in payload && size(payload.emails) > 0 && payload.emails[0] != null && 'value' in payload.emails[0] ? payload.emails[0].value : ""))`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, expr: `'emails' in payload && size(payload.emails) > 0 && payload.emails[0] != null && 'value' in payload.emails[0] && payload.emails[0].value != "" ? payload.emails[0].value : ('userName' in payload ? payload.userName : "")`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, expr: `'displayName' in payload && payload.displayName != "" ? payload.displayName : ('name' in payload && payload.name != null && 'givenName' in payload.name && 'familyName' in payload.name ? payload.name.givenName + " " + payload.name.familyName : ('userName' in payload ? payload.userName : ""))`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, expr: `'name' in payload && payload.name != null && 'givenName' in payload.name ? payload.name.givenName : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, expr: `'name' in payload && payload.name != null && 'familyName' in payload.name ? payload.name.familyName : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, expr: `dyn(action == "delete" ? "DELETED" : ('active' in payload ? (payload.active ? "ACTIVE" : "INACTIVE") : "ACTIVE"))`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountMfaState, expr: `dyn("UNKNOWN")`},
	{key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, expr: "payload"},
})

var mapExprDirectoryGroup = celMapExpr([]celMapEntry{
	{key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, expr: `'externalId' in payload && payload.externalId != "" ? payload.externalId : ('displayName' in payload ? payload.displayName : "")`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, expr: `'displayName' in payload ? payload.displayName : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupClassification, expr: `dyn("TEAM")`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupStatus, expr: `dyn(action == "delete" ? "DELETED" : ('active' in payload ? (payload.active ? "ACTIVE" : "INACTIVE") : "ACTIVE"))`},
	{key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, expr: "payload"},
})

var mapExprDirectoryMembership = celMapExpr([]celMapEntry{
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, expr: `'member' in payload && payload.member != null && 'value' in payload.member ? payload.member.value : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, expr: `'group' in payload && payload.group != null && 'externalId' in payload.group ? payload.group.externalId : ""`},
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, expr: `dyn("MEMBER")`},
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipSource, expr: `dyn("scim")`},
	{key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, expr: "payload"},
})

// scimMappings returns the built-in SCIM ingest mappings
func scimMappings() []types.MappingRegistration {
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
