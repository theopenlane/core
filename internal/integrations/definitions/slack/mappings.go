package slack

import (
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Slack workspace user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyDirectoryAccountExternalID, Expr: `payload.id`},
	{Key: entityops.InputKeyDirectoryAccountCanonicalEmail, Expr: `'email' in payload ? payload.email : ""`},
	{Key: entityops.InputKeyDirectoryAccountDisplayName, Expr: `'display_name' in payload && payload.display_name != "" ? payload.display_name : ('real_name' in payload && payload.real_name != "" ? payload.real_name : ('name' in payload ? payload.name : ""))`},
	{Key: entityops.InputKeyDirectoryAccountGivenName, Expr: `'first_name' in payload ? payload.first_name : ""`},
	{Key: entityops.InputKeyDirectoryAccountFamilyName, Expr: `'last_name' in payload ? payload.last_name : ""`},
	{Key: entityops.InputKeyDirectoryAccountJobTitle, Expr: `'title' in payload ? payload.title : ""`},
	{Key: entityops.InputKeyDirectoryAccountAvatarRemoteURL, Expr: `'avatar_url' in payload ? payload.avatar_url : ""`},
	{Key: entityops.InputKeyDirectoryAccountMfaState, Expr: `dyn(payload.has_2fa ? "ENABLED" : "DISABLED")`},
	{Key: entityops.InputKeyDirectoryAccountStatus, Expr: `dyn(payload.deleted ? "INACTIVE" : "ACTIVE")`},
	{Key: entityops.InputKeyDirectoryAccountDirectoryInstanceID, Expr: `payload.team_id`},
	{Key: entityops.InputKeyDirectoryAccountProfile, Expr: "payload"},
	{Key: entityops.InputKeyDirectoryAccountAccountType, Expr: `payload.is_bot ? "SERVICE" : payload.is_external ? "GUEST" : "USER"`},
})
