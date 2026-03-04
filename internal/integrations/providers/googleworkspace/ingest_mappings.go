package googleworkspace

import (
	"strconv"
	"strings"

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
)

type celMapEntry struct {
	key  string
	expr string
}

// celMapExpr renders CEL map entries into a CEL object literal string.
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

// googleWorkspaceDirectoryAccountMappings returns the built-in directory account mapping specs for Google Workspace.
func googleWorkspaceDirectoryAccountMappings() map[string]integrationtypes.MappingSpec {
	return map[string]integrationtypes.MappingSpec{
		"": {
			FilterExpr: `payload.id != "" || payload.primaryEmail != ""`,
			MapExpr:    mapExprGoogleWorkspaceDirectoryAccount,
		},
	}
}
