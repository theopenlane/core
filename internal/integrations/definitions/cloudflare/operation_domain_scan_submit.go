package cloudflare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/url_scanner"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// maxDomainsAllowed is the max domains a single domain scan request should contain
const maxDomainsAllowed = 10

// DomainScanSubmit submits domains to Cloudflare's URL Scanner for scanning
type DomainScanSubmit struct {
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
	return providerkit.WithClientRequestConfig(cloudflareClient, DomainScanSubmitOp, ErrOperationConfigInvalid, func(ctx context.Context, _ types.OperationRequest, client *CloudflareClient, cfg DomainScanSubmit) (json.RawMessage, error) {
		result, err := s.Run(ctx, client, cfg)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(result, ErrResultEncode)
	})
}

// Run submits the domains to Cloudflare's URL Scanner, with a max of 10 domains in any scan request
// Cloudflare bulk scan allows up to 100 in a single request, so we do not need to batch these request
func (DomainScanSubmit) Run(ctx context.Context, client *CloudflareClient, cfg DomainScanSubmit) (DomainScanSubmitResult, error) {
	if len(cfg.Domains) > maxDomainsAllowed {
		logx.FromContext(ctx).Warn().Strs("domains", cfg.Domains).Msg("cloudflare: max domains surpassed, only running on first 10")

		cfg.Domains = cfg.Domains[:maxDomainsAllowed]
	}

	if client.Config.DomainScan.BulkScan {
		return runBulkScan(ctx, client, cfg.Domains)
	}

	return runScansSequentially(ctx, client, cfg.Domains)
}

func runBulkScan(ctx context.Context, client *CloudflareClient, domains []string) (DomainScanSubmitResult, error) {
	scans := make([]url_scanner.ScanBulkNewResponse, 0, len(domains))

	body := make([]url_scanner.ScanBulkNewParamsBody, len(domains))
	for idx, domain := range domains {
		body[idx] = url_scanner.ScanBulkNewParamsBody{
			URL:            cf.F(domain),
			AgentReadiness: cf.F(true),
			Visibility:     cf.F(url_scanner.ScanBulkNewParamsBodyVisibilityUnlisted),
		}
	}

	params := url_scanner.ScanBulkNewParams{
		AccountID: cf.F(client.Config.AccountID),
		Body:      body,
	}

	result, err := client.URLScanner.Scans.BulkNew(ctx, params)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Strs("domains", domains).
			Str("account_id", client.Config.AccountID).
			Msg("cloudflare: error submitting domain scan batch")

		return DomainScanSubmitResult{Scans: scans}, fmt.Errorf("%w: %w", ErrDomainScanSubmitFailed, err)
	}

	scans = append(scans, *result...)

	return DomainScanSubmitResult{Scans: scans}, nil
}

// runScansSequentially submits domains sequentially through the single URL Scanner endpoint
func runScansSequentially(ctx context.Context, client *CloudflareClient, domains []string) (DomainScanSubmitResult, error) {
	scans := make([]url_scanner.ScanBulkNewResponse, 0, len(domains))

	var errs error

	for _, domain := range domains {
		result, err := client.URLScanner.Scans.New(ctx, url_scanner.ScanNewParams{
			AccountID:      cf.F(client.Config.AccountID),
			URL:            cf.F(domain),
			AgentReadiness: cf.F(true),
			Visibility:     cf.F(url_scanner.ScanNewParamsVisibilityUnlisted),
		})
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("%s: %w", domain, err))

			logx.FromContext(ctx).Error().Err(err).Str("domain", domain).Str("account_id", client.Config.AccountID).Msg("cloudflare: error submitting individual domain scan")

			continue
		}

		scans = append(scans, url_scanner.ScanBulkNewResponse{
			API:        result.API,
			Result:     result.Result,
			URL:        result.URL,
			UUID:       result.UUID,
			Visibility: url_scanner.ScanBulkNewResponseVisibility(result.Visibility),
			Options: url_scanner.ScanBulkNewResponseOptions{
				Useragent: result.Options.Useragent,
			},
		})
	}

	if len(scans) == 0 && errs != nil {
		return DomainScanSubmitResult{Scans: scans}, fmt.Errorf("%w: %w", ErrDomainScanSubmitFailed, errs)
	}

	return DomainScanSubmitResult{Scans: scans}, nil
}
