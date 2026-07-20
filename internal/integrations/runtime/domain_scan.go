package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"time"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
	"github.com/riverqueue/river"
	"github.com/theopenlane/iam/auth"
	"golang.org/x/sync/errgroup"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/scan"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/vendorenrich"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/domainscan"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// errDomainScanTaskFailed is returned when Cloudflare reports the scan task itself failed
var errDomainScanTaskFailed = errors.New("domain scan: cloudflare scan task failed")

// errDomainScanMaxAttemptsReached is returned when a scan never completes within the poll budget
var errDomainScanMaxAttemptsReached = errors.New("domain scan: max poll attempts reached")

// domainScanEnrichmentMetadataKey is the Scan.Metadata key used to carry the enrichment data
// gathered concurrently with URL Scanner submission and polling, until the poll cycle
// finishes and needs it to build the final report
const domainScanEnrichmentMetadataKey = "enrichment"

// HandleDomainScanCreate submits an organization's domains to Cloudflare's URL Scanner and creates a Scan object in the system to track the scan
func (r *Runtime) HandleDomainScanCreate(ctx context.Context, envelope operations.DomainScanCreateEnvelope) error {
	ctx = logx.WithFields(ctx, logx.LogFields{
		"organization_id": envelope.OrganizationID,
		"domains":         envelope.Domains,
	})

	systemCtx := domainScanSystemContext(ctx, envelope.OrganizationID)

	now, err := models.ToDateTime(time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	scanIDs := make(map[string]string, len(envelope.Domains))

	// the listener on Scan creation (listeners_scan_domain.go) would otherwise also try to
	// submit each of these individually; skip it since this batch submits them all together below
	createCtx := workflows.SkipEventEmission(systemCtx)

	for _, domain := range envelope.Domains {
		scanRecord, err := r.DB().Scan.Create().
			SetOwnerID(envelope.OrganizationID).
			SetTarget(domain).
			SetScanType(enums.ScanTypeDomain).
			SetScanDate(*now).
			SetPerformedBy(operations.DomainScanPerformedBy).
			SetStatus(enums.ScanStatusPending).
			Save(createCtx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("domain", domain).Msg("domain scan: failed creating scan record")
			return err
		}

		scanIDs[domain] = scanRecord.ID
	}

	return r.submitAndScheduleDomainScans(ctx, envelope.OrganizationID, scanIDs, envelope.ForceRefresh)
}

// HandleDomainScanSubmit submits a single already-created Scan record to domain scanner and schedules its poll cycle
func (r *Runtime) HandleDomainScanSubmit(ctx context.Context, organizationID, scanID, domain string, forceRefresh bool) error {
	ctx = logx.WithFields(ctx, logx.LogFields{
		"organization_id": organizationID,
		"domain":          domain,
	})

	return r.submitAndScheduleDomainScans(ctx, organizationID, map[string]string{domain: scanID}, forceRefresh)
}

// submitAndScheduleDomainScans submits the scan records in scanID to the domain scan together and schedules a poll cycle for each.
// Any domain that isn't returned marked failed rather than left stuck in "processing" forever
func (r *Runtime) submitAndScheduleDomainScans(ctx context.Context, organizationID string, scanIDs map[string]string, forceRefresh bool) error {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	// snapshotted before the loop below starts deleting entries, so every sibling (including
	// this one) knows the full group to check against once it's this one's turn to finish
	siblingScanIDs := slices.Collect(maps.Values(scanIDs))

	// mark every scan processing as soon as this job actually starts working, rather than
	// leaving them in "pending" (indistinguishable from "not yet picked up") for however long
	// the Cloudflare submission and enrichment gathering below take
	if err := r.DB().Scan.Update().
		Where(scan.IDIn(siblingScanIDs...)).
		SetStatus(enums.ScanStatusProcessing).
		Exec(systemCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed marking scans processing before submission")
	}

	domains := make([]string, 0, len(scanIDs))
	for domain := range scanIDs {
		domains = append(domains, domain)
	}

	config, err := json.Marshal(cloudflare.DomainScanSubmit{
		Domains: domains,
	})
	if err != nil {
		return err
	}

	response, err := r.ExecuteRuntimeOperation(ctx, cloudflare.DefinitionID.ID(), cloudflare.DomainScanSubmitOp.Name(), config)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed submitting scans to cloudflare")
		r.markDomainScansFailed(ctx, organizationID, scanIDs)

		return errors.Join(err, r.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs))
	}

	var result cloudflare.DomainScanSubmitResult
	if err := json.Unmarshal(response, &result); err != nil {
		r.markDomainScansFailed(ctx, organizationID, scanIDs)
		return errors.Join(err, r.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs))
	}

	enrichments := r.gatherDomainScanEnrichments(ctx, result.Scans, forceRefresh)

	// each domain is handled independently: one domain's failure marks only that domain failed
	// and moves on, rather than aborting the loop and leaving the remaining domains stuck with no
	// poll cycle ever scheduled for them
	for i, scan := range result.Scans {
		domain := hostFromURL(scan.URL)

		internalScanID, ok := scanIDs[domain]
		if !ok {
			logx.FromContext(ctx).Warn().Str("domain", domain).Msg("domain scan: cloudflare returned an unexpected domain, skipping")
			continue
		}

		delete(scanIDs, domain)

		if err := r.DB().Scan.UpdateOneID(internalScanID).
			SetMetadata(map[string]any{domainScanEnrichmentMetadataKey: enrichments[i]}).
			Exec(systemCtx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("scan_id", scan.UUID).Msg("domain scan: failed updating scan record with enrichment")
			r.markDomainScanFailed(ctx, organizationID, internalScanID)

			if notifyErr := r.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs); notifyErr != nil {
				logx.FromContext(ctx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after enrichment update failure")
			}

			continue
		}

		receipt := r.Gala().EmitWithHeaders(ctx, operations.DomainScanPollTopic, operations.DomainScanPollEnvelope{
			OrganizationID: organizationID,
			ScanResultID:   scan.UUID,
			InternalScanID: internalScanID,
			SiblingScanIDs: siblingScanIDs,
		}, gala.Headers{})
		if receipt.Err != nil {
			logx.FromContext(ctx).Error().Err(receipt.Err).Str("scan_id", scan.UUID).Msg("domain scan: failed scheduling poll cycle")
			r.markDomainScanFailed(ctx, organizationID, internalScanID)

			if notifyErr := r.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs); notifyErr != nil {
				logx.FromContext(ctx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after poll scheduling failure")
			}
		}
	}

	// any domain cloudflare didn't return a scan for is left with a Scan record that will never be polled, so mark it failed rather than
	// leaving it stuck forever
	if len(scanIDs) > 0 {
		logx.FromContext(ctx).Warn().Int("count", len(scanIDs)).Msg("domain scan: some domains were not submitted to cloudflare, marking as failed")
		r.markDomainScansFailed(ctx, organizationID, scanIDs)

		if err := r.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed checking group completion after partial submission failure")
		}
	}

	logx.FromContext(ctx).Info().Int("scan_count", len(result.Scans)).Msg("domain scan: submission jobs scheduled")

	return nil
}

// domainScanSystemContext builds a context authorized to create/update Scan and Notification records
// for organizationID on behalf of the system. CapBypassOrgFilter is deliberately omitted so every
// query made through this context stays auto-scoped to organizationID - every use of this context
// operates on a single org, so nothing should ever need to see across orgs
func domainScanSystemContext(ctx context.Context, organizationID string) context.Context {
	return auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		OrganizationID: organizationID,
		Capabilities:   auth.CapBypassFGA | auth.CapInternalOperation,
	})
}

// hostFromURL returns rawURL's host, falling back to rawURL unchanged if it doesn't parse
func hostFromURL(rawURL string) string {
	if parsed, err := url.Parse(rawURL); err == nil && parsed.Host != "" {
		return parsed.Host
	}

	return rawURL
}

// gatherDomainScanEnrichments gathers company profile, compliance, and DNS vendor data for
// every submitted scan concurrently, so enrichment overlaps with URL Scanner processing
// instead of waiting for it to complete. Each lookup is best-effort: a failure is logged and
// that scan's Enrichment is left zero-valued rather than failing the whole batch
func (r *Runtime) gatherDomainScanEnrichments(ctx context.Context, scans []url_scanner.ScanBulkNewResponse, forceRefresh bool) []domainscan.Enrichment {
	enrichments := make([]domainscan.Enrichment, len(scans))

	var g errgroup.Group

	for i, scan := range scans {
		g.Go(func() error {
			domain := hostFromURL(scan.URL)

			config, err := json.Marshal(cloudflare.DomainScanGatherEnrichment{
				Domain:       domain,
				ForceRefresh: forceRefresh,
			})
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("domain", domain).Msg("domain scan: failed encoding enrichment gather config")
				return nil
			}

			response, err := r.ExecuteRuntimeOperation(ctx, cloudflare.DefinitionID.ID(), cloudflare.DomainScanGatherEnrichmentOp.Name(), config)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("domain", domain).Msg("domain scan: failed gathering enrichment")
				return nil
			}

			var result cloudflare.DomainScanGatherEnrichmentResult
			if err := json.Unmarshal(response, &result); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("domain", domain).Msg("domain scan: failed decoding gathered enrichment")
				return nil
			}

			enrichments[i] = result.Enrichment

			return nil
		})
	}

	_ = g.Wait()

	return enrichments
}

// HandleDomainScanPoll processes one poll cycle for a submitted scan: re-emitting itself for
// another attempt while the scan is still processing, giving up after the attempt budget is
// exhausted, and finalizing the scan once ready
func (r *Runtime) HandleDomainScanPoll(ctx context.Context, envelope operations.DomainScanPollEnvelope) (bool, error) {
	ctx = logx.WithFields(ctx, logx.LogFields{
		"organization_id": envelope.OrganizationID,
		"scan_result_id":  envelope.ScanResultID,
		"attempt":         envelope.Attempt,
	},
	)

	config, err := json.Marshal(cloudflare.DomainScanPoll{
		ScanResultID: envelope.ScanResultID,
	})
	if err != nil {
		return true, river.JobCancel(err)
	}

	response, err := r.ExecuteRuntimeOperation(ctx, cloudflare.DefinitionID.ID(), cloudflare.DomainScanPollOp.Name(), config)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed polling cloudflare for scan result")
		r.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

		if notifyErr := r.maybeNotifyDomainScanGroup(ctx, envelope.OrganizationID, envelope.SiblingScanIDs); notifyErr != nil {
			logx.FromContext(ctx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after poll failure")
		}

		return true, river.JobCancel(err)
	}

	var result cloudflare.DomainScanPollResult
	if err := json.Unmarshal(response, &result); err != nil {
		r.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

		if notifyErr := r.maybeNotifyDomainScanGroup(ctx, envelope.OrganizationID, envelope.SiblingScanIDs); notifyErr != nil {
			logx.FromContext(ctx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after decode failure")
		}

		return true, river.JobCancel(err)
	}

	if len(result.TaskErrors) > 0 {
		taskErr := fmt.Errorf("%w: %s", errDomainScanTaskFailed, result.TaskErrors.Error())
		// logged as info because the report still completes from whatever enrichment was gathered
		logx.FromContext(ctx).Info().Err(taskErr).Msg("domain scan: cloudflare scan task failed, finalizing from enrichment alone")

		if err := r.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, envelope.SiblingScanIDs, nil); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed finalizing scan after task failure")
			return true, err
		}

		return true, river.JobCancel(taskErr)
	}

	if result.NotReady || !result.Result.Task.Success {
		if envelope.Attempt >= operations.DomainScanMaxAttempts {
			logx.FromContext(ctx).Warn().Msg("domain scan: max poll attempts reached, finalizing from enrichment alone")

			if err := r.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, envelope.SiblingScanIDs, nil); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed finalizing scan after max attempts")
				return true, err
			}

			return true, river.JobCancel(errDomainScanMaxAttemptsReached)
		}

		scheduledAt := time.Now().Add(operations.DomainScanPollBackoff(envelope.Attempt))

		receipt := r.Gala().EmitWithHeaders(ctx, operations.DomainScanPollTopic, operations.DomainScanPollEnvelope{
			OrganizationID: envelope.OrganizationID,
			ScanResultID:   envelope.ScanResultID,
			InternalScanID: envelope.InternalScanID,
			Attempt:        envelope.Attempt + 1,
			SiblingScanIDs: envelope.SiblingScanIDs,
		}, gala.Headers{ScheduledAt: &scheduledAt})
		if receipt.Err != nil {
			logx.FromContext(ctx).Error().Err(receipt.Err).Msg("domain scan: failed scheduling next poll cycle")
			return true, receipt.Err
		}

		logx.FromContext(ctx).Info().Msg("domain scan: result not ready, poll cycle scheduled")

		return false, nil
	}

	if err := r.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, envelope.SiblingScanIDs, &result); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed finalizing scan")
		return true, err
	}

	logx.FromContext(ctx).Info().Msg("domain scan: finalized successfully")

	return true, nil
}

// markDomainScanFailed marks a Scan record as failed when its poll cycle gives up.
// Best-effort: failures updating the record are logged but not propagated, since the caller is already
// returning the original poll failure
func (r *Runtime) markDomainScanFailed(ctx context.Context, organizationID, internalScanID string) {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	if err := r.DB().Scan.UpdateOneID(internalScanID).SetStatus(enums.ScanStatusFailed).Exec(systemCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("internal_scan_id", internalScanID).Msg("domain scan: failed marking scan as failed")
	}
}

// markDomainScansFailed marks every Scan record in scanIDs (keyed by domain) as failed, used
// when submission never reached Cloudflare or a domain fell out of Cloudflare's response
func (r *Runtime) markDomainScansFailed(ctx context.Context, organizationID string, scanIDs map[string]string) {
	for _, internalScanID := range scanIDs {
		r.markDomainScanFailed(ctx, organizationID, internalScanID)
	}
}

// finalizeDomainScan builds the structured scan report and marks the Scan record completed, then
// notifies the organization once every sibling in the group has finished. result is nil when the
// URL Scanner task itself never produced a usable result (it errored, or the poll budget was
// exhausted) - the report is still built from whatever enrichment data was already gathered, so
// the scan completes with partial data instead of a bare failure
func (r *Runtime) finalizeDomainScan(ctx context.Context, organizationID, internalScanID string, siblingScanIDs []string, result *cloudflare.DomainScanPollResult) error {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	scanRecord, err := r.DB().Scan.Get(systemCtx, internalScanID)
	if err != nil {
		return err
	}

	// enrichment was already gathered concurrently with URL Scanner submission and polling
	// (see gatherDomainScanEnrichments); round-trip it since ent's field.JSON decodes the
	// stored struct back as a map[string]any
	var enrichment domainscan.Enrichment
	if err := jsonx.RoundTrip(scanRecord.Metadata[domainScanEnrichmentMetadataKey], &enrichment); err != nil {
		return err
	}

	var resultJSON json.RawMessage

	if result != nil {
		resultJSON, err = json.Marshal(result.Result)
		if err != nil {
			return err
		}
	}

	config, err := json.Marshal(cloudflare.DomainScanBuildReport{
		InternalScanID: internalScanID,
		Result:         resultJSON,
		Enrichment:     enrichment,
	})
	if err != nil {
		return err
	}

	response, err := r.ExecuteRuntimeOperation(ctx, cloudflare.DefinitionID.ID(), cloudflare.DomainScanBuildReportOp.Name(), config)
	if err != nil {
		return err
	}

	var enriched cloudflare.DomainScanBuildReportResult
	if err := json.Unmarshal(response, &enriched); err != nil {
		return err
	}

	enriched.Data = vendorenrich.EnrichVendors(systemCtx, r.DB(), enriched.Data)

	if err := r.DB().Scan.UpdateOneID(internalScanID).
		SetStatus(enums.ScanStatusCompleted).
		SetMetadata(enriched.Data).
		Exec(systemCtx); err != nil {
		return err
	}

	return r.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs)
}

// maybeNotifyDomainScanGroup checks whether every sibling scan in siblingScanIDs (a single-element
// slice for a one-off scan) has reached a terminal state (completed or failed); if any is still
// processing it returns nil without doing anything, since it'll be called again when that one
// finishes. Once every sibling is terminal, it combines their reports and sends one Notification
// for the whole group, so a submission of N domains produces exactly one notification, not N
func (r *Runtime) maybeNotifyDomainScanGroup(ctx context.Context, organizationID string, siblingScanIDs []string) error {
	if organizationID == "" {
		return nil
	}

	systemCtx := domainScanSystemContext(ctx, organizationID)

	siblings, err := r.DB().Scan.Query().Where(scan.IDIn(siblingScanIDs...)).All(systemCtx)
	if err != nil {
		return err
	}

	for _, sibling := range siblings {
		if sibling.Status == enums.ScanStatusPending || sibling.Status == enums.ScanStatusProcessing {
			return nil
		}
	}

	results := make([]domainscan.DomainScanResult, 0, len(siblings))
	reports := make([]map[string]any, 0, len(siblings))

	var completed, failed int

	for _, sibling := range siblings {
		domainResult := domainscan.DomainScanResult{
			Domain:         sibling.Target,
			InternalScanID: sibling.ID,
			Status:         "failed",
		}

		if sibling.Status == enums.ScanStatusCompleted {
			domainResult.Status = "completed"
			completed++

			report := sibling.Metadata
			reports = append(reports, report)

			if externalScanID, ok := report["external_scan_id"].(string); ok {
				domainResult.ExternalScanID = externalScanID
			}

			if reportURL, ok := report["url"].(string); ok {
				domainResult.URL = reportURL
			}
		} else {
			failed++
		}

		results = append(results, domainResult)
	}

	merged := domainscan.MergeReports(results, reports)

	data, err := jsonx.ToMap(merged)
	if err != nil {
		return err
	}

	body := fmt.Sprintf("Scan completed for %d domain(s), see the results to import your detected vendors, findings, and more", completed)
	if failed > 0 {
		body = fmt.Sprintf("%s (%d domain(s) failed)", body, failed)
	}

	_, err = r.DB().Notification.Create().
		SetOwnerID(organizationID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType("scan.created").
		SetTitle("Domain scan completed").
		SetBody(body).
		SetData(data).
		SetTopic(enums.NotificationTopicDomainScan).
		Save(systemCtx)

	return err
}
