package cloudflare

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/domainscan"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// domainScanEnrichmentTimeout bounds how long to spend gathering company profile,
// branding, compliance, and DNS vendor data before finalizing the scan without it
const domainScanEnrichmentTimeout = 5 * time.Minute

// DomainScanGatherEnrichment gathers company profile, branding, compliance, and DNS vendor data for a domain
type DomainScanGatherEnrichment struct {
	// Domain is the domain to gather enrichment for
	Domain string `json:"domain"`
	// ForceRefresh bypasses Cloudflare's Browser Rendering cache, forcing a fresh render
	ForceRefresh bool `json:"forceRefresh,omitempty"`
	// BrandDesignOnly instructs the scanning process to only extract the brand design not the entire
	// domain scan
	BrandDesignOnly bool `json:"brandDesignOnly,omitempty"`
}

// DomainScanGatherEnrichmentResult carries the gathered enrichment data
type DomainScanGatherEnrichmentResult struct {
	// Enrichment is the gathered company profile, branding, compliance, and DNS vendor data
	Enrichment domainscan.Enrichment `json:"enrichment"`
}

// Handle adapts domain scan enrichment gathering to the generic operation registration boundary
func (e DomainScanGatherEnrichment) Handle() types.OperationHandler {
	return providerkit.WithClientConfig(cloudflareClient, DomainScanEnrichmentOp, ErrOperationConfigInvalid, func(ctx context.Context, client *CloudflareClient, cfg DomainScanGatherEnrichment) (json.RawMessage, error) {
		result, err := e.Run(ctx, client, cfg)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(result, ErrResultEncode)
	})
}

// Run gathers company profile, branding, compliance, and DNS vendor data for the domain
func (DomainScanGatherEnrichment) Run(ctx context.Context, client *CloudflareClient, cfg DomainScanGatherEnrichment) (DomainScanGatherEnrichmentResult, error) {
	ctx = logx.WithFields(ctx, logx.LogFields{
		"domain": cfg.Domain,
	})

	cacheTTL := client.Config.DomainScan.ScanTTL
	if cfg.ForceRefresh {
		cacheTTL = 0
	}

	enrichmentCfg := domainscan.Config{
		APIToken:  client.Config.APIToken,
		AccountID: client.Config.AccountID,
		CacheTTL:  cacheTTL,
	}

	if cfg.BrandDesignOnly {
		brandDesignCtx, cancel := context.WithTimeout(ctx, domainScanEnrichmentTimeout)
		defer cancel()

		branding, err := enrichmentCfg.GetBrandingData(brandDesignCtx, cfg.Domain)
		if err != nil {
			logx.FromContext(ctx).Warn().Err(err).Msg("domain scan: failed to get brand design data")

			return DomainScanGatherEnrichmentResult{}, nil
		}

		if branding.IsEmpty() {
			return DomainScanGatherEnrichmentResult{}, nil
		}

		return DomainScanGatherEnrichmentResult{
			Enrichment: domainscan.Enrichment{Branding: branding},
		}, nil
	}

	enrichment, enrichmentErrs := enrichmentCfg.GatherEnrichment(ctx, cfg.Domain, domainScanEnrichmentTimeout)
	logDomainScanEnrichmentErrors(ctx, enrichmentErrs)

	return DomainScanGatherEnrichmentResult{Enrichment: enrichment}, nil
}

// logDomainScanEnrichmentErrors logs any per-lookup enrichment failures through the structured
// logger; each is best-effort (the report is built without that section) so these are warnings, not errors
func logDomainScanEnrichmentErrors(ctx context.Context, errs domainscan.EnrichmentErrors) {
	if errs.Company != nil {
		logx.FromContext(ctx).Warn().Err(errs.Company).Msg("domain scan: failed to get company profile")
	}

	if errs.Branding != nil {
		logx.FromContext(ctx).Warn().Err(errs.Branding).Msg("domain scan: failed to get branding data")
	}

	if errs.Compliance != nil {
		logx.FromContext(ctx).Warn().Err(errs.Compliance).Msg("domain scan: failed to get compliance data")
	}

	if errs.DNS != nil {
		logx.FromContext(ctx).Warn().Err(errs.DNS).Msg("domain scan: failed to get dns vendor info")
	}
}
