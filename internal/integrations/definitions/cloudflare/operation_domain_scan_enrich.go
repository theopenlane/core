package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/domainscan"
	"github.com/theopenlane/core/pkg/logx"
)

// DomainScanBuildReport combines a completed URL Scanner result with previously-gathered
// enrichment data (company profile, compliance, and DNS vendor data) and builds the
// structured onboarding domain scan report
type DomainScanBuildReport struct {
	// InternalScanID is the openlane Scan record id the built report belongs to
	InternalScanID string `json:"internalScanId"`
	// Result is the completed URL Scanner task result to build the report from, allowed to be empty
	// so if the url scanner failed the enrichment data can still be used for a partial report
	Result json.RawMessage `json:"result,omitempty"`
	// Enrichment is the company profile, compliance, and DNS vendor data gathered via
	// DomainScanGatherEnrichment
	Enrichment domainscan.Enrichment `json:"enrichment"`
}

// DomainScanBuildReportResult carries the structured report built from the scan result and enrichment
type DomainScanBuildReportResult struct {
	// Data is the structured scan report, ready to persist on the Scan record
	Data map[string]any `json:"data"`
}

// Handle adapts domain scan report building to the generic operation registration boundary
func (b DomainScanBuildReport) Handle() types.OperationHandler {
	return providerkit.WithClientConfig(cloudflareClient, DomainScanBuildReportOp, ErrOperationConfigInvalid, func(ctx context.Context, client *CloudflareClient, cfg DomainScanBuildReport) (json.RawMessage, error) {
		result, err := b.Run(ctx, client, cfg)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(result, ErrResultEncode)
	})
}

// Run builds the structured scan report from the submitted URL Scanner result plus the
// already-gathered enrichment data
func (DomainScanBuildReport) Run(ctx context.Context, client *CloudflareClient, cfg DomainScanBuildReport) (DomainScanBuildReportResult, error) {
	var scanResult *url_scanner.ScanGetResponse

	if len(cfg.Result) > 0 {
		scanResult = &url_scanner.ScanGetResponse{}
		if err := json.Unmarshal(cfg.Result, scanResult); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("domainscan: invalid config")
			return DomainScanBuildReportResult{}, fmt.Errorf("%w: %w", ErrOperationConfigInvalid, err)
		}
	}

	data := domainscan.BuildScanReport(scanResult, cfg.Enrichment, client.Config.DomainScan.NonVendorCategories, client.Config.DomainScan.DeniedVendorNames)
	data["internal_scan_id"] = cfg.InternalScanID

	return DomainScanBuildReportResult{Data: data}, nil
}
