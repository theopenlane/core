package authentik

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount is the CEL mapping expression for Authentik user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountExternalID, Expr: `'uid' in payload ? payload.uid : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, Expr: `'email' in payload && payload.email != null ? payload.email : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountDisplayName, Expr: `'name' in payload && payload.name != null && payload.name != "" ? payload.name : ('username' in payload ? payload.username : "")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountStatus, Expr: `dyn('is_active' in payload ? (payload.is_active ? "ACTIVE" : "INACTIVE") : "INACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAccountType, Expr: `dyn('type' in payload && payload.type != null ? (payload.type == "internal" ? "USER" : (payload.type == "external" ? "GUEST" : (payload.type == "service_account" ? "SERVICE" : (payload.type == "internal_service_account" ? "SERVICE" : "USER")))) : "USER")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountAddedAt, Expr: `'date_joined' in payload ? payload.date_joined`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountLastSeenAt, Expr: `'last_login' in payload ? payload.last_login`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountObservedAt, Expr: `'last_updated' in payload ? payload.last_updated`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountMetadata, Expr: `'attributes' in payload ? payload.attributes : {}`},
	{Key: integrationgenerated.IntegrationMappingDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup is the CEL mapping expression for Authentik group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupExternalID, Expr: `'pk' in payload ? payload.pk : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupDisplayName, Expr: `'name' in payload && payload.name != null ? payload.name : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupMetadata, Expr: `'attributes' in payload ? payload.attributes : {}`},
	{Key: integrationgenerated.IntegrationMappingDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Authentik membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID, Expr: `'uid' in payload ? payload.uid : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID, Expr: `resource != "" ? resource : ""`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, Expr: "payload"},
})

// authentikMappings returns the built-in Authentik ingest mappings
func authentikMappings() []types.MappingRegistration {
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
