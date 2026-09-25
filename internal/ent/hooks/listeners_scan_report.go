package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/pdftext"
)

// init registers the report scan listeners so gala setup picks them up automatically
func init() { registerListeners(ReportScanListeners) }

// ReportScanListeners submits pending report scans to the parser
func ReportScanListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpCreate},
			Match: []entityops.FieldMatch{
				{Field: scan.FieldScanType, In: []string{string(enums.ScanTypeReport)}},
				{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusPending)}},
				{Field: scan.FieldPerformedBy, In: []string{gemini.PerformedBy}},
			},
			Handle: entityops.RequireDep(handleScanReportCreated),
		},
		entityops.MutationListener{
			Concern:    entityops.MutationConcernNotification,
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{scan.FieldStatus},
			Match:      []entityops.FieldMatch{{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusCompleted)}}},
			RowMatch:   reportScanRowMatch,
			Caller:     internalCaller,
			Notify: &entityops.NotifySpec{
				Recipients: entityops.RecipientsFromField(scan.FieldCreatedBy),
				Content: entityops.NotificationContent{
					Type:          enums.NotificationTypeUser,
					Topic:         enums.NotificationTopicReportScan,
					TitleTemplate: "Report scan completed",
					BodyTemplate:  "Report {{ .Entity.target }} finished parsing, review the results to import your controls, vendors, findings, and more",
					Data:          map[string]any{"scan_id": "{{ .EntityID }}", "report_name": "{{ .Entity.target }}", "url": reportScanConsolePath},
					Channels:      []enums.Channel{enums.ChannelInApp, enums.ChannelEmail},
				},
				Email: entityops.EmailVia(email.DefinitionID, email.BrandedMessageOp, reportScanCompletedEmail).Batched(),
			},
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{scan.FieldStatus},
			Match:      []entityops.FieldMatch{{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusCompleted)}}},
			RowMatch:   reportScanRowMatch,
			Caller:     internalCaller,
			Handle:     handleScanReportCompleted,
		},
		entityops.MutationListener{
			Concern:    entityops.MutationConcernNotification,
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{scan.FieldStatus},
			Match:      []entityops.FieldMatch{{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusFailed)}}},
			RowMatch:   reportScanRowMatch,
			Caller:     internalCaller,
			Notify: &entityops.NotifySpec{
				Recipients: entityops.RecipientsFromField(scan.FieldCreatedBy),
				Content: entityops.NotificationContent{
					Type:          enums.NotificationTypeUser,
					Topic:         enums.NotificationTopicReportScan,
					TitleTemplate: "Report scan failed",
					BodyTemplate:  "Report {{ .Entity.target }} could not be parsed, {{ .Entity.metadata.error }}. Upload it again to retry",
					Data:          map[string]any{"scan_id": "{{ .EntityID }}", "report_name": "{{ .Entity.target }}", "error": "{{ .Entity.metadata.error }}", "url": reportScanConsolePath},
					Channels:      []enums.Channel{enums.ChannelInApp, enums.ChannelEmail},
				},
				Email: entityops.EmailVia(email.DefinitionID, email.BrandedMessageOp, reportScanFailedEmail).Batched(),
			},
		},
	}
}

// reportScanCompletedEmail builds the branded email telling the submitter their report is ready to review
func reportScanCompletedEmail(_ entityops.Invocation, _ entityops.MutationPayload, row json.RawMessage, recipient entityops.EmailRecipient) (email.BrandedMessageRequest, error) {
	reportName := entityops.FieldValue(row, scan.FieldTarget)
	consolePath, _ := recipient.Data["url"].(string)

	return email.BrandedMessageRequest{
		RecipientInfo: emailRecipientInfo(recipient.Users...),
		Subject:       "Your report is ready to review",
		Preheader:     "Parsing finished for " + reportName,
		Title:         "Your report is ready to review",
		Intros:        []string{"We finished parsing " + reportName + ". Here is what we found in each section."},
		Tables:        reportSummaryTables(row),
		ButtonText:    "Review Results",
		ButtonLink:    consolePath,
	}, nil
}

// reportSummaryTables renders the scan's report summary as one table of section, item count,
// and status so the reader can see what was extracted without opening the console
func reportSummaryTables(row json.RawMessage) []email.MessageTable {
	metadata, ok := jsonx.DecodeObjectKey[map[string]any](row, scan.FieldMetadata)
	if !ok {
		return nil
	}

	summary, _ := metadata[gemini.SummaryMetadataKey].(map[string]any)
	if len(summary) == 0 {
		return nil
	}

	sections := lo.Keys(summary)
	slices.Sort(sections)

	rows := make([][]string, 0, len(sections))

	for _, section := range sections {
		entry, _ := summary[section].(map[string]any)
		count, _ := entry["count"].(float64)
		status, _ := entry["status"].(string)

		if message, ok := entry["error"].(string); ok && message != "" {
			status += ", " + message
		}

		rows = append(rows, []string{section, strconv.Itoa(int(count)), status})
	}

	return []email.MessageTable{{
		Columns: []string{"Section", "Items", "Status"},
		Rows:    rows,
		Footer:  "Review the results and choose what to import into your organization.",
	}}
}

// reportScanFailedEmail builds the branded email telling the submitter their report could not be parsed
func reportScanFailedEmail(_ entityops.Invocation, _ entityops.MutationPayload, row json.RawMessage, recipient entityops.EmailRecipient) (email.BrandedMessageRequest, error) {
	reportName := entityops.FieldValue(row, scan.FieldTarget)
	reason, _ := recipient.Data["error"].(string)
	consolePath, _ := recipient.Data["url"].(string)

	return email.BrandedMessageRequest{
		RecipientInfo: emailRecipientInfo(recipient.Users...),
		Subject:       "Your report could not be parsed",
		Preheader:     "Parsing failed for " + reportName,
		Title:         "Your report could not be parsed",
		Intros: []string{
			"We were unable to parse " + reportName + ": " + reason + ".",
			"You can upload the report again to retry.",
		},
		ButtonText: "View Scan",
		ButtonLink: consolePath,
	}, nil
}

// reportScanConsolePath is the console page for a scan, rendered into notification data since
// the scan schema declares no console route of its own
const reportScanConsolePath = "/exposure/scans/{{ .EntityID }}"

// reportScanRateLimitedReason is the user-facing failure reason stored when the submit rate limit rejects a scan
const reportScanRateLimitedReason = "a report was already submitted for this organization within the last 24 hours, try again later"

// reportScanRowMatch gates status listeners on the stored row being a system-parsed report scan,
// since a status-only update does not propose scan_type or performed_by
var reportScanRowMatch = []entityops.FieldMatch{
	{Field: scan.FieldScanType, In: []string{string(enums.ScanTypeReport)}},
	{Field: scan.FieldPerformedBy, In: []string{gemini.PerformedBy}},
}

// internalCaller grants the internal-operation capability so the listener passes privacy
func internalCaller(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
	return restored.WithCapabilities(auth.CapInternalOperation)
}

// handleScanReportCompleted files the parsed report as a not-visible trust center document so the
// organization's own report is on hand once its controls and reviews are imported; organizations
// without the trust center module or a trust center are skipped
func handleScanReportCompleted(inv entityops.Invocation, _ entityops.MutationPayload) error {
	ctx := logx.WithFields(inv.Context, map[string]any{"scan_id": inv.EntityID})

	if utils.ModulesEnabled(inv.Client) {
		ok, err := rule.HasFeature(ctx, models.CatalogTrustCenterModule.String())
		if err != nil {
			logx.FromContext(ctx).Debug().Err(err).Msg("report scan: trust center module lookup failed")

			return err
		}

		if !ok {
			return nil
		}
	}

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	scanRecord, ok, err := entityops.LoadEntity(ctx, inv.EntityID, inv.Client.Scan.Get)
	if err != nil || !ok {
		return err
	}

	trustCenterID, err := inv.Client.TrustCenter.Query().Where(trustcenter.OwnerID(scanRecord.OwnerID)).OnlyID(ctx)
	if generated.IsNotFound(err) {
		logx.FromContext(ctx).Debug().Msg("report scan: no trust center, skipping document")

		return nil
	}

	if err != nil {
		return err
	}

	reportFile, err := inv.Client.Scan.QueryFiles(scanRecord).Where(file.DetectedContentTypeEQ(pdftext.ContentType)).First(ctx)
	if err != nil {
		return err
	}

	create := inv.Client.TrustCenterDoc.Create().
		SetTrustCenterID(trustCenterID).
		SetTitle(scanRecord.Target).
		SetOriginalFileID(reportFile.ID).
		SetTags([]string{"soc2"}).
		SetTrustCenterDocKindName("compliance").
		SetVisibility(enums.TrustCenterDocumentVisibilityNotVisible)

	if standardID, ok := soc2StandardID(ctx, inv.Client); ok {
		create.SetStandardID(standardID)
	}

	doc, err := create.Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("report scan: trust center document failed to be created")

		return nil
	}

	logx.FromContext(ctx).Debug().Str("trust_center_doc_id", doc.ID).Msg("report scan: trust center document created")

	return nil
}

// soc2StandardID resolves the latest active system SOC 2 standard, reporting false when none exists
func soc2StandardID(ctx context.Context, client *generated.Client) (string, bool) {
	standardID, err := client.Standard.Query().
		Where(
			standard.FrameworkEQ(soc2Framework),
			standard.StatusEQ(enums.StandardActive),
			standard.SystemOwned(true),
		).
		Order(standard.ByVersion(sql.OrderDesc()), standard.ByID()).
		FirstID(ctx)
	if err != nil {
		return "", false
	}

	return standardID, true
}

// soc2Framework is the framework key of the system SOC 2 standard
const soc2Framework = "soc2"

// handleScanReportCreated runs the parse request for a newly created report-type scan inline so the
// operation's per-organization rate limit is applied here; a rejected scan is marked failed with the
// cause rather than left pending
func handleScanReportCreated(inv entityops.Invocation, _ entityops.MutationPayload, rt *intruntime.Runtime) error {
	scanRecord, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Scan.Get)
	if err != nil || !ok {
		return err
	}

	config, err := json.Marshal(gemini.ReportScanRequest{
		ScanID:         scanRecord.ID,
		OrganizationID: scanRecord.OwnerID,
	})
	if err != nil {
		return err
	}

	_, err = rt.ExecuteRuntimeOperation(inv.Context, gemini.DefinitionID.ID(), gemini.ReportScanRequestOp.Name(), config)
	if !errors.Is(err, intruntime.ErrOperationRateLimited) {
		return err
	}

	logx.FromContext(inv.Context).Warn().Str("scan_id", scanRecord.ID).Msg("report scan: rate limited, marking scan failed")

	metadata := scanRecord.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadata[gemini.ErrorMetadataKey] = reportScanRateLimitedReason

	return inv.Client.Scan.UpdateOneID(scanRecord.ID).
		SetStatus(enums.ScanStatusFailed).
		SetMetadata(metadata).
		Exec(privacy.DecisionContext(inv.Context, privacy.Allow))
}
