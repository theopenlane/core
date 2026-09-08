package tailscale

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Tailscale user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'loginName' in payload ? payload.loginName : ""`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'displayName' in payload && payload.displayName != "" ? payload.displayName : ('loginName' in payload ? payload.loginName : "")`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn('status' in payload ? (payload.status == "active" ? "ACTIVE" : (payload.status == "suspended" ? "INACTIVE" : "INACTIVE")) : "INACTIVE")`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
)

// mapExprDirectoryGroup is the CEL mapping expression for Tailscale role group payloads mapped to DirectoryGroup
// Payload is tailscaleGroupPayload — a known struct, so direct field access is safe without 'key' in payload guards
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`payload.id`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`payload.name != "" ? payload.name : payload.id`),
	entityops.DirectoryGroupFields.Status.Expr(`dyn("ACTIVE")`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
)

// mapExprDirectoryMembership is the CEL mapping expression for Tailscale membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`'user_id' in payload ? payload.user_id : ""`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`'group_id' in payload ? payload.group_id : ""`),
	entityops.DirectoryMembershipFields.Role.Expr(`dyn("MEMBER")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
)

// mapExprAsset is the CEL mapping expression for Tailscale device payloads mapped to Asset
var mapExprAsset = providerkit.CelMapExpr(
	entityops.AssetFields.SourceIdentifier.Expr(`'id' in payload ? payload.id : ""`),
	entityops.AssetFields.Name.Expr(`'name' in payload && payload.name != "" ? payload.name : ('hostname' in payload ? payload.hostname : "")`),
	entityops.AssetFields.DisplayName.Expr(`'hostname' in payload && payload.hostname != "" ? payload.hostname : ('name' in payload ? payload.name : "")`),
	entityops.AssetFields.Description.Expr(`'os' in payload && payload.os != "" ? "OS: " + payload.os : ""`),
	entityops.AssetFields.Tags.Expr(`'tags' in payload && payload.tags != null ? payload.tags : []`),
	entityops.AssetFields.AssetType.Expr(`"DEVICE"`),
	entityops.AssetFields.InternalOwner.Expr(`'user' in payload && payload.user != "" ? payload.user : 'creator' in payload && payload.creator != "" ? payload.creator : null`),
)
