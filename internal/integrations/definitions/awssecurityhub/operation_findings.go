package awssecurityhub

import (
	"context"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	securityhubtypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// defaultPageSize is the number of Security Hub findings requested per paginated API call
	defaultPageSize = int32(100)
	// vul
	vulnerabilityType = "Software and Configuration Checks/Vulnerabilities/CVE"
)

// FindingSync holds per-invocation execution controls for the vulnerabilities.collect operation
type FindingSync struct{}

// FindingsCollect collects AWS Security Hub findings
type FindingsCollect struct{}

// IngestHandle adapts vulnerabilities collection to the ingest operation registration boundary
func (v FindingsCollect) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(securityHubClient, findingsCollectOperation, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *securityhub.Client, cfg FindingSync) ([]types.IngestPayloadSet, error) {
		return v.Run(ctx, client, request.Credentials, cfg, request.LastRunAt)
	})
}

// Run collects Security Hub findings
func (FindingsCollect) Run(ctx context.Context, c *securityhub.Client, credentials types.CredentialBindings, cfg FindingSync, lastRunAt *time.Time) ([]types.IngestPayloadSet, error) {
	var (
		findingEnvelopes       []types.MappingEnvelope
		vulnerabilityEnvelopes []types.MappingEnvelope
		nextToken              *string
	)

	filters, err := buildFilters(ctx, credentials, lastRunAt)
	if err != nil {
		return nil, err
	}

	if filters.AwsAccountId != nil {
		logx.FromContext(ctx).Debug().Interface("account filters", filters.AwsAccountId).Msg("awssecurityhub: using the account filter")
	} else {
		logx.FromContext(ctx).Debug().Msg("awssecurityhub: no account filter added, pulling all results allowed by service account")
	}

	for {
		input := &securityhub.GetFindingsInput{
			MaxResults: aws.Int32(defaultPageSize),
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
			envelope, err := buildFindingEnvelope(resp.Findings[i])
			if err != nil {
				return nil, err
			}

			if slices.Contains(finding.Types, vulnerabilityType) {
				vulnerabilityEnvelopes = append(vulnerabilityEnvelopes, envelope)
			} else {
				findingEnvelopes = append(findingEnvelopes, envelope)
			}
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

func buildFilters(ctx context.Context, creds types.CredentialBindings, lastRunAt *time.Time) (*securityhubtypes.AwsSecurityFindingFilters, error) {
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

	if lastRunAt != nil {
		filters.UpdatedAt = []securityhubtypes.DateFilter{
			{
				Start: aws.String(lastRunAt.UTC().Format(time.RFC3339)),
				End:   aws.String(time.Now().UTC().Format(time.RFC3339)),
			},
		}
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
