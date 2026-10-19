package graphapi

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/docextract/soc2"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
	"github.com/theopenlane/core/v2/internal/graphapi/model"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
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

	envelope, err := buildReportScanImportEnvelope(organizationID, input)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("report scan import: failed to read import input")
		return nil, ErrReportScanImportFailed
	}

	config, err := json.Marshal(envelope)
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

// buildReportScanImportEnvelope translates the GraphQL input into the envelope the async import handler consumes;
// edited items are create inputs, which encode under the same field names the import reads
func buildReportScanImportEnvelope(organizationID string, input model.ImportReportScanInput) (gemini.ReportScanImport, error) {
	envelope := gemini.ReportScanImport{
		OrganizationID: organizationID,
		ScanID:         input.ScanID,
		Accepted:       map[string][]string{},
		Edited:         map[string][]map[string]any{},
	}

	sections := map[string]*model.ImportReportScanSectionInput{
		soc2.ProgramsPart: input.Programs,
		soc2.ControlsPart: input.Controls,
		soc2.EntitiesPart: input.Entities,
		soc2.AssetsPart:   input.Assets,
		soc2.GroupsPart:   input.Groups,
		soc2.ContactsPart: input.Contacts,
		soc2.ReviewsPart:  input.Reviews,
		soc2.FindingsPart: input.Findings,
	}

	for section, accepted := range sections {
		if accepted != nil {
			envelope.Accepted[section] = accepted.Keys
		}
	}

	if input.Platforms != nil {
		envelope.Accepted[soc2.PlatformsPart] = input.Platforms.Keys

		var edited []map[string]any
		if err := jsonx.RoundTrip(input.Platforms.Edited, &edited); err != nil {
			return envelope, err
		}

		if len(edited) > 0 {
			envelope.Edited[soc2.PlatformsPart] = edited
		}
	}

	if input.SystemDetails != nil {
		envelope.Accepted[soc2.SystemDetailsPart] = input.SystemDetails.Keys

		for _, edit := range input.SystemDetails.Edited {
			item, err := systemDetailEditItem(edit)
			if err != nil {
				return envelope, err
			}

			envelope.Edited[soc2.SystemDetailsPart] = append(envelope.Edited[soc2.SystemDetailsPart], item)
		}
	}

	return envelope, nil
}

// systemDetailEditItem encodes an edited system detail with the platform it runs on, as the import reads it
func systemDetailEditItem(edit *model.ReportScanSystemDetailEdit) (map[string]any, error) {
	var item map[string]any
	if err := jsonx.RoundTrip(edit.SystemDetail, &item); err != nil {
		return nil, err
	}

	if edit.PlatformName != nil {
		item[soc2.PlatformNameKey] = *edit.PlatformName
	}

	return item, nil
}
