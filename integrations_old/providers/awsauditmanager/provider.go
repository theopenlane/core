package awsauditmanager

import (
	"context"
	"encoding/json"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/awskit"
	"github.com/theopenlane/core/internal/integrations/providers/awssts"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

// TypeAWSAuditManager identifies the AWS Audit Manager provider
const TypeAWSAuditManager types.ProviderType = "awsauditmanager"

const (
	// ClientAWSAuditManager identifies the AWS Audit Manager client descriptor
	ClientAWSAuditManager types.ClientName = "auditmanager"

	auditManagerDefaultSession       = "openlane-auditmanager"
	auditListMaxOne            int32 = 1
)

// awsAuditManagerCredentialsSchema is the JSON Schema for AWS Audit Manager credentials.
var awsAuditManagerCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["roleArn","externalId","homeRegion"],"allOf":[{"if":{"properties":{"accountScope":{"const":"specific"}},"required":["accountScope"]},"then":{"required":["accountIds"]}}],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly identifier for this AWS integration."},"roleArn":{"type":"string","title":"IAM Role ARN","description":"Cross-account role Openlane should assume in the tenant environment."},"externalId":{"type":"string","title":"External ID","description":"External ID required in the tenant role trust policy to prevent confused deputy attacks."},"homeRegion":{"type":"string","title":"Home Region","description":"Primary AWS region where Audit Manager data is managed.","default":"us-east-1"},"region":{"type":"string","title":"AWS Region (Legacy)","description":"Legacy alias for home region; use Security Hub Home Region when possible.","default":"us-east-1"},"linkedRegions":{"type":"array","title":"Linked Regions","description":"Optional list of regions to filter findings by region.","items":{"type":"string"}},"organizationId":{"type":"string","title":"AWS Organization ID","description":"Optional AWS Organizations identifier (for traceability and scoping)."},"accountScope":{"type":"string","title":"Account Scope","description":"Use all accessible accounts from the delegated admin role, or limit to specific account IDs.","default":"all","enum":["all","specific"]},"accountIds":{"type":"array","title":"Account IDs","description":"Required when Account Scope is set to specific.","items":{"type":"string"}},"sessionDuration":{"type":"string","title":"Session Duration","description":"Optional session duration (Go duration string, e.g. 1h30m)."},"sessionName":{"type":"string","title":"Session Name","description":"Optional session name override for STS AssumeRole calls."},"accessKeyId":{"type":"string","title":"Access Key ID","description":"Optional source credential key when Openlane cannot use runtime IAM credentials."},"secretAccessKey":{"type":"string","title":"Secret Access Key","description":"Optional source credential secret paired with Access Key ID.","secret":true},"sessionToken":{"type":"string","title":"Session Token","description":"Optional source session token when using temporary source credentials.","secret":true},"auditManagerAssessmentId":{"type":"string","title":"Audit Manager Assessment ID","description":"Optional assessment identifier to scope evidence pulls."},"systemsManagerDocument":{"type":"string","title":"Systems Manager Document","description":"Optional SSM document used when triggering compliance scans."},"configAggregatorName":{"type":"string","title":"AWS Config Aggregator","description":"Aggregator name that should be queried for compliance status."},"controlTowerRegions":{"type":"array","title":"Regions to Inspect","description":"Optional list of AWS regions that should be included when evaluating compliance.","items":{"type":"string"}},"tags":{"type":"object","title":"Default Tags","description":"Optional key/value map added to generated integrations for traceability.","additionalProperties":{"type":"string"}}}}`)

// Builder returns the AWS Audit Manager provider builder
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeAWSAuditManager,
		SpecFunc:     awsAuditManagerSpec,
		BuildFunc: func(ctx context.Context, s spec.ProviderSpec) (types.Provider, error) {
			return awssts.Builder(
				TypeAWSAuditManager,
				awssts.WithOperations(auditManagerOperations()),
				awssts.WithClientDescriptors(auditManagerClientDescriptors()),
			).Build(ctx, s)
		},
	}
}

// awsAuditManagerSpec returns the static provider specification for the AWS Audit Manager provider.
func awsAuditManagerSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "awsauditmanager",
		DisplayName: "AWS Audit & Compliance",
		Category:    "compliance",
		AuthType:    types.AuthKindAWSFederation,
		Active:      lo.ToPtr(true),
		Visible:     lo.ToPtr(false),
		LogoURL:     "",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws/overview",
		Labels: map[string]string{
			"vendor":  "aws",
			"service": "audit-manager",
		},
		CredentialsSchema: awsAuditManagerCredentialsSchema,
		Description:       "Collect AWS Audit Manager assessment metadata for compliance posture checks using STS role assumption.",
	}
}

// auditManagerOperations returns the operation descriptors for the AWS Audit Manager provider
func auditManagerOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        "audit_manager.assessments.list",
			Kind:        types.OperationKindScanSettings,
			Description: "List Audit Manager assessments to validate access.",
			Client:      ClientAWSAuditManager,
			Run:         runAuditAssessments,
		},
	}
}

// auditManagerClientDescriptors returns the client descriptors for the AWS Audit Manager provider
func auditManagerClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeAWSAuditManager, ClientAWSAuditManager, "AWS Audit Manager client", pooledAuditManagerClient)
}

// pooledAuditManagerClient builds the AWS Audit Manager client for pooling
func pooledAuditManagerClient(ctx context.Context, credential types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	client, _, err := buildAuditManagerClient(ctx, credential)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(client), nil
}

// newAuditManagerSDKClient wraps auditmanager.NewFromConfig for use with generic helpers
func newAuditManagerSDKClient(cfg awssdk.Config) *auditmanager.Client {
	return auditmanager.NewFromConfig(cfg)
}

// metadataFromCredential extracts and validates AWS metadata from a credential set
func metadataFromCredential(credential types.CredentialSet) (awskit.Metadata, error) {
	if len(credential.ProviderData) == 0 {
		return awskit.Metadata{}, ErrMetadataMissing
	}

	parsed, err := awskit.MetadataFromProviderData(credential.ProviderData, auditManagerDefaultSession)
	if err != nil {
		return awskit.Metadata{}, err
	}

	if parsed.RoleARN == "" {
		return awskit.Metadata{}, ErrRoleARNMissing
	}

	if parsed.Region == "" {
		return awskit.Metadata{}, ErrRegionMissing
	}

	return parsed, nil
}

// buildAuditManagerClient constructs an Audit Manager client from the stored credential
func buildAuditManagerClient(ctx context.Context, credential types.CredentialSet) (*auditmanager.Client, awskit.Metadata, error) {
	var zero *auditmanager.Client

	meta, err := metadataFromCredential(credential)
	if err != nil {
		return zero, awskit.Metadata{}, err
	}

	cfg, err := awskit.BuildAWSConfig(ctx, meta.Region, awskit.CredentialsFromMetadata(meta), awskit.AssumeRole{
		RoleARN:         meta.RoleARN,
		ExternalID:      meta.ExternalID,
		SessionName:     meta.SessionName,
		SessionDuration: meta.SessionDuration,
	})
	if err != nil {
		return zero, meta, err
	}

	return newAuditManagerSDKClient(cfg), meta, nil
}

// resolveAuditManagerClient returns a pooled client when available or builds one on demand
func resolveAuditManagerClient(ctx context.Context, input types.OperationInput) (*auditmanager.Client, awskit.Metadata, error) {
	if client, ok := types.ClientInstanceAs[*auditmanager.Client](input.Client); ok {
		meta, err := metadataFromCredential(input.Credential)
		if err != nil {
			return nil, awskit.Metadata{}, err
		}

		return client, meta, nil
	}

	return buildAuditManagerClient(ctx, input.Credential)
}

type auditManagerDetails struct {
	RoleArn   string `json:"roleArn,omitempty"`
	Region    string `json:"region,omitempty"`
	AccountID string `json:"accountId,omitempty"`
}

// runAuditAssessments validates AWS Audit Manager access via ListAssessments
func runAuditAssessments(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, meta, err := resolveAuditManagerClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	_, err = client.ListAssessments(ctx, &auditmanager.ListAssessmentsInput{
		MaxResults: awssdk.Int32(auditListMaxOne),
	})
	if err != nil {
		return providerkit.OperationFailure("AWS Audit Manager list assessments failed", err, auditManagerDetails{
			Region: meta.Region,
		})
	}

	details := auditManagerDetails{
		RoleArn: meta.RoleARN,
		Region:  meta.Region,
	}

	if meta.AccountID != "" {
		details.AccountID = meta.AccountID
	}

	summary := "AWS Audit Manager reachable"
	if meta.AccountID != "" {
		summary = fmt.Sprintf("AWS Audit Manager reachable for account %s", meta.AccountID)
	}

	return providerkit.OperationSuccess(summary, details), nil
}
