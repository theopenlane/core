package ingest

import (
	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	googleworkspaceprovider "github.com/theopenlane/core/internal/integrations/providers/googleworkspace"
)

var normalizedDirectoryAccountSchema = normalizeMappingKey(mappingSchemaDirectoryAccount)

var mapExprGoogleWorkspaceDirectoryAccount = celMapExpr([]celMapEntry{
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountExternalID,
		expr: `payload.id != "" ? payload.id : payload.primaryEmail`,
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail,
		expr: "payload.primaryEmail",
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountDisplayName,
		expr: "payload.name.fullName",
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountGivenName,
		expr: "payload.name.givenName",
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountFamilyName,
		expr: "payload.name.familyName",
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountAvatarRemoteURL,
		expr: "payload.thumbnailPhotoUrl",
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountJobTitle,
		expr: `payload.organizations != null && size(payload.organizations) > 0 ? payload.organizations[0].title : ""`,
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountDepartment,
		expr: `payload.organizations != null && size(payload.organizations) > 0 ? payload.organizations[0].department : ""`,
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountOrganizationUnit,
		expr: "payload.orgUnitPath",
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountAccountType,
		expr: `"USER"`,
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountStatus,
		expr: `payload.suspended == true ? "SUSPENDED" : (payload.archived == true ? "INACTIVE" : "ACTIVE")`,
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountMfaState,
		expr: `payload.isEnforcedIn2Sv == true ? "ENFORCED" : (payload.isEnrolledIn2Sv == true ? "ENABLED" : "UNKNOWN")`,
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountLastLoginAt,
		expr: "payload.lastLoginTime",
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountObservedAt,
		expr: "payload.observedAt",
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountDirectoryName,
		expr: `"googleworkspace"`,
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountProfile,
		expr: "payload",
	},
	{
		key:  integrationgenerated.IntegrationMappingDirectoryAccountProfileHash,
		expr: `payload.id != "" ? payload.id : payload.primaryEmail`,
	},
})

var googleWorkspaceDirectoryAccountMapping = openapi.IntegrationMappingOverride{
	FilterExpr: `payload.id != "" || payload.primaryEmail != ""`,
	MapExpr:    mapExprGoogleWorkspaceDirectoryAccount,
}

// directoryAccountMappingSpec selects built-in directory account mappings for supported providers
func directoryAccountMappingSpec(provider integrationtypes.ProviderType, variant string) (openapi.IntegrationMappingOverride, bool) {
	_ = variant

	switch provider {
	case googleworkspaceprovider.TypeGoogleWorkspace:
		return googleWorkspaceDirectoryAccountMapping, true
	default:
		return openapi.IntegrationMappingOverride{}, false
	}
}
