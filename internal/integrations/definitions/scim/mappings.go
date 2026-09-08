package scim

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for SCIM user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'externalId' in payload && payload.externalId != "" ? payload.externalId : ('userName' in payload && payload.userName != "" ? payload.userName : ('emails' in payload && size(payload.emails) > 0 && payload.emails[0] != null && 'value' in payload.emails[0] ? payload.emails[0].value : ""))`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'emails' in payload && size(payload.emails) > 0 && payload.emails[0] != null && 'value' in payload.emails[0] && payload.emails[0].value != "" ? payload.emails[0].value : ('userName' in payload ? payload.userName : "")`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'displayName' in payload && payload.displayName != "" ? payload.displayName : ('name' in payload && payload.name != null && 'givenName' in payload.name && 'familyName' in payload.name ? payload.name.givenName + " " + payload.name.familyName : ('userName' in payload ? payload.userName : ""))`),
	entityops.DirectoryAccountFields.GivenName.Expr(`'name' in payload && payload.name != null && 'givenName' in payload.name ? payload.name.givenName : ""`),
	entityops.DirectoryAccountFields.FamilyName.Expr(`'name' in payload && payload.name != null && 'familyName' in payload.name ? payload.name.familyName : ""`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn(action == "delete" ? "DELETED" : ('active' in payload ? (payload.active ? "ACTIVE" : "INACTIVE") : "ACTIVE"))`),
	entityops.DirectoryAccountFields.MfaState.Expr(`dyn("UNKNOWN")`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
)

// mapExprDirectoryGroup is the CEL mapping expression for SCIM group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`'externalId' in payload && payload.externalId != "" ? payload.externalId : ('displayName' in payload ? payload.displayName : "")`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`'displayName' in payload ? payload.displayName : ""`),
	entityops.DirectoryGroupFields.Classification.Expr(`dyn("TEAM")`),
	entityops.DirectoryGroupFields.Status.Expr(`dyn(action == "delete" ? "DELETED" : ('active' in payload ? (payload.active ? "ACTIVE" : "INACTIVE") : "ACTIVE"))`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
)

// mapExprDirectoryMembership is the CEL mapping expression for SCIM group membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`'member' in payload && payload.member != null && 'value' in payload.member ? payload.member.value : ""`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`'group' in payload && payload.group != null && 'externalId' in payload.group ? payload.group.externalId : ""`),
	entityops.DirectoryMembershipFields.Role.Expr(`dyn("MEMBER")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
)
