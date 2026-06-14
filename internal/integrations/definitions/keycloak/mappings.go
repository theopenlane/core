package keycloak

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Keycloak user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'email' in payload && payload.email != "" ? payload.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'username' in payload ? payload.username : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountGivenName, Expr: `'firstName' in payload ? payload.firstName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountFamilyName, Expr: `'lastName' in payload ? payload.lastName : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('enabled' in payload ? (payload.enabled ? "ACTIVE" : "INACTIVE") : "INACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAccountType, Expr: `dyn('serviceAccountClientId' in payload && payload.serviceAccountClientId != "" ? "SERVICE" : "USER")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAddedAt, Expr: `'createdTimestamp' in payload ? timestamp(int(payload.createdTimestamp) / 1000) : null`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountMetadata, Expr: `'attributes' in payload ? payload.attributes : {}`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup is the CEL mapping expression for Keycloak group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, Expr: `'name' in payload ? payload.name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupMetadata, Expr: `'attributes' in payload ? payload.attributes : {}`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Keycloak membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, Expr: `resource != "" ? resource : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, Expr: "payload"},
})

// keycloakMappings returns the built-in Keycloak ingest mappings
func keycloakMappings() []types.MappingRegistration {
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
