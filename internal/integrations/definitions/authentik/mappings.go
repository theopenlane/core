package authentik

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Authentik user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `'uid' in payload ? payload.uid : ""`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'email' in payload && payload.email != null ? payload.email : ""`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'name' in payload && payload.name != null && payload.name != "" ? payload.name : ('username' in payload ? payload.username : "")`},
	{Key: entityops.InputKeyDirectoryAccountStatus, Expr: `dyn('is_active' in payload ? (payload.is_active ? "ACTIVE" : "INACTIVE") : "INACTIVE")`},
	{Key: entityops.InputKeyDirectoryAccountAccountType, Expr: `dyn('type' in payload && payload.type != null ? (payload.type == "internal" ? "USER" : (payload.type == "external" ? "GUEST" : (payload.type == "service_account" ? "SERVICE" : (payload.type == "internal_service_account" ? "SERVICE" : "USER")))) : "USER")`},
	{Key: entityops.InputKeyDirectoryAccountAddedAt, Expr: `'date_joined' in payload ? payload.date_joined : null`},
	{Key: entityops.InputKeyDirectoryAccountLastSeenAt, Expr: `'last_login' in payload ? payload.last_login : null`},
	{Key: entityops.InputKeyDirectoryAccountObservedAt, Expr: `'last_updated' in payload ? payload.last_updated : null`},
	{Key: entityops.InputKeyDirectoryAccountMetadata, Expr: `'attributes' in payload ? payload.attributes : {}`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
})

// mapExprDirectoryGroup is the CEL mapping expression for Authentik group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryGroupExternalID, Expr: `'pk' in payload ? payload.pk : ""`},
	{Key: entityops.InputKeyDirectoryGroupDisplayName, Expr: `'name' in payload && payload.name != null ? payload.name : ""`},
	{Key: entityops.InputKeyDirectoryGroupStatus, Expr: `dyn("ACTIVE")`},
	{Key: entityops.InputKeyDirectoryGroupMetadata, Expr: `'attributes' in payload ? payload.attributes : {}`},
	{Key: entityops.InputKeyDirectoryGroupProfile, Expr: "payload"},
})

// mapExprDirectoryMembership is the CEL mapping expression for Authentik membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryMembershipDirectoryAccountID, Expr: `'uid' in payload ? payload.uid : ""`},
	{Key: entityops.InputKeyDirectoryMembershipDirectoryGroupID, Expr: `resource != "" ? resource : ""`},
	{Key: entityops.InputKeyDirectoryMembershipRole, Expr: `dyn("MEMBER")`},
	{Key: entityops.InputKeyDirectoryMembershipMetadata, Expr: "payload"},
})
