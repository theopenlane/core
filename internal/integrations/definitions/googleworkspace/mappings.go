package googleworkspace

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Google Workspace user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'primaryEmail' in payload ? payload.primaryEmail : ""`),
	entityops.DirectoryAccountFields.PhoneNumber.Expr(`'recoveryPhone' in payload ? payload.recoveryPhone : ""`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'name' in payload && 'fullName' in payload.name ? payload.name.fullName : ""`),
	entityops.DirectoryAccountFields.GivenName.Expr(`'name' in payload && 'givenName' in payload.name ? payload.name.givenName : ""`),
	entityops.DirectoryAccountFields.FamilyName.Expr(`'name' in payload && 'familyName' in payload.name ? payload.name.familyName : ""`),

	entityops.DirectoryAccountFields.OrganizationUnit.Expr(`'orgUnitPath' in payload ? payload.orgUnitPath : ""`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn('deletionTime' in payload && payload.deletionTime != "" ? "DELETED" : ('suspended' in payload && payload.suspended ? "SUSPENDED" : ('archived' in payload && payload.archived ? "INACTIVE" : "ACTIVE")))`),
	entityops.DirectoryAccountFields.Department.Expr(`'organizations' in payload && size(payload.organizations.filter(o, ('primary' in o) && o.primary == true)) > 0 ? ('department' in payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0] ? payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0].department : "") : ""`),
	entityops.DirectoryAccountFields.JobTitle.Expr(`'organizations' in payload && size(payload.organizations.filter(o, ('primary' in o) && o.primary == true)) > 0 ? ('title' in payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0] ? payload.organizations.filter(o, ('primary' in o) && o.primary == true)[0].title : "") : ""`),
	entityops.DirectoryAccountFields.AccountType.Expr(`'type' in payload ? payload.type : "USER"`),
	entityops.DirectoryAccountFields.MfaState.Expr(`dyn('isEnforcedIn2Sv' in payload && payload.isEnforcedIn2Sv ? "ENFORCED" : ('isEnrolledIn2Sv' in payload && payload.isEnrolledIn2Sv ? "ENABLED" : "DISABLED"))`),
	entityops.DirectoryAccountFields.EmailAliases.Expr(`dyn('emails' in payload ? payload.emails.filter(e, !('primary' in e) || e.primary != true).map(e, e.address) : [])`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.AddedAt.Expr(`'creationTime' in payload ? payload.creationTime : ""`),
	entityops.DirectoryAccountFields.AvatarRemoteURL.Expr(`'thumbnailPhotoUrl' in payload ? payload.thumbnailPhotoUrl : null`),
	entityops.DirectoryAccountFields.DirectoryName.Expr("installation.name"),
	entityops.DirectoryAccountFields.PrimarySource.Expr("installation.primary_directory"),
)

// mapExprDirectoryGroup is the CEL mapping expression for Google Workspace group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryGroupFields.Email.Expr(`'email' in payload ? payload.email : ""`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`'name' in payload ? payload.name : ""`),
	entityops.DirectoryGroupFields.Classification.Expr(`dyn('adminCreated' in payload && payload.adminCreated ? "TEAM" : "DISTRIBUTION")`),
	entityops.DirectoryGroupFields.Status.Expr(`"ACTIVE"`),
	entityops.DirectoryGroupFields.SourceVersion.Expr(`'etag' in payload ? payload.etag : ""`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
	entityops.DirectoryGroupFields.DirectoryName.Expr("installation.name"),
)

// mapExprDirectoryMembership is the CEL mapping expression for Google Workspace membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`resource`),
	entityops.DirectoryMembershipFields.Role.Expr(`dyn('role' in payload && payload.role != "" ? payload.role : "MEMBER")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
	entityops.DirectoryMembershipFields.DirectoryName.Expr("installation.name"),
)
