package cloudflare

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprDirectoryAccount is the CEL mapping expression for Cloudflare account member payloads mapped to DirectoryAccount
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`'user_id' in payload && payload.user_id != "" ? payload.user_id : ('email' in payload ? payload.email : "")`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`'email' in payload ? payload.email : ""`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`'first_name' in payload && payload.first_name != "" ? (payload.first_name + ('last_name' in payload && payload.last_name != "" ? " " + payload.last_name : "")) : ('email' in payload ? payload.email : "")`),
	entityops.DirectoryAccountFields.GivenName.Expr(`'first_name' in payload ? payload.first_name : ""`),
	entityops.DirectoryAccountFields.FamilyName.Expr(`'last_name' in payload ? payload.last_name : ""`),
	entityops.DirectoryAccountFields.MfaState.Expr(`dyn('two_factor_enabled' in payload && payload.two_factor_enabled ? "ENABLED" : "DISABLED")`),
	entityops.DirectoryAccountFields.Status.Expr(`dyn('status' in payload && payload.status == "accepted" ? "ACTIVE" : "INACTIVE")`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.DirectoryName.Expr("installation.name"),
	entityops.DirectoryAccountFields.PrimarySource.Expr("installation.primary_directory"),
)

// mapExprDirectoryGroup is the CEL mapping expression for Cloudflare groups and roles payloads mapped to DirectoryGroup
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`payload.id`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`payload.name`),
	entityops.DirectoryGroupFields.Profile.Expr(`payload`),
	entityops.DirectoryGroupFields.DirectoryName.Expr("installation.name"),
)

// mapExprDirectoryMembership is the CEL mapping expression for Cloudflare policy payloads mapped to DirectoryMembership
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`payload.group_id`),
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`'user_id' in payload && payload.user_id != "" ? payload.user_id : ('email' in payload ? payload.email : "")`),
	entityops.DirectoryMembershipFields.Metadata.Expr("payload"),
	entityops.DirectoryMembershipFields.DirectoryName.Expr("installation.name"),
)

// mapExprFinding is the CEL mapping expression for Cloudflare Security Center insight payloads mapped to Finding
var mapExprFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`'id' in payload ? payload.id : ""`),
	entityops.FindingFields.ExternalOwnerID.Expr(`resource`),
	entityops.FindingFields.DisplayName.Expr(`'issue_class' in payload ? payload.issue_class : ""`),
	entityops.FindingFields.ResourceName.Expr(`'subject' in payload ? payload.subject : ""`),
	entityops.FindingFields.TargetDetails.Expr(`'subject' in payload && payload.subject != "" ? {"affected_endpoints": [payload.subject]} : {}`),
	entityops.FindingFields.Targets.Expr(`'subject' in payload && payload.subject != "" ? [payload.subject] : []`),
	entityops.FindingFields.Category.Expr(`'issue_type' in payload ? payload.issue_type : ""`),
	entityops.FindingFields.SourceUpdatedAt.Expr(`'since' in payload && payload.since != "" ? payload.since : null`),
	entityops.FindingFields.RecommendedActions.Expr(`'resolve_text' in payload ? payload.resolve_text : ""`),
	entityops.FindingFields.Open.Expr(`'status' in payload ? payload.status == "active" : false`),
	entityops.FindingFields.FindingStatusName.Expr(`'dismissed' in payload && payload.dismissed ? "Dismissed" : ('user_classification' in payload && payload.user_classification == "false_positive" ? "False Positive" : ('status' in payload ? payload.status : ""))`),
	entityops.FindingFields.State.Expr(`'status' in payload ? payload.status : ""`),
	entityops.FindingFields.References.Expr(`'resolve_link' in payload && payload.resolve_link != "" ? [payload.resolve_link] : []`),
	entityops.FindingFields.Description.Expr(`paragraphs('payload' in payload && 'detection_method' in payload.payload ? payload.payload.detection_method : "")`),
	entityops.FindingFields.EventTime.Expr(`'timestamp' in payload && payload.timestamp != "" ? payload.timestamp : null`),
	entityops.FindingFields.Severity.Expr(`'severity' in payload ? payload.severity : ""`),
	entityops.FindingFields.ExternalURI.Expr(`'resolve_link' in payload ? payload.resolve_link : ""`),
	entityops.FindingFields.RawPayload.Expr("payload"),
)

// mapExprAsset is the CEL mapping expression for Cloudflare Registrar domain payloads mapped to Asset
var mapExprAsset = providerkit.CelMapExpr(
	entityops.AssetFields.SourceIdentifier.Expr(`'domain_name' in payload ? payload.domain_name : ""`),
	entityops.AssetFields.DisplayName.Expr(`'domain_name' in payload ? payload.domain_name : ""`),
	entityops.AssetFields.Name.Expr(`'domain_name' in payload ? payload.domain_name : ""`),
	entityops.AssetFields.AssetType.Expr(`"DOMAIN"`),
	entityops.AssetFields.ObservedAt.Expr(`'created_at' in payload && payload.created_at != "" ? payload.created_at : null`),
)
