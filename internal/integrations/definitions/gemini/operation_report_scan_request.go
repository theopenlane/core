package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/docextract/soc2"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/modelarmor"
)

var supportedDocumentKinds = []string{soc2.KindName}

// Handle adapts ReportScanRequest to the generic operation registration boundary
func (r ReportScanRequest) Handle() types.OperationHandler {
	return providerkit.WithClientRequestConfig(geminiClient, ReportScanRequestOp, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *Client, cfg ReportScanRequest) (json.RawMessage, error) {
		result, err := r.Run(ctx, request, client, cfg)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(result, ErrResultEncode)
	})
}

// Run marks the pending scan processing, seeds the per-section tracking metadata, and emits one
// durable part job per configured section; the last part job to finish assembles the report
func (r ReportScanRequest) Run(ctx context.Context, request types.OperationRequest, client *Client, cfg ReportScanRequest) (ReportScanRequestResult, error) {
	if cfg.OrganizationID == "" {
		return ReportScanRequestResult{}, ErrOrganizationRequired
	}

	if len(r.parts) == 0 {
		return ReportScanRequestResult{}, ErrNoPartsConfigured
	}

	systemCtx := scanSystemContext(ctx, cfg.OrganizationID)

	scanRecord, err := request.DB.Scan.Query().Where(
		scan.ID(cfg.ScanID),
		scan.OwnerID(cfg.OrganizationID),
		scan.ScanTypeEQ(enums.ScanTypeReport),
		scan.PerformedBy(PerformedBy),
	).Only(systemCtx)
	if err != nil {
		return ReportScanRequestResult{}, err
	}

	ctx = logx.WithFields(ctx, logx.LogFields{"scan_id": cfg.ScanID, "kind": scanRecord.DocumentKindName})

	if scanRecord.Status != enums.ScanStatusPending {
		return ReportScanRequestResult{Message: "report scan already processed", ScanID: scanRecord.ID}, nil
	}

	if !slices.Contains(supportedDocumentKinds, scanRecord.DocumentKindName) {
		err = fmt.Errorf("%w: %q", ErrUnsupportedDocumentKind, scanRecord.DocumentKindName)
		return failScan(systemCtx, request.DB, scanRecord.ID, err, scanRecord.Metadata, nil)
	}

	// validate the upload before any parsing is scheduled so an obviously wrong file fails fast
	pdf, reportFileID, err := downloadReport(systemCtx, request.DB, scanRecord)
	if err != nil {
		return ReportScanRequestResult{}, err
	}

	var (
		plan reportScanPlan
	)

	switch scanRecord.DocumentKindName {
	case soc2.KindName:
		plan, err = r.planSOC2Report(pdf, scanRecord.Metadata)
	default:
		err = ErrUnsupportedDocumentKind
	}

	if err != nil {
		return failScan(systemCtx, request.DB, scanRecord.ID, err, scanRecord.Metadata, &plan.validation)
	}

	// the document is untrusted input; screen it for injected instructions before the model ever sees it
	if client.Screener != nil {
		verdict, err := client.Screener.ScreenDocument(ctx, pdf)
		if err != nil {
			return ReportScanRequestResult{}, err
		}

		logx.FromContext(ctx).Debug().Bool("blocked", verdict.Blocked).Strs("filters", verdict.Filters).Strs("reported", verdict.Reported).Str("template", client.Screener.Template()).Msg("report scan: upload screened")

		if verdict.Blocked {
			return failScan(systemCtx, request.DB, scanRecord.ID, modelarmor.ErrDocumentBlocked, scanRecord.Metadata, &plan.validation)
		}
	}

	summary := make(map[string]any, len(plan.parts))

	for _, name := range plan.parts {
		summary[name] = map[string]any{"status": PartStatePending, "count": 0}
	}

	metadata := scanRecord.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadata[reportMetadataKey] = map[string]any{}
	metadata[summaryMetadataKey] = summary
	metadata[validationMetadataKey] = plan.validation
	metadata[reportFileMetadataKey] = reportFileID

	delete(metadata, ErrorMetadataKey)

	if err := request.DB.Scan.UpdateOneID(scanRecord.ID).SetStatus(enums.ScanStatusProcessing).SetDocumentKindName(plan.validation.Kind).SetMetadata(metadata).Exec(systemCtx); err != nil {
		return ReportScanRequestResult{}, err
	}

	cacheName, err := client.CacheDocument(ctx, bytes.NewReader(pdf), documentCacheTTL)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("report scan: document cache unavailable, parts will send the document")
	}

	saga := reportScanSaga{services: request.Services}
	if err := saga.scheduleParts(ctx, cfg.OrganizationID, scanRecord.ID, reportFileID, plan.parts, cacheName); err != nil {
		return ReportScanRequestResult{}, err
	}

	return ReportScanRequestResult{Message: "report scan submitted", ScanID: scanRecord.ID}, nil
}

func failScan(ctx context.Context, db *generated.Client, id string, err error, metadata map[string]any, validation *docextract.Validation) (ReportScanRequestResult, error) {
	reason := docextract.ValidationReason(err)

	logx.FromContext(ctx).Warn().Err(err).Msg("report scan: failed")

	if metadata == nil {
		metadata = map[string]any{}
	}

	metadata[ErrorMetadataKey] = reason

	if validation != nil {
		metadata[validationMetadataKey] = *validation
	}

	updateErr := db.Scan.UpdateOneID(id).SetStatus(enums.ScanStatusFailed).SetMetadata(metadata).Exec(ctx)
	if updateErr != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("report scan: failed to mark scan as failed")
	}

	return ReportScanRequestResult{Message: reason, ScanID: id}, nil
}

// reportScanPlan is what a document kind contributes to a report scan submit: the validation that
// accepted the upload and the sections to parse from it
type reportScanPlan struct {
	validation docextract.Validation
	parts      []string
}

// planSOC2Report validates the upload against the SOC 2 profile and narrows the configured sections
// to the ones the scan asked for; the validation is returned on failure so the reason can be stored
func (r ReportScanRequest) planSOC2Report(pdf []byte, metadata map[string]any) (reportScanPlan, error) {
	validation, err := docextract.Validate(pdf, soc2.KindName)
	if err != nil {
		return reportScanPlan{validation: validation}, err
	}

	parts, err := r.selectParts(metadata)
	if err != nil {
		return reportScanPlan{validation: validation}, err
	}

	return reportScanPlan{validation: validation, parts: parts}, nil
}

// selectParts narrows the configured sections to the ones the scan asked for; no request means
// every configured section, and a request naming only unconfigured sections is an error
func (r ReportScanRequest) selectParts(metadata map[string]any) ([]string, error) {
	requested, err := RequestedParts(metadata)
	if err != nil {
		return nil, err
	}

	if len(requested) == 0 {
		return r.parts, nil
	}

	selected := lo.Filter(r.parts, func(name string, _ int) bool { return slices.Contains(requested, name) })
	if len(selected) == 0 {
		return nil, fmt.Errorf("%w: requested %s, configured %s", ErrRequestedPartsUnavailable, strings.Join(requested, ", "), strings.Join(r.parts, ", "))
	}

	return selected, nil
}

// scanSystemContext builds a context authorized to update Scan and Notification records for
// organizationID on behalf of the system
func scanSystemContext(ctx context.Context, organizationID string) context.Context {
	return auth.WithCaller(ctx, &auth.Caller{
		OrganizationID: organizationID,
		Capabilities:   auth.CapInternalOperation,
	})
}
