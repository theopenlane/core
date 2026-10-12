package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/review"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/docextract/soc2"
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
		// on creation of report_scan of status pending and performed by the openlane_report_parse
		// kick off the job to begin the scan
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
		// on creation of a pending vendor report review, kick off the review under its own rate limit
		entityops.MutationListener{
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpCreate},
			Match: []entityops.FieldMatch{
				{Field: scan.FieldScanType, In: []string{string(enums.ScanTypeReport)}},
				{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusPending)}},
				{Field: scan.FieldPerformedBy, In: []string{gemini.VendorReviewPerformedBy}},
			},
			Handle: entityops.RequireDep(handleVendorReviewCreated),
		},
		// on scan update when the status is completed and is a report parse, send both an in-app and email notification
		entityops.MutationListener{
			Concern:    entityops.MutationConcernNotification,
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{scan.FieldStatus},
			Match:      []entityops.FieldMatch{{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusCompleted)}}},
			RowMatch:   reportScanRowMatchForKind(soc2.KindName, gemini.PerformedBy),
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
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
		// on update when the status is completed, upload the report to the trust center as a not visible document so it can be added to the trust center by the user later
		entityops.MutationListener{
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{scan.FieldStatus},
			Match:      []entityops.FieldMatch{{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusCompleted)}}},
			RowMatch:   reportScanRowMatchForKind(soc2.KindName, gemini.PerformedBy),
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleScanReportSOC2Completed,
		},
		// on update when the status is completed, copy a vendor review's assessment onto its vendor and record it as a review
		entityops.MutationListener{
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{scan.FieldStatus},
			Match:      []entityops.FieldMatch{{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusCompleted)}}},
			RowMatch:   reportScanRowMatchForKind(soc2.KindName, gemini.VendorReviewPerformedBy),
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleVendorReviewCompleted,
		},
		// on scan update when the status fails, send scan failure notification and email
		entityops.MutationListener{
			Concern:    entityops.MutationConcernNotification,
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{scan.FieldStatus},
			Match:      []entityops.FieldMatch{{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusFailed)}}},
			RowMatch:   reportScanRowMatch(gemini.ReportPerformers...),
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
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

	summary := gemini.ReportSummary(metadata)
	if len(summary) == 0 {
		return nil
	}

	sections := lo.Keys(summary)
	slices.Sort(sections)

	rows := make([][]string, 0, len(sections))

	for _, section := range sections {
		entry := summary[section]

		status := string(entry.Status)
		if entry.Error != "" {
			status += ", " + entry.Error
		}

		rows = append(rows, []string{section, strconv.Itoa(entry.Count), status})
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
// since a status-only update does not propose scan_type or performed_by; the performers pick the
// organization's own imports, vendor reviews, or both
func reportScanRowMatch(performedBy ...string) []entityops.FieldMatch {
	return []entityops.FieldMatch{
		{Field: scan.FieldScanType, In: []string{string(enums.ScanTypeReport)}},
		{Field: scan.FieldPerformedBy, In: performedBy},
	}
}

// reportScanRowMatchForKind narrows the report scan gate to a report the given document kind
// accepted, so each kind carries its own listeners and copy; the kind is recorded when the upload is
// validated, so a row that never passed validation fails the match
func reportScanRowMatchForKind(kind string, performedBy ...string) []entityops.FieldMatch {
	return slices.Concat(reportScanRowMatch(performedBy...), []entityops.FieldMatch{
		{Field: scan.FieldDocumentKindName, In: []string{kind}},
	})
}

// handleScanReportSOC2Completed files the parsed SOC 2 report as a not-visible trust center document
// so the organization's own report is on hand once its controls and reviews are imported; the
// listener gates on the document kind, so organizations without the trust center module or a trust
// center are the only ones skipped here
func handleScanReportSOC2Completed(inv entityops.Invocation, _ entityops.MutationPayload) error {
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

	reportFileID, err := reportScanFileID(ctx, inv.Client, scanRecord)
	if err != nil {
		return err
	}

	// a scan started from a report already uploaded to the trust center must not file it a second time
	alreadyFiled, err := inv.Client.TrustCenterDoc.Query().
		Where(trustcenterdoc.Or(trustcenterdoc.OriginalFileID(reportFileID), trustcenterdoc.FileID(reportFileID))).
		Exist(ctx)
	if err != nil || alreadyFiled {
		return err
	}

	create := inv.Client.TrustCenterDoc.Create().
		SetTrustCenterID(trustCenterID).
		SetTitle("SOC 2 Report").
		SetOriginalFileID(reportFileID).
		SetTags([]string{"soc2"}).
		SetTrustCenterDocKindName("compliance").
		SetVisibility(enums.TrustCenterDocumentVisibilityNotVisible)

	if standardID, ok := soc2StandardID(ctx, inv.Client); ok {
		create.SetStandardID(standardID)
	}

	if err = create.Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("report scan: trust center document failed to be created")
	}

	return nil
}

// handleVendorReviewCompleted marks the vendors a completed scan reviewed as having a SOC 2 report
// and copies the assessment's risk rating and report period end onto them
func handleVendorReviewCompleted(inv entityops.Invocation, _ entityops.MutationPayload) error {
	ctx := logx.WithFields(inv.Context, map[string]any{"scan_id": inv.EntityID})

	scanRecord, ok, err := entityops.LoadEntity(ctx, inv.EntityID, inv.Client.Scan.Get)
	if err != nil || !ok {
		return err
	}

	entityIDs, err := scanRecord.QueryEntities().IDs(ctx)
	if err != nil {
		return err
	}

	assessment, ok := gemini.Assessment(scanRecord.Metadata)
	if !ok {
		logx.FromContext(ctx).Warn().Strs("entity_ids", entityIDs).Msg("report scan: vendor review completed without an assessment, vendors not updated")

		return nil
	}

	// a vendor review with no vendor linked finds or creates the vendor the report names
	if len(entityIDs) == 0 {
		vendorID, err := linkReviewedVendor(ctx, inv.Client, scanRecord, assessment)
		if err != nil {
			return err
		}

		entityIDs = []string{vendorID}
	}

	update := inv.Client.Entity.Update().Where(entity.IDIn(entityIDs...)).SetHasSoc2(true)

	// vendor risk ratings are stored uppercase, matching the scale the console offers
	if assessment.RiskRating != "" {
		update.SetRiskRating(strings.ToUpper(assessment.RiskRating))
	}

	if periodEnd, err := models.ToDateTime(assessment.PeriodEnd); err == nil {
		update.SetSoc2PeriodEnd(*periodEnd)
	} else {
		logx.FromContext(ctx).Debug().Err(err).Str("period_end", assessment.PeriodEnd).Msg("report scan: assessment period end unreadable, leaving vendor soc2 period end unchanged")
	}

	if err := update.Exec(ctx); err != nil {
		return err
	}

	if err := addProvidedServices(ctx, inv.Client, entityIDs, assessment.VendorServices); err != nil {
		return err
	}

	created, err := createVendorReviewRecord(ctx, inv.Client, scanRecord, entityIDs, assessment)
	if err != nil || !created {
		return err
	}

	return notifyVendorReviewCompleted(ctx, inv.Client, scanRecord, entityIDs[0])
}

// vendorRiskReviewConsolePath is the console page for a vendor's risk review tab
const vendorRiskReviewConsolePath = "/registry/vendors/%s?tab=risk-review"

// notifyVendorReviewCompleted tells the submitter the review's results are on the vendor, linking to
// its risk review tab; sent here rather than declaratively since the vendor is only known once linked
func notifyVendorReviewCompleted(ctx context.Context, client *generated.Client, scanRecord *generated.Scan, vendorID string) error {
	if scanRecord.CreatedBy == "" {
		return nil
	}

	topic := enums.NotificationTopicReportScan

	return entityops.CreateNotifications(ctx, client, []string{scanRecord.CreatedBy}, &generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeUser,
		Title:            "Vendor report review completed",
		Body:             fmt.Sprintf("Review of %s finished, the auditor's opinion, risk rating, and findings are on the vendor's risk review", scanRecord.Target),
		Data:             map[string]any{"scan_id": scanRecord.ID, "report_name": scanRecord.Target, "vendor_id": vendorID, "url": fmt.Sprintf(vendorRiskReviewConsolePath, vendorID)},
		Channels:         []enums.Channel{enums.ChannelInApp},
		Topic:            &topic,
		ObjectType:       entityops.SchemaScan.Name,
		OwnerID:          &scanRecord.OwnerID,
	})
}

// addProvidedServices adds the report's services to each vendor, skipping ones it already lists
// regardless of case so services entered by hand are not duplicated
func addProvidedServices(ctx context.Context, client *generated.Client, entityIDs, services []string) error {
	if len(services) == 0 {
		return nil
	}

	vendors, err := client.Entity.Query().Where(entity.IDIn(entityIDs...)).Select(entity.FieldID, entity.FieldProvidedServices).All(ctx)
	if err != nil {
		return err
	}

	for _, vendor := range vendors {
		missing := missingServices(vendor.ProvidedServices, services)
		if len(missing) == 0 {
			continue
		}

		if err := client.Entity.UpdateOneID(vendor.ID).AppendProvidedServices(missing).Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}

// missingServices are the trimmed services not already in existing, compared without case
func missingServices(existing, services []string) []string {
	seen := make(map[string]bool, len(existing)+len(services))
	for _, service := range existing {
		seen[strings.ToLower(strings.TrimSpace(service))] = true
	}

	var missing []string

	for _, service := range services {
		service = strings.TrimSpace(service)
		key := strings.ToLower(service)

		if service == "" || seen[key] {
			continue
		}

		seen[key] = true
		missing = append(missing, service)
	}

	return missing
}

const (
	// vendorReviewCategory is the category of reviews recorded from a vendor's SOC 2 report
	vendorReviewCategory = "SOC 2 report"
	// vendorReviewReporter is the reporter shown on reviews the system recorded from a report
	vendorReviewReporter = "Openlane"
)

// createVendorReviewRecord records a completed vendor review as a review on the vendors it covers,
// keyed by the scan so a retried listener does not record it twice; reports whether it was created
func createVendorReviewRecord(ctx context.Context, client *generated.Client, scanRecord *generated.Scan, entityIDs []string, assessment gemini.ReportAssessment) (bool, error) {
	exists, err := client.Review.Query().Where(review.ExternalID(scanRecord.ID)).Exist(ctx)
	if err != nil || exists {
		return false, err
	}

	validation, _ := gemini.Validation(scanRecord.Metadata)
	findings := gemini.Findings(scanRecord.Metadata)

	err = client.Review.Create().
		SetTitle(vendorReviewTitle(validation.ReportType, assessment.PeriodEnd)).
		SetSummary(assessment.Summary).
		SetDetails(vendorReviewDetails(assessment, findings)).
		SetCategory(vendorReviewCategory).
		SetClassification(soc2.OpinionLabel(assessment.Opinion)).
		SetStatus(enums.ReviewStatusCompleted).
		SetReporter(vendorReviewReporter).
		SetSource(gemini.PerformedBy).
		SetExternalID(scanRecord.ID).
		SetReviewedAt(models.DateTime(time.Now())).
		SetMetadata(map[string]any{
			"scan_id":        scanRecord.ID,
			"report_type":    validation.ReportType.String(),
			"opinion":        assessment.Opinion.String(),
			"risk_rating":    assessment.RiskRating,
			"period_end":     assessment.PeriodEnd,
			"findings_count": len(findings),
		}).
		AddEntityIDs(entityIDs...).
		Exec(ctx)

	return err == nil, err
}

// vendorReviewTitle names the review after the report type and the end of the period it covers
func vendorReviewTitle(reportType docextract.ReportType, periodEnd string) string {
	title := "SOC 2 report review"

	switch reportType {
	case soc2.ReportTypeOne:
		title = "SOC 2 Type 1 report review"
	case soc2.ReportTypeTwo:
		title = "SOC 2 Type 2 report review"
	}

	if periodEnd != "" {
		title += ", period ending " + periodEnd
	}

	return title
}

// vendorReviewDetails writes the risk rationale, the opinion, period, and rating, and every finding as
// a markdown table for the review's notes
func vendorReviewDetails(assessment gemini.ReportAssessment, findings []gemini.ReportFinding) string {
	var details strings.Builder

	if rationale := strings.TrimSpace(assessment.RiskRationale); rationale != "" {
		details.WriteString(rationale + "\n\n")
	}

	// a trailing double space is a markdown line break, keeping the three facts together as one block
	fmt.Fprintf(&details, "**Auditor opinion:** %s  \n", soc2.OpinionLabel(assessment.Opinion))

	if assessment.PeriodEnd != "" {
		fmt.Fprintf(&details, "**Report period end:** %s  \n", assessment.PeriodEnd)
	}

	fmt.Fprintf(&details, "**Risk rating:** %s\n\n", strings.ToUpper(assessment.RiskRating))

	if len(findings) == 0 {
		details.WriteString("**Findings:** The auditor noted no exceptions.")

		return details.String()
	}

	details.WriteString("**Findings:**\n\n| Summary | Criticality | Related Controls |\n| --- | --- | --- |\n")

	for _, finding := range findings {
		fmt.Fprintf(&details, "| %s | %s | %s |\n", markdownCell(finding.Description), markdownCell(finding.Severity), markdownCell(strings.Join(finding.RefCodes, ", ")))
	}

	return strings.TrimSuffix(details.String(), "\n")
}

// markdownCell keeps text inside one markdown table cell by escaping pipes and flattening line breaks
func markdownCell(text string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(text, "|", `\|`)), " ")
}

// linkReviewedVendor finds or creates the vendor a create-from-report review names, then links the
// scan to it and files the report as one of its documents
func linkReviewedVendor(ctx context.Context, client *generated.Client, scanRecord *generated.Scan, assessment gemini.ReportAssessment) (string, error) {
	vendorID, err := findOrCreateReviewedVendor(ctx, client, scanRecord.OwnerID, assessment)
	if err != nil {
		return "", err
	}

	if err := client.Scan.UpdateOneID(scanRecord.ID).AddEntityIDs(vendorID).Exec(ctx); err != nil {
		return "", err
	}

	reportFileID, err := reportScanFileID(ctx, client, scanRecord)
	if err != nil {
		return "", err
	}

	return vendorID, client.Entity.UpdateOneID(vendorID).AddFileIDs(reportFileID).Exec(ctx)
}

// findOrCreateReviewedVendor matches an existing vendor by the report's domains, then by name, and
// creates a draft vendor when none matches
func findOrCreateReviewedVendor(ctx context.Context, client *generated.Client, organizationID string, assessment gemini.ReportAssessment) (string, error) {
	name := strings.TrimSpace(assessment.VendorName)
	if name == "" {
		return "", ErrVendorReviewNameMissing
	}

	typeID, hasVendorType, err := orgVendorEntityTypeID(ctx, client, organizationID)
	if err != nil {
		return "", err
	}

	scope := []predicate.Entity{entity.OwnerIDEQ(organizationID)}
	if hasVendorType {
		scope = append(scope, entity.EntityTypeIDEQ(typeID))
	}

	domains := normalizedDomains(assessment.VendorDomains)

	for _, domain := range domains {
		if id, found, err := firstEntityID(ctx, client, append(scope, entityDomainContains(domain))...); err != nil || found {
			return id, err
		}
	}

	if id, found, err := firstEntityID(ctx, client, append(scope, entity.Or(entity.NameEqualFold(name), entity.DisplayNameEqualFold(name)))...); err != nil || found {
		return id, err
	}

	create := client.Entity.Create().SetName(name).SetStatus(enums.EntityStatusDraft).SetDomains(domains)

	if description := strings.TrimSpace(assessment.VendorDescription); description != "" {
		create.SetDescription(description)
	}

	if hasVendorType {
		create.SetEntityTypeID(typeID)
	}

	created, err := create.Save(ctx)
	if err != nil {
		return "", err
	}

	logx.FromContext(ctx).Info().Str("entity_id", created.ID).Str("vendor_name", name).Msg("report scan: created draft vendor from soc2 report")

	return created.ID, nil
}

// firstEntityID is the first entity matching the predicates, reporting false when none does
func firstEntityID(ctx context.Context, client *generated.Client, preds ...predicate.Entity) (string, bool, error) {
	id, err := client.Entity.Query().Where(preds...).FirstID(ctx)

	switch {
	case generated.IsNotFound(err):
		return "", false, nil
	case err != nil:
		return "", false, err
	default:
		return id, true, nil
	}
}

// entityDomainContains matches entities whose domains list holds the domain
func entityDomainContains(domain string) predicate.Entity {
	return predicate.Entity(func(s *sql.Selector) {
		s.Where(sqljson.ValueContains(entity.FieldDomains, domain))
	})
}

// normalizedDomains lowercases and trims the domains, dropping blanks and repeats
func normalizedDomains(domains []string) []string {
	normalized := make([]string, 0, len(domains))

	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain != "" && !slices.Contains(normalized, domain) {
			normalized = append(normalized, domain)
		}
	}

	return normalized
}

// reportScanFileID is the file the report was parsed from, taken from the id recorded at submit and
// falling back to the oldest attached pdf for scans submitted before the id was recorded
func reportScanFileID(ctx context.Context, client *generated.Client, scanRecord *generated.Scan) (string, error) {
	if fileID, ok := gemini.ReportFileID(scanRecord.Metadata); ok {
		return fileID, nil
	}

	reportFile, err := client.Scan.QueryFiles(scanRecord).
		Where(file.DetectedContentTypeEQ(pdftext.ContentType)).
		Order(file.ByCreatedAt(), file.ByID()).
		First(ctx)
	if err != nil {
		return "", err
	}

	return reportFile.ID, nil
}

// soc2Framework is the framework identifier of the system SOC 2 standard
const soc2Framework = "soc2"

// soc2StandardID resolves the latest active system SOC 2 standard, reporting false when none exists
func soc2StandardID(ctx context.Context, client *generated.Client) (string, bool) {
	standardID, err := client.Standard.Query().
		Where(
			standard.FrameworkEQ(soc2Framework),
			standard.StatusEQ(enums.StandardActive),
			standard.IsPublic(true),
			standard.SystemOwned(true),
		).
		Order(standard.ByVersion(sql.OrderDesc()), standard.ByID()).
		FirstID(ctx)
	if err != nil {
		return "", false
	}

	return standardID, true
}

// handleScanReportCreated runs the parse request for a newly created report-type scan inline so the
// operation's per-organization rate limit is applied here
func handleScanReportCreated(inv entityops.Invocation, _ entityops.MutationPayload, rt *intruntime.Runtime) error {
	return runReportScanOperation(inv, rt, gemini.ReportScanRequestOp.Name(), reportScanRateLimitedReason)
}

// handleVendorReviewCreated runs the review of a newly created vendor report scan inline under the
// vendor review operation, which carries its own per-organization rate limit
func handleVendorReviewCreated(inv entityops.Invocation, _ entityops.MutationPayload, rt *intruntime.Runtime) error {
	return runReportScanOperation(inv, rt, gemini.VendorReportScanRequestOp.Name(), vendorReviewRateLimitedReason)
}

// vendorReviewRateLimitedReason is the user-facing failure reason stored when the vendor review rate limit rejects a scan
const vendorReviewRateLimitedReason = "a vendor report was already reviewed for this organization within the last hour, try again later"

// runReportScanOperation runs the named request operation for the scan, marking the scan failed
// with the given reason when the operation's rate limit rejects it
func runReportScanOperation(inv entityops.Invocation, rt *intruntime.Runtime, operation, rateLimitedReason string) error {
	scanRecord, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Scan.Get)
	if err != nil || !ok {
		return err
	}

	// both request operations take the scan and the organization it belongs to
	config, err := json.Marshal(gemini.ReportScanRequest{
		ScanID:         scanRecord.ID,
		OrganizationID: scanRecord.OwnerID,
	})
	if err != nil {
		return err
	}

	_, err = rt.ExecuteRuntimeOperation(inv.Context, gemini.DefinitionID.ID(), operation, config)
	if !errors.Is(err, intruntime.ErrOperationRateLimited) {
		return err
	}

	return failRateLimitedReportScan(inv, scanRecord, rateLimitedReason)
}

// failRateLimitedReportScan marks a report scan failed with the reason the rate limit rejected it
func failRateLimitedReportScan(inv entityops.Invocation, scanRecord *generated.Scan, reason string) error {
	logx.FromContext(inv.Context).Warn().Str("scan_id", scanRecord.ID).Str("reason", reason).Msg("report scan: rate limited, marking scan failed")

	metadata := scanRecord.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadata[gemini.ErrorMetadataKey] = reason

	return inv.Client.Scan.UpdateOneID(scanRecord.ID).
		SetStatus(enums.ScanStatusFailed).
		SetMetadata(metadata).
		Exec(inv.Context)
}
