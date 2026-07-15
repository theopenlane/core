package cloudflare

import (
	"context"
	"encoding/json"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/url_scanner"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// domainScanMaxBatchSize is the largest batch Cloudflare's URL Scanner bulk
// endpoint accepts in a single request; larger domain lists are split into
// batches of this size
const domainScanMaxBatchSize = 100

// DomainScanSubmit submits domains to Cloudflare's URL Scanner for scanning
type DomainScanSubmit struct {
	// AccountID is the Cloudflare account the scans are submitted under
	AccountID string `json:"accountId"`
	// Domains is the list of domains to submit for scanning
	Domains []string `json:"domains"`
}

// DomainScanSubmitResult is the set of scan tasks created by a submission
type DomainScanSubmitResult struct {
	// Scans is the list of scan tasks created by Cloudflare's URL Scanner
	Scans []url_scanner.ScanBulkNewResponse `json:"scans"`
}

// Handle adapts domain scan submission to the generic operation registration boundary
func (s DomainScanSubmit) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(cloudflareClient, DomainScanSubmitOp, ErrOperationConfigInvalid, func(ctx context.Context, _ types.OperationRequest, client *cf.Client, cfg DomainScanSubmit) (json.RawMessage, error) {
		result, err := s.Run(ctx, client, cfg)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(result, ErrResultEncode)
	})
}

// Run submits the domains to Cloudflare's URL Scanner, batching requests as needed
func (DomainScanSubmit) Run(ctx context.Context, client *cf.Client, cfg DomainScanSubmit) (DomainScanSubmitResult, error) {
	scans := make([]url_scanner.ScanBulkNewResponse, 0, len(cfg.Domains))

	for i := 0; i < len(cfg.Domains); i += domainScanMaxBatchSize {
		end := i + domainScanMaxBatchSize
		if end > len(cfg.Domains) {
			end = len(cfg.Domains)
		}

		batch := cfg.Domains[i:end]

		body := make([]url_scanner.ScanBulkNewParamsBody, len(batch))
		for j, d := range batch {
			body[j] = url_scanner.ScanBulkNewParamsBody{
				URL:            cf.F(d),
				AgentReadiness: cf.F(true),
				Visibility:     cf.F(url_scanner.ScanBulkNewParamsBodyVisibilityUnlisted),
			}
		}

		params := url_scanner.ScanBulkNewParams{
			AccountID: cf.F(cfg.AccountID),
			Body:      body,
		}

		result, err := client.URLScanner.Scans.BulkNew(ctx, params)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("cloudflare: error submitting domain scan batch")

			return DomainScanSubmitResult{Scans: scans}, ErrDomainScanSubmitFailed
		}

		scans = append(scans, *result...)
	}

	return DomainScanSubmitResult{Scans: scans}, nil
}
