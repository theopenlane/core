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
var mapExprVulnerability = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyVulnerabilityExternalID, Expr: `'id' in payload && payload.id != null ? string(payload.id) : ""`},
	{Key: entityops.InputKeyVulnerabilityCveID, Expr: `'cve' in payload && payload.cve != null ? payload.cve : ""`},
	{Key: entityops.InputKeyVulnerabilityDisplayName, Expr: `'cve' in payload && payload.cve != null && payload.cve != "" ? payload.cve : ('title' in payload && payload.title != null ? payload.title : "")`},
	{Key: entityops.InputKeyVulnerabilitySummary, Expr: `'title' in payload && payload.title != null ? payload.title : ""`},
	{Key: entityops.InputKeyVulnerabilityDescription, Expr: `'details' in payload && payload.details != null ? payload.details : ""`},
	{Key: entityops.InputKeyVulnerabilitySeverity, Expr: `dyn('severity' in payload && payload.severity != null ? payload.severity.upperAscii() : "")`},
	{Key: entityops.InputKeyVulnerabilityScore, Expr: `'cvss' in payload && payload.cvss != null ? double(payload.cvss) : 0.0`},
	{Key: entityops.InputKeyVulnerabilityVector, Expr: `'cvssVector' in payload && payload.cvssVector != null ? payload.cvssVector : ""`},
	{Key: entityops.InputKeyVulnerabilityCweIds, Expr: `'cwes' in payload && payload.cwes != null ? payload.cwes : []`},
	{Key: entityops.InputKeyVulnerabilityExploitability, Expr: `'epss' in payload && payload.epss != null && 'score' in payload.epss && payload.epss.score != null ? double(payload.epss.score) : 0.0`},
	{Key: entityops.InputKeyVulnerabilityPackageName, Expr: `'source' in payload && payload.source != null && 'name' in payload.source ? payload.source.name : ""`},
	{Key: entityops.InputKeyVulnerabilityPackageEcosystem, Expr: `'source' in payload && payload.source != null && 'packageManager' in payload.source ? payload.source.packageManager : ""`},
	{Key: entityops.InputKeyVulnerabilityVulnerableVersionRange, Expr: `'affectedVersionRanges' in payload && payload.affectedVersionRanges != null && size(payload.affectedVersionRanges) > 0 ? payload.affectedVersionRanges.join(", ") : ""`},
	{Key: entityops.InputKeyVulnerabilityFixAvailable, Expr: `'remediation' in payload && payload.remediation != null && 'completeFix' in payload.remediation && payload.remediation.completeFix != null && payload.remediation.completeFix != ""`},
	{Key: entityops.InputKeyVulnerabilityFirstPatchedVersion, Expr: `'remediation' in payload && payload.remediation != null && 'completeFix' in payload.remediation && payload.remediation.completeFix != null ? payload.remediation.completeFix : ""`},
	{Key: entityops.InputKeyVulnerabilityDependencyScope, Expr: `dyn('depths' in payload && payload.depths != null && 'direct' in payload.depths && payload.depths.direct > 0 ? "DIRECT" : "TRANSITIVE")`},
	{Key: entityops.InputKeyVulnerabilityReferences, Expr: `'references' in payload && payload.references != null ? payload.references : []`},
	{Key: entityops.InputKeyVulnerabilityExternalURI, Expr: `'url' in payload && payload.url != null ? payload.url : ""`},
	{Key: entityops.InputKeyVulnerabilityExternalOwnerID, Expr: `resource != "" ? resource : ""`},
	{Key: entityops.InputKeyVulnerabilityCategory, Expr: `'type' in payload && payload.type != null ? payload.type : ""`},
	{Key: entityops.InputKeyVulnerabilityOpen, Expr: `'statuses' in payload && payload.statuses != null && 'active' in payload.statuses ? payload.statuses.active > 0 : false`},
	{Key: entityops.InputKeyVulnerabilityVulnerabilityStatusName, Expr: `dyn('statuses' in payload && payload.statuses != null && 'active' in payload.statuses && payload.statuses.active > 0 ? "ACTIVE" : "IGNORED")`},
	{Key: entityops.InputKeyVulnerabilityPublishedAt, Expr: `'published' in payload ? payload.published : null`},
	{Key: entityops.InputKeyVulnerabilityDiscoveredAt, Expr: `'createdAt' in payload ? payload.createdAt : null`},
	{Key: entityops.InputKeyVulnerabilitySourceUpdatedAt, Expr: `'projects' in payload && payload.projects != null && size(payload.projects) > 0 && payload.projects[0] != null && 'scannedAt' in payload.projects[0] ? payload.projects[0].scannedAt : null`},
	// metadata carries the remediation guidance and scoring detail that has no dedicated field:
	// the partial fix and upgrade distances, the EPSS percentile, and the CVSS metric breakdown
	{Key: entityops.InputKeyVulnerabilityMetadata, Expr: `{
		"remediation": 'remediation' in payload && payload.remediation != null ? payload.remediation : {},
		"epss": 'epss' in payload && payload.epss != null ? payload.epss : {},
		"cvssMetrics": 'metrics' in payload && payload.metrics != null ? payload.metrics : [],
		"fossaVulnId": 'vulnId' in payload && payload.vulnId != null ? payload.vulnId : "",
		"cveStatus": 'cveStatus' in payload && payload.cveStatus != null ? payload.cveStatus : "",
		"fossaExploitability": 'exploitability' in payload && payload.exploitability != null ? payload.exploitability : "",
		"patchedVersionRanges": 'patchedVersionRanges' in payload && payload.patchedVersionRanges != null ? payload.patchedVersionRanges : []
	}`},
	{Key: entityops.InputKeyVulnerabilitySource, Expr: `"FOSSA"`},
	{Key: entityops.InputKeyVulnerabilityRawPayload, Expr: "payload"},
})

// mapExprFinding is the CEL mapping expression for FOSSA OSS license compliance issues.
//
// Licensing issues carry no severity or score, so those fields are deliberately left unmapped
// rather than given a fabricated value.
var mapExprFinding = providerkit.CelMapExpr([]providerkit.CelMapEntry{
	{Key: entityops.InputKeyFindingExternalID, Expr: `'id' in payload && payload.id != null ? string(payload.id) : ""`},
	{Key: entityops.InputKeyFindingDisplayName, Expr: `('license' in payload && payload.license != null && payload.license != "" ? payload.license : "License policy issue") + ('source' in payload && payload.source != null && 'name' in payload.source && payload.source.name != "" ? " in " + payload.source.name : "")`},
	{Key: entityops.InputKeyFindingDescription, Expr: `'details' in payload && payload.details != null ? payload.details : ""`},
	{Key: entityops.InputKeyFindingCategory, Expr: `'type' in payload && payload.type != null ? payload.type : ""`},
	{Key: entityops.InputKeyFindingCategories, Expr: `'type' in payload && payload.type != null ? [payload.type] : []`},
	{Key: entityops.InputKeyFindingResourceName, Expr: `'source' in payload && payload.source != null && 'name' in payload.source ? payload.source.name : ""`},
	{Key: entityops.InputKeyFindingExternalOwnerID, Expr: `resource != "" ? resource : ""`},
	{Key: entityops.InputKeyFindingExternalURI, Expr: `'url' in payload && payload.url != null ? payload.url : ""`},
	{Key: entityops.InputKeyFindingOpen, Expr: `'statuses' in payload && payload.statuses != null && 'active' in payload.statuses ? payload.statuses.active > 0 : false`},
	{Key: entityops.InputKeyFindingFindingStatusName, Expr: `dyn('statuses' in payload && payload.statuses != null && 'active' in payload.statuses && payload.statuses.active > 0 ? "ACTIVE" : "IGNORED")`},
	{Key: entityops.InputKeyFindingReportedAt, Expr: `'createdAt' in payload ? payload.createdAt : null`},
	{Key: entityops.InputKeyFindingSourceUpdatedAt, Expr: `'projects' in payload && payload.projects != null && size(payload.projects) > 0 && payload.projects[0] != null && 'scannedAt' in payload.projects[0] ? payload.projects[0].scannedAt : null`},
	{Key: entityops.InputKeyFindingTargets, Expr: `'projects' in payload && payload.projects != null && size(payload.projects) > 0 ? payload.projects.filter(p, p != null && 'id' in p).map(p, p.id) : []`},
	{Key: entityops.InputKeyFindingTargetDetails, Expr: `'projects' in payload && payload.projects != null && size(payload.projects) > 0 ? indexBy(payload.projects.filter(p, p != null && 'id' in p), "id") : {}`},
	{Key: entityops.InputKeyFindingSource, Expr: `"FOSSA"`},
	{Key: entityops.InputKeyFindingRawPayload, Expr: "payload"},
})
