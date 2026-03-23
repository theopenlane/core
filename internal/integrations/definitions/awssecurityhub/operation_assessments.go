package awssecurityhub

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	auditmanagertypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// assessmentsPageSize is the number of assessments requested per paginated API call.
const assessmentsPageSize = int32(100)

// AssessmentsConfig holds per-invocation parameters for the assessments.list operation.
type AssessmentsConfig struct {
	// Status filters assessments by enrollment status. Valid values: ACTIVE, INACTIVE. Empty returns all.
	Status string `json:"status,omitempty" jsonschema:"title=Status Filter,description=Filter assessments by status (ACTIVE INACTIVE). Empty returns all."`
	// MaxAssessments caps the total number of assessments returned; 0 means no limit.
	MaxAssessments int `json:"maxAssessments,omitempty" jsonschema:"title=Max Assessments,description=Maximum number of assessments to return. 0 means no limit."`
}

// AssessmentSummary captures the fields from AssessmentMetadataItem useful for future compliance ingest.
type AssessmentSummary struct {
	// ID is the Audit Manager assessment identifier.
	ID string `json:"id,omitempty"`
	// Name is the assessment display name.
	Name string `json:"name,omitempty"`
	// ComplianceType is the compliance framework type for the assessment.
	ComplianceType string `json:"complianceType,omitempty"`
	// Status is the current assessment status.
	Status string `json:"status,omitempty"`
	// DelegationCount is the number of active delegations for the assessment.
	DelegationCount int32 `json:"delegationCount,omitempty"`
	// CreationTime is when the assessment was created.
	CreationTime time.Time `json:"creationTime,omitempty"`
	// LastUpdated is when the assessment was last updated.
	LastUpdated time.Time `json:"lastUpdated,omitempty"`
}

// AssessmentsList lists and returns AWS Audit Manager assessments.
//
// TODO: map these summaries into ingest payloads and add upsert contracts once the target schema is defined.
type AssessmentsList struct {
	// Region is the AWS region used for the session.
	Region string `json:"region"`
	// RoleARN is the assumed role ARN when present.
	RoleARN string `json:"roleArn,omitempty"`
	// AccountID is the AWS account identifier when provided in the credential input.
	AccountID string `json:"accountId,omitempty"`
	// Total is the total number of returned assessments.
	Total int `json:"total"`
	// Assessments is the collected assessment list.
	Assessments []AssessmentSummary `json:"assessments"`
}

// Handle adapts assessments listing to the generic operation registration boundary.
func (a AssessmentsList) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(
		AuditManagerClient,
		AssessmentsListOperation,
		ErrOperationConfigInvalid,
		func(ctx context.Context, request types.OperationRequest, client *auditmanager.Client, cfg AssessmentsConfig) (json.RawMessage, error) {
			return a.Run(ctx, request.Credentials, client, cfg)
		},
	)
}

// Run paginates through all Audit Manager assessments.
func (AssessmentsList) Run(ctx context.Context, credentials types.CredentialBindings, c *auditmanager.Client, cfg AssessmentsConfig) (json.RawMessage, error) {
	awsCredential, err := resolveAssumeRoleCredential(credentials)
	if err != nil {
		return nil, err
	}

	input := &auditmanager.ListAssessmentsInput{
		MaxResults: awssdk.Int32(assessmentsPageSize),
	}

	if status := strings.ToUpper(cfg.Status); status != "" {
		input.Status = auditmanagertypes.AssessmentStatus(status)
	}

	var (
		summaries []AssessmentSummary
		nextToken *string
	)

collectLoop:
	for {
		if nextToken != nil {
			input.NextToken = nextToken
		}

		resp, err := c.ListAssessments(ctx, input)
		if err != nil {
			return nil, ErrListAssessmentsFailed
		}

		for i := range resp.AssessmentMetadata {
			if cfg.MaxAssessments > 0 && len(summaries) >= cfg.MaxAssessments {
				break collectLoop
			}

			item := resp.AssessmentMetadata[i]
			summary := AssessmentSummary{
				ID:              awssdk.ToString(item.Id),
				Name:            awssdk.ToString(item.Name),
				ComplianceType:  awssdk.ToString(item.ComplianceType),
				Status:          string(item.Status),
				DelegationCount: int32(len(item.Delegations)),
			}

			if item.CreationTime != nil {
				summary.CreationTime = *item.CreationTime
			}

			if item.LastUpdated != nil {
				summary.LastUpdated = *item.LastUpdated
			}

			summaries = append(summaries, summary)
		}

		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}

		nextToken = resp.NextToken
	}

	return providerkit.EncodeResult(AssessmentsList{
		Region:      awsCredential.HomeRegion,
		RoleARN:     awsCredential.RoleARN,
		AccountID:   awsCredential.AccountID,
		Total:       len(summaries),
		Assessments: summaries,
	}, ErrResultEncode)
}
