package graphapi

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
	"github.com/theopenlane/core/v2/internal/graphapi/model"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

var (
	// ErrReportScanImportNotReport is returned when the submitted scan is not a report scan
	ErrReportScanImportNotReport = errors.New("report scan import: scan is not a report scan")

	// ErrReportScanImportNotParsed is returned when the report scan has not finished parsing
	ErrReportScanImportNotParsed = errors.New("report scan import: report has not finished parsing")

	// ErrReportScanImportFailed is returned when the accepted report could not be queued for import;
	// the underlying cause is internal and gets logged, not surfaced
	ErrReportScanImportFailed = errors.New("report scan import: failed to queue import")
)

// importReportScan validates the parsed report scan, resolves the organization it belongs to,
// and emits the accepted items for asynchronous import
func (r *mutationResolver) importReportScan(ctx context.Context, input model.ImportReportScanInput) (*model.ImportReportScanPayload, error) {
	scanRecord, err := withTransactionalMutation(ctx).Scan.Get(ctx, input.ScanID)
	if err != nil {
		return nil, err
	}

	if scanRecord.ScanType != enums.ScanTypeReport {
		return nil, ErrReportScanImportNotReport
	}

	if _, ok := gemini.ReportFileID(scanRecord.Metadata); !ok || scanRecord.Status != enums.ScanStatusCompleted {
		return nil, ErrReportScanImportNotParsed
	}

	organizationID := scanRecord.OwnerID

	ctx, err = common.SetOrganizationInAuthContext(ctx, &organizationID)
	if err != nil {
		return nil, err
	}

	rt := intruntime.Default()
	if rt == nil {
		return nil, ErrCampaignDispatchRuntimeRequired
	}

	config, err := json.Marshal(buildReportScanImportEnvelope(organizationID, input))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("report scan import: failed to encode import config")
		return nil, ErrReportScanImportFailed
	}

	if _, err := rt.Dispatch(ctx, types.DispatchRequest{
		DefinitionID: gemini.DefinitionID.ID(),
		OwnerID:      organizationID,
		Operation:    gemini.ReportScanImportOp.Name(),
		Config:       config,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
	}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("report scan import: failed to queue import")
		return nil, ErrReportScanImportFailed
	}

	return &model.ImportReportScanPayload{Accepted: true}, nil
}

// buildReportScanImportEnvelope translates the GraphQL input into the envelope the async import handler consumes
func buildReportScanImportEnvelope(organizationID string, input model.ImportReportScanInput) gemini.ReportScanImport {
	envelope := gemini.ReportScanImport{
		OrganizationID: organizationID,
		ScanID:         input.ScanID,
		Accepted:       make(map[string][]string, len(input.Sections)),
		Edited:         map[string][]map[string]any{},
	}

	for _, section := range input.Sections {
		envelope.Accepted[section.Section] = append(envelope.Accepted[section.Section], section.Keys...)

		if len(section.Edited) > 0 {
			envelope.Edited[section.Section] = append(envelope.Edited[section.Section], section.Edited...)
		}
	}

	return envelope
}
