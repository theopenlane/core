package awssecurityhub

import (
	"context"
	"math"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	auditmanagertypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// assessmentsPageSize is the number of assessments requested per paginated API call
const assessmentsPageSize = int32(100)

// assessmentVariant is the mapping variant for Audit Manager assessment payloads
const assessmentVariant = "assessment"

// AssessmentsConfig holds per-invocation parameters for the assessments.collect operation
type AssessmentsConfig struct {
	// Status filters assessments by enrollment status. Valid values: ACTIVE, INACTIVE. Empty returns all
	Status string `json:"status,omitempty" jsonschema:"title=Status Filter,description=Filter assessments by status (ACTIVE INACTIVE). Empty returns all."`
	// MaxAssessments caps the total number of assessments returned; 0 means no limit
	MaxAssessments int `json:"maxAssessments,omitempty" jsonschema:"title=Max Assessments,description=Maximum number of assessments to return. 0 means no limit."`
}

// AssessmentPayload is the normalized assessment payload emitted for Finding ingest
type AssessmentPayload struct {
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
	CreationTime time.Time `json:"creationTime"`
	// LastUpdated is when the assessment was last updated
	LastUpdated time.Time `json:"lastUpdated"`
	// AccountID is the AWS account identifier
	AccountID string `json:"accountId,omitempty"`
	// Region is the AWS region
	Region string `json:"region,omitempty"`
}

// AssessmentsCollect collects AWS Audit Manager assessments for Finding ingest
type AssessmentsCollect struct{}

// IngestHandle adapts assessment collection to the ingest operation registration boundary
func (a AssessmentsCollect) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(auditManagerClient, assessmentsCollectOperation, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *auditmanager.Client, cfg AssessmentsConfig) ([]types.IngestPayloadSet, error) {
		return a.Run(ctx, request.Credentials, client, cfg)
	})
}

// Run paginates through Audit Manager assessments and emits Finding ingest payloads
func (AssessmentsCollect) Run(ctx context.Context, credentials types.CredentialBindings, c *auditmanager.Client, cfg AssessmentsConfig) ([]types.IngestPayloadSet, error) {
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
		envelopes []types.MappingEnvelope
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
			if cfg.MaxAssessments > 0 && len(envelopes) >= cfg.MaxAssessments {
				break collectLoop
			}

			payload := mapAssessmentPayload(resp.AssessmentMetadata[i], awsCredential)

			envelope, err := providerkit.MarshalEnvelopeVariant(assessmentVariant, payload.AccountID, payload, ErrAssessmentEncode)
			if err != nil {
				return nil, err
			}

			envelopes = append(envelopes, envelope)
		}

		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}

		nextToken = resp.NextToken
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaFinding,
			Envelopes: envelopes,
		},
	}, nil
}

// mapAssessmentPayload converts an Audit Manager assessment metadata item to the ingest payload shape
func mapAssessmentPayload(item auditmanagertypes.AssessmentMetadataItem, credential AssumeRoleCredentialSchema) AssessmentPayload {
	payload := AssessmentPayload{
		ID:              awssdk.ToString(item.Id),
		Name:            awssdk.ToString(item.Name),
		ComplianceType:  awssdk.ToString(item.ComplianceType),
		Status:          string(item.Status),
		DelegationCount: int32(min(len(item.Delegations), math.MaxInt32)), //nolint:gosec // G115: bounded by min
		AccountID:       credential.AccountID,
		Region:          credential.HomeRegion,
	}

	if item.CreationTime != nil {
		payload.CreationTime = *item.CreationTime
	}

	if item.LastUpdated != nil {
		payload.LastUpdated = *item.LastUpdated
	}

	return payload
}
