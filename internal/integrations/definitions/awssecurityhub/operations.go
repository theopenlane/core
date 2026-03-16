package awssecurityhub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultPageSize = int32(100)
	maxPageSize     = int32(100)
)

// FindingsConfig holds per-invocation parameters for the vulnerabilities.collect operation
// Severity, RecordState, and WorkflowStatus are pushed to the Security Hub server-side
// filter so that only matching findings are transferred over the wire
type FindingsConfig struct {
	// Severity filters by ASFF severity label. Valid values: INFORMATIONAL, LOW, MEDIUM, HIGH, CRITICAL
	Severity string `json:"severity,omitempty" jsonschema:"title=Severity Filter,description=ASFF severity label filter (INFORMATIONAL LOW MEDIUM HIGH CRITICAL)."`
	// RecordState filters by finding record state. Valid values: ACTIVE, ARCHIVED
	RecordState string `json:"recordState,omitempty" jsonschema:"title=Record State,description=Finding record state filter (ACTIVE ARCHIVED)."`
	// WorkflowStatus filters by finding workflow status. Valid values: NEW, NOTIFIED, RESOLVED, SUPPRESSED
	WorkflowStatus string `json:"workflowStatus,omitempty" jsonschema:"title=Workflow Status,description=Finding workflow status filter (NEW NOTIFIED RESOLVED SUPPRESSED)."`
	// MaxFindings caps the total number of findings returned; 0 means no limit
	MaxFindings int `json:"maxFindings,omitempty" jsonschema:"title=Max Findings,description=Maximum number of findings to collect. 0 means no limit."`
}

// HealthCheck holds the result of an AWS Security Hub health check
type HealthCheck struct {
	// Region is the AWS region used for the session
	Region string `json:"region"`
	// RoleARN is the assumed role ARN when present
	RoleARN string `json:"roleArn,omitempty"`
	// HubARN is the Security Hub ARN
	HubARN string `json:"hubArn,omitempty"`
	// SubscribedAt is the Security Hub subscription timestamp
	SubscribedAt string `json:"subscribedAt,omitempty"`
}

// VulnerabilitiesCollect collects AWS Security Hub findings
type VulnerabilitiesCollect struct{}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, request.Credential, c)
	}
}

// Run validates Security Hub access by calling DescribeHub
func (HealthCheck) Run(ctx context.Context, credential types.CredentialSet, c *securityhub.Client) (json.RawMessage, error) {
	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, err
	}

	resp, err := c.DescribeHub(ctx, &securityhub.DescribeHubInput{})
	if err != nil {
		return nil, fmt.Errorf("awssecurityhub: DescribeHub failed: %w", err)
	}

	details := HealthCheck{
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

// Handle adapts vulnerabilities collection to the generic operation registration boundary
func (v VulnerabilitiesCollect) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		var cfg FindingsConfig
		if err := jsonx.UnmarshalIfPresent(request.Config, &cfg); err != nil {
			return nil, err
		}

		return v.Run(ctx, request.Credential, c, cfg)
	}
}

// Run collects Security Hub findings using server-side filters
func (VulnerabilitiesCollect) Run(ctx context.Context, credential types.CredentialSet, c *securityhub.Client, cfg FindingsConfig) (json.RawMessage, error) {
	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, err
	}

	pageSize := defaultPageSize
	if cfg.MaxFindings > 0 && int32(cfg.MaxFindings) < maxPageSize { //nolint:gosec // bounded by maxPageSize
		pageSize = int32(cfg.MaxFindings) //nolint:gosec
	}

	filters := buildSecurityHubFilters(meta, cfg)

	var (
		envelopes []types.MappingEnvelope
		total     int
		nextToken *string
	)

	if cfg.MaxFindings > 0 {
		envelopes = make([]types.MappingEnvelope, 0, cfg.MaxFindings)
	}

collectLoop:
	for {
		input := &securityhub.GetFindingsInput{
			MaxResults: awssdk.Int32(pageSize),
			Filters:    filters,
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		resp, err := c.GetFindings(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("awssecurityhub: GetFindings failed: %w", err)
		}

		for i := range resp.Findings {
			if cfg.MaxFindings > 0 && total >= cfg.MaxFindings {
				break collectLoop
			}

			envelope, err := buildFindingEnvelope(resp.Findings[i])
			if err != nil {
				return nil, err
			}

			envelopes = append(envelopes, envelope)
			total++
		}

		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}

		nextToken = resp.NextToken
	}

	return jsonx.ToRawMessage([]types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: envelopes,
		},
	})
}

// buildSecurityHubFilters constructs a server-side filter from credential metadata and per-invocation config
// Account and region scoping come from the credential; severity, record state, and workflow status
// come from the operation config. All filters are applied server-side via the GetFindings API
func buildSecurityHubFilters(meta awskit.Metadata, cfg FindingsConfig) *securityhubtypes.AwsSecurityFindingFilters {
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

// toStringFilters converts string values into equality filters
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

// buildFindingEnvelope serializes one Security Hub finding into an ingest envelope
func buildFindingEnvelope(finding securityhubtypes.AwsSecurityFinding) (types.MappingEnvelope, error) {
	rawPayload, err := json.Marshal(finding)
	if err != nil {
		return types.MappingEnvelope{}, fmt.Errorf("awssecurityhub: finding serialization failed: %w", err)
	}

	return types.MappingEnvelope{
		Resource: resolveFindingResource(finding),
		Payload:  rawPayload,
	}, nil
}

// resolveFindingResource chooses the best resource identifier for one finding
func resolveFindingResource(finding securityhubtypes.AwsSecurityFinding) string {
	for _, resource := range finding.Resources {
		if resource.Id != nil && *resource.Id != "" {
			return *resource.Id
		}
	}

	if finding.AwsAccountId != nil {
		return awssdk.ToString(finding.AwsAccountId)
	}

	return ""
}
