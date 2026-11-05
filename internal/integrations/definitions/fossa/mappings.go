package fossa

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprVulnerability is the CEL mapping expression for FOSSA security vulnerability issues
//
// Three provider quirks drive the shape of these expressions:
//   - cvss arrives as a whole number for integral scores, so it is coerced with double()
//   - FOSSA reports exploitability as a string enum, which cannot satisfy the numeric
//     exploitability field, so the EPSS probability is used instead
//   - affectedVersionRanges is a list while vulnerable_version_range is a single string
//
// Issues have their null members removed by dropNulls before mapping, orValue does not replace a null
var mapExprVulnerability = providerkit.CelMapExpr(
	entityops.VulnerabilityFields.ExternalID.Expr(`string(payload.?id.orValue(""))`),
	entityops.VulnerabilityFields.CveID.Expr(`payload.?cve.orValue("")`),
	entityops.VulnerabilityFields.DisplayName.Expr(`payload.?cve.orValue("") != "" ? payload.cve : payload.?title.orValue("")`),
	entityops.VulnerabilityFields.Summary.Expr(`payload.?title.orValue("")`),
	entityops.VulnerabilityFields.Description.Expr(`paragraphs(payload.?details.orValue(""))`),
	entityops.VulnerabilityFields.Severity.Expr(`payload.?severity.orValue("").upperAscii()`),
	entityops.VulnerabilityFields.Score.Expr(`double(payload.?cvss.orValue(0))`),
	entityops.VulnerabilityFields.Vector.Expr(`payload.?cvssVector.orValue("")`),
	entityops.VulnerabilityFields.CweIds.Expr(`payload.?cwes.orValue([])`),
	entityops.VulnerabilityFields.Exploitability.Expr(`double(payload.?epss.?score.orValue(0))`),
	entityops.VulnerabilityFields.PackageName.Expr(`payload.?source.?name.orValue("")`),
	entityops.VulnerabilityFields.PackageEcosystem.Expr(`payload.?source.?packageManager.orValue("")`),
	entityops.VulnerabilityFields.VulnerableVersionRange.Expr(`payload.?affectedVersionRanges.orValue([]).join(", ")`),
	entityops.VulnerabilityFields.FixAvailable.Expr(`payload.?remediation.?completeFix.orValue("") != ""`),
	entityops.VulnerabilityFields.FirstPatchedVersion.Expr(`payload.?remediation.?completeFix.orValue("")`),
	entityops.VulnerabilityFields.DependencyScope.Expr(`payload.?depths.?direct.orValue(0) > 0 ? "DIRECT" : "TRANSITIVE"`),
	entityops.VulnerabilityFields.References.Expr(`payload.?references.orValue([])`),
	entityops.VulnerabilityFields.ExternalURI.Expr(`payload.?url.orValue("")`),
	entityops.VulnerabilityFields.ExternalOwnerID.Expr(`resource != "" ? resource : ""`),
	entityops.VulnerabilityFields.Category.Expr(`payload.?type.orValue("")`),
	entityops.VulnerabilityFields.Open.Expr(`payload.?statuses.?active.orValue(0) > 0`),
	entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(`payload.?statuses.?active.orValue(0) > 0 ? "ACTIVE" : "IGNORED"`),
	entityops.VulnerabilityFields.PublishedAt.Expr(`payload.?published.orValue(null)`),
	entityops.VulnerabilityFields.DiscoveredAt.Expr(`payload.?createdAt.orValue(null)`),
	entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`payload.?projects[?0].?scannedAt.orValue(null)`),
	// metadata carries the remediation guidance and scoring detail that has no dedicated field:
	// the partial fix and upgrade distances, the EPSS percentile, and the CVSS metric breakdown
	entityops.VulnerabilityFields.Metadata.Expr(`{
		"remediation": payload.?remediation.orValue({}),
		"epss": payload.?epss.orValue({}),
		"cvssMetrics": payload.?metrics.orValue([]),
		"fossaVulnId": payload.?vulnId.orValue(""),
		"cveStatus": payload.?cveStatus.orValue(""),
		"fossaExploitability": payload.?exploitability.orValue(""),
		"patchedVersionRanges": payload.?patchedVersionRanges.orValue([])
	}`),
	entityops.VulnerabilityFields.Source.Expr(providerkit.ExprInstallationName),
	entityops.VulnerabilityFields.RawPayload.Expr("payload"),
)

// mapExprFinding is the CEL mapping expression for FOSSA OSS license compliance issues
//
// Licensing issues carry no severity or score, so those fields are deliberately left unmapped
// rather than given a fabricated value
var mapExprFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`string(payload.?id.orValue(""))`),
	entityops.FindingFields.DisplayName.Expr(`(payload.?license.orValue("") != "" ? payload.license : "License policy issue") + (payload.?source.?name.orValue("") != "" ? " in " + payload.source.name : "")`),
	entityops.FindingFields.Description.Expr(`paragraphs(payload.?details.orValue(""))`),
	entityops.FindingFields.Category.Expr(`payload.?type.orValue("")`),
	entityops.FindingFields.Categories.Expr(`payload.?type.optMap(t, [t]).orValue([])`),
	entityops.FindingFields.ResourceName.Expr(`payload.?source.?name.orValue("")`),
	entityops.FindingFields.ExternalOwnerID.Expr(`resource != "" ? resource : ""`),
	entityops.FindingFields.ExternalURI.Expr(`payload.?url.orValue("")`),
	entityops.FindingFields.Open.Expr(`payload.?statuses.?active.orValue(0) > 0`),
	entityops.FindingFields.FindingStatusName.Expr(`payload.?statuses.?active.orValue(0) > 0 ? "ACTIVE" : "IGNORED"`),
	entityops.FindingFields.ReportedAt.Expr(`payload.?createdAt.orValue(null)`),
	entityops.FindingFields.SourceUpdatedAt.Expr(`payload.?projects[?0].?scannedAt.orValue(null)`),
	entityops.FindingFields.Targets.Expr(`payload.?projects.orValue([]).filter(p, p != null && 'id' in p).map(p, p.id)`),
	entityops.FindingFields.TargetDetails.Expr(`indexBy(payload.?projects.orValue([]).filter(p, p != null && 'id' in p), "id")`),
	entityops.FindingFields.Source.Expr(providerkit.ExprInstallationName),
	entityops.FindingFields.RawPayload.Expr("payload"),
)
