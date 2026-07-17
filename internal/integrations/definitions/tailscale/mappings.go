package tailscale

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Tailscale user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'loginName' in payload ? payload.loginName : ""`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'displayName' in payload && payload.displayName != "" ? payload.displayName : ('loginName' in payload ? payload.loginName : "")`},
	{Key: entityops.InputKeyDirectoryAccountStatus, Expr: `dyn('status' in payload ? (payload.status == "active" ? "ACTIVE" : (payload.status == "suspended" ? "INACTIVE" : "INACTIVE")) : "INACTIVE")`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup is the CEL mapping expression for Tailscale role group payloads mapped to DirectoryGroup
// Payload is tailscaleGroupPayload — a known struct, so direct field access is safe without 'key' in payload guards
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryGroupExternalID, Expr: `payload.id`},
	{Key: entityops.InputKeyDirectoryGroupDisplayName, Expr: `payload.name != "" ? payload.name : payload.id`},
	{Key: entityops.InputKeyDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: entityops.InputKeyDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Tailscale membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryMembershipDirectoryAccountID, Expr: `'user_id' in payload ? payload.user_id : ""`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryGroupID, Expr: `'group_id' in payload ? payload.group_id : ""`},
	{Key: entityops.InputKeyDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: entityops.InputKeyDirectoryMembershipMetadata, Expr: "payload"},
})

// mapExprAsset is the CEL mapping expression for Tailscale device payloads mapped to Asset
var mapExprAsset = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyAssetSourceIdentifier, Expr: `'id' in payload ? payload.id : ""`},
	{Key: entityops.InputKeyAssetName, Expr: `'name' in payload && payload.name != "" ? payload.name : ('hostname' in payload ? payload.hostname : "")`},
	{Key: entityops.InputKeyAssetDisplayName, Expr: `'hostname' in payload && payload.hostname != "" ? payload.hostname : ('name' in payload ? payload.name : "")`},
	{Key: entityops.InputKeyAssetDescription, Expr: `'os' in payload && payload.os != "" ? "OS: " + payload.os : ""`},
	{Key: entityops.InputKeyAssetTags, Expr: `'tags' in payload && payload.tags != null ? payload.tags : []`},
	{Key: entityops.InputKeyAssetAssetType, Expr: `"DEVICE"`},
	{Key: entityops.InputKeyAssetInternalOwner, Expr: `'user' in payload && payload.user != "" ? payload.user : 'creator' in payload && payload.creator != "" ? payload.creator : null`},
})
