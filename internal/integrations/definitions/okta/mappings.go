package okta

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Okta user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'profile' in payload && payload.profile != null && 'email' in payload.profile ? payload.profile.email : ""`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'profile' in payload && payload.profile != null && 'displayName' in payload.profile && payload.profile.displayName != null && payload.profile.displayName != "" ? payload.profile.displayName : ('profile' in payload && payload.profile != null && 'login' in payload.profile ? payload.profile.login : "")`),
	entityops.DirectoryAccountFields.GivenName.Expr(`'profile' in payload && payload.profile != null && 'firstName' in payload.profile && payload.profile.firstName != null ? payload.profile.firstName : ""`),
	entityops.DirectoryAccountFields.FamilyName.Expr(`'profile' in payload && payload.profile != null && 'lastName' in payload.profile && payload.profile.lastName != null ? payload.profile.lastName : ""`),
	entityops.DirectoryAccountFields.Department.Expr(`'profile' in payload && payload.profile != null && 'department' in payload.profile ? payload.profile.department : ""`),
	entityops.DirectoryAccountFields.JobTitle.Expr(`'profile' in payload && payload.profile != null && 'title' in payload.profile && payload.profile.title != null ? payload.profile.title : ""`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn('status' in payload ? (payload.status == "DEPROVISIONED" ? "DELETED" : (payload.status == "SUSPENDED" ? "SUSPENDED" : (payload.status == "STAGED" || payload.status == "PROVISIONED" ? "INACTIVE" : "ACTIVE"))) : "ACTIVE")`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.DirectoryName.Expr(providerkit.ExprInstallationName),
	entityops.DirectoryAccountFields.PrimarySource.Expr("installation.primary_directory"),
)

// mapExprDirectoryGroup is the CEL mapping expression for Okta group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryGroupFields.Email.Expr(`""`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`'profile' in payload && payload.profile != null && 'name' in payload.profile ? payload.profile.name : ('id' in payload ? payload.id : "")`),
	entityops.DirectoryGroupFields.Classification.Expr(`dyn('type' in payload && payload.type == "OKTA_GROUP" ? "TEAM" : "DISTRIBUTION")`),
	entityops.DirectoryGroupFields.Status.Expr(`dyn("ACTIVE")`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
	entityops.DirectoryGroupFields.DirectoryName.Expr(providerkit.ExprInstallationName),
)

// mapExprDirectoryMembership is the CEL mapping expression for Okta membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`'member' in payload && payload.member != null && 'id' in payload.member ? payload.member.id : ""`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`'group' in payload && payload.group != null && 'id' in payload.group ? payload.group.id : ""`),
	entityops.DirectoryMembershipFields.Role.Expr(`dyn("MEMBER")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
	entityops.DirectoryMembershipFields.DirectoryName.Expr(providerkit.ExprInstallationName),
)
