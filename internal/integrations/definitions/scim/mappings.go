package scim

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for SCIM user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `'externalId' in payload && payload.externalId != "" ? payload.externalId : ('userName' in payload && payload.userName != "" ? payload.userName : ('emails' in payload && size(payload.emails) > 0 && payload.emails[0] != null && 'value' in payload.emails[0] ? payload.emails[0].value : ""))`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'emails' in payload && size(payload.emails) > 0 && payload.emails[0] != null && 'value' in payload.emails[0] && payload.emails[0].value != "" ? payload.emails[0].value : ('userName' in payload ? payload.userName : "")`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'displayName' in payload && payload.displayName != "" ? payload.displayName : ('name' in payload && payload.name != null && 'givenName' in payload.name && 'familyName' in payload.name ? payload.name.givenName + " " + payload.name.familyName : ('userName' in payload ? payload.userName : ""))`},
	{Key: entityops.InputKeyDirectoryAccountGivenName, Expr: `'name' in payload && payload.name != null && 'givenName' in payload.name ? payload.name.givenName : ""`},
	{Key: entityops.InputKeyDirectoryAccountFamilyName, Expr: `'name' in payload && payload.name != null && 'familyName' in payload.name ? payload.name.familyName : ""`},
	{Key: entityops.InputKeyDirectoryAccountStatus, Expr: `dyn(action == "delete" ? "DELETED" : ('active' in payload ? (payload.active ? "ACTIVE" : "INACTIVE") : "ACTIVE"))`},
	{Key: entityops.InputKeyDirectoryAccountMfaState, Expr: `dyn("UNKNOWN")`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup is the CEL mapping expression for SCIM group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryGroupExternalID, Expr: `'externalId' in payload && payload.externalId != "" ? payload.externalId : ('displayName' in payload ? payload.displayName : "")`},
	{Key: entityops.InputKeyDirectoryGroupDisplayName, Expr: `'displayName' in payload ? payload.displayName : ""`},
	{Key: entityops.InputKeyDirectoryGroupClassification, Expr: `dyn("TEAM")`},
	{Key: entityops.InputKeyDirectoryGroupStatus, Expr: `dyn(action == "delete" ? "DELETED" : ('active' in payload ? (payload.active ? "ACTIVE" : "INACTIVE") : "ACTIVE"))`},
	{Key: entityops.InputKeyDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for SCIM group membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryMembershipDirectoryAccountID, Expr: `'member' in payload && payload.member != null && 'value' in payload.member ? payload.member.value : ""`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryGroupID, Expr: `'group' in payload && payload.group != null && 'externalId' in payload.group ? payload.group.externalId : ""`},
	{Key: entityops.InputKeyDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: entityops.InputKeyDirectoryMembershipMetadata, Expr: "payload"},
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
