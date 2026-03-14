package awssecurityhub

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

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/awskit"
	"github.com/theopenlane/core/internal/integrations/providers/awssts"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TypeAWSSecurityHub identifies the AWS Security Hub provider
const TypeAWSSecurityHub types.ProviderType = "awssecurityhub"

const (
	// ClientAWSSecurityHub identifies the AWS Security Hub client descriptor
	ClientAWSSecurityHub types.ClientName = "securityhub"

	securityHubDefaultSession   = "openlane-securityhub"
	securityHubAlertTypeFinding = "finding"
	securityHubMaxPageSize      = 100
	securityHubDefaultPageSize  = 100
)

// awsSecurityHubCredentialsSchema is the JSON Schema for AWS Security Hub credentials.
var awsSecurityHubCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["roleArn","externalId","homeRegion"],"allOf":[{"if":{"properties":{"accountScope":{"const":"specific"}},"required":["accountScope"]},"then":{"required":["accountIds"]}}],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly identifier for this AWS integration."},"roleArn":{"type":"string","title":"IAM Role ARN","description":"Cross-account role Openlane should assume in the tenant environment."},"externalId":{"type":"string","title":"External ID","description":"External ID required in the tenant role trust policy to prevent confused deputy attacks."},"homeRegion":{"type":"string","title":"Security Hub Home Region","description":"Primary AWS region where Security Hub aggregation is managed.","default":"us-east-1"},"region":{"type":"string","title":"AWS Region (Legacy)","description":"Legacy alias for home region; use Security Hub Home Region when possible.","default":"us-east-1"},"linkedRegions":{"type":"array","title":"Linked Regions","description":"Optional list of regions to filter findings by region.","items":{"type":"string"}},"organizationId":{"type":"string","title":"AWS Organization ID","description":"Optional AWS Organizations identifier (for traceability and scoping)."},"accountScope":{"type":"string","title":"Account Scope","description":"Use all accessible accounts from the delegated admin role, or limit to specific account IDs.","default":"all","enum":["all","specific"]},"accountIds":{"type":"array","title":"Account IDs","description":"Required when Account Scope is set to specific.","items":{"type":"string"}},"sessionDuration":{"type":"string","title":"Session Duration","description":"Optional session duration (Go duration string, e.g. 1h30m)."},"sessionName":{"type":"string","title":"Session Name","description":"Optional session name override for STS AssumeRole calls."},"accessKeyId":{"type":"string","title":"Access Key ID","description":"Optional source credential key when Openlane cannot use runtime IAM credentials."},"secretAccessKey":{"type":"string","title":"Secret Access Key","description":"Optional source credential secret paired with Access Key ID.","secret":true},"sessionToken":{"type":"string","title":"Session Token","description":"Optional source session token when using temporary source credentials.","secret":true},"accountId":{"type":"string","title":"Account ID","description":"Optional AWS account identifier for reference."},"tags":{"type":"object","title":"Default Tags","description":"Optional key/value map added to generated integrations for traceability.","additionalProperties":{"type":"string"}}}}`)

// Builder returns the AWS Security Hub provider builder
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeAWSSecurityHub,
		SpecFunc:     awsSecurityHubSpec,
		BuildFunc: func(ctx context.Context, s spec.ProviderSpec) (types.Provider, error) {
			return awssts.Builder(
				TypeAWSSecurityHub,
				awssts.WithOperations(securityHubOperations()),
				awssts.WithClientDescriptors(securityHubClientDescriptors()),
			).Build(ctx, s)
		},
	}
}

// awsSecurityHubSpec returns the static provider specification for the AWS Security Hub provider.
func awsSecurityHubSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "awssecurityhub",
		DisplayName: "AWS Security Hub",
		Category:    "cloud",
		AuthType:    types.AuthKindAWSFederation,
		Active:      lo.ToPtr(true),
		Visible:     lo.ToPtr(false),
		LogoURL:     "",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws/overview",
		Labels: map[string]string{
			"vendor":  "aws",
			"service": "security-hub",
		},
		CredentialsSchema: awsSecurityHubCredentialsSchema,
		Description:       "Collect AWS Security Hub findings for vulnerability ingestion using STS role assumption in a tenant AWS environment.",
	}
}

// securityHubClientDescriptors returns the client descriptors for the AWS Security Hub provider
func securityHubClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeAWSSecurityHub, ClientAWSSecurityHub, "AWS Security Hub client", pooledSecurityHubClient)
}

// pooledSecurityHubClient builds the AWS Security Hub client for pooling
func pooledSecurityHubClient(ctx context.Context, credential types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	client, _, err := buildSecurityHubClient(ctx, credential)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(client), nil
}

// newSecurityHubSDKClient wraps securityhub.NewFromConfig for use with generic helpers
func newSecurityHubSDKClient(cfg awssdk.Config) *securityhub.Client {
	return securityhub.NewFromConfig(cfg)
}

// metadataFromCredential extracts and validates AWS metadata from a credential set
func metadataFromCredential(credential types.CredentialSet) (awskit.Metadata, error) {
	if len(credential.ProviderData) == 0 {
		return awskit.Metadata{}, ErrMetadataMissing
	}

	parsed, err := awskit.MetadataFromProviderData(credential.ProviderData, securityHubDefaultSession)
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

// buildSecurityHubClient constructs a Security Hub client from the stored credential
func buildSecurityHubClient(ctx context.Context, credential types.CredentialSet) (*securityhub.Client, awskit.Metadata, error) {
	var zero *securityhub.Client

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

	return newSecurityHubSDKClient(cfg), meta, nil
}

// resolveSecurityHubClient returns a pooled client when available or builds one on demand
func resolveSecurityHubClient(ctx context.Context, input types.OperationInput) (*securityhub.Client, awskit.Metadata, error) {
	if client, ok := types.ClientInstanceAs[*securityhub.Client](input.Client); ok {
		meta, err := metadataFromCredential(input.Credential)
		if err != nil {
			return nil, awskit.Metadata{}, err
		}

		return client, meta, nil
	}

	return buildSecurityHubClient(ctx, input.Credential)
}

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

type securityHubFindingsDetails struct {
	Region          string                `json:"region"`
	AlertsTotal     int                   `json:"alerts_total"`
	AlertTypeCounts map[string]int        `json:"alert_type_counts"`
	Alerts          []types.AlertEnvelope `json:"alerts,omitempty"`
}

type securityHubFailureDetails struct {
	Region string `json:"region,omitempty"`
}

var securityHubFindingsSchema = providerkit.SchemaFrom[securityHubFindingsConfig]()

// securityHubOperations returns the operation descriptors for the AWS Security Hub provider
func securityHubOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:         types.OperationVulnerabilitiesCollect,
			Kind:         types.OperationKindCollectFindings,
			Description:  "Collect AWS Security Hub findings for vulnerability ingestion.",
			Client:       ClientAWSSecurityHub,
			Run:          runSecurityHubFindings,
			ConfigSchema: securityHubFindingsSchema,
			Ingest: []types.IngestContract{
				{
					Schema:         types.MappingSchema(integrationgenerated.IntegrationMappingSchemaVulnerability),
					EnsurePayloads: true,
				},
			},
		},
	}
}

// runSecurityHubFindings collects Security Hub findings for ingestion
func runSecurityHubFindings(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, meta, err := resolveSecurityHubClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	cfg := securityHubFindingsConfig{PageSize: securityHubDefaultPageSize}
	if err := jsonx.UnmarshalIfPresent(input.Config, &cfg); err != nil {
		return types.OperationResult{}, err
	}

	pageSize := cfg.PageSize
	if pageSize <= 0 {
		pageSize = securityHubDefaultPageSize
	}

	if pageSize > securityHubMaxPageSize {
		pageSize = securityHubMaxPageSize
	}

	maxFindings := cfg.MaxFindings
	severityFilter := cfg.Severity.String()
	recordStateFilter := cfg.RecordState.String()
	workflowFilter := cfg.WorkflowStatus.String()

	filters := securityHubFiltersFromMetadata(meta)

	fetch := func(ctx context.Context, pageToken string) (providerkit.PageResult[securityhubtypes.AwsSecurityFinding], error) {
		var nextToken *string
		if pageToken != "" {
			nextToken = &pageToken
		}

		resp, err := client.GetFindings(ctx, &securityhub.GetFindingsInput{
			MaxResults: awssdk.Int32(int32(min(pageSize, math.MaxInt32))), //nolint:gosec // bounds checked via min
			NextToken:  nextToken,
			Filters:    filters,
		})
		if err != nil {
			return providerkit.PageResult[securityhubtypes.AwsSecurityFinding]{}, err
		}

		result := providerkit.PageResult[securityhubtypes.AwsSecurityFinding]{Items: resp.Findings}
		if resp.NextToken != nil && *resp.NextToken != "" {
			result.NextToken = *resp.NextToken
		}

		return result, nil
	}

	allFindings, err := providerkit.CollectAll(ctx, fetch, 0)
	if err != nil {
		return providerkit.OperationFailure("AWS Security Hub findings fetch failed", err, securityHubFailureDetails{
			Region: meta.Region,
		})
	}

	var (
		envelopes []types.AlertEnvelope
		total     int
	)

	for _, finding := range allFindings {
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
			return providerkit.OperationFailure("AWS Security Hub finding serialization failed", err, securityHubFailureDetails{
				Region: meta.Region,
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
			AlertType: securityHubAlertTypeFinding,
			Resource:  resourceID,
			Payload:   payload,
		})
		total++
	}

	details := securityHubFindingsDetails{
		Region:      meta.Region,
		AlertsTotal: total,
		AlertTypeCounts: map[string]int{
			securityHubAlertTypeFinding: total,
		},
	}

	if cfg.IncludePayloads {
		details.Alerts = envelopes
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Collected %d Security Hub findings", total), details), nil
}

func securityHubFiltersFromMetadata(meta awskit.Metadata) *securityhubtypes.AwsSecurityFindingFilters {
	var filters securityhubtypes.AwsSecurityFindingFilters

	if meta.AccountScope == awskit.AccountScopeSpecific {
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
