package azureentraid

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// mapExprDirectoryAccount maps Azure Entra ID user payloads to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'mail' in payload && payload.mail != "" ? payload.mail : ('userPrincipalName' in payload ? payload.userPrincipalName : "")`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'displayName' in payload && payload.displayName != "" ? payload.displayName : ('userPrincipalName' in payload ? payload.userPrincipalName : "")`},
	{Key: entityops.InputKeyDirectoryAccountGivenName, Expr: `'givenName' in payload ? payload.givenName : ""`},
	{Key: entityops.InputKeyDirectoryAccountFamilyName, Expr: `'surname' in payload ? payload.surname : ""`},
	{Key: entityops.InputKeyDirectoryAccountDepartment, Expr: `'department' in payload ? payload.department : ""`},
	{Key: entityops.InputKeyDirectoryAccountJobTitle, Expr: `'jobTitle' in payload ? payload.jobTitle : ""`},
	{Key: entityops.InputKeyDirectoryAccountEmailAliases, Expr: `'otherMails' in payload && payload.otherMails != null ? payload.otherMails : []`},
	{Key: entityops.InputKeyDirectoryAccountPhoneNumber, Expr: `'phone' in payload ? payload.phone : ""`},
	{Key: entityops.InputKeyDirectoryAccountAddedAt, Expr: `'employeeHireDate' in payload && payload.employeeHireDate != null ? payload.employeeHireDate : null`},
	{Key: entityops.InputKeyDirectoryAccountRemovedAt, Expr: `'employeeLeaveDateTime' in payload && payload.employeeLeaveDateTime != null ? payload.employeeLeaveDateTime : null`},
	{Key: entityops.InputKeyDirectoryAccountStatus, Expr: `dyn('accountEnabled' in payload && payload.accountEnabled ? "ACTIVE" : "INACTIVE")`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup maps Azure Entra ID group payloads to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryGroupExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryGroupEmail, Expr: `'mail' in payload ? payload.mail : ""`},
	{Key: entityops.InputKeyDirectoryGroupDisplayName, Expr: `'displayName' in payload && payload.displayName != "" ? payload.displayName : ('mail' in payload ? payload.mail : "")`},
	{Key: entityops.InputKeyDirectoryGroupClassification, Expr: `dyn('groupTypes' in payload && payload.groupTypes != null && payload.groupTypes.exists(t, t == "Unified") ? "TEAM" : ('securityEnabled' in payload && payload.securityEnabled ? "SECURITY" : "DISTRIBUTION"))`},
	{Key: entityops.InputKeyDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: entityops.InputKeyDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership maps Azure Entra ID membership payloads to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryMembershipDirectoryAccountID, Expr: `'member' in payload && payload.member != null && 'id' in payload.member && payload.member.id != "" ? payload.member.id : ('member' in payload && payload.member != null && 'email' in payload.member ? payload.member.email : "")`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryGroupID, Expr: `'group' in payload && payload.group != null && 'id' in payload.group && payload.group.id != "" ? payload.group.id : ('group' in payload && payload.group != null && 'email' in payload.group ? payload.group.email : "")`},
	{Key: entityops.InputKeyDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: entityops.InputKeyDirectoryMembershipMetadata, Expr: "payload"},
})

// entraIDMappings returns the built-in Azure Entra ID ingest mappings
func entraIDMappings() []types.MappingRegistration {
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
