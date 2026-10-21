package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/precis/docextract"
	"github.com/theopenlane/precis/docextract/gemini/modelarmor"
	"github.com/theopenlane/precis/docextract/kinds/soc2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

var supportedDocumentKinds = []string{soc2.KindName}

var (
	// ReportScanRequestOp is the operation ref for kicking off the parse of a pending report scan; its handler is bound by the builder from the runtime config
	ReportScanRequestOp = types.OperationPayloadOf[ReportScanRequest]().Policy(types.ExecutionPolicy{SkipRunRecord: true}).CustomerSelectable(false).Internal() //nolint:revive
	// VendorReportScanRequestOp is the operation ref for kicking off the review of a vendor's report,
	// a separate operation so it carries its own rate limit
	VendorReportScanRequestOp = types.OperationPayloadOf[VendorReportScanRequest]().Policy(types.ExecutionPolicy{SkipRunRecord: true}).CustomerSelectable(false).Internal() //nolint:revive
)

// handleVendor runs the same parse under the vendor review operation, so the vendor rate limit applies
func (r ReportScanRequest) handleVendor(ctx context.Context, request types.OperationRequest, client *Client, cfg VendorReportScanRequest) (json.RawMessage, error) {
	return r.runEncoded(ctx, request, client, ReportScanRequest{ScanID: cfg.ScanID, OrganizationID: cfg.OrganizationID})
}

// runEncoded runs the request and encodes its result for the operation response
func (r ReportScanRequest) runEncoded(ctx context.Context, request types.OperationRequest, client *Client, cfg ReportScanRequest) (json.RawMessage, error) {
	result, err := r.Run(ctx, request, client, cfg)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(result, ErrResultEncode)
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
		scan.PerformedByIn(ReportPerformers...),
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
		plan, err = r.planSOC2Report(ctx, client.SOC2, pdf, scanRecord.Metadata, performerProfiles[scanRecord.PerformedBy])
	default:
		err = ErrUnsupportedDocumentKind
	}

	// a rejected upload fails the scan with its reason rather than leaving it pending
	if err != nil {
		return failScan(systemCtx, request.DB, scanRecord.ID, err, scanRecord.Metadata, &plan.validation)
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

	// the document is untrusted input, so caching screens it for injected instructions before the model ever sees it
	cacheName, err := client.CacheDocument(ctx, bytes.NewReader(pdf), documentCacheTTL)

	switch {
	case errors.Is(err, docextract.ErrDocumentBlocked):
		return failScan(systemCtx, request.DB, scanRecord.ID, err, metadata, &plan.validation)
	case errors.Is(err, modelarmor.ErrSanitizeFailed):
		return ReportScanRequestResult{}, err
	case err != nil:
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

	event := logx.FromContext(ctx).Warn().Err(err)
	if validation != nil {
		event = event.Int("pages", validation.Pages).Int("text_length", validation.TextLength).Strs("strong", validation.Strong).Strs("supporting", validation.Supporting)
	}

	event.Msg("report scan: failed")

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
func (r ReportScanRequest) planSOC2Report(ctx context.Context, kind soc2.Kind, pdf []byte, metadata map[string]any, profile soc2.Profile) (reportScanPlan, error) {
	validation, err := kind.Validate(ctx, pdf)
	if err != nil {
		return reportScanPlan{validation: validation}, err
	}

	parts, err := r.selectParts(metadata, profile)
	if err != nil {
		return reportScanPlan{validation: validation}, err
	}

	return reportScanPlan{validation: validation, parts: parts}, nil
}

// scanSystemContext builds a context authorized to update Scan and Notification records for
// organizationID on behalf of the system
func scanSystemContext(ctx context.Context, organizationID string) context.Context {
	return auth.WithCaller(ctx, &auth.Caller{
		OrganizationID: organizationID,
		Capabilities:   auth.CapInternalOperation,
	})
}
