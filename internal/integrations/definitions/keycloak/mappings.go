package keycloak

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Keycloak user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'email' in payload && payload.email != "" ? payload.email : ""`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'username' in payload ? payload.username : ""`},
	{Key: entityops.InputKeyDirectoryAccountGivenName, Expr: `'firstName' in payload ? payload.firstName : ""`},
	{Key: entityops.InputKeyDirectoryAccountFamilyName, Expr: `'lastName' in payload ? payload.lastName : ""`},
	{Key: entityops.InputKeyDirectoryAccountStatus, Expr: `dyn('enabled' in payload ? (payload.enabled ? "ACTIVE" : "INACTIVE") : "INACTIVE")`},
	{Key: entityops.InputKeyDirectoryAccountAccountType, Expr: `dyn('serviceAccountClientId' in payload && payload.serviceAccountClientId != "" ? "SERVICE" : "USER")`},
	{Key: entityops.InputKeyDirectoryAccountAddedAt, Expr: `'createdTimestamp' in payload ? timestamp(int(payload.createdTimestamp) / 1000) : null`},
	{Key: entityops.InputKeyDirectoryAccountMetadata, Expr: `'attributes' in payload ? payload.attributes : {}`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
	{Key: entityops.InputKeyDirectoryAccountLastSeenAt, Expr: `'lastLogin' in payload && payload.lastLogin != null ? timestamp(int(payload.lastLogin) / 1000) : null`},
})

// mapExprDirectoryGroup is the CEL mapping expression for Keycloak group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryGroupDisplayName, Expr: `'name' in payload ? payload.name : ""`},
	{Key: entityops.InputKeyDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: entityops.InputKeyDirectoryGroupMetadata, Expr: `'attributes' in payload ? payload.attributes : {}`},
	{Key: entityops.InputKeyDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Keycloak membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryMembershipDirectoryAccountID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryGroupID, Expr: `resource != "" ? resource : ""`},
	{Key: entityops.InputKeyDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: entityops.InputKeyDirectoryMembershipMetadata, Expr: "payload"},
})

// keycloakMappings returns the built-in Keycloak ingest mappings
func keycloakMappings() []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: entityops.SchemaDirectoryAccount.Name,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryAccount,
			},
		},
		{
			Schema: entityops.SchemaDirectoryGroup.Name,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryGroup,
			},
		},
		{
			Schema: entityops.SchemaDirectoryMembership.Name,
			Spec: types.MappingOverride{
				FilterExpr: "true",
				MapExpr:    mapExprDirectoryMembership,
			},
		},
	}
}
