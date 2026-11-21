package aikido

import (
	"strconv"
	"strings"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/remediation"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
)

const (
	// exprIssueOpen treats snoozed issues as still open since they have not been resolved or accepted
	exprIssueOpen = `payload.?status.orValue("") in ["open", "snoozed"]`
	// exprIssueStatus is the Aikido issue status
	exprIssueStatus = `payload.?status.orValue("")`
	// exprIssueSeverity is the upper-cased Aikido severity
	exprIssueSeverity = `payload.?severity.orValue("").upperAscii()`
	// exprIssueScore rescales the 1-100 Aikido severity score to the 0-10 CVSS range
	exprIssueScore = `double(payload.?severity_score.orValue(0)) / 10.0`
	// exprIssueFirstDetected is when Aikido first detected the issue
	exprIssueFirstDetected = `'first_detected_at' in payload ? timestamp(int(payload.first_detected_at)) : null`
	// exprIssueSLADays is the remediation SLA in days
	exprIssueSLADays = `payload.?sla_days.orValue(null)`
	// exprAssetObserved is when Aikido last saw the asset, a scan where -1 means never, a device heartbeat, or a subdomain DNS resolution
	exprAssetObserved = `'last_scanned_at' in payload && int(payload.last_scanned_at) > 0 ? timestamp(int(payload.last_scanned_at)) : ('last_heartbeat_at' in payload ? timestamp(int(payload.last_heartbeat_at)) : ('dns_last_resolved_at' in payload ? timestamp(int(payload.dns_last_resolved_at)) : null))`
	// exprAssetWebsite links a domain asset, adding the scheme when Aikido reports a bare host
	exprAssetWebsite = `resource.startsWith("http") ? resource : "https://" + resource`
	// exprAssetTags are the Aikido label names on the asset
	exprAssetTags = `payload.?labels.orValue([]).map(l, l.?name.orValue("")).filter(n, n != "")`
	// exprAssetAccessModel is the Aikido internet connectivity as the access model, unknown leaves it unset
	exprAssetAccessModel = `payload.?connectivity.orValue("") != "unknown" ? payload.?connectivity.orValue("") : ""`
	// exprAssetDataClassification is the Aikido sensitivity as the data classification, no_data leaves it unset
	exprAssetDataClassification = `payload.?sensitivity.orValue("") != "no_data" ? payload.?sensitivity.orValue("") : ""`
	// exprAssetName is the asset name issues refer to, carried as the envelope resource
	exprAssetName = "resource"
)

// exprIssueURI links to the issue group detail page in Aikido, anchored on the CVE when there is one
var exprIssueURI = `'group_id' in payload ? ` + strconv.Quote(oauthAppURL+"/issues/") + ` + string(payload.group_id) + "/detail" + (payload.?cve_id.orValue("") != "" ? "#" + payload.cve_id : "") : ""`

const (
	// exprIssueDescription is the description of the issue group the issue belongs to
	exprIssueDescription = `paragraphs(payload.?group.?description.orValue(""))`
	// exprIssueHowToFix is the remediation guidance of the issue group the issue belongs to
	exprIssueHowToFix = `paragraphs(payload.?group.?how_to_fix.orValue(""))`
	// exprVulnerabilityDescription prefers the Aikido Intel tldr, else the issue group description unless it only repeats the rule used as the summary
	exprVulnerabilityDescription = `payload.?cve_details.?tldr.orValue("") != "" ? paragraphs(payload.cve_details.tldr) : (payload.?group.?description.orValue("") != payload.?rule.orValue("") ? ` + exprIssueDescription + ` : "")`
	// exprFindingDisplayName is the rule, else the affected package with its license for license issues, else the issue group title, else the issue type
	exprFindingDisplayName = `payload.?rule.orValue("") != "" ? payload.rule : (payload.?affected_package.orValue("") != "" ? payload.affected_package + (payload.?license_type.orValue("") != "" ? " (" + payload.license_type + ")" : "") : (payload.?group.?title.orValue("") != "" ? payload.group.title : payload.?type.orValue("")))`
	// exprFindingDescription prefers the pentest attack analysis description over the issue group description
	exprFindingDescription = `payload.?attack_analysis.?description.orValue("") != "" ? paragraphs(payload.attack_analysis.description) : ` + exprIssueDescription
	// exprFindingRecommendation prefers the pentest attack analysis remediation steps over the issue group guidance
	exprFindingRecommendation = `payload.?attack_analysis.?remediation.orValue([]).size() > 0 ? payload.attack_analysis.remediation.join("\n") : ` + exprIssueHowToFix
	// exprFindingStepsToReproduce lists the pentest attack analysis reproduction steps, each with its command
	exprFindingStepsToReproduce = `payload.?attack_analysis.?reproduction_steps.orValue([]).map(s, s.?step_info.orValue("") + (s.?code_block.orValue("") != "" ? "\n" + s.code_block : ""))`
)

// exprIssueReferences links the Aikido Intel advisory for the issue id and its CVE alias, which covers both CVE and AIKIDO identifiers
const exprIssueReferences = `[payload.?cve_id.orValue(""), payload.?cve_id_alias.orValue("")].filter(id, id != "").map(id, "https://intel.aikido.dev/cve/" + id)`

// exprIssueCVE prefers the CVE alias the bulk details carry for AIKIDO identifiers
const exprIssueCVE = `payload.?cve_id_alias.orValue("") != "" ? payload.cve_id_alias : payload.?cve_id.orValue("")`

// exprIssueIgnoreReasons joins the reasons Aikido gives for ignoring the issue
const exprIssueIgnoreReasons = `payload.?ignore_reasons.orValue([]).map(r, r.?reason.orValue("")).filter(r, r != "").join(", ")`

// exprIssueMetadata carries the issue detail that has no dedicated field
const exprIssueMetadata = `{
	"groupId": payload.?group_id.orValue(0),
	"ruleId": payload.?rule_id.orValue(""),
	"attackSurface": payload.?attack_surface.orValue(""),
	"severityScore": payload.?severity_score.orValue(0),
	"originalCvssSeverityScore": payload.?original_cvss_severity_score.orValue(0),
	"exploitability": payload.?exploitability.orValue(""),
	"installedVersion": payload.?installed_version.orValue(""),
	"patchedVersions": payload.?patched_versions.orValue([]),
	"affectedFile": payload.?affected_file.orValue(""),
	"startLine": payload.?start_line.orValue(0),
	"endLine": payload.?end_line.orValue(0),
	"licenseType": payload.?license_type.orValue(""),
	"programmingLanguage": payload.?programming_language.orValue(""),
	"branch": payload.?code_repo_branch_name.orValue(""),
	"slaRemediateBy": payload.?sla_remediate_by.orValue(0),
	"ignoredBy": payload.?ignored_by.orValue(""),
	"snoozeUntil": payload.?snooze_until.orValue(0),
	"snoozeReason": payload.?snooze_reason.orValue(""),
	"cveIdAlias": payload.?cve_id_alias.orValue(""),
	"ignoreReasons": payload.?ignore_reasons.orValue([]),
	"relatedCveIds": payload.?group.?related_cve_ids.orValue([]),
	"timeToFixMinutes": payload.?group.?time_to_fix_minutes.orValue(0),
	"assessmentId": payload.?assessment_id.orValue(""),
	"attackSummary": payload.?attack_analysis.?summary.orValue(""),
	"risk": payload.?attack_analysis.?risk.orValue(""),
	"rootCauseAnalysis": payload.?attack_analysis.?root_cause_analysis.orValue(""),
	"doesThisAffectMe": payload.?cve_details.?does_this_affect_me.orValue(""),
	"howToFix": payload.?group.?how_to_fix.orValue("")
}`

// mapExprVulnerability maps Aikido open source dependency issues
var mapExprVulnerability = providerkit.CelMapExpr(
	entityops.VulnerabilityFields.ExternalID.Expr(`string(payload.?id.orValue(""))`),
	entityops.VulnerabilityFields.CveID.Expr(exprIssueCVE),
	entityops.VulnerabilityFields.DisplayName.Expr(`payload.?cve_id.orValue("") != "" ? payload.cve_id : payload.?rule.orValue("")`),
	entityops.VulnerabilityFields.Summary.Expr(`payload.?rule.orValue("")`),
	entityops.VulnerabilityFields.Description.Expr(exprVulnerabilityDescription),
	entityops.VulnerabilityFields.PublishedAt.Expr(`payload.?cve_details.?published.orValue("") != "" ? timestamp(payload.cve_details.published + "T00:00:00Z") : null`),
	entityops.VulnerabilityFields.Severity.Expr(exprIssueSeverity),
	entityops.VulnerabilityFields.Score.Expr(exprIssueScore),
	entityops.VulnerabilityFields.CweIds.Expr(`payload.?cwe_classes.orValue([])`),
	entityops.VulnerabilityFields.PackageName.Expr(`payload.?affected_package.orValue("")`),
	entityops.VulnerabilityFields.ManifestPath.Expr(`payload.?affected_file.orValue("")`),
	entityops.VulnerabilityFields.FixAvailable.Expr(`payload.?patched_versions.orValue([]).size() > 0`),
	entityops.VulnerabilityFields.FirstPatchedVersion.Expr(`payload.?patched_versions[?0].orValue("")`),
	entityops.VulnerabilityFields.ExternalOwnerID.Expr("resource"),
	providerkit.EdgeNameEntry(entityops.SchemaVulnerability, vulnerability.EdgeAssets, `resource != "" ? [resource] : []`),
	entityops.VulnerabilityFields.ExternalURI.Expr(exprIssueURI),
	entityops.VulnerabilityFields.References.Expr(exprIssueReferences),
	entityops.VulnerabilityFields.Category.Expr(`payload.?type.orValue("")`),
	entityops.VulnerabilityFields.Open.Expr(exprIssueOpen),
	entityops.VulnerabilityFields.VulnerabilityStatusName.Expr(exprIssueStatus),
	entityops.VulnerabilityFields.DiscoveredAt.Expr(exprIssueFirstDetected),
	entityops.VulnerabilityFields.DismissedAt.Expr(`'ignored_at' in payload ? timestamp(int(payload.ignored_at)) : null`),
	entityops.VulnerabilityFields.DismissedReason.Expr(`payload.?ignore_reasons[?0].?kind.orValue("")`),
	entityops.VulnerabilityFields.DismissedComment.Expr(exprIssueIgnoreReasons),
	entityops.VulnerabilityFields.FixedAt.Expr(`'closed_at' in payload ? timestamp(int(payload.closed_at)) : null`),
	entityops.VulnerabilityFields.RemediationSLA.Expr(exprIssueSLADays),
	entityops.VulnerabilityFields.Metadata.Expr(exprIssueMetadata),
	entityops.VulnerabilityFields.Source.Expr(providerkit.ExprInstallationName),
	entityops.VulnerabilityFields.RawPayload.Expr("payload"),
)

// mapExprFinding maps every Aikido issue type other than open source dependencies, such as SAST, secrets, IaC and cloud
var mapExprFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`string(payload.?id.orValue(""))`),
	entityops.FindingFields.DisplayName.Expr(exprFindingDisplayName),
	entityops.FindingFields.Description.Expr(exprFindingDescription),
	entityops.FindingFields.RecommendedActions.Expr(exprFindingRecommendation),
	entityops.FindingFields.StepsToReproduce.Expr(exprFindingStepsToReproduce),
	providerkit.EdgeExternalIDEntry(entityops.SchemaFinding, finding.EdgeScans, `payload.?assessment_id.optMap(id, [id]).orValue([])`),
	entityops.FindingFields.Category.Expr(`payload.?type.orValue("")`),
	entityops.FindingFields.Categories.Expr(`payload.?type.optMap(t, [t]).orValue([])`),
	entityops.FindingFields.Severity.Expr(exprIssueSeverity),
	entityops.FindingFields.Score.Expr(exprIssueScore),
	entityops.FindingFields.NumericSeverity.Expr(`double(payload.?severity_score.orValue(0))`),
	entityops.FindingFields.ResourceName.Expr("resource"),
	entityops.FindingFields.Targets.Expr(`resource != "" ? [resource] : []`),
	entityops.FindingFields.ExternalOwnerID.Expr("resource"),
	entityops.FindingFields.ExternalURI.Expr(exprIssueURI),
	entityops.FindingFields.References.Expr(exprIssueReferences),
	entityops.FindingFields.Open.Expr(exprIssueOpen),
	entityops.FindingFields.FindingStatusName.Expr(exprIssueStatus),
	entityops.FindingFields.ReportedAt.Expr(exprIssueFirstDetected),
	entityops.FindingFields.RemediationSLA.Expr(exprIssueSLADays),
	entityops.FindingFields.Metadata.Expr(exprIssueMetadata),
	entityops.FindingFields.Source.Expr(providerkit.ExprInstallationName),
	entityops.FindingFields.RawPayload.Expr("payload"),
)

// mapExprDataleakFinding maps credentials leaked for a monitored domain, the resource is the domain
var mapExprDataleakFinding = providerkit.CelMapExpr(
	entityops.FindingFields.ExternalID.Expr(`"dataleak:" + payload.?id.orValue("")`),
	entityops.FindingFields.DisplayName.Expr(`"Leaked credentials for " + payload.?email.orValue("")`),
	entityops.FindingFields.Description.Expr(`"Credentials for " + payload.?email.orValue("") + " were found in a data breach"`),
	entityops.FindingFields.Category.Expr(strconv.Quote(findingVariantDataleak)),
	entityops.FindingFields.Categories.Expr(`[`+strconv.Quote(findingVariantDataleak)+`]`),
	entityops.FindingFields.ResourceName.Expr("resource"),
	entityops.FindingFields.Targets.Expr(`resource != "" ? [resource] : []`),
	entityops.FindingFields.ExternalOwnerID.Expr("resource"),
	entityops.FindingFields.Open.Expr(`payload.?status.orValue("") == "OPEN"`),
	entityops.FindingFields.FindingStatusName.Expr(`payload.?status.orValue("")`),
	entityops.FindingFields.ReportedAt.Expr(`'last_breach_date' in payload ? timestamp(int(payload.last_breach_date)) : null`),
	entityops.FindingFields.Source.Expr(providerkit.ExprInstallationName),
	entityops.FindingFields.RawPayload.Expr("payload"),
)

const (
	// exprAutofixStatus maps the autofix task status onto the remediation status, a fix preview nobody applied is still open
	// and a failed or closed task will not land
	exprAutofixStatus = `{
	"preview_available": "OPEN",
	"creating": "IN_PROGRESS",
	"pr_open": "IN_REVIEW",
	"pr_created": "IN_REVIEW",
	"pr_merged": "COMPLETED",
	"pr_closed": "WONT_DO",
	"no_fix": "WONT_DO",
	"autofix_failed": "WONT_DO",
	"preview_failed": "WONT_DO"
}[?payload.?status.orValue("")].orValue("OPEN")`
	// exprAutofixInstructions describes the autofix and where it applies
	exprAutofixInstructions = `"Aikido " + payload.?type.orValue("") + " autofix" + (payload.?autofix_target.?location.orValue("") != "" ? " in " + payload.autofix_target.location : "")`
	// exprAutofixFinished is when the autofix task finished
	exprAutofixFinished = `('finished_at' in payload ? timestamp(int(payload.finished_at)) : null)`
	// exprAutofixIssueIDs are the ids of the issues the task targets, matching the external id of both vulnerabilities and findings
	exprAutofixIssueIDs = `payload.?` + autofixIssueIDsKey + `.orValue([])`
)

// mapExprRemediation maps Aikido autofix tasks to remediations linked to the issues they target
var mapExprRemediation = providerkit.CelMapExpr(
	entityops.RemediationFields.ExternalID.Expr(`string(payload.?id.orValue(""))`),
	entityops.RemediationFields.Title.Expr(`payload.?autofix_target.?title.orValue("")`),
	entityops.RemediationFields.Intent.Expr(`payload.?type.orValue("")`),
	entityops.RemediationFields.Instructions.Expr(exprAutofixInstructions),
	entityops.RemediationFields.Status.Expr(exprAutofixStatus),
	entityops.RemediationFields.State.Expr(`payload.?status.orValue("")`),
	entityops.RemediationFields.Error.Expr(`payload.?status.orValue("") in ["autofix_failed", "preview_failed"] ? payload.status : ""`),
	entityops.RemediationFields.PullRequestURI.Expr(`payload.?pull_request.?url.orValue("")`),
	entityops.RemediationFields.PrGeneratedAt.Expr(`'pull_request' in payload ? `+exprAutofixFinished+` : null`),
	entityops.RemediationFields.CompletedAt.Expr(`payload.?status.orValue("") == "pr_merged" ? `+exprAutofixFinished+` : null`),
	entityops.RemediationFields.Source.Expr(providerkit.ExprInstallationName),
	entityops.RemediationFields.Metadata.Expr(`{
	"location": payload.?autofix_target.?location.orValue(""),
	"codeRepoId": payload.?autofix_target.?code_repo_id.orValue(0),
	"triggeredBy": payload.?triggered_by.?type.orValue(""),
	"triggeredAt": payload.?triggered_at.orValue(0),
	"closeReason": payload.?pull_request.?close_reason.orValue("")
}`),
	providerkit.EdgeExternalIDEntry(entityops.SchemaRemediation, remediation.EdgeVulnerabilities, exprAutofixIssueIDs),
	providerkit.EdgeExternalIDEntry(entityops.SchemaRemediation, remediation.EdgeFindings, exprAutofixIssueIDs),
)

// mapExprScan maps Aikido AI pentest and code analysis assessments to scans, keyed by the assessment id their issues carry
var mapExprScan = providerkit.CelMapExpr(
	entityops.ScanFields.ExternalID.Expr(`payload.?id.orValue("")`),
	entityops.ScanFields.Target.Expr(`payload.?name.orValue("") != "" ? payload.name : payload.?id.orValue("")`),
	entityops.ScanFields.ScanType.Expr(`"VULNERABILITY"`),
	entityops.ScanFields.Status.Expr(`payload.?status.orValue("") == "completed" ? "COMPLETED" : (payload.?status.orValue("") in ["failed", "cancelled"] ? "FAILED" : (payload.?status.orValue("") in ["running", "paused"] ? "PROCESSING" : "PENDING"))`),
	entityops.ScanFields.ScanDate.Expr(`'created_at' in payload ? timestamp(int(payload.created_at)) : null`),
	entityops.ScanFields.PerformedBy.Expr(`"Aikido"`),
	entityops.ScanFields.Tags.Expr(`payload.?`+scanKindKey+`.optMap(k, [k]).orValue([])`),
	entityops.ScanFields.Metadata.Expr(`{
	"kind": payload.?`+scanKindKey+`.orValue(""),
	"status": payload.?status.orValue("")
}`),
)

// mapExprCheckResult maps one compliance report rule to a check result, the framework and control reference in tags link it to its control
var mapExprCheckResult = providerkit.CelMapExpr(
	entityops.CheckResultFields.ParentExternalID.Expr(`"aikido:" + payload.?framework.orValue("") + ":" + payload.?control_code.orValue("") + ":" + payload.?title.orValue("")`),
	entityops.CheckResultFields.Status.Expr(`payload.?status.orValue("") == "complying" ? "PASS" : (payload.?status.orValue("") == "failing" ? "FAIL" : "UNKNOWN")`),
	entityops.CheckResultFields.Details.Expr(`payload.?measure_title.orValue("") != "" ? payload.?title.orValue("") + ": " + payload.measure_title : payload.?title.orValue("")`),
	entityops.CheckResultFields.Tags.Expr(`[payload.?standard.orValue(""), payload.?ref_code.orValue("")].filter(t, t != "")`),
	entityops.CheckResultFields.Source.Expr(providerkit.ExprInstallationName),
)

// mapExprDirectoryAccount maps Aikido workspace users to directory accounts
var mapExprDirectoryAccount = providerkit.CelMapExpr(
	entityops.DirectoryAccountFields.ExternalID.Expr(`string(payload.?id.orValue(""))`),
	entityops.DirectoryAccountFields.CanonicalEmail.Expr(`payload.?email.orValue("")`),
	entityops.DirectoryAccountFields.DisplayName.Expr(`payload.?full_name.orValue("") != "" ? payload.full_name : payload.?email.orValue("")`),
	entityops.DirectoryAccountFields.AccountType.Expr(`"USER"`),
	entityops.DirectoryAccountFields.Status.Expr(`int(payload.?active.orValue(0)) == 1 ? "ACTIVE" : "INACTIVE"`),
	entityops.DirectoryAccountFields.Metadata.Expr(`{
	"role": payload.?role.orValue(""),
	"authType": payload.?auth_type.orValue(""),
	"lastLoginAt": payload.?last_login_timestamp.orValue(0)
}`),
	entityops.DirectoryAccountFields.Profile.Expr("payload"),
	entityops.DirectoryAccountFields.DirectoryName.Expr(providerkit.ExprInstallationName),
	entityops.DirectoryAccountFields.PrimarySource.Expr("installation.primary_directory"),
)

// mapExprDirectoryGroup maps Aikido teams to directory groups
var mapExprDirectoryGroup = providerkit.CelMapExpr(
	entityops.DirectoryGroupFields.ExternalID.Expr(`string(payload.?id.orValue(""))`),
	entityops.DirectoryGroupFields.DisplayName.Expr(`payload.?name.orValue("")`),
	entityops.DirectoryGroupFields.Status.Expr(`payload.?active.orValue(false) ? "ACTIVE" : "INACTIVE"`),
	entityops.DirectoryGroupFields.MemberCount.Expr(`int(payload.?`+teamMemberCountKey+`.orValue(0))`),
	entityops.DirectoryGroupFields.Metadata.Expr(`{
	"externalSource": payload.?external_source.orValue(""),
	"externalSourceId": payload.?external_source_id.orValue(""),
	"responsibilities": payload.?responsibilities.orValue([])
}`),
	entityops.DirectoryGroupFields.Profile.Expr("payload"),
	entityops.DirectoryGroupFields.DirectoryName.Expr(providerkit.ExprInstallationName),
)

// mapExprDirectoryMembership maps Aikido team members to directory memberships linked by the user and team external ids
var mapExprDirectoryMembership = providerkit.CelMapExpr(
	entityops.DirectoryMembershipFields.DirectoryAccountID.Expr(`payload.?`+membershipUserIDKey+`.orValue("")`),
	entityops.DirectoryMembershipFields.DirectoryGroupID.Expr(`payload.?`+membershipTeamIDKey+`.orValue("")`),
	entityops.DirectoryMembershipFields.Role.Expr(`"MEMBER"`),
	entityops.DirectoryMembershipFields.DirectoryName.Expr(providerkit.ExprInstallationName),
)

// assetEntries returns the mapping entries shared by every Aikido asset kind
func assetEntries(variant, assetType string, extras ...entityops.MappingEntry) string {
	entries := []entityops.MappingEntry{
		entityops.AssetFields.SourceIdentifier.Expr(strconv.Quote(assetSourcePrefix(variant)) + ` + string(payload.?id.orValue(""))`),
		entityops.AssetFields.Name.Expr(exprAssetName),
		entityops.AssetFields.DisplayName.Expr(exprAssetName),
		entityops.AssetFields.AssetType.Expr(strconv.Quote(assetType)),
		entityops.AssetFields.ObservedAt.Expr(exprAssetObserved),
		entityops.AssetFields.Tags.Expr(exprAssetTags),
	}

	return providerkit.CelMapExpr(append(entries, extras...)...)
}

// assetSourcePrefix is the source identifier prefix of an Aikido asset kind, followed by the Aikido id
func assetSourcePrefix(variant string) string {
	return "aikido:" + variant + ":"
}

// connectedAsset is the source identifier of the asset of variant whose Aikido id is at key, empty when the key is absent
func connectedAsset(variant, key string) string {
	return `payload.?` + key + `.optMap(id, ` + strconv.Quote(assetSourcePrefix(variant)) + ` + string(id)).orValue("")`
}

// connectedAssets binds the source identifiers of the assets an Aikido asset connects to
func connectedAssets(exprs ...string) entityops.MappingEntry {
	return providerkit.ConnectedAssetsEntry(`[` + strings.Join(exprs, ", ") + `].filter(s, s != "")`)
}

// exprDomainLinkedResource is the source identifier of the code repository or container a domain is linked to
var exprDomainLinkedResource = `payload.?linked_resource.?id.optMap(id, (payload.linked_resource.?type.orValue("") == "code_repo" ? ` +
	strconv.Quote(assetSourcePrefix(assetVariantCodeRepository)) + ` : ` + strconv.Quote(assetSourcePrefix(assetVariantContainer)) + `) + string(id)).orValue("")`

var (
	// mapExprCodeRepositoryAsset maps Aikido code repositories to assets
	mapExprCodeRepositoryAsset = assetEntries(assetVariantCodeRepository, "REPOSITORY",
		entityops.AssetFields.Identifier.Expr(`payload.?external_repo_id.orValue("")`),
		entityops.AssetFields.Website.Expr(`payload.?url.orValue("")`),
		entityops.AssetFields.Categories.Expr(`["repository"] + payload.?provider.optMap(p, [p]).orValue([])`),
		entityops.AssetFields.AccessModelName.Expr(exprAssetAccessModel),
		entityops.AssetFields.AssetDataClassificationName.Expr(exprAssetDataClassification),
	)
	// mapExprContainerAsset maps Aikido container images to assets connected to their source repository and cloud
	mapExprContainerAsset = assetEntries(assetVariantContainer, "TECHNOLOGY",
		entityops.AssetFields.Categories.Expr(`["container"] + payload.?provider.optMap(p, [p]).orValue([]) + payload.?distro.optMap(d, [d]).orValue([])`),
		connectedAssets(connectedAsset(assetVariantCodeRepository, "linked_code_repo_id"), connectedAsset(assetVariantCloud, "cloud_id")),
		entityops.AssetFields.AccessModelName.Expr(exprAssetAccessModel),
		entityops.AssetFields.AssetDataClassificationName.Expr(exprAssetDataClassification),
	)
	// mapExprCloudAsset maps Aikido cloud environments to assets, the identifier is the provider account ID
	mapExprCloudAsset = assetEntries(assetVariantCloud, "TECHNOLOGY",
		entityops.AssetFields.Identifier.Expr(`payload.?external_id.orValue("")`),
		entityops.AssetFields.Categories.Expr(`["cloud"] + payload.?provider.optMap(p, [p]).orValue([]) + payload.?environment.optMap(e, [e]).orValue([])`),
	)
	// mapExprDomainAsset maps Aikido monitored domains to assets connected to the repository or container they serve
	mapExprDomainAsset = assetEntries(assetVariantDomain, "DOMAIN",
		entityops.AssetFields.Identifier.Expr(exprAssetName),
		entityops.AssetFields.Website.Expr(exprAssetWebsite),
		entityops.AssetFields.Categories.Expr(`["domain"] + payload.?kind.optMap(k, [k]).orValue([])`),
		connectedAssets(exprDomainLinkedResource),
	)
	// mapExprSubdomainAsset maps subdomains discovered under a monitored domain to assets connected to that domain
	mapExprSubdomainAsset = assetEntries(assetVariantSubdomain, "DOMAIN",
		entityops.AssetFields.Identifier.Expr(exprAssetName),
		entityops.AssetFields.Website.Expr(exprAssetWebsite),
		entityops.AssetFields.Categories.Expr(`["subdomain"]`),
		connectedAssets(connectedAsset(assetVariantDomain, subdomainDomainIDKey)),
	)
	// mapExprVirtualMachineAsset maps Aikido cloud virtual machines to assets connected to their cloud, the identifier is the provider instance ID
	mapExprVirtualMachineAsset = assetEntries(assetVariantVirtualMachine, "TECHNOLOGY",
		entityops.AssetFields.Identifier.Expr(`payload.?external_id.orValue("")`),
		entityops.AssetFields.Region.Expr(`payload.?region.orValue("")`),
		entityops.AssetFields.Categories.Expr(`["virtual_machine"] + payload.?cloud_kind.optMap(k, [k]).orValue([]) + payload.?environment.optMap(e, [e]).orValue([])`),
		connectedAssets(connectedAsset(assetVariantCloud, "cloud_id")),
	)
	// mapExprDeviceAsset maps Aikido endpoint protection devices to assets owned by the device user, the identifier is the serial number
	mapExprDeviceAsset = assetEntries(assetVariantDevice, "DEVICE",
		entityops.AssetFields.Identifier.Expr(`payload.?serial_number.orValue("")`),
		entityops.AssetFields.InternalOwner.Expr(`payload.?user.orValue("")`),
		entityops.AssetFields.Categories.Expr(`["device"] + payload.?os.optMap(o, [o]).orValue([]) + payload.?status.optMap(s, [s]).orValue([])`),
	)
)
