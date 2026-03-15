package awsauditmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	auditmanagertypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultSessionName  = "openlane-awsauditmanager"
	assessmentsPageSize = int32(100)
)

// assessmentsConfig holds per-invocation parameters for the assessments.list operation
type assessmentsConfig struct {
	// Status filters assessments by enrollment status. Valid values: ACTIVE, INACTIVE. Empty returns all.
	Status string `json:"status,omitempty" jsonschema:"title=Status Filter,description=Filter assessments by status (ACTIVE INACTIVE). Empty returns all."`
	// MaxAssessments caps the total number of assessments returned; 0 means no limit.
	MaxAssessments int `json:"maxAssessments,omitempty" jsonschema:"title=Max Assessments,description=Maximum number of assessments to return. 0 means no limit."`
}

type healthDetails struct {
	Region        string `json:"region"`
	RoleARN       string `json:"roleArn,omitempty"`
	AccountID     string `json:"accountId,omitempty"`
	AccountStatus string `json:"accountStatus"`
}

// assessmentSummary captures the fields from AssessmentMetadataItem that are useful for compliance posture
type assessmentSummary struct {
	ID              string    `json:"id,omitempty"`
	Name            string    `json:"name,omitempty"`
	ComplianceType  string    `json:"complianceType,omitempty"`
	Status          string    `json:"status,omitempty"`
	DelegationCount int32     `json:"delegationCount,omitempty"`
	CreationTime    time.Time `json:"creationTime,omitempty"`
	LastUpdated     time.Time `json:"lastUpdated,omitempty"`
}

type assessmentsResult struct {
	Region      string              `json:"region"`
	RoleARN     string              `json:"roleArn,omitempty"`
	Total       int                 `json:"total"`
	Assessments []assessmentSummary `json:"assessments"`
}

// buildAuditManagerClient constructs an AWS Audit Manager client using STS AssumeRole
func buildAuditManagerClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	if len(req.Credential.ProviderData) == 0 {
		return nil, ErrCredentialMetadataRequired
	}

	meta, err := awskit.MetadataFromProviderData(req.Credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, fmt.Errorf("awsauditmanager: metadata decode failed: %w", err)
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
		return nil, fmt.Errorf("awsauditmanager: aws config build failed: %w", err)
	}

	return auditmanager.NewFromConfig(cfg), nil
}

// runHealthOperation validates Audit Manager access via GetAccountStatus.
// This confirms both that the STS credentials work and that Audit Manager is
// enrolled for the account. An INACTIVE status is reported as a non-error since
// it indicates successful access to the service.
func runHealthOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	amClient, ok := client.(*auditmanager.Client)
	if !ok {
		return nil, ErrClientType
	}

	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, err
	}

	resp, err := amClient.GetAccountStatus(ctx, &auditmanager.GetAccountStatusInput{})
	if err != nil {
		return nil, fmt.Errorf("awsauditmanager: GetAccountStatus failed: %w", err)
	}

	details := healthDetails{
		Region:        meta.Region,
		RoleARN:       meta.RoleARN,
		AccountStatus: string(resp.Status),
	}

	if meta.AccountID != "" {
		details.AccountID = meta.AccountID
	}

	return jsonx.ToRawMessage(details)
}

// runAssessmentsListOperation paginates through all Audit Manager assessments and returns
// their compliance type, status, and evidence counts for compliance posture reporting.
// The v1 implementation fetched exactly one assessment and returned only connectivity metadata;
// this implementation performs full collection with optional status filtering.
func runAssessmentsListOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, config json.RawMessage) (json.RawMessage, error) {
	amClient, ok := client.(*auditmanager.Client)
	if !ok {
		return nil, ErrClientType
	}

	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, err
	}

	var cfg assessmentsConfig
	if err := jsonx.UnmarshalIfPresent(config, &cfg); err != nil {
		return nil, err
	}

	input := &auditmanager.ListAssessmentsInput{
		MaxResults: awssdk.Int32(assessmentsPageSize),
	}

	if status := strings.ToUpper(cfg.Status); status != "" {
		input.Status = auditmanagertypes.AssessmentStatus(status)
	}

	var (
		summaries []assessmentSummary
		nextToken *string
	)

collectLoop:
	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}

		resp, err := amClient.ListAssessments(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("awsauditmanager: ListAssessments failed: %w", err)
		}

		for i := range resp.AssessmentMetadata {
			if cfg.MaxAssessments > 0 && len(summaries) >= cfg.MaxAssessments {
				break collectLoop
			}

			item := resp.AssessmentMetadata[i]
			s := assessmentSummary{
				ID:              awssdk.ToString(item.Id),
				Name:            awssdk.ToString(item.Name),
				ComplianceType:  awssdk.ToString(item.ComplianceType),
				Status:          string(item.Status),
				DelegationCount: int32(len(item.Delegations)),
			}

			if item.CreationTime != nil {
				s.CreationTime = *item.CreationTime
			}

			if item.LastUpdated != nil {
				s.LastUpdated = *item.LastUpdated
			}

			summaries = append(summaries, s)
		}

		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}

		nextToken = resp.NextToken
	}

	return jsonx.ToRawMessage(assessmentsResult{
		Region:      meta.Region,
		RoleARN:     meta.RoleARN,
		Total:       len(summaries),
		Assessments: summaries,
	})
}
