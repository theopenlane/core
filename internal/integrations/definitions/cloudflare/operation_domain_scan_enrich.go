package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/domainscan"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// DomainScanBuildReport builds the structured domain scan report from a result and enrichment
type DomainScanBuildReport struct {
	// InternalScanID is the openlane Scan record id the built report belongs to
	InternalScanID string `json:"internalScanId"`
	// Result is the completed URL Scanner task result to build the report from, may be empty
	Result json.RawMessage `json:"result,omitempty"`
	// Enrichment is the data gathered via DomainScanGatherEnrichment
	Enrichment domainscan.Enrichment `json:"enrichment"`
}

// DomainScanBuildReportResult carries the structured report built from the result and enrichment
type DomainScanBuildReportResult struct {
	// Data is the structured scan report, ready to persist on the Scan record
	Data map[string]any `json:"data"`
}

// DomainScanBuildReportOp is the operation ref for building the scan report
//
//nolint:revive
var DomainScanBuildReportOp = types.OperationPayloadOf[DomainScanBuildReport]().
	Handles(runDomainScanBuildReport).
	Policy(types.ExecutionPolicy{SkipRunRecord: true}).
	CustomerSelectable(false).
	Internal()

// runDomainScanBuildReport builds the structured scan report and encodes it
func runDomainScanBuildReport(ctx context.Context, _ types.OperationRequest, client *CloudflareClient, cfg DomainScanBuildReport) (json.RawMessage, error) {
	result, err := DomainScanBuildReport{}.Run(ctx, client, cfg)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(result, ErrResultEncode)
}

// Run builds the structured scan report from the URL Scanner result plus gathered enrichment
func (DomainScanBuildReport) Run(ctx context.Context, client *CloudflareClient, cfg DomainScanBuildReport) (DomainScanBuildReportResult, error) {
	var scanResult *url_scanner.ScanGetResponse

	if len(cfg.Result) > 0 {
		scanResult = &url_scanner.ScanGetResponse{}
		if err := json.Unmarshal(cfg.Result, scanResult); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("domainscan: invalid config")
			return DomainScanBuildReportResult{}, fmt.Errorf("%w: %w", types.ErrOperationConfigInvalid, err)
		}
	}

	data := domainscan.BuildScanReport(scanResult, cfg.Enrichment, client.Config.DomainScan.NonVendorCategories, client.Config.DomainScan.DeniedVendorNames)
	data["internal_scan_id"] = cfg.InternalScanID

	return DomainScanBuildReportResult{Data: data}, nil
}
