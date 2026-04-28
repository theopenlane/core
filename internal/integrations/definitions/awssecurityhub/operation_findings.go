package awssecurityhub

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// defaultPageSize is the default number of Security Hub findings requested per paginated API call
	defaultPageSize = int32(100)
	// maxPageSize is the maximum number of findings that can be requested per API page
	maxPageSize = int32(100)
	// vul
	vulnerabilityType = "Software and Configuration Checks/Vulnerabilities/CVE"
)

// FindingSync holds per-invocation execution controls for the vulnerabilities.collect operation
type FindingSync struct {
	// MaxFindings caps the total number of findings returned; 0 means no limit
	MaxFindings int `json:"maxFindings,omitempty" jsonschema:"title=Max Findings,description=Maximum number of findings to collect. 0 means no limit."`
}

// FindingsCollect collects AWS Security Hub findings
type FindingsCollect struct{}

// IngestHandle adapts vulnerabilities collection to the ingest operation registration boundary
func (v FindingsCollect) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(securityHubClient, findingsCollectOperation, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *securityhub.Client, cfg FindingSync) ([]types.IngestPayloadSet, error) {
		return v.Run(ctx, client, request.Credentials, cfg)
	})
}

// Run collects Security Hub findings
func (FindingsCollect) Run(ctx context.Context, c *securityhub.Client, credentials types.CredentialBindings, cfg FindingSync) ([]types.IngestPayloadSet, error) {
	pageSize := defaultPageSize

	if cfg.MaxFindings > 0 && int32(cfg.MaxFindings) < maxPageSize { //nolint:gosec // bounded by maxPageSize
		pageSize = int32(cfg.MaxFindings) //nolint:gosec
	}

	var (
		findingEnvelopes       []types.MappingEnvelope
		vulnerabilityEnvelopes []types.MappingEnvelope
		total                  int
		nextToken              *string
	)

	if cfg.MaxFindings > 0 {
		findingEnvelopes = make([]types.MappingEnvelope, 0, cfg.MaxFindings)
		vulnerabilityEnvelopes = make([]types.MappingEnvelope, 0, cfg.MaxFindings)
	}

	filters, err := buildFilters(ctx, credentials)
	if err != nil {
		return nil, err
	}

	if filters.AwsAccountId != nil {
		logx.FromContext(ctx).Debug().Interface("account filters", filters.AwsAccountId).Msg("awssecurityhub: using the account filter")
	} else {
		logx.FromContext(ctx).Debug().Msg("awssecurityhub: no account filter added, pulling all results allowed by service account")
	}

collectLoop:
	for {
		input := &securityhub.GetFindingsInput{
			MaxResults: aws.Int32(pageSize),
			Filters:    filters,
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		resp, err := c.GetFindings(ctx, input)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("awssecurityhub: error fetching findings")
			return nil, ErrFindingsFetchFailed
		}

		for i, finding := range resp.Findings {
			if cfg.MaxFindings > 0 && total >= cfg.MaxFindings {
				break collectLoop
			}

			envelope, err := buildFindingEnvelope(resp.Findings[i])
			if err != nil {
				return nil, err
			}

			if slices.Contains(finding.Types, vulnerabilityType) {
				vulnerabilityEnvelopes = append(vulnerabilityEnvelopes, envelope)
			} else {
				findingEnvelopes = append(findingEnvelopes, envelope)
			}

			total++

		}

		if resp.NextToken == nil || *resp.NextToken == "" {
			break
		}

		nextToken = resp.NextToken
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaFinding,
			Envelopes: findingEnvelopes,
		},
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaVulnerability,
			Envelopes: vulnerabilityEnvelopes,
		},
	}, nil
}

func buildFilters(ctx context.Context, creds types.CredentialBindings) (*securityhubtypes.AwsSecurityFindingFilters, error) {
	filters := &securityhubtypes.AwsSecurityFindingFilters{}
	meta, err := resolveAssumeRoleCredential(creds)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("awssecurityhub: error resolving credentials for filter")

		return nil, err
	}

	if meta.AccountScope != AccountScopeAll {
		if len(meta.AccountIDs) > 0 {
			accountFilters := make([]securityhubtypes.StringFilter, len(meta.AccountIDs))
			for i, id := range meta.AccountIDs {
				accountFilters[i] = securityhubtypes.StringFilter{
					Value:      aws.String(id),
					Comparison: securityhubtypes.StringFilterComparisonEquals,
				}
			}

			filters.AwsAccountId = accountFilters
		} else if meta.AccountID != "" {
			accountFilters := make([]securityhubtypes.StringFilter, 1)

			accountFilters[0] = securityhubtypes.StringFilter{
				Value:      aws.String(meta.AccountID),
				Comparison: securityhubtypes.StringFilterComparisonEquals,
			}

			filters.AwsAccountId = accountFilters
		}
	}

	if len(meta.LinkedRegions) > 0 {
		regionFilters := make([]securityhubtypes.StringFilter, len(meta.LinkedRegions))
		for i, reg := range meta.LinkedRegions {
			regionFilters[i] = securityhubtypes.StringFilter{
				Value:      aws.String(reg),
				Comparison: securityhubtypes.StringFilterComparisonEquals,
			}
		}

		filters.Region = regionFilters
	}

	return filters, nil
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
		return aws.ToString(finding.AwsAccountId)
	}

	return ""
}
