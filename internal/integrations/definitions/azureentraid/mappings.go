package azureentraid

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprDirectoryAccount maps Azure Entra ID user payloads to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'mail' in payload && payload.mail != "" ? payload.mail : ('userPrincipalName' in payload ? payload.userPrincipalName : "")`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'displayName' in payload && payload.displayName != "" ? payload.displayName : ('userPrincipalName' in payload ? payload.userPrincipalName : "")`),
	entityops.DirectoryAccountFields.GivenName.Expr(`'givenName' in payload ? payload.givenName : ""`),
	entityops.DirectoryAccountFields.FamilyName.Expr(`'surname' in payload ? payload.surname : ""`),
	entityops.DirectoryAccountFields.Department.Expr(`'department' in payload ? payload.department : ""`),
	entityops.DirectoryAccountFields.JobTitle.Expr(`'jobTitle' in payload ? payload.jobTitle : ""`),
	entityops.DirectoryAccountFields.EmailAliases.Expr(`'otherMails' in payload && payload.otherMails != null ? payload.otherMails : []`),
	entityops.DirectoryAccountFields.PhoneNumber.Expr(`'phone' in payload ? payload.phone : ""`),
	entityops.DirectoryAccountFields.AddedAt.Expr(`'employeeHireDate' in payload && payload.employeeHireDate != null ? payload.employeeHireDate : null`),
	entityops.DirectoryAccountFields.RemovedAt.Expr(`'employeeLeaveDateTime' in payload && payload.employeeLeaveDateTime != null ? payload.employeeLeaveDateTime : null`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn('accountEnabled' in payload && payload.accountEnabled ? "ACTIVE" : "INACTIVE")`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.DirectoryName.Expr(providerkit.ExprInstallationName),
	entityops.DirectoryAccountFields.PrimarySource.Expr("installation.primary_directory"),
)

// mapExprDirectoryGroup maps Azure Entra ID group payloads to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryGroupFields.Email.Expr(`'mail' in payload ? payload.mail : ""`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`'displayName' in payload && payload.displayName != "" ? payload.displayName : ('mail' in payload ? payload.mail : "")`),
	entityops.DirectoryGroupFields.Classification.Expr(`dyn('groupTypes' in payload && payload.groupTypes != null && payload.groupTypes.exists(t, t == "Unified") ? "TEAM" : ('securityEnabled' in payload && payload.securityEnabled ? "SECURITY" : "DISTRIBUTION"))`),
	entityops.DirectoryGroupFields.Status.Expr(`dyn("ACTIVE")`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
	entityops.DirectoryGroupFields.DirectoryName.Expr(providerkit.ExprInstallationName),
)

// mapExprDirectoryMembership maps Azure Entra ID membership payloads to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`'member' in payload && payload.member != null && 'id' in payload.member ? payload.member.id : ""`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`'group' in payload && payload.group != null && 'id' in payload.group ? payload.group.id : ""`),
	entityops.DirectoryMembershipFields.Role.Expr(`dyn("MEMBER")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
	entityops.DirectoryMembershipFields.DirectoryName.Expr(providerkit.ExprInstallationName),
)
