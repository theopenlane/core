package authentik

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Authentik user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'uid' in payload ? payload.uid : ""`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'email' in payload && payload.email != null ? payload.email : ""`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'name' in payload && payload.name != null && payload.name != "" ? payload.name : ('username' in payload ? payload.username : "")`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn('is_active' in payload ? (payload.is_active ? "ACTIVE" : "INACTIVE") : "INACTIVE")`),
	entityops.DirectoryAccountFields.AccountType.Expr(`dyn('type' in payload && payload.type != null ? (payload.type == "internal" ? "USER" : (payload.type == "external" ? "GUEST" : (payload.type == "service_account" ? "SERVICE" : (payload.type == "internal_service_account" ? "SERVICE" : "USER")))) : "USER")`),
	entityops.DirectoryAccountFields.AddedAt.Expr(`'date_joined' in payload ? payload.date_joined : null`),
	entityops.DirectoryAccountFields.LastSeenAt.Expr(`'last_login' in payload ? payload.last_login : null`),
	entityops.DirectoryAccountFields.ObservedAt.Expr(`'last_updated' in payload ? payload.last_updated : null`),
	entityops.DirectoryAccountFields.Metadata.Expr(`'attributes' in payload ? payload.attributes : {}`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
)

// mapExprDirectoryGroup is the CEL mapping expression for Authentik group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`'pk' in payload ? payload.pk : ""`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`'name' in payload && payload.name != null ? payload.name : ""`),
	entityops.DirectoryGroupFields.Status.Expr(`dyn("ACTIVE")`),
	entityops.DirectoryGroupFields.Metadata.Expr(`'attributes' in payload ? payload.attributes : {}`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
)

// mapExprDirectoryMembership is the CEL mapping expression for Authentik membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`'uid' in payload ? payload.uid : ""`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`resource != "" ? resource : ""`),
	entityops.DirectoryMembershipFields.Role.Expr(`dyn("MEMBER")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
)
