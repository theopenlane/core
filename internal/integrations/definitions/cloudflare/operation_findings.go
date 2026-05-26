package cloudflare

import (
	"context"
	"time"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/cloudflare/cloudflare-go/v7/security_center"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// defaultPageSize is the maximum number of records requested per page
	defaultPageSize = 1000
)

// FindingsCollect collects Cloudflare Security Center insights for ingest as findings
type FindingsCollect struct{}

// IngestHandle adapts findings collection to the ingest operation registration boundary
func (f FindingsCollect) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(cloudflareClient, findingsSyncOperation, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *cf.Client, _ FindingsSync) ([]types.IngestPayloadSet, error) {
		return f.Run(ctx, request.Credentials, client, request.LastRunAt)
	})
}

// Run collects Cloudflare Security Center insights and emits finding ingest payloads
func (FindingsCollect) Run(ctx context.Context, credentials types.CredentialBindings, client *cf.Client, lastRunAt *time.Time) ([]types.IngestPayloadSet, error) {
	meta, err := resolveCredential(credentials)
	if err != nil {
		return nil, err
	}

	if meta.AccountID == "" {
		return nil, ErrAccountIDMissing
	}

	issues, err := fetchSecurityInsights(ctx, client, meta.AccountID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("cloudflare: error fetching Security Center insights")
		return nil, ErrFindingsFetchFailed
	}

	envelopes := make([]types.MappingEnvelope, 0, len(issues))
	for _, issue := range issues {
		if !insightUpdatedSince(issue, lastRunAt) {
			continue
		}

		envelope, err := providerkit.MarshalEnvelope(meta.AccountID, issue, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaFinding,
			Envelopes: envelopes,
		},
	}, nil
}

// the sdk returns the wrong json structure so use a wrapper
type cloudflareInsightsResponse struct {
	Result security_center.InsightListResponse `json:"result"`
}

func fetchSecurityInsights(ctx context.Context, client *cf.Client, accountID string) ([]security_center.InsightListResponseIssue, error) {
	issues := make([]security_center.InsightListResponseIssue, 0)

	defaultPage := int64(1)

	for page := defaultPage; ; page++ {
		var response cloudflareInsightsResponse
		if _, err := client.SecurityCenter.Insights.List(ctx, security_center.InsightListParams{
			AccountID: cf.F(accountID),
			Page:      cf.F(page),
			PerPage:   cf.F(int64(defaultPageSize)),
		}, option.WithResponseBodyInto(&response)); err != nil {
			return nil, err
		}

		currIssues := response.Result.Issues
		count := response.Result.Count
		perPage := response.Result.PerPage

		issues = append(issues, currIssues...)

		if perPage <= 0 {
			perPage = defaultPageSize
		}

		if len(currIssues) == 0 || count > 0 && page*perPage >= count {
			break
		}
	}

	return issues, nil
}

func insightUpdatedSince(issue security_center.InsightListResponseIssue, lastRunAt *time.Time) bool {
	if lastRunAt == nil || issue.Timestamp.IsZero() {
		return true
	}

	return !issue.Timestamp.Before(lastRunAt.UTC())
}
