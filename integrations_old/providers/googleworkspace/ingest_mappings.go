package googleworkspace

import (
	"strconv"
	"strings"

	"github.com/samber/lo"

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

// mappingSchemaDirectoryAccount is the schema identifier for directory account ingest
const mappingSchemaDirectoryAccount = types.MappingSchema(integrationgenerated.IntegrationMappingSchemaDirectoryAccount)

// googleWorkspaceDirectoryAccountMappings returns the built-in directory account mapping specs for Google Workspace
func googleWorkspaceDirectoryAccountMappings() map[string]types.MappingOverride {
	return map[string]types.MappingOverride{
		"": {
			FilterExpr: `payload.id != "" || payload.primaryEmail != ""`,
			MapExpr:    mapExprGoogleWorkspaceDirectoryAccount,
		},
	}
}

// googleWorkspaceDefaultMappings returns all built-in ingest mappings for Google Workspace
func googleWorkspaceDefaultMappings() []types.MappingRegistration {
	return lo.MapToSlice(googleWorkspaceDirectoryAccountMappings(), func(variant string, spec types.MappingOverride) types.MappingRegistration {
		return types.MappingRegistration{
			Schema:  mappingSchemaDirectoryAccount,
			Variant: variant,
			Spec:    spec,
		}
	})
}
