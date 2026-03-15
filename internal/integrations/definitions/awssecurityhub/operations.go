package awssecurityhub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultSessionName = "openlane-awssecurityhub"
	defaultPageSize    = int32(100)
	maxPageSize        = int32(100)
)

// findingsConfig holds per-invocation parameters for the vulnerabilities.collect operation.
// Severity, RecordState, and WorkflowStatus are pushed to the Security Hub server-side
// filter so that only matching findings are transferred over the wire.
type findingsConfig struct {
	// Severity filters by ASFF severity label. Valid values: INFORMATIONAL, LOW, MEDIUM, HIGH, CRITICAL.
	Severity string `json:"severity,omitempty" jsonschema:"title=Severity Filter,description=ASFF severity label filter (INFORMATIONAL LOW MEDIUM HIGH CRITICAL)."`
	// RecordState filters by finding record state. Valid values: ACTIVE, ARCHIVED.
	RecordState string `json:"recordState,omitempty" jsonschema:"title=Record State,description=Finding record state filter (ACTIVE ARCHIVED)."`
	// WorkflowStatus filters by finding workflow status. Valid values: NEW, NOTIFIED, RESOLVED, SUPPRESSED.
	WorkflowStatus string `json:"workflowStatus,omitempty" jsonschema:"title=Workflow Status,description=Finding workflow status filter (NEW NOTIFIED RESOLVED SUPPRESSED)."`
	// MaxFindings caps the total number of findings returned; 0 means no limit.
	MaxFindings int `json:"maxFindings,omitempty" jsonschema:"title=Max Findings,description=Maximum number of findings to collect. 0 means no limit."`
	// IncludePayloads controls whether raw finding JSON is included in the output.
	IncludePayloads bool `json:"includePayloads,omitempty" jsonschema:"title=Include Payloads,description=Include raw finding payloads in the operation output."`
}

type hubHealthDetails struct {
	Region       string `json:"region"`
	RoleARN      string `json:"roleArn,omitempty"`
	HubARN       string `json:"hubArn,omitempty"`
	SubscribedAt string `json:"subscribedAt,omitempty"`
}

type findingsResult struct {
	Region         string            `json:"region"`
	TotalCollected int               `json:"totalCollected"`
	SeverityCounts map[string]int    `json:"severityCounts"`
	Findings       []json.RawMessage `json:"findings,omitempty"`
}

// buildSecurityHubClient constructs an AWS Security Hub client using STS AssumeRole
func buildSecurityHubClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	if len(req.Credential.ProviderData) == 0 {
		return nil, ErrCredentialMetadataRequired
	}

	meta, err := awskit.MetadataFromProviderData(req.Credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, fmt.Errorf("awssecurityhub: metadata decode failed: %w", err)
	}

	if meta.RoleARN == "" {
		return nil, ErrRoleARNMissing
	}

	if meta.Region == "" {
		return nil, ErrRegionMissing
	}

	cfg, err := awskit.BuildAWSConfig(ctx, meta.Region, awskit.CredentialsFromMetadata(meta), awskit.AssumeRole{
		RoleARN:         meta.RoleARN,
		ExternalID:      meta.ExternalID,
		SessionName:     meta.SessionName,
		SessionDuration: meta.SessionDuration,
	})
	if err != nil {
		return nil, fmt.Errorf("awssecurityhub: aws config build failed: %w", err)
	}

	return securityhub.NewFromConfig(cfg), nil
}

// runHealthOperation validates Security Hub access by calling DescribeHub.
// This confirms both that the STS credentials work and that Security Hub is
// enabled and reachable in the configured home region.
func runHealthOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	shClient, ok := client.(*securityhub.Client)
	if !ok {
		return nil, ErrClientType
	}

	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, err
	}

	resp, err := shClient.DescribeHub(ctx, &securityhub.DescribeHubInput{})
	if err != nil {
		return nil, fmt.Errorf("awssecurityhub: DescribeHub failed: %w", err)
	}

	details := hubHealthDetails{
		Region:  meta.Region,
		RoleARN: meta.RoleARN,
	}

	if resp.HubArn != nil {
		details.HubARN = *resp.HubArn
	}

	if resp.SubscribedAt != nil {
		details.SubscribedAt = *resp.SubscribedAt
	}

	return jsonx.ToRawMessage(details)
}

// runVulnerabilitiesCollectOperation collects Security Hub findings using server-side filters.
// Severity, record state, and workflow status are applied via the GetFindings Filters field
// rather than post-fetch client-side filtering. Pagination stops as soon as MaxFindings is reached.
func runVulnerabilitiesCollectOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, config json.RawMessage) (json.RawMessage, error) {
	shClient, ok := client.(*securityhub.Client)
	if !ok {
		return nil, ErrClientType
	}

	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, err
	}

	var cfg findingsConfig
	if err := jsonx.UnmarshalIfPresent(config, &cfg); err != nil {
		return nil, err
	}

	pageSize := defaultPageSize
	if cfg.MaxFindings > 0 && int32(cfg.MaxFindings) < maxPageSize { //nolint:gosec // bounded by maxPageSize
		pageSize = int32(cfg.MaxFindings) //nolint:gosec
	}

	filters := buildSecurityHubFilters(meta, cfg)

	var (
		envelopes      []json.RawMessage
		total          int
		nextToken      *string
		severityCounts = map[string]int{}
	)

collectLoop:
	for {
		input := &securityhub.GetFindingsInput{
			MaxResults: awssdk.Int32(pageSize),
			Filters:    filters,
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		resp, err := shClient.GetFindings(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("awssecurityhub: GetFindings failed: %w", err)
		}

		for i := range resp.Findings {
			if cfg.MaxFindings > 0 && total >= cfg.MaxFindings {
				break collectLoop
			}

			finding := resp.Findings[i]

			if finding.Severity != nil {
				label := strings.ToLower(string(finding.Severity.Label))
				if label != "" {
					severityCounts[label]++
				}
			}

			if cfg.IncludePayloads {
				payload, err := json.Marshal(finding)
				if err != nil {
					return nil, fmt.Errorf("awssecurityhub: finding serialization failed: %w", err)
				}

				envelopes = append(envelopes, payload)
			}

			total++
		}

		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}

		nextToken = resp.NextToken
	}

	result := findingsResult{
		Region:         meta.Region,
		TotalCollected: total,
		SeverityCounts: severityCounts,
	}

	if cfg.IncludePayloads {
		result.Findings = envelopes
	}

	return jsonx.ToRawMessage(result)
}

// buildSecurityHubFilters constructs a server-side filter from credential metadata and per-invocation config.
// Account and region scoping come from the credential; severity, record state, and workflow status
// come from the operation config. All filters are applied server-side via the GetFindings API.
func buildSecurityHubFilters(meta awskit.Metadata, cfg findingsConfig) *securityhubtypes.AwsSecurityFindingFilters {
	var filters securityhubtypes.AwsSecurityFindingFilters

	if meta.AccountScope == awskit.AccountScopeSpecific {
		filters.AwsAccountId = toStringFilters(meta.AccountIDs)
	}

	filters.Region = toStringFilters(meta.LinkedRegions)

	if sev := strings.ToUpper(cfg.Severity); sev != "" {
		filters.SeverityLabel = []securityhubtypes.StringFilter{{
			Comparison: securityhubtypes.StringFilterComparisonEquals,
			Value:      awssdk.String(sev),
		}}
	}

	if rs := strings.ToUpper(cfg.RecordState); rs != "" {
		filters.RecordState = []securityhubtypes.StringFilter{{
			Comparison: securityhubtypes.StringFilterComparisonEquals,
			Value:      awssdk.String(rs),
		}}
	}

	if ws := strings.ToUpper(cfg.WorkflowStatus); ws != "" {
		filters.WorkflowStatus = []securityhubtypes.StringFilter{{
			Comparison: securityhubtypes.StringFilterComparisonEquals,
			Value:      awssdk.String(ws),
		}}
	}

	if len(filters.AwsAccountId) == 0 && len(filters.Region) == 0 &&
		len(filters.SeverityLabel) == 0 && len(filters.RecordState) == 0 &&
		len(filters.WorkflowStatus) == 0 {
		return nil
	}

	return &filters
}

func toStringFilters(values []string) []securityhubtypes.StringFilter {
	out := make([]securityhubtypes.StringFilter, 0, len(values))

	for _, v := range values {
		if v == "" {
			continue
		}

		out = append(out, securityhubtypes.StringFilter{
			Comparison: securityhubtypes.StringFilterComparisonEquals,
			Value:      awssdk.String(v),
		})
	}

	if len(out) == 0 {
		return nil
	}

	return out
}
