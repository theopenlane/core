package gemini

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// Handle adapts ReportScanReleaseRequest to the generic operation registration boundary
func (r ReportScanReleaseRequest) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(geminiClient, ReportScanReleaseOp, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *Client, cfg ReportScanReleaseRequest) (json.RawMessage, error) {
		result, err := r.Run(ctx, request, client, cfg)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(result, ErrResultEncode)
	})
}

// Run deletes the cached report recorded on the scan and clears its metadata entry
func (ReportScanReleaseRequest) Run(ctx context.Context, request types.OperationRequest, client *Client, cfg ReportScanReleaseRequest) (ReportScanReleaseResult, error) {
	ctx = logx.WithFields(ctx, logx.LogFields{"organization_id": cfg.OrganizationID, "scan_id": cfg.ScanID})
	systemCtx := scanSystemContext(ctx, cfg.OrganizationID)

	if cfg.Cache == "" {
		return ReportScanReleaseResult{}, nil
	}

	// the scan is looked up only to confirm the caller owns it before touching the cache
	if _, err := request.DB.Scan.Query().Where(
		scan.ID(cfg.ScanID),
		scan.OwnerID(cfg.OrganizationID),
		scan.ScanTypeEQ(enums.ScanTypeReport),
	).Only(systemCtx); err != nil {
		return ReportScanReleaseResult{}, err
	}

	if err := client.ReleaseDocumentCache(ctx, cfg.Cache); err != nil {
		return ReportScanReleaseResult{}, err
	}

	return ReportScanReleaseResult{Released: true}, nil
}
