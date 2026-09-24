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
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/docextract/soc2"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/modelarmor"
)

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

	ctx = logx.WithFields(ctx, logx.LogFields{"organization_id": cfg.OrganizationID, "scan_id": cfg.ScanID})
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

	if scanRecord.Status != enums.ScanStatusPending {
		return ReportScanRequestResult{Message: "report scan already processed", ScanID: scanRecord.ID}, nil
	}

	// validate the upload before any parsing is scheduled so an obviously wrong file fails fast
	pdf, err := downloadReport(systemCtx, request.DB, scanRecord)
	if err != nil {
		return ReportScanRequestResult{}, err
	}

	validation, err := docextract.Validate(pdf, soc2.KindName)
	if err != nil {
		reason := docextract.ValidationReason(err)

		logx.FromContext(ctx).Warn().Err(err).Int("pages", validation.Pages).Strs("strong", validation.Strong).Msg("report scan: rejected upload")

		if err := failScan(systemCtx, request.DB, scanRecord, reason, validation); err != nil {
			return ReportScanRequestResult{}, err
		}

		return ReportScanRequestResult{Message: reason, ScanID: scanRecord.ID}, nil
	}

	logx.FromContext(ctx).Debug().Int("pages", validation.Pages).Str("confidence", validation.Confidence.String()).Str("report_type", validation.ReportType.String()).Strs("strong", validation.Strong).Strs("supporting", validation.Supporting).Msg("report scan: upload validated")

	// the document is untrusted input; screen it for injected instructions before the model ever sees it
	if client.Screener != nil {
		verdict, err := client.Screener.ScreenDocument(ctx, pdf)
		if err != nil {
			return ReportScanRequestResult{}, err
		}

		logx.FromContext(ctx).Debug().Bool("blocked", verdict.Blocked).Strs("filters", verdict.Filters).Str("template", client.Screener.Template()).Msg("report scan: upload screened")

		if verdict.Blocked {
			if err := failScan(systemCtx, request.DB, scanRecord, modelarmor.ErrDocumentBlocked.Error(), validation); err != nil {
				return ReportScanRequestResult{}, err
			}

			return ReportScanRequestResult{Message: modelarmor.ErrDocumentBlocked.Error(), ScanID: scanRecord.ID}, nil
		}
	}

	partNames, err := r.selectParts(scanRecord.Metadata)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("report scan: rejected requested sections")

		if err := failScan(systemCtx, request.DB, scanRecord, err.Error(), validation); err != nil {
			return ReportScanRequestResult{}, err
		}

		return ReportScanRequestResult{Message: err.Error(), ScanID: scanRecord.ID}, nil
	}

	summary := make(map[string]any, len(partNames))

	for _, name := range partNames {
		summary[name] = map[string]any{"status": PartStatePending, "count": 0}
	}

	metadata := scanRecord.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadata[ReportMetadataKey] = map[string]any{}
	metadata[ReportTypeMetadataKey] = soc2.KindName
	metadata[SummaryMetadataKey] = summary
	metadata[ValidationMetadataKey] = validation

	delete(metadata, ErrorMetadataKey)

	if err := request.DB.Scan.UpdateOneID(scanRecord.ID).SetStatus(enums.ScanStatusProcessing).SetMetadata(metadata).Exec(systemCtx); err != nil {
		return ReportScanRequestResult{}, err
	}

	cacheName, err := client.CacheDocument(ctx, bytes.NewReader(pdf), DocumentCacheTTL)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("report scan: document cache unavailable, parts will send the document")
	}

	deferDependents := slices.Contains(partNames, soc2.ControlsPart)

	for _, name := range partNames {
		if deferDependents && slices.Contains(soc2.ControlScopedParts, name) {
			continue
		}

		if _, err := request.Services.Gala().EmitWithHeaders(ctx, reportScanPartTopic.Name, ReportScanPartEnvelope{
			OrganizationID: cfg.OrganizationID,
			ScanID:         scanRecord.ID,
			Part:           name,
			Cache:          cacheName,
		}, gala.Headers{UniqueOnce: true}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("part", name).Msg("report scan: failed scheduling part job")

			return ReportScanRequestResult{}, err
		}
	}

	logx.FromContext(ctx).Debug().Int("parts", len(partNames)).Msg("report scan: part jobs scheduled")

	return ReportScanRequestResult{Message: "report scan submitted", ScanID: scanRecord.ID}, nil
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

// failScan marks the scan failed with a user-facing reason before any part tracking exists
func failScan(ctx context.Context, db *generated.Client, scanRecord *generated.Scan, reason string, validation docextract.Validation) error {
	metadata := scanRecord.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadata[ErrorMetadataKey] = reason
	metadata[ValidationMetadataKey] = validation

	return db.Scan.UpdateOneID(scanRecord.ID).SetStatus(enums.ScanStatusFailed).SetMetadata(metadata).Exec(ctx)
}

// scanSystemContext builds a context authorized to update Scan and Notification records for
// organizationID on behalf of the system
func scanSystemContext(ctx context.Context, organizationID string) context.Context {
	return auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		OrganizationID: organizationID,
		Capabilities:   auth.CapBypassFGA | auth.CapInternalOperation,
	})
}
