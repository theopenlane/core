package awsauditmanager

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	auditmanagertypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"

	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const assessmentsPageSize = int32(100)

// AssessmentsConfig holds per-invocation parameters for the assessments.list operation
type AssessmentsConfig struct {
	// Status filters assessments by enrollment status. Valid values: ACTIVE, INACTIVE. Empty returns all
	Status string `json:"status,omitempty" jsonschema:"title=Status Filter,description=Filter assessments by status (ACTIVE INACTIVE). Empty returns all."`
	// MaxAssessments caps the total number of assessments returned; 0 means no limit
	MaxAssessments int `json:"maxAssessments,omitempty" jsonschema:"title=Max Assessments,description=Maximum number of assessments to return. 0 means no limit."`
}

// AssessmentSummary captures the fields from AssessmentMetadataItem useful for compliance posture
type AssessmentSummary struct {
	// ID is the Audit Manager assessment identifier
	ID string `json:"id,omitempty"`
	// Name is the assessment display name
	Name string `json:"name,omitempty"`
	// ComplianceType is the compliance framework type for the assessment
	ComplianceType string `json:"complianceType,omitempty"`
	// Status is the current assessment status
	Status string `json:"status,omitempty"`
	// DelegationCount is the number of active delegations for the assessment
	DelegationCount int32 `json:"delegationCount,omitempty"`
	// CreationTime is when the assessment was created
	CreationTime time.Time `json:"creationTime,omitempty"`
	// LastUpdated is when the assessment was last updated
	LastUpdated time.Time `json:"lastUpdated,omitempty"`
}

// AssessmentsList lists and returns AWS Audit Manager assessments
type AssessmentsList struct {
	// Region is the AWS region used for the session
	Region string `json:"region"`
	// RoleARN is the assumed role ARN when present
	RoleARN string `json:"roleArn,omitempty"`
	// Total is the total number of returned assessments
	Total int `json:"total"`
	// Assessments is the collected assessment list
	Assessments []AssessmentSummary `json:"assessments"`
}

// Handle adapts assessments listing to the generic operation registration boundary
func (a AssessmentsList) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		var cfg AssessmentsConfig
		if err := jsonx.UnmarshalIfPresent(request.Config, &cfg); err != nil {
			return nil, ErrOperationConfigInvalid
		}

		return a.Run(ctx, request.Credential, c, cfg)
	}
}

// Run paginates through all Audit Manager assessments
func (AssessmentsList) Run(ctx context.Context, credential types.CredentialSet, c *auditmanager.Client, cfg AssessmentsConfig) (json.RawMessage, error) {
	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, ErrCredentialMetadataInvalid
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
			s := AssessmentSummary{
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

	return providerkit.EncodeResult(AssessmentsList{
		Region:      meta.Region,
		RoleARN:     meta.RoleARN,
		Total:       len(summaries),
		Assessments: summaries,
	}, ErrResultEncode)
}
