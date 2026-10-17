package azuresecuritycenter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// portalRecommendationURLFormat is the Defender for Cloud portal link for one assessment key
	portalRecommendationURLFormat = "https://portal.azure.com/#blade/Microsoft_Azure_Security/RecommendationsBlade/assessmentKey/%s"
	// securityNamespace is the resource provider namespace assessments live under
	securityNamespace = "Microsoft.Security"
	// cveDetailsKey is the additional data key holding the CVE list of a vulnerability recommendation
	cveDetailsKey = "CvesDetails"
	// fixedVersionKey is the additional data key holding the version that fixes a vulnerability recommendation
	fixedVersionKey = "FixedVersion"
	// softwareNameKey is the additional data key holding the vulnerable package name
	softwareNameKey = "SoftwareName"
	// softwareVendorKey is the additional data key holding the package vendor, e.g. debian
	softwareVendorKey = "SoftwareVendor"
	// detectedVersionsKey is the additional data key holding the installed package versions
	detectedVersionsKey = "DetectedSoftwareVersions"
	// maxCvssScoreKey is the additional data key holding the highest CVSS score across the recommendation's CVEs
	maxCvssScoreKey = "MaxCvssScore"
	// cveExternalIDFormat makes a per-CVE external ID from the recommendation ID and CVE ID
	cveExternalIDFormat = "%s/cves/%s"
	// recommendationFindingIDFormat is the subscription-scoped external ID of a finding covering every affected resource of one recommendation
	recommendationFindingIDFormat = "/subscriptions/%s/providers/Microsoft.Security/assessments/%s"
	// unknownValue is the placeholder Azure reports for an unclassified category, never stored
	unknownValue = "Unknown"
	// securityEntityDataSegment separates a parent resource ID from a scanned entity such as a container image
	securityEntityDataSegment = "/securityentitydata/"
	// imageRepositoryPrefix starts a container image entity name
	imageRepositoryPrefix = "repositories-"
	// imageDigestSeparator separates the repository from the digest in a container image entity name
	imageDigestSeparator = "-images-"
)

// Client queries Defender for Cloud through Azure Resource Graph for one consented tenant
type Client struct {
	// arm calls Resource Manager and Resource Graph with the tenant credential
	arm *microsoft.ARMClient
}

// clientBuilder returns the Azure Security Center client builder bound to the operator config
func clientBuilder(cfg Config) func(context.Context, types.ConnectionRequest[securityCenterCred]) (*Client, error) {
	return func(ctx context.Context, req types.ConnectionRequest[securityCenterCred]) (*Client, error) {
		if req.Credential.TenantID == "" {
			return nil, ErrTenantIDMissing
		}

		ctx = logx.WithField(ctx, "tenant_id", req.Credential.TenantID)

		azCred, err := microsoft.NewTokenCredential(cfg, req.Credential.TenantID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create client secret credential")
			return nil, ErrCredentialBuildFailed
		}

		armClient, err := microsoft.NewARMClient(cfg, azCred)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create arm client")
			return nil, ErrCredentialBuildFailed
		}

		return &Client{arm: armClient}, nil
	}
}

// collect runs an assessments query and returns the ingest envelopes shape produces from the assessment payloads
func (c *Client) collect(ctx context.Context, query string, requested []string, shape func([]AssessmentPayload) []AssessmentPayload) ([]types.MappingEnvelope, error) {
	assessments, err := c.queryRows[assessment](ctx, query, requested)
	if err != nil {
		return nil, err
	}

	payloads := make([]AssessmentPayload, 0, len(assessments))

	for _, a := range assessments {
		payload, err := buildPayload(a)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("assessment_id", a.ID).Msg("azuresecuritycenter: failed to decode assessment metadata")
			return nil, ErrAssessmentDecode
		}

		payloads = append(payloads, payload)
	}

	shaped := shape(payloads)
	envelopes := make([]types.MappingEnvelope, 0, len(shaped))

	for _, p := range shaped {
		envelope, err := providerkit.MarshalEnvelope(p.ID, p, ErrIngestPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("assessment_id", p.ID).Msg("azuresecuritycenter: failed marshal envelope")
			return nil, ErrIngestPayloadEncode
		}

		envelopes = append(envelopes, envelope)
	}

	return envelopes, nil
}

// recommendationFindings merges per-resource assessments into one finding per subscription and recommendation, with every affected resource in targets
func recommendationFindings(payloads []AssessmentPayload) []AssessmentPayload {
	var findings []AssessmentPayload

	index := map[string]int{}

	for _, p := range payloads {
		id := fmt.Sprintf(recommendationFindingIDFormat, p.SubscriptionID, p.AssessmentKey)

		i, seen := index[id]
		if !seen {
			index[id] = len(findings)

			p.ID = id
			p.ExternalURI = portalRecommendationURL(p.AssessmentKey)
			findings = append(findings, p)

			continue
		}

		mergeRecommendation(&findings[i], p)
	}

	return findings
}

// mergeRecommendation folds another affected resource's assessment into a recommendation finding
func mergeRecommendation(finding *AssessmentPayload, p AssessmentPayload) {
	finding.Resources = lo.Uniq(append(finding.Resources, p.Resources...))
	finding.AffectedResources = append(finding.AffectedResources, p.AffectedResources...)
	finding.Categories = lo.Uniq(append(finding.Categories, p.Categories...))
	finding.Standards = lo.Uniq(append(finding.Standards, p.Standards...))
	finding.GovernanceOwner = lo.CoalesceOrEmpty(finding.GovernanceOwner, p.GovernanceOwner)

	// per-resource fields no longer describe the finding once it spans several resources; affected resources keep each one's details
	finding.ResourceID = ""
	finding.ResourceDetails = nil
	finding.AdditionalData = nil

	if p.DiscoveredAt != nil && (finding.DiscoveredAt == nil || p.DiscoveredAt.Before(*finding.DiscoveredAt)) {
		finding.DiscoveredAt = p.DiscoveredAt
	}

	if p.UpdatedAt != nil && (finding.UpdatedAt == nil || p.UpdatedAt.After(*finding.UpdatedAt)) {
		finding.UpdatedAt = p.UpdatedAt
	}
}

// splitCVEs expands every vulnerability recommendation into one payload per CVE
func splitCVEs(payloads []AssessmentPayload) []AssessmentPayload {
	var split []AssessmentPayload

	for _, p := range payloads {
		split = append(split, cvePayloads(p)...)
	}

	return split
}

// cvePayloads splits a vulnerability recommendation into one payload per CVE, using each CVE's own severity and fixed version when present,
// and keeps the recommendation as is when it lists none
func cvePayloads(payload AssessmentPayload) []AssessmentPayload {
	if len(payload.CVEDetails) == 0 {
		return []AssessmentPayload{payload}
	}

	payloads := make([]AssessmentPayload, 0, len(payload.CVEDetails))

	for _, d := range payload.CVEDetails {
		p := payload
		p.ID = fmt.Sprintf(cveExternalIDFormat, payload.ID, d.CveID)
		p.CveID = d.CveID
		p.Severity = lo.CoalesceOrEmpty(d.Severity, payload.Severity)
		p.FixedVersion = lo.CoalesceOrEmpty(d.FixedVersion, payload.FixedVersion)

		payloads = append(payloads, p)
	}

	return payloads
}

// buildPayload extracts all mappable fields from an assessment, including CVE details on vulnerability recommendations
func buildPayload(a assessment) (AssessmentPayload, error) {
	props := a.Properties

	payload := AssessmentPayload{
		ID:              a.ID,
		AssessmentKey:   a.Name,
		GovernanceOwner: a.GovernanceOwner,
		StatusCode:      props.Status.Code,
		StatusCause:     props.Status.Cause,
		DisplayName:     props.DisplayName,
		DiscoveredAt:    props.Status.FirstEvaluationDate,
		UpdatedAt:       props.Status.StatusChangeDate,
		ExternalURI:     lo.CoalesceOrEmpty(props.Links.AzurePortalURI, portalRecommendationURL(a.Name)),
		AdditionalData:  props.AdditionalData,
		CVEDetails:      cveDetails(props.AdditionalData[cveDetailsKey]),
	}

	payload.FixedVersion, _ = props.AdditionalData[fixedVersionKey].(string)
	payload.PackageName, _ = props.AdditionalData[softwareNameKey].(string)
	payload.PackageEcosystem, _ = props.AdditionalData[softwareVendorKey].(string)
	payload.DetectedVersions, _ = decodeAdditional[[]string](props.AdditionalData[detectedVersionsKey])

	if score, ok := decodeAdditional[float64](props.AdditionalData[maxCvssScoreKey]); ok {
		payload.CVSSScore = &score
	}

	if len(props.Metadata) > 0 {
		var meta assessmentMetadata
		if err := json.Unmarshal(props.Metadata, &meta); err != nil {
			return AssessmentPayload{}, err
		}

		payload.Metadata = props.Metadata
		payload.DisplayName = lo.CoalesceOrEmpty(payload.DisplayName, meta.DisplayName)
		payload.Severity = meta.Severity
		payload.Description = meta.Description
		payload.Remediation = meta.RemediationDescription
		payload.AssessmentType = meta.AssessmentType
		payload.RecommendationCategory = knownValue(meta.RecommendationCategory)
		payload.Categories = knownCategories(meta.Categories)
		payload.Threats = meta.Threats

		// individual recommendations often carry no classic category, so fall back to the recommendation category
		if len(payload.Categories) == 0 && payload.RecommendationCategory != "" {
			payload.Categories = []string{payload.RecommendationCategory}
		}

		if len(payload.Categories) > 0 {
			payload.Category = payload.Categories[0]
		}
	}

	// compliance control IDs link the finding to every control that references the assessment
	if len(a.ComplianceControls) > 0 {
		payload.Categories = lo.Uniq(append(payload.Categories, a.ComplianceControls...))
	}

	payload.Standards = a.ComplianceStandards

	resourceDetails := newResourceDetails(props.ResourceDetails, payload.ID)
	payload.ResourceDetails = &resourceDetails
	payload.AffectedResources = []ResourceDetails{resourceDetails}
	payload.ResourceID = resourceDetails.ID

	if parsed, err := arm.ParseResourceID(a.ID); err == nil {
		payload.SubscriptionID = parsed.SubscriptionID
	}

	if payload.ResourceID != "" {
		payload.Resources = []string{payload.ResourceID}
	}

	return payload, nil
}

// knownValue returns the value, or empty when Azure reports it as Unknown
func knownValue(value string) string {
	if strings.EqualFold(value, unknownValue) {
		return ""
	}

	return value
}

// knownCategories returns the categories without empty or Unknown entries
func knownCategories(categories []string) []string {
	var known []string

	for _, c := range categories {
		if knownValue(c) != "" {
			known = append(known, c)
		}
	}

	return known
}

// decodeAdditional decodes an additional data value into T; Resource Graph returns these values either as JSON-encoded strings or as decoded JSON
func decodeAdditional[T any](value any) (T, bool) {
	var decoded T

	if value == nil {
		return decoded, false
	}

	if s, ok := value.(string); ok {
		decoded, err := jsonx.Decode[T](json.RawMessage(s))

		return decoded, err == nil
	}

	return decoded, jsonx.RoundTrip(value, &decoded) == nil
}

// cveDetails extracts the CVEs with an identifier from a CvesDetails value
func cveDetails(value any) []cveDetail {
	details, _ := decodeAdditional[[]cveDetail](value)

	var known []cveDetail

	for _, d := range details {
		if d.CveID != "" {
			known = append(known, d)
		}
	}

	return known
}

// newResourceDetails describes the assessed resource, filling name, type, subscription, and resource group from ARM IDs
func newResourceDetails(details assessmentResourceDetails, assessmentID string) ResourceDetails {
	rd := ResourceDetails{
		ID:     lo.CoalesceOrEmpty(details.ID, details.SourceComputerID),
		Source: details.Source,
		Name:   details.MachineName,
	}

	// fall back to the resource the assessment ID is scoped to when the details carry no ID
	if rd.ID == "" {
		rd.ID = assessedResourceID(assessmentID)
	}

	resourceID := rd.ID

	// scanned entities such as container images sit under their parent resource's securityentitydata path
	if i := strings.Index(strings.ToLower(rd.ID), securityEntityDataSegment); i >= 0 {
		resourceID = rd.ID[:i]
		rd.ParentResourceID = resourceID
		rd.Repository, rd.Digest = splitImageEntity(rd.ID[i+len(securityEntityDataSegment):])
	}

	if parsed, err := arm.ParseResourceID(resourceID); err == nil {
		rd.SubscriptionID = parsed.SubscriptionID
		rd.ResourceGroup = parsed.ResourceGroupName

		if rd.ParentResourceID != "" {
			rd.ParentResourceName = parsed.Name
		} else {
			rd.Name = parsed.Name
			rd.Type = parsed.ResourceType.String()
		}
	}

	rd.Name = lo.CoalesceOrEmpty(details.ResourceName, rd.Name)
	rd.Type = lo.CoalesceOrEmpty(rd.Type, details.ResourceType)
	rd.Provider = details.ResourceProvider

	return rd
}

// splitImageEntity splits a container image entity name such as repositories-nginx-images-sha256:abc into its repository and digest
func splitImageEntity(entity string) (string, string) {
	name := strings.TrimPrefix(entity, imageRepositoryPrefix)

	i := strings.LastIndex(name, imageDigestSeparator)
	if i < 0 {
		return "", ""
	}

	return name[:i], name[i+len(imageDigestSeparator):]
}

// assessedResourceID returns the resource an assessment ID is scoped to by walking up past the Microsoft.Security segments
func assessedResourceID(assessmentID string) string {
	parsed, err := arm.ParseResourceID(assessmentID)
	if err != nil {
		return ""
	}

	for parsed != nil && strings.EqualFold(parsed.ResourceType.Namespace, securityNamespace) {
		parsed = parsed.Parent
	}

	if parsed == nil || parsed.Parent == nil {
		return ""
	}

	return parsed.String()
}

// portalRecommendationURL returns the Defender for Cloud portal link for an assessment key, empty when the key is unknown
func portalRecommendationURL(assessmentKey string) string {
	if assessmentKey == "" {
		return ""
	}

	return fmt.Sprintf(portalRecommendationURLFormat, assessmentKey)
}
