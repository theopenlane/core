package awssecurityhub

import (
	"context"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// defaultPageSize is the default number of Security Hub findings requested per paginated API call
	defaultPageSize = int32(100)
	// maxPageSize is the maximum number of findings that can be requested per API page
	maxPageSize = int32(100)
)

// FindingsConfig holds per-invocation execution controls for the vulnerabilities.collect operation
type FindingsConfig struct {
	// MaxFindings caps the total number of findings returned; 0 means no limit
	MaxFindings int `json:"maxFindings,omitempty" jsonschema:"title=Max Findings,description=Maximum number of findings to collect. 0 means no limit."`
}

// VulnerabilitiesCollect collects AWS Security Hub findings
type VulnerabilitiesCollect struct{}

// IngestHandle adapts vulnerabilities collection to the ingest operation registration boundary
func (v VulnerabilitiesCollect) IngestHandle() types.IngestHandler {
	return providerkit.IngestWithClientRequestConfig(
		SecurityHubClient,
		VulnerabilitiesCollectOperation,
		ErrOperationConfigInvalid,
		func(ctx context.Context, request types.OperationRequest, client *securityhub.Client, cfg FindingsConfig) ([]types.IngestPayloadSet, error) {
			return v.Run(ctx, request.Credentials, client, cfg)
		},
	)
}

// Run collects Security Hub findings using credential-defined collection scope
func (VulnerabilitiesCollect) Run(ctx context.Context, credentials types.CredentialBindings, c *securityhub.Client, cfg FindingsConfig) ([]types.IngestPayloadSet, error) {
	awsCredential, err := resolveAssumeRoleCredential(credentials)
	if err != nil {
		return nil, err
	}

	pageSize := defaultPageSize
	if cfg.MaxFindings > 0 && int32(cfg.MaxFindings) < maxPageSize { //nolint:gosec // bounded by maxPageSize
		pageSize = int32(cfg.MaxFindings) //nolint:gosec
	}

	filters := buildSecurityHubFilters(awsCredential)

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
			return nil, ErrFindingsFetchFailed
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

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: envelopes,
		},
	}, nil
}

// buildSecurityHubFilters constructs a server-side filter from credential metadata
// so account and region scope stay tied to integration setup.
func buildSecurityHubFilters(credential AssumeRoleCredentialSchema) *securityhubtypes.AwsSecurityFindingFilters {
	var filters securityhubtypes.AwsSecurityFindingFilters

	if credential.AccountScope == AccountScopeSpecific {
		filters.AwsAccountId = toStringFilters(credential.AccountIDs)
	}

	filters.Region = toStringFilters(credential.LinkedRegions)

	if len(filters.AwsAccountId) == 0 && len(filters.Region) == 0 {
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
	return providerkit.MarshalEnvelope(resolveFindingResource(finding), finding, ErrFindingEncode)
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
