package cloudflare

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
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/scan"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/vendorenrich"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/domainscan"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// domainScanEnrichmentMetadataKey is the Scan.Metadata key used to carry the enrichment data
// gathered concurrently with URL Scanner submission and polling, until the poll cycle
// finishes and needs it to build the final report
const domainScanEnrichmentMetadataKey = "enrichment"

// domainScanSaga orchestrates the durable domain scan flow: submit, poll, finalize
type domainScanSaga struct {
	// services exposes runtime execution, persistence, and event capabilities
	services types.RuntimeServices
}

// domainScanListeners declares the standalone gala listeners implementing the domain scan saga
func domainScanListeners() types.GalaListenerRegistration {
	return types.GalaListenerRegistration{
		Name: "cloudflare.domainscan",
		Register: func(registry *gala.Registry, services types.RuntimeServices) ([]gala.ListenerID, error) {
			saga := domainScanSaga{services: services}

			createIDs, err := gala.RegisterListeners(registry, gala.Definition[operations.DomainScanCreateEnvelope]{
				Topic: gala.Topic[operations.DomainScanCreateEnvelope]{Name: operations.DomainScanCreateTopic},
				Name:  operations.DomainScanCreateListenerName,
				Handle: func(hc gala.HandlerContext, envelope operations.DomainScanCreateEnvelope) error {
					return saga.handleCreate(hc.Context, envelope)
				},
			})
			if err != nil {
				return nil, err
			}

			pollIDs, err := gala.RegisterListeners(registry, gala.Definition[operations.DomainScanPollEnvelope]{
				Topic: gala.Topic[operations.DomainScanPollEnvelope]{Name: operations.DomainScanPollTopic},
				Name:  operations.DomainScanPollListenerName,
				Handle: func(hc gala.HandlerContext, envelope operations.DomainScanPollEnvelope) error {
					_, err := saga.handlePoll(hc.Context, envelope)
					return err
				},
			})
			if err != nil {
				return nil, err
			}

			submitExistingIDs, err := gala.RegisterListeners(registry, gala.Definition[operations.DomainScanSubmitEnvelope]{
				Topic: gala.Topic[operations.DomainScanSubmitEnvelope]{Name: operations.DomainScanSubmitTopic},
				Name:  operations.DomainScanSubmitListenerName,
				Handle: func(hc gala.HandlerContext, envelope operations.DomainScanSubmitEnvelope) error {
					return saga.submitAndScheduleDomainScan(hc.Context, envelope.OrganizationID, envelope.ScanID, envelope.Domain, envelope.ForceRefresh)
				},
			})
			if err != nil {
				return nil, err
			}

			importIDs, err := gala.RegisterListeners(registry, gala.Definition[operations.ImportDomainScanReviewEnvelope]{
				Topic: gala.Topic[operations.ImportDomainScanReviewEnvelope]{Name: operations.DomainScanImportTopic},
				Name:  operations.DomainScanImportListenerName,
				Handle: func(hc gala.HandlerContext, envelope operations.ImportDomainScanReviewEnvelope) error {
					return saga.HandleImportDomainScanReview(hc.Context, envelope)
				},
			})
			if err != nil {
				return nil, err
			}

			return append(append(append(createIDs, pollIDs...), submitExistingIDs...), importIDs...), nil
		},
	}
}

// handleCreate submits an organization's domains to Cloudflare's URL Scanner
func (s domainScanSaga) handleCreate(ctx context.Context, envelope operations.DomainScanCreateEnvelope) error {
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
		// a retry of this same job (e.g. after a transient failure) re-runs from the top, so reuse
		// a still-pending scan from an earlier attempt instead of creating another one for it
		scanRecord, err := s.services.DB().Scan.Query().
			Where(
				scan.OwnerID(envelope.OrganizationID),
				scan.Target(domain),
				scan.ScanTypeEQ(enums.ScanTypeDomain),
				scan.PerformedBy(operations.DomainScanPerformedBy),
				scan.StatusEQ(enums.ScanStatusPending),
			).
			First(createCtx)
		if err != nil && !generated.IsNotFound(err) {
			logx.FromContext(ctx).Error().Err(err).Str("domain", domain).Msg("domain scan: failed checking for existing pending scan record")
			return err
		}

		if scanRecord == nil {
			scanRecord, err = s.services.DB().Scan.Create().
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
		}

		scanIDs[domain] = scanRecord.ID
	}

	return s.submitAndScheduleDomainScans(ctx, envelope.OrganizationID, scanIDs, envelope.ForceRefresh)
}

// submitAndScheduleDomainScan submits a single already-created Scan record to domain scanner and schedules its poll cycle
func (s domainScanSaga) submitAndScheduleDomainScan(ctx context.Context, organizationID, scanID, domain string, forceRefresh bool) error {
	ctx = logx.WithFields(ctx, logx.LogFields{
		"organization_id": organizationID,
		"domain":          domain,
	})

	return s.submitAndScheduleDomainScans(ctx, organizationID, map[string]string{domain: scanID}, forceRefresh)
}

// submitAndScheduleDomainScans submits the scan records in scanID to the domain scan together and schedules a poll cycle for each.
// Any domain that isn't returned marked failed rather than left stuck in "processing" forever
func (s domainScanSaga) submitAndScheduleDomainScans(ctx context.Context, organizationID string, scanIDs map[string]string, forceRefresh bool) error {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	// snapshotted before the loop below starts deleting entries, so every sibling (including
	// this one) knows the full group to check against once it's this one's turn to finish
	siblingScanIDs := slices.Collect(maps.Values(scanIDs))

	// mark every scan processing as soon as this job actually starts working, rather than
	// leaving them in "pending" (indistinguishable from "not yet picked up") for however long
	// the Cloudflare submission and enrichment gathering below take
	if err := s.services.DB().Scan.Update().
		Where(scan.IDIn(siblingScanIDs...)).
		SetStatus(enums.ScanStatusProcessing).
		Exec(systemCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed marking scans processing before submission")
	}

	domains := make([]string, 0, len(scanIDs))
	for domain := range scanIDs {
		domains = append(domains, domain)
	}

	config, err := json.Marshal(DomainScanSubmit{
		Domains: domains,
	})
	if err != nil {
		return err
	}

	response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanSubmitOp.Name(), config)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed submitting scans to cloudflare")
		s.markDomainScansFailed(ctx, organizationID, scanIDs)

		return errors.Join(err, s.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs))
	}

	var result DomainScanSubmitResult
	if err := json.Unmarshal(response, &result); err != nil {
		s.markDomainScansFailed(ctx, organizationID, scanIDs)
		return errors.Join(err, s.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs))
	}

	enrichments := s.gatherDomainScanEnrichments(ctx, result.Scans, forceRefresh)

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

		if err := s.services.DB().Scan.UpdateOneID(internalScanID).
			SetMetadata(map[string]any{domainScanEnrichmentMetadataKey: enrichments[i]}).
			Exec(systemCtx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("scan_id", scan.UUID).Msg("domain scan: failed updating scan record with enrichment")
			s.markDomainScanFailed(ctx, organizationID, internalScanID)

			if notifyErr := s.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs); notifyErr != nil {
				logx.FromContext(ctx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after enrichment update failure")
			}

			continue
		}

		receipt := s.services.Gala().EmitWithHeaders(ctx, operations.DomainScanPollTopic, operations.DomainScanPollEnvelope{
			OrganizationID: organizationID,
			ScanResultID:   scan.UUID,
			InternalScanID: internalScanID,
			SiblingScanIDs: siblingScanIDs,
		}, gala.Headers{})
		if receipt.Err != nil {
			logx.FromContext(ctx).Error().Err(receipt.Err).Str("scan_id", scan.UUID).Msg("domain scan: failed scheduling poll cycle")
			s.markDomainScanFailed(ctx, organizationID, internalScanID)

			if notifyErr := s.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs); notifyErr != nil {
				logx.FromContext(ctx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after poll scheduling failure")
			}
		}
	}

	// any domain cloudflare didn't return a scan for is left with a Scan record that will never be polled, so mark it failed rather than
	// leaving it stuck forever
	if len(scanIDs) > 0 {
		logx.FromContext(ctx).Warn().Int("count", len(scanIDs)).Msg("domain scan: some domains were not submitted to cloudflare, marking as failed")
		s.markDomainScansFailed(ctx, organizationID, scanIDs)

		if err := s.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs); err != nil {
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
func (s domainScanSaga) gatherDomainScanEnrichments(ctx context.Context, scans []url_scanner.ScanBulkNewResponse, forceRefresh bool) []domainscan.Enrichment {
	enrichments := make([]domainscan.Enrichment, len(scans))

	var g errgroup.Group

	for i, scan := range scans {
		g.Go(func() error {
			domain := hostFromURL(scan.URL)

			config, err := json.Marshal(DomainScanGatherEnrichment{
				Domain:       domain,
				ForceRefresh: forceRefresh,
			})
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("domain", domain).Msg("domain scan: failed encoding enrichment gather config")
				return nil
			}

			response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanEnrichmentOp.Name(), config)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("domain", domain).Msg("domain scan: failed gathering enrichment")
				return nil
			}

			var result DomainScanGatherEnrichmentResult
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

// handlePoll processes one poll cycle for a submitted scan: re-emitting itself for
// another attempt while the scan is still processing, giving up after the attempt budget is
// exhausted, and finalizing the scan once ready
func (s domainScanSaga) handlePoll(ctx context.Context, envelope operations.DomainScanPollEnvelope) (bool, error) {
	ctx = logx.WithFields(ctx, logx.LogFields{
		"organization_id": envelope.OrganizationID,
		"scan_result_id":  envelope.ScanResultID,
		"attempt":         envelope.Attempt,
	},
	)

	config, err := json.Marshal(DomainScanPoll{
		ScanResultID: envelope.ScanResultID,
	})
	if err != nil {
		return true, river.JobCancel(err)
	}

	response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanPollOp.Name(), config)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed polling cloudflare for scan result")

		if notifyErr := s.maybeNotifyDomainScanGroup(ctx, envelope.OrganizationID, envelope.SiblingScanIDs); notifyErr != nil {
			logx.FromContext(ctx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after poll failure")
		}

		s.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

		return true, river.JobCancel(err)
	}

	var result DomainScanPollResult
	if err := json.Unmarshal(response, &result); err != nil {
		s.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

		if notifyErr := s.maybeNotifyDomainScanGroup(ctx, envelope.OrganizationID, envelope.SiblingScanIDs); notifyErr != nil {
			logx.FromContext(ctx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after decode failure")
		}

		return true, river.JobCancel(err)
	}

	if len(result.TaskErrors) > 0 {
		taskErr := fmt.Errorf("%w: %s", ErrDomainScanTaskFailed, result.TaskErrors.Error())
		// logged as info because the report still completes from whatever enrichment was gathered
		logx.FromContext(ctx).Info().Err(taskErr).Msg("domain scan: cloudflare scan task failed, finalizing from enrichment alone")

		if err := s.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, envelope.SiblingScanIDs, nil); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed finalizing scan after task failure")
			return true, err
		}

		s.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

		return true, river.JobCancel(taskErr)
	}

	if result.NotReady || !result.Result.Task.Success {
		if envelope.Attempt >= operations.DomainScanMaxAttempts {
			logx.FromContext(ctx).Warn().Msg("domain scan: max poll attempts reached, finalizing from enrichment alone")
			s.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

			if err := s.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, envelope.SiblingScanIDs, nil); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed finalizing scan after max attempts")
				return true, err
			}

			return true, river.JobCancel(ErrDomainScanMaxAttemptsReached)
		}

		scheduledAt := time.Now().Add(operations.DomainScanPollBackoff(envelope.Attempt))

		receipt := s.services.Gala().EmitWithHeaders(ctx, operations.DomainScanPollTopic, operations.DomainScanPollEnvelope{
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

	if err := s.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, envelope.SiblingScanIDs, &result); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed finalizing scan")
		return true, err
	}

	logx.FromContext(ctx).Info().Msg("domain scan: finalized successfully")

	return true, nil
}

// markDomainScanFailed marks a Scan record as failed when its poll cycle gives up
func (s domainScanSaga) markDomainScanFailed(ctx context.Context, organizationID, internalScanID string) {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	if err := s.services.DB().Scan.UpdateOneID(internalScanID).SetStatus(enums.ScanStatusFailed).Exec(systemCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("internal_scan_id", internalScanID).Msg("domain scan: failed marking scan as failed")
	}
}

// markDomainScansFailed marks every Scan record in scanIDs as failed
func (s domainScanSaga) markDomainScansFailed(ctx context.Context, organizationID string, scanIDs map[string]string) {
	for _, internalScanID := range scanIDs {
		s.markDomainScanFailed(ctx, organizationID, internalScanID)
	}
}

// finalizeDomainScan builds the structured scan report and marks the Scan record completed, then
// notifies the organization once every sibling in the group has finished. result is nil when the
// URL Scanner task itself never produced a usable result (it errored, or the poll budget was
// exhausted) - the report is still built from whatever enrichment data was already gathered, so
// the scan completes with partial data instead of a bare failure
func (s domainScanSaga) finalizeDomainScan(ctx context.Context, organizationID, internalScanID string, siblingScanIDs []string, result *DomainScanPollResult) error {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	scanRecord, err := s.services.DB().Scan.Get(systemCtx, internalScanID)
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

	config, err := json.Marshal(DomainScanBuildReport{
		InternalScanID: internalScanID,
		Result:         resultJSON,
		Enrichment:     enrichment,
	})
	if err != nil {
		return err
	}

	response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanBuildReportOp.Name(), config)
	if err != nil {
		return err
	}

	var enriched DomainScanBuildReportResult
	if err := json.Unmarshal(response, &enriched); err != nil {
		return err
	}

	enriched.Data = vendorenrich.EnrichVendors(systemCtx, s.services.DB(), enriched.Data)

	if err := s.services.DB().Scan.UpdateOneID(internalScanID).
		SetStatus(enums.ScanStatusCompleted).
		SetMetadata(enriched.Data).
		Exec(systemCtx); err != nil {
		return err
	}

	return s.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs)
}

// maybeNotifyDomainScanGroup checks whether every sibling scan in siblingScanIDs (a single-element
// slice for a one-off scan) has reached a terminal state (completed or failed); if any is still
// processing it returns nil without doing anything, since it'll be called again when that one
// finishes. Once every sibling is terminal, it combines their reports and sends one Notification
// for the whole group, so a submission of N domains produces exactly one notification, not N
func (s domainScanSaga) maybeNotifyDomainScanGroup(ctx context.Context, organizationID string, siblingScanIDs []string) error {
	if organizationID == "" {
		return nil
	}

	systemCtx := domainScanSystemContext(ctx, organizationID)

	siblings, err := s.services.DB().Scan.Query().Where(scan.IDIn(siblingScanIDs...)).All(systemCtx)
	if err != nil {
		return err
	}

	for _, sibling := range siblings {
		if sibling.Status == enums.ScanStatusPending || sibling.Status == enums.ScanStatusProcessing {
			return nil
		}
	}

	results := make([]domainscan.Result, 0, len(siblings))
	reports := make([]domainscan.ScanReportInput, 0, len(siblings))

	var completed, failed int

	for _, sibling := range siblings {
		domainResult := domainscan.Result{
			Domain:         sibling.Target,
			InternalScanID: sibling.ID,
			Status:         "failed",
		}

		if sibling.Status == enums.ScanStatusCompleted {
			domainResult.Status = "completed"
			completed++

			var report domainscan.ScanReport
			if err := jsonx.RoundTrip(sibling.Metadata, &report); err != nil {
				return err
			}

			reports = append(reports, domainscan.ScanReportInput{Domain: domainResult.Domain, Report: report})

			domainResult.ExternalScanID = report.ExternalScanID
			domainResult.URL = report.URL
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

	_, err = s.services.DB().Notification.Create().
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
