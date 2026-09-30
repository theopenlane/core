package keycloak

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Keycloak user payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'email' in payload && payload.email != "" ? payload.email : ""`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'username' in payload ? payload.username : ""`),
	entityops.DirectoryAccountFields.GivenName.Expr(`'firstName' in payload ? payload.firstName : ""`),
	entityops.DirectoryAccountFields.FamilyName.Expr(`'lastName' in payload ? payload.lastName : ""`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn('enabled' in payload ? (payload.enabled ? "ACTIVE" : "INACTIVE") : "INACTIVE")`),
	entityops.DirectoryAccountFields.AccountType.Expr(`dyn('serviceAccountClientId' in payload && payload.serviceAccountClientId != "" ? "SERVICE" : "USER")`),
	entityops.DirectoryAccountFields.AddedAt.Expr(`'createdTimestamp' in payload ? timestamp(int(payload.createdTimestamp) / 1000) : null`),
	entityops.DirectoryAccountFields.Metadata.Expr(`'attributes' in payload ? payload.attributes : {}`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.DirectoryName.Expr(providerkit.ExprInstallationName),
	entityops.DirectoryAccountFields.PrimarySource.Expr("installation.primary_directory"),
)

// mapExprDirectoryGroup is the CEL mapping expression for Keycloak group payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`'name' in payload ? payload.name : ""`),
	entityops.DirectoryGroupFields.Status.Expr(`dyn("ACTIVE")`),
	entityops.DirectoryGroupFields.Metadata.Expr(`'attributes' in payload ? payload.attributes : {}`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
	entityops.DirectoryGroupFields.DirectoryName.Expr(providerkit.ExprInstallationName),
)

// mapExprDirectoryMembership is the CEL mapping expression for Keycloak membership payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`resource != "" ? resource : ""`),
	entityops.DirectoryMembershipFields.Role.Expr(`dyn("MEMBER")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
	entityops.DirectoryMembershipFields.DirectoryName.Expr(providerkit.ExprInstallationName),
)
