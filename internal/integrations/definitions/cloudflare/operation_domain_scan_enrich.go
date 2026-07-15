package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/domainscan"
	"github.com/theopenlane/core/pkg/logx"
)

// domainScanEnrichmentTimeout bounds how long to spend gathering company profile,
// compliance, and DNS vendor data before finalizing the scan without it
const domainScanEnrichmentTimeout = 5 * time.Minute

// DomainScanEnrich enriches a completed URL Scanner result with company profile, compliance,
// and DNS vendor data and builds the structured onboarding domain scan report
type DomainScanEnrich struct {
	// Domain is the scanned domain to enrich
	Domain string `json:"domain"`
	// InternalScanID is the openlane Scan record id the built report belongs to
	InternalScanID string `json:"internalScanId"`
	// Result is the completed URL Scanner task result to build the report from
	Result json.RawMessage `json:"result"`
	// ForceRefresh bypasses Cloudflare's Browser Rendering cache, forcing a fresh render
	// instead of reusing one from a previous scan of the same domain
	ForceRefresh bool `json:"forceRefresh,omitempty"`
}

// DomainScanEnrichResult carries the structured report built from the scan result and enrichment
type DomainScanEnrichResult struct {
	// Data is the structured scan report, ready to persist on the Scan record
	Data map[string]any `json:"data"`
}

// Handle adapts domain scan enrichment to the generic operation registration boundary
func (e DomainScanEnrich) Handle() types.OperationHandler {
	return providerkit.WithClientConfig(cloudflareClient, DomainScanEnrichOp, ErrOperationConfigInvalid, func(ctx context.Context, client *CloudflareClient, cfg DomainScanEnrich) (json.RawMessage, error) {
		result, err := e.Run(ctx, client, cfg)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(result, ErrResultEncode)
	})
}

// Run gathers enrichment for the domain and builds the structured scan report from it plus the
// submitted URL Scanner result
func (DomainScanEnrich) Run(ctx context.Context, client *CloudflareClient, cfg DomainScanEnrich) (DomainScanEnrichResult, error) {
	var scanResult url_scanner.ScanGetResponse
	if err := json.Unmarshal(cfg.Result, &scanResult); err != nil {
		return DomainScanEnrichResult{}, fmt.Errorf("%w: %w", ErrOperationConfigInvalid, err)
	}

	cacheTTL := client.Config.DomainScan.ScanTTL
	if cfg.ForceRefresh {
		cacheTTL = 0
	}

	enrichmentCfg := domainscan.Config{
		APIToken:  client.Config.APIToken,
		AccountID: client.Config.AccountID,
		CacheTTL:  cacheTTL,
	}

	enrichment, enrichmentErrs := enrichmentCfg.GatherEnrichment(ctx, cfg.Domain, domainScanEnrichmentTimeout)
	logDomainScanEnrichmentErrors(ctx, cfg.Domain, enrichmentErrs)

	data := domainscan.BuildScanReport(&scanResult, enrichment, client.Config.DomainScan.NonVendorCategories, client.Config.DomainScan.DeniedVendorNames)
	data["internal_scan_id"] = cfg.InternalScanID

	return DomainScanEnrichResult{Data: data}, nil
}

// logDomainScanEnrichmentErrors logs any per-lookup enrichment failures through the structured
// logger; each is best-effort (the report is built without that section) so these are warnings, not errors
func logDomainScanEnrichmentErrors(ctx context.Context, domain string, errs domainscan.EnrichmentErrors) {
	logger := logx.FromContext(ctx).With().Str("domain", domain).Logger()

	if errs.Company != nil {
		logger.Warn().Err(errs.Company).Msg("domain scan: failed to get company profile")
	}

	if errs.Compliance != nil {
		logger.Warn().Err(errs.Compliance).Msg("domain scan: failed to get compliance data")
	}

	if errs.DNS != nil {
		logger.Warn().Err(errs.DNS).Msg("domain scan: failed to get dns vendor info")
	}
}
