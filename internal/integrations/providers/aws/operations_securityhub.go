package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	awsSecurityHubAlertTypeFinding = "finding"
	awsSecurityHubMaxPageSize      = 100
	awsSecurityHubDefaultPageSize  = 100
)

type securityHubFindingsConfig struct {
	// PageSize overrides the page size per request
	PageSize int `json:"page_size,omitempty" jsonschema:"description=Optional page size override (max 100)."`
	// MaxFindings limits the total number of findings returned
	MaxFindings int `json:"max_findings,omitempty" jsonschema:"description=Optional cap on total findings returned."`
	// Severity filters findings by severity label
	Severity types.LowerString `json:"severity,omitempty" jsonschema:"description=Optional severity label filter (low, medium, high, critical)."`
	// RecordState filters findings by record state
	RecordState types.UpperString `json:"record_state,omitempty" jsonschema:"description=Optional record state filter (ACTIVE, ARCHIVED)."`
	// WorkflowStatus filters findings by workflow status
	WorkflowStatus types.UpperString `json:"workflow_status,omitempty" jsonschema:"description=Optional workflow status filter (NEW, NOTIFIED, RESOLVED, SUPPRESSED)."`
	// IncludePayloads controls whether raw payloads are returned
	IncludePayloads bool `json:"include_payloads,omitempty" jsonschema:"description=Return raw finding payloads in the response (defaults to false)."`
}

var securityHubFindingsSchema = operations.SchemaFrom[securityHubFindingsConfig]()

// awsSecurityHubOperations lists the AWS Security Hub operations supported by this provider.
func awsSecurityHubOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:         types.OperationVulnerabilitiesCollect,
			Kind:         types.OperationKindCollectFindings,
			Description:  "Collect AWS Security Hub findings for vulnerability ingestion.",
			Client:       ClientAWSSecurityHub,
			Run:          runAWSSecurityHubFindings,
			ConfigSchema: securityHubFindingsSchema,
		},
	}
}

// runAWSSecurityHubFindings collects Security Hub findings for ingestion.
func runAWSSecurityHubFindings(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, meta, err := resolveSecurityHubClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	cfg := securityHubFindingsConfig{PageSize: awsSecurityHubDefaultPageSize}
	if err := operations.DecodeConfig(input.Config, &cfg); err != nil {
		return types.OperationResult{}, err
	}

	pageSize := cfg.PageSize
	if pageSize <= 0 {
		pageSize = awsSecurityHubDefaultPageSize
	}
	if pageSize > awsSecurityHubMaxPageSize {
		pageSize = awsSecurityHubMaxPageSize
	}

	maxFindings := cfg.MaxFindings
	severityFilter := cfg.Severity.String()
	recordStateFilter := cfg.RecordState.String()
	workflowFilter := cfg.WorkflowStatus.String()

	var (
		envelopes []types.AlertEnvelope
		total     int
		nextToken *string
		filters   = securityHubFiltersFromMetadata(meta)
	)

	for {
		resp, err := client.GetFindings(ctx, &securityhub.GetFindingsInput{
			MaxResults: awssdk.Int32(int32(min(pageSize, math.MaxInt32))), //nolint:gosec // bounds checked via min
			NextToken:  nextToken,
			Filters:    filters,
		})
		if err != nil {
			return operations.OperationFailure("AWS Security Hub findings fetch failed", err, map[string]any{
				"region": meta.Region,
			})
		}

		for _, finding := range resp.Findings {
			if maxFindings > 0 && total >= maxFindings {
				break
			}

			severityLabel := ""
			if finding.Severity != nil {
				severityLabel = strings.ToLower(string(finding.Severity.Label))
			}

			recordState := string(finding.RecordState)
			workflowStatus := ""
			if finding.Workflow != nil {
				workflowStatus = string(finding.Workflow.Status)
			}

			if severityFilter != "" && severityLabel != severityFilter {
				continue
			}

			if recordStateFilter != "" && recordState != recordStateFilter {
				continue
			}

			if workflowFilter != "" && workflowStatus != workflowFilter {
				continue
			}

			payload, err := json.Marshal(finding)
			if err != nil {
				return operations.OperationFailure("AWS Security Hub finding serialization failed", err, map[string]any{
					"region": meta.Region,
				})
			}

			resourceID := ""
			for _, resource := range finding.Resources {
				if resource.Id == nil || *resource.Id == "" {
					continue
				}
				resourceID = *resource.Id
				break
			}
			envelopes = append(envelopes, types.AlertEnvelope{
				AlertType: awsSecurityHubAlertTypeFinding,
				Resource:  resourceID,
				Payload:   payload,
			})
			total++
		}

		if maxFindings > 0 && total >= maxFindings {
			break
		}

		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}
		nextToken = resp.NextToken
	}

	alertTypeCounts := map[string]int{
		awsSecurityHubAlertTypeFinding: total,
	}
	details := map[string]any{
		"region":            meta.Region,
		"alerts_total":      total,
		"alert_type_counts": alertTypeCounts,
	}
	details = operations.AddPayloadIf(details, cfg.IncludePayloads, "alerts", envelopes)

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d Security Hub findings", total),
		Details: details,
	}, nil
}

// newSecurityHubClient wraps securityhub.NewFromConfig for use with generic helpers
func newSecurityHubClient(cfg awssdk.Config) *securityhub.Client {
	return securityhub.NewFromConfig(cfg)
}

// resolveSecurityHubClient returns a pooled client when supplied or builds one on demand.
func resolveSecurityHubClient(ctx context.Context, input types.OperationInput) (*securityhub.Client, auth.AWSMetadata, error) {
	return resolveAWSClient(ctx, input, newSecurityHubClient)
}

// buildSecurityHubClient builds a Security Hub client from stored credentials.
func buildSecurityHubClient(ctx context.Context, payload types.CredentialPayload) (*securityhub.Client, auth.AWSMetadata, error) {
	return buildAWSClient(ctx, payload, newSecurityHubClient)
}

func securityHubFiltersFromMetadata(meta auth.AWSMetadata) *securityhubtypes.AwsSecurityFindingFilters {
	var filters securityhubtypes.AwsSecurityFindingFilters

	if meta.AccountScope == auth.AWSAccountScopeSpecific {
		filters.AwsAccountId = toSecurityHubStringFilters(meta.AccountIDs)
	}

	filters.Region = toSecurityHubStringFilters(meta.LinkedRegions)

	if len(filters.AwsAccountId) == 0 && len(filters.Region) == 0 {
		return nil
	}

	return &filters
}

func toSecurityHubStringFilters(values []string) []securityhubtypes.StringFilter {
	filters := lo.FilterMap(values, func(value string, _ int) (securityhubtypes.StringFilter, bool) {
		if value == "" {
			return securityhubtypes.StringFilter{}, false
		}
		return securityhubtypes.StringFilter{
			Comparison: securityhubtypes.StringFilterComparisonEquals,
			Value:      awssdk.String(value),
		}, true
	})
	if len(filters) == 0 {
		return nil
	}
	return filters
}
