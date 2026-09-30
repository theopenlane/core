package cloudflare

import (
	"context"
	"time"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/cloudflare/cloudflare-go/v7/security_center"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// defaultPageSize is the maximum number of records requested per page
	defaultPageSize = 1000
)

// findingsSyncOperation is the operation ref for the Security Center insights finding sync operation
var findingsSyncOperation = types.OperationRefOf[FindingsSync]().Ingests(cloudflareClient, runFindingsCollect)

// runFindingsCollect collects Cloudflare Security Center insights and emits finding ingest payloads
func runFindingsCollect(ctx context.Context, request types.OperationRequest, client *CloudflareClient, _ FindingsSync) ([]types.IngestPayloadSet, error) {
	meta, err := resolveCredential(request.Credentials)
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
		if !insightUpdatedSince(issue, request.LastRunAt) {
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
			Schema:    entityops.SchemaFinding.Name,
			Envelopes: envelopes,
		},
	}, nil
}

// cloudflareInsightsResponse wraps the insight list response, correcting the SDK's JSON structure
type cloudflareInsightsResponse struct {
	Result security_center.InsightListResponse `json:"result"`
}

func fetchSecurityInsights(ctx context.Context, client *CloudflareClient, accountID string) ([]security_center.InsightListResponseIssue, error) {
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
