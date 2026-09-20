package fossa

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

// mapExprVulnerability is the CEL mapping expression for FOSSA security vulnerability issues.
//
// Three provider quirks drive the shape of these expressions:
//   - cvss arrives as a whole number for integral scores, so it is coerced with double()
//   - FOSSA reports exploitability as a string enum, which cannot satisfy the numeric
//     exploitability field, so the EPSS probability is used instead
//   - affectedVersionRanges is a list while vulnerable_version_range is a single string
var mapExprVulnerability = providerkit.CelMapExpr(
	entityops.VulnerabilityFields.ExternalID.Expr(`'id' in payload && payload.id != null ? string(payload.id) : ""`),
	entityops.VulnerabilityFields.CveID.Expr(`'cve' in payload && payload.cve != null ? payload.cve : ""`),
	entityops.VulnerabilityFields.DisplayName.Expr(`'cve' in payload && payload.cve != null && payload.cve != "" ? payload.cve : ('title' in payload && payload.title != null ? payload.title : "")`),
	entityops.VulnerabilityFields.Summary.Expr(`'title' in payload && payload.title != null ? payload.title : ""`),
	entityops.VulnerabilityFields.Description.Expr(`'details' in payload && payload.details != null ? payload.details : ""`),
	entityops.VulnerabilityFields.Severity.Expr(`'severity' in payload && payload.severity != null ? payload.severity.upperAscii() : ""`),
	entityops.VulnerabilityFields.Score.Expr(`'cvss' in payload && payload.cvss != null ? double(payload.cvss) : 0.0`),
	entityops.VulnerabilityFields.Vector.Expr(`'cvssVector' in payload && payload.cvssVector != null ? payload.cvssVector : ""`),
	entityops.VulnerabilityFields.CweIds.Expr(`'cwes' in payload && payload.cwes != null ? payload.cwes : []`),
	entityops.VulnerabilityFields.Exploitability.Expr(`'epss' in payload && payload.epss != null && 'score' in payload.epss && payload.epss.score != null ? double(payload.epss.score) : 0.0`),
	entityops.VulnerabilityFields.PackageName.Expr(`'source' in payload && payload.source != null && 'name' in payload.source ? payload.source.name : ""`),
	entityops.VulnerabilityFields.PackageEcosystem.Expr(`'source' in payload && payload.source != null && 'packageManager' in payload.source ? payload.source.packageManager : ""`),
	entityops.VulnerabilityFields.VulnerableVersionRange.Expr(`'affectedVersionRanges' in payload && payload.affectedVersionRanges != null && size(payload.affectedVersionRanges) > 0 ? payload.affectedVersionRanges.join(", ") : ""`),
	entityops.VulnerabilityFields.FixAvailable.Expr(`'remediation' in payload && payload.remediation != null && 'completeFix' in payload.remediation && payload.remediation.completeFix != null && payload.remediation.completeFix != ""`),
	entityops.VulnerabilityFields.FirstPatchedVersion.Expr(`'remediation' in payload && payload.remediation != null && 'completeFix' in payload.remediation && payload.remediation.completeFix != null ? payload.remediation.completeFix : ""`),
	entityops.VulnerabilityFields.DependencyScope.Expr(`'depths' in payload && payload.depths != null && 'direct' in payload.depths && payload.depths.direct > 0 ? "DIRECT" : "TRANSITIVE"`),
	entityops.VulnerabilityFields.References.Expr(`'references' in payload && payload.references != null ? payload.references : []`),
	entityops.VulnerabilityFields.ExternalURI.Expr(`'url' in payload && payload.url != null ? payload.url : ""`),
	entityops.VulnerabilityFields.ExternalOwnerID.Expr(`resource != "" ? resource : ""`),
	entityops.VulnerabilityFields.Category.Expr(`'type' in payload && payload.type != null ? payload.type : ""`),
	entityops.VulnerabilityFields.Open.Expr(`'statuses' in payload && payload.statuses != null && 'active' in payload.statuses ? payload.statuses.active > 0 : false`),
	entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(`'statuses' in payload && payload.statuses != null && 'active' in payload.statuses && payload.statuses.active > 0 ? "ACTIVE" : "IGNORED"`),
	entityops.VulnerabilityFields.PublishedAt.Expr(`'published' in payload ? payload.published : null`),
	entityops.VulnerabilityFields.DiscoveredAt.Expr(`'createdAt' in payload ? payload.createdAt : null`),
	entityops.VulnerabilityFields.SourceUpdatedAt.Expr(`'projects' in payload && payload.projects != null && size(payload.projects) > 0 && payload.projects[0] != null && 'scannedAt' in payload.projects[0] ? payload.projects[0].scannedAt : null`),
	// metadata carries the remediation guidance and scoring detail that has no dedicated field:
	// the partial fix and upgrade distances, the EPSS percentile, and the CVSS metric breakdown
	entityops.VulnerabilityFields.Metadata.Expr(`{
		"remediation": 'remediation' in payload && payload.remediation != null ? payload.remediation : {},
		"epss": 'epss' in payload && payload.epss != null ? payload.epss : {},
		"cvssMetrics": 'metrics' in payload && payload.metrics != null ? payload.metrics : [],
		"fossaVulnId": 'vulnId' in payload && payload.vulnId != null ? payload.vulnId : "",
		"cveStatus": 'cveStatus' in payload && payload.cveStatus != null ? payload.cveStatus : "",
		"fossaExploitability": 'exploitability' in payload && payload.exploitability != null ? payload.exploitability : "",
		"patchedVersionRanges": 'patchedVersionRanges' in payload && payload.patchedVersionRanges != null ? payload.patchedVersionRanges : []
	}`),
	entityops.VulnerabilityFields.Source.Expr(`"FOSSA"`),
	entityops.VulnerabilityFields.RawPayload.Expr("payload"),
)

// mapExprFinding is the CEL mapping expression for FOSSA OSS license compliance issues.
//
// Licensing issues carry no severity or score, so those fields are deliberately left unmapped
// rather than given a fabricated value.
var mapExprFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`'id' in payload && payload.id != null ? string(payload.id) : ""`),
	entityops.FindingFields.DisplayName.Expr(`('license' in payload && payload.license != null && payload.license != "" ? payload.license : "License policy issue") + ('source' in payload && payload.source != null && 'name' in payload.source && payload.source.name != "" ? " in " + payload.source.name : "")`),
	entityops.FindingFields.Description.Expr(`'details' in payload && payload.details != null ? payload.details : ""`),
	entityops.FindingFields.Category.Expr(`'type' in payload && payload.type != null ? payload.type : ""`),
	entityops.FindingFields.Categories.Expr(`'type' in payload && payload.type != null ? [payload.type] : []`),
	entityops.FindingFields.ResourceName.Expr(`'source' in payload && payload.source != null && 'name' in payload.source ? payload.source.name : ""`),
	entityops.FindingFields.ExternalOwnerID.Expr(`resource != "" ? resource : ""`),
	entityops.FindingFields.ExternalURI.Expr(`'url' in payload && payload.url != null ? payload.url : ""`),
	entityops.FindingFields.Open.Expr(`'statuses' in payload && payload.statuses != null && 'active' in payload.statuses ? payload.statuses.active > 0 : false`),
	entityops.FindingFields.FindingStatusName.Expr(`'statuses' in payload && payload.statuses != null && 'active' in payload.statuses && payload.statuses.active > 0 ? "ACTIVE" : "IGNORED"`),
	entityops.FindingFields.ReportedAt.Expr(`'createdAt' in payload ? payload.createdAt : null`),
	entityops.FindingFields.SourceUpdatedAt.Expr(`'projects' in payload && payload.projects != null && size(payload.projects) > 0 && payload.projects[0] != null && 'scannedAt' in payload.projects[0] ? payload.projects[0].scannedAt : null`),
	entityops.FindingFields.Targets.Expr(`'projects' in payload && payload.projects != null && size(payload.projects) > 0 ? payload.projects.filter(p, p != null && 'id' in p).map(p, p.id) : []`),
	entityops.FindingFields.TargetDetails.Expr(`'projects' in payload && payload.projects != null && size(payload.projects) > 0 ? indexBy(payload.projects.filter(p, p != null && 'id' in p), "id") : {}`),
	entityops.FindingFields.Source.Expr(`"FOSSA"`),
	entityops.FindingFields.RawPayload.Expr("payload"),
)
