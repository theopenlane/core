package awssecurityhub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	awsSecurityHubHealth types.OperationName = "health.default"

	awsSecurityHubAlertTypeFinding = "finding"
	awsSecurityHubMaxPageSize      = 100
	awsSecurityHubDefaultPageSize  = 100
	awsSecurityHubDefaultSession   = "openlane-securityhub"
)

type securityHubFindingsConfig struct {
	PageSize        int                 `json:"page_size,omitempty" jsonschema:"description=Optional page size override (max 100)."`
	MaxFindings     int                 `json:"max_findings,omitempty" jsonschema:"description=Optional cap on total findings returned."`
	Severity        helpers.LowerString `json:"severity,omitempty" jsonschema:"description=Optional severity label filter (low, medium, high, critical)."`
	RecordState     helpers.UpperString `json:"record_state,omitempty" jsonschema:"description=Optional record state filter (ACTIVE, ARCHIVED)."`
	WorkflowStatus  helpers.UpperString `json:"workflow_status,omitempty" jsonschema:"description=Optional workflow status filter (NEW, NOTIFIED, RESOLVED, SUPPRESSED)."`
	IncludePayloads bool                `json:"include_payloads,omitempty" jsonschema:"description=Return raw finding payloads in the response (defaults to false)."`
}

var securityHubFindingsSchema = helpers.SchemaFrom[securityHubFindingsConfig]()

// awsSecurityHubOperations lists the AWS Security Hub operations supported by this provider
func awsSecurityHubOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		helpers.HealthOperation(awsSecurityHubHealth, "Validate AWS Security Hub access by listing a finding.", ClientAWSSecurityHub, runAWSSecurityHubHealth),
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

// runAWSSecurityHubHealth validates Security Hub access via GetFindings
func runAWSSecurityHubHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, meta, err := resolveSecurityHubClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	_, err = client.GetFindings(ctx, &securityhub.GetFindingsInput{
		MaxResults: aws.Int32(1),
	})
	if err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "AWS Security Hub list findings failed",
			Details: map[string]any{
				"region": meta.Region,
				"error":  err.Error(),
			},
		}, err
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("AWS Security Hub reachable for region %s", meta.Region),
		Details: map[string]any{
			"region": meta.Region,
		},
	}, nil
}

// runAWSSecurityHubFindings collects Security Hub findings for ingestion
func runAWSSecurityHubFindings(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, meta, err := resolveSecurityHubClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	cfg := securityHubFindingsConfig{PageSize: awsSecurityHubDefaultPageSize}
	if err := helpers.DecodeConfig(input.Config, &cfg); err != nil {
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
	severityFilter := string(cfg.Severity)
	recordStateFilter := string(cfg.RecordState)
	workflowFilter := string(cfg.WorkflowStatus)

	var (
		envelopes []types.AlertEnvelope
		total     int
		nextToken *string
	)

	for {
		resp, err := client.GetFindings(ctx, &securityhub.GetFindingsInput{
			MaxResults: aws.Int32(int32(pageSize)),
			NextToken:  nextToken,
		})
		if err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "AWS Security Hub findings fetch failed",
				Details: map[string]any{
					"region": meta.Region,
					"error":  err.Error(),
				},
			}, err
		}

		for _, finding := range resp.Findings {
			if maxFindings > 0 && total >= maxFindings {
				break
			}

			severityLabel := ""
			if finding.Severity != nil {
				severityLabel = strings.ToLower(helpers.StringFromAny(finding.Severity.Label))
			}

			recordState := strings.ToUpper(helpers.StringFromAny(finding.RecordState))
			workflowStatus := ""
			if finding.Workflow != nil {
				workflowStatus = strings.ToUpper(helpers.StringFromAny(finding.Workflow.Status))
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
				return types.OperationResult{
					Status:  types.OperationStatusFailed,
					Summary: "AWS Security Hub finding serialization failed",
					Details: map[string]any{
						"region": meta.Region,
						"error":  err.Error(),
					},
				}, err
			}

			resourceID := ""
			for _, resource := range finding.Resources {
				if id := helpers.StringFromAny(resource.Id); id != "" {
					resourceID = id
					break
				}
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
	details = helpers.AddPayloadIf(details, cfg.IncludePayloads, "alerts", envelopes)

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d Security Hub findings", total),
		Details: details,
	}, nil
}

type awsSecurityHubMetadata = helpers.AWSMetadata

// resolveSecurityHubClient returns a pooled client when supplied or builds one on demand.
func resolveSecurityHubClient(ctx context.Context, input types.OperationInput) (*securityhub.Client, awsSecurityHubMetadata, error) {
	if client, ok := input.Client.(*securityhub.Client); ok && client != nil {
		meta, err := awsSecurityHubMetadataFromPayload(input.Credential)
		if err != nil {
			return nil, awsSecurityHubMetadata{}, err
		}
		return client, meta, nil
	}

	return buildSecurityHubClient(ctx, input.Credential)
}

// buildSecurityHubClient builds a Security Hub client from stored credentials
func buildSecurityHubClient(ctx context.Context, payload types.CredentialPayload) (*securityhub.Client, awsSecurityHubMetadata, error) {
	meta, err := awsSecurityHubMetadataFromPayload(payload)
	if err != nil {
		return nil, awsSecurityHubMetadata{}, err
	}

	cfg, err := helpers.BuildAWSConfig(ctx, meta.Region, helpers.AWSCredentialsFromPayload(payload), helpers.AWSAssumeRole{
		RoleARN:         meta.RoleARN,
		ExternalID:      meta.ExternalID,
		SessionName:     meta.SessionName,
		SessionDuration: meta.SessionDuration,
	})
	if err != nil {
		return nil, meta, err
	}

	return securityhub.NewFromConfig(cfg), meta, nil
}

// awsSecurityHubMetadataFromPayload extracts required AWS metadata from the payload
func awsSecurityHubMetadataFromPayload(payload types.CredentialPayload) (awsSecurityHubMetadata, error) {
	meta := payload.Data.ProviderData
	if len(meta) == 0 {
		return awsSecurityHubMetadata{}, ErrMetadataMissing
	}

	parsed := helpers.AWSMetadataFromProviderData(meta, awsSecurityHubDefaultSession)
	if err := helpers.RequireString(parsed.RoleARN, ErrRoleARNMissing); err != nil {
		return awsSecurityHubMetadata{}, err
	}
	if err := helpers.RequireString(parsed.Region, ErrRegionMissing); err != nil {
		return awsSecurityHubMetadata{}, err
	}

	return awsSecurityHubMetadata(parsed), nil
}
