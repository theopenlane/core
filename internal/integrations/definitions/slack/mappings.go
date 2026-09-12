package slack

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Slack workspace user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`payload.id`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'email' in payload ? payload.email : ""`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'display_name' in payload && payload.display_name != "" ? payload.display_name : ('real_name' in payload && payload.real_name != "" ? payload.real_name : ('name' in payload ? payload.name : ""))`),
	entityops.DirectoryAccountFields.GivenName.Expr(`'first_name' in payload ? payload.first_name : ""`),
	entityops.DirectoryAccountFields.FamilyName.Expr(`'last_name' in payload ? payload.last_name : ""`),
	entityops.DirectoryAccountFields.JobTitle.Expr(`'title' in payload ? payload.title : ""`),
	entityops.DirectoryAccountFields.AvatarRemoteURL.Expr(`'avatar_url' in payload ? payload.avatar_url : ""`),
	entityops.DirectoryAccountFields.MfaState.Expr(`dyn(payload.has_2fa ? "ENABLED" : "DISABLED")`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn(payload.deleted ? "INACTIVE" : "ACTIVE")`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.AccountType.Expr(`payload.is_bot ? "SERVICE" : payload.is_external ? "GUEST" : "USER"`),
	entityops.DirectoryAccountFields.DirectoryName.Expr("installation.name"),
	entityops.DirectoryAccountFields.PrimarySource.Expr("installation.primary_directory"),
)
