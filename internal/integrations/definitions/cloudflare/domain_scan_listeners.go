package cloudflare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
	"github.com/riverqueue/river"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"golang.org/x/sync/errgroup"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/vendorenrich"
	"github.com/theopenlane/core/v2/pkg/domainscan"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

// domainScanEnrichmentMetadataKey is the Scan.Metadata key holding gathered enrichment data
const domainScanEnrichmentMetadataKey = "enrichment"

// domainScanSaga orchestrates the durable domain scan flow: submit, poll, finalize
type domainScanSaga struct {
	// services exposes runtime execution, persistence, and event capabilities
	services types.RuntimeServices
}

type brandDesignScanOpts struct {
	organizationID            string
	scanID                    string
	domain                    string
	applyBrandDesignToPreview bool
	applyBrandDesignToLive    bool
}

// runBrandDesignScan runs only the brand design portion of the scan
func (s domainScanSaga) runBrandDesignScan(ctx context.Context, opts brandDesignScanOpts) error {
	systemCtx := domainScanSystemContext(ctx, opts.organizationID)

	if err := s.services.DB().Scan.UpdateOneID(opts.scanID).
		SetStatus(enums.ScanStatusProcessing).
		Exec(systemCtx); err != nil {
		return err
	}

	config, err := json.Marshal(DomainScanGatherEnrichment{
		Domain:          opts.domain,
		ForceRefresh:    true,
		BrandDesignOnly: true,
	})
	if err != nil {
		return err
	}

	response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanEnrichmentOp.Name(), config)
	if err != nil {
		return err
	}

	var result DomainScanGatherEnrichmentResult
	if err := json.Unmarshal(response, &result); err != nil {
		return err
	}

	if result.Enrichment.Branding == nil {
		logx.FromContext(ctx).Info().Msg("domain scan: no brand design found")

		err := s.services.DB().Scan.UpdateOneID(opts.scanID).
			SetStatus(enums.ScanStatusCompleted).
			Exec(systemCtx)
		if err != nil {
			return err
		}

		return s.sendBrandDesignNotification(ctx, opts)
	}

	if result.Enrichment.Branding.Error == "" {
		result.Enrichment.Branding.ApplyToPreviewTrustcenter = opts.applyBrandDesignToPreview
		result.Enrichment.Branding.ApplyToLiveTrustcenter = opts.applyBrandDesignToLive

		if _, err := applyBrandingToTrustCenter(systemCtx, s.services.DB(), *result.Enrichment.Branding); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed applying brand design to trust center")
		}
	}

	metadata := map[string]any{
		"url": opts.domain,
		"branding": domainscan.Branding{
			Error: result.Enrichment.Branding.Error,
			Favicon: domainscan.Favicon{
				URL: result.Enrichment.Branding.FaviconURL,
			},
			LogoURL:                  result.Enrichment.Branding.LogoURL,
			PrimaryColor:             result.Enrichment.Branding.PrimaryColor,
			Font:                     result.Enrichment.Branding.Font,
			ForegroundColor:          result.Enrichment.Branding.ForegroundColor,
			BackgroundColor:          result.Enrichment.Branding.BackgroundColor,
			AccentColor:              result.Enrichment.Branding.AccentColor,
			SecondaryBackgroundColor: result.Enrichment.Branding.SecondaryBackgroundColor,
			SecondaryForegroundColor: result.Enrichment.Branding.SecondaryForegroundColor,
		},
	}

	metadata[DomainScanApplyBrandDesignToPreviewMetadataKey] = opts.applyBrandDesignToPreview
	metadata[DomainScanApplyBrandDesignToLiveMetadataKey] = opts.applyBrandDesignToLive

	if err := s.services.DB().Scan.UpdateOneID(opts.scanID).
		SetStatus(enums.ScanStatusCompleted).
		SetMetadata(metadata).
		Exec(systemCtx); err != nil {
		return err
	}

	return s.sendBrandDesignNotification(ctx, opts)
}

func (s domainScanSaga) sendBrandDesignNotification(ctx context.Context, opts brandDesignScanOpts) error {
	if opts.organizationID == "" {
		return nil
	}

	var url string
	body := "We finished scanning your domain. Your branding is now ready to review and use."

	// if applying to any env, we should link to the branding page
	if opts.applyBrandDesignToPreview || opts.applyBrandDesignToLive {
		body = "We finished scanning your domain. You can now review the changes in your preview environment, then publish."
		url = fmt.Sprintf("%s/%s", entityops.ConsoleLanding(entityops.SchemaTrustCenter.String()), "branding")
	}

	result := &domainscan.Result{
		URL:            url,
		InternalScanID: opts.scanID,
	}
	data, err := jsonx.ToMap(result)
	if err != nil {
		return err
	}

	_, err = s.services.DB().Notification.Create().
		SetOwnerID(opts.organizationID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType("scan.created").
		SetTitle("Domain Branding is ready").
		SetBody(body).
		SetData(data).
		SetTopic(enums.NotificationTopicDomainScan).
		Save(domainScanSystemContext(ctx, opts.organizationID))

	return err
}

// domainScanListeners declares the standalone gala listeners implementing the domain scan saga
func domainScanListeners() types.GalaListenerRegistration {
	return types.GalaListenerRegistration{
		Name: "cloudflare.domainscan",
		Register: func(g *gala.Gala, services types.RuntimeServices) ([]gala.ListenerID, error) {
			saga := domainScanSaga{services: services}

			return gala.Register(g, gala.Definition[DomainScanPollEnvelope]{
				Topic: domainScanPollTopic,
				LogFields: func(envelope DomainScanPollEnvelope) map[string]any {
					return map[string]any{
						"organization_id":  envelope.OrganizationID,
						"scan_result_id":   envelope.ScanResultID,
						"internal_scan_id": envelope.InternalScanID,
						"attempt":          envelope.Attempt,
					}
				},
				Handle: func(hc gala.HandlerContext, envelope DomainScanPollEnvelope) error {
					_, err := saga.handlePoll(hc.Context, envelope)
					return err
				},
			})
		},
	}
}

// domainScanGroupSiblings returns every Scan ID sharing internalScanID's group, or itself alone
func (s domainScanSaga) domainScanGroupSiblings(ctx context.Context, organizationID, internalScanID string) ([]string, error) {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	scanRecord, err := s.services.DB().Scan.Get(systemCtx, internalScanID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("internal_scan_id", internalScanID).Msg("domain scan: failed fetching scan record for group resolution")

		return nil, err
	}

	groupID, ok := scanRecord.Metadata[DomainScanGroupMetadataKey].(string)
	if !ok || groupID == "" {
		return []string{internalScanID}, nil
	}

	return s.services.DB().Scan.Query().
		Where(
			scan.OwnerID(organizationID),
			func(sel *sql.Selector) {
				sel.Where(sqljson.ValueEQ(scan.FieldMetadata, groupID, sqljson.Path(DomainScanGroupMetadataKey)))
			},
		).
		IDs(systemCtx)
}

// submitAndScheduleDomainScan submits one Scan record to domain scanner and schedules its poll
func (s domainScanSaga) submitAndScheduleDomainScan(ctx context.Context, organizationID, scanID, domain string, forceRefresh bool) error {
	ctx = logx.WithFields(ctx, logx.LogFields{
		"organization_id": organizationID,
		"domain":          domain,
	})

	siblingScanIDs, err := s.domainScanGroupSiblings(ctx, organizationID, scanID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed resolving scan group, notifying for this scan alone")
		siblingScanIDs = []string{scanID}
	}

	return s.submitAndScheduleDomainScans(ctx, organizationID, map[string]string{domain: scanID}, forceRefresh, siblingScanIDs)
}

// submitAndScheduleDomainScans submits the scan records together and schedules a poll per scan
func (s domainScanSaga) submitAndScheduleDomainScans(ctx context.Context, organizationID string, scanIDs map[string]string, forceRefresh bool, siblingScanIDs []string) error {
	systemCtx := domainScanSystemContext(ctx, organizationID)

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
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed submitting scans to cloudflare, finalizing from enrichment alone")

		return errors.Join(err, s.finalizeDomainScansFromEnrichment(ctx, organizationID, scanIDs, forceRefresh, siblingScanIDs))
	}

	var result DomainScanSubmitResult
	if err := json.Unmarshal(response, &result); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed decoding submit result, finalizing from enrichment alone")

		return errors.Join(err, s.finalizeDomainScansFromEnrichment(ctx, organizationID, scanIDs, forceRefresh, siblingScanIDs))
	}

	acceptedDomains := lo.Map(result.Scans, func(scan url_scanner.ScanBulkNewResponse, _ int) string {
		return hostFromURL(scan.URL)
	})

	enrichments := s.gatherDomainScanEnrichments(ctx, acceptedDomains, forceRefresh)

	for i, scan := range result.Scans {
		domain := acceptedDomains[i]
		domainCtx := logx.WithFields(ctx, map[string]any{"domain": domain, "scan_id": scan.UUID})

		internalScanID, ok := scanIDs[domain]
		if !ok {
			logx.FromContext(domainCtx).Warn().Msg("domain scan: cloudflare returned an unexpected domain, skipping")

			continue
		}

		delete(scanIDs, domain)

		if err := s.persistDomainScanEnrichment(domainCtx, organizationID, internalScanID, enrichments[i], siblingScanIDs); err != nil {
			continue
		}

		if _, err := s.services.Gala().EmitWithHeaders(ctx, domainScanPollTopic.Name, DomainScanPollEnvelope{
			OrganizationID: organizationID,
			ScanResultID:   scan.UUID,
			InternalScanID: internalScanID,
			SiblingScanIDs: siblingScanIDs,
		}, gala.Headers{UniqueOnce: true}); err != nil {
			logx.FromContext(domainCtx).Error().Err(err).Msg("domain scan: failed scheduling poll cycle")
			s.markDomainScanFailed(domainCtx, organizationID, internalScanID)

			if notifyErr := s.maybeNotifyDomainScanGroup(domainCtx, organizationID, siblingScanIDs); notifyErr != nil {
				logx.FromContext(domainCtx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after poll scheduling failure")
			}
		}
	}

	if len(scanIDs) > 0 {
		logx.FromContext(ctx).Warn().Int("count", len(scanIDs)).Msg("domain scan: some domains were not submitted to cloudflare, finalizing from enrichment alone")

		if err := s.finalizeDomainScansFromEnrichment(ctx, organizationID, scanIDs, forceRefresh, siblingScanIDs); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed finalizing unsubmitted domains from enrichment")
		}
	}

	logx.FromContext(ctx).Info().Int("scan_count", len(result.Scans)).Msg("domain scan: submission jobs scheduled")

	return nil
}

// finalizeDomainScansFromEnrichment finalizes scans that will never get a URL Scanner result
func (s domainScanSaga) finalizeDomainScansFromEnrichment(ctx context.Context, organizationID string, scanIDs map[string]string, forceRefresh bool, siblingScanIDs []string) error {
	domains := lo.Keys(scanIDs)
	enrichments := s.gatherDomainScanEnrichments(ctx, domains, forceRefresh)

	var errs []error

	for i, domain := range domains {
		internalScanID := scanIDs[domain]
		domainCtx := logx.WithFields(ctx, map[string]any{"domain": domain})

		if err := s.persistDomainScanEnrichment(domainCtx, organizationID, internalScanID, enrichments[i], siblingScanIDs); err != nil {
			errs = append(errs, err)

			continue
		}

		if _, err := s.finalizeDomainScan(domainCtx, organizationID, internalScanID, siblingScanIDs, nil); err != nil {
			logx.FromContext(domainCtx).Error().Err(err).Msg("domain scan: failed finalizing scan from enrichment")
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// persistDomainScanEnrichment stores the enrichment on the scan record
func (s domainScanSaga) persistDomainScanEnrichment(ctx context.Context, organizationID, internalScanID string, enrichment domainscan.Enrichment, siblingScanIDs []string) error {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	scanRecord, err := s.services.DB().Scan.Get(systemCtx, internalScanID)
	if err != nil {
		s.markDomainScanFailed(ctx, organizationID, internalScanID)

		return err
	}

	applyBrandDesignToPreview, _ := scanRecord.Metadata[DomainScanApplyBrandDesignToPreviewMetadataKey].(bool)
	applyBrandDesignToLive, _ := scanRecord.Metadata[DomainScanApplyBrandDesignToLiveMetadataKey].(bool)

	if err := s.services.DB().Scan.UpdateOneID(internalScanID).
		SetMetadata(map[string]any{domainScanEnrichmentMetadataKey: enrichment}).
		Exec(systemCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed updating scan record with enrichment")
		s.markDomainScanFailed(ctx, organizationID, internalScanID)

		if notifyErr := s.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs); notifyErr != nil {
			logx.FromContext(ctx).Error().Err(notifyErr).Msg("domain scan: failed checking group completion after enrichment update failure")
		}

		return err
	}

	if enrichment.Branding == nil || enrichment.Branding.Error != "" {
		logx.FromContext(ctx).Info().Msg("domain scan: no brand design found, skipping trust center update")

		return nil
	}

	enrichment.Branding.ApplyToPreviewTrustcenter = applyBrandDesignToPreview
	enrichment.Branding.ApplyToLiveTrustcenter = applyBrandDesignToLive

	if _, err := applyBrandingToTrustCenter(systemCtx, s.services.DB(), *enrichment.Branding); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed applying brand design to trust center")
	}

	return nil
}

// domainScanSystemContext builds a context authorized to update Scan/Notification records
func domainScanSystemContext(ctx context.Context, organizationID string) context.Context {
	return auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		OrganizationID: organizationID,
		Capabilities:   auth.CapBypassFGA | auth.CapInternalOperation,
	})
}

// hostFromURL returns rawURL's host, falling back to rawURL unchanged if it doesn't parse
func hostFromURL(rawURL string) string {
	if parsed, err := urlx.Parse(rawURL); err == nil {
		return parsed.Host
	}

	return rawURL
}

// gatherDomainScanEnrichments gathers enrichment data for every domain concurrently
func (s domainScanSaga) gatherDomainScanEnrichments(ctx context.Context, domains []string, forceRefresh bool) []domainscan.Enrichment {
	enrichments := make([]domainscan.Enrichment, len(domains))

	var g errgroup.Group

	for i, domain := range domains {
		g.Go(func() error {
			domainCtx := logx.WithFields(ctx, map[string]any{"domain": domain})

			config, err := json.Marshal(DomainScanGatherEnrichment{
				Domain:       domain,
				ForceRefresh: forceRefresh,
			})
			if err != nil {
				logx.FromContext(domainCtx).Error().Err(err).Msg("domain scan: failed encoding enrichment gather config")

				return nil
			}

			response, err := s.services.ExecuteRuntimeOperation(domainCtx, DefinitionID.ID(), DomainScanEnrichmentOp.Name(), config)
			if err != nil {
				logx.FromContext(domainCtx).Error().Err(err).Msg("domain scan: failed gathering enrichment")

				return nil
			}

			var result DomainScanGatherEnrichmentResult
			if err := json.Unmarshal(response, &result); err != nil {
				logx.FromContext(domainCtx).Error().Err(err).Msg("domain scan: failed decoding gathered enrichment")

				return nil
			}

			enrichments[i] = result.Enrichment

			return nil
		})
	}

	_ = g.Wait()

	return enrichments
}

// handlePoll processes one poll cycle for a submitted scan
func (s domainScanSaga) handlePoll(ctx context.Context, envelope DomainScanPollEnvelope) (bool, error) {
	config, err := json.Marshal(DomainScanPoll{
		ScanResultID: envelope.ScanResultID,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed encoding poll config")

		return true, river.JobCancel(err)
	}

	response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanPollOp.Name(), config)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed polling cloudflare for scan result, finalizing from enrichment alone")

		return s.finalizeDomainScanWithoutResult(ctx, envelope, err)
	}

	var result DomainScanPollResult
	if err := json.Unmarshal(response, &result); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed decoding poll result, finalizing from enrichment alone")

		return s.finalizeDomainScanWithoutResult(ctx, envelope, err)
	}

	if len(result.TaskErrors) > 0 {
		taskErr := fmt.Errorf("%w: %s", ErrDomainScanTaskFailed, result.TaskErrors.Error())
		logx.FromContext(ctx).Info().Err(taskErr).Interface("task_errors", result.TaskErrors).Msg("domain scan: cloudflare scan task failed, finalizing from enrichment alone")

		return s.finalizeDomainScanWithoutResult(ctx, envelope, taskErr)
	}

	if result.NotReady || !result.Result.Task.Success {
		if envelope.Attempt >= DomainScanMaxAttempts {
			logx.FromContext(ctx).Warn().Msg("domain scan: max poll attempts reached, finalizing from enrichment alone")

			return s.finalizeDomainScanWithoutResult(ctx, envelope, ErrDomainScanMaxAttemptsReached)
		}

		scheduledAt := time.Now().Add(DomainScanPollBackoff(envelope.Attempt))

		if _, err := s.services.Gala().EmitWithHeaders(ctx, domainScanPollTopic.Name, DomainScanPollEnvelope{
			OrganizationID: envelope.OrganizationID,
			ScanResultID:   envelope.ScanResultID,
			InternalScanID: envelope.InternalScanID,
			Attempt:        envelope.Attempt + 1,
			SiblingScanIDs: envelope.SiblingScanIDs,
		}, gala.Headers{ScheduledAt: &scheduledAt, UniqueOnce: true}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed scheduling next poll cycle")

			return true, err
		}

		logx.FromContext(ctx).Info().Msg("domain scan: result not ready, poll cycle scheduled")

		return false, nil
	}

	if _, err := s.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, envelope.SiblingScanIDs, &result); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed finalizing scan")

		return true, err
	}

	logx.FromContext(ctx).Info().Msg("domain scan: finalized successfully")

	return true, nil
}

// finalizeDomainScanWithoutResult finalizes a scan whose URL Scanner result never became available
func (s domainScanSaga) finalizeDomainScanWithoutResult(ctx context.Context, envelope DomainScanPollEnvelope, cause error) (bool, error) {
	status, err := s.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, envelope.SiblingScanIDs, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan: failed finalizing scan without cloudflare result")

		return true, err
	}

	if status == enums.ScanStatusCompleted {
		logx.FromContext(ctx).Info().Msg("domain scan: finalized from enrichment without cloudflare result")

		return true, nil
	}

	return true, river.JobCancel(cause)
}

// markDomainScanFailed marks a Scan record as failed when its poll cycle gives up
func (s domainScanSaga) markDomainScanFailed(ctx context.Context, organizationID, internalScanID string) {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	if err := s.services.DB().Scan.UpdateOneID(internalScanID).SetStatus(enums.ScanStatusFailed).Exec(systemCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("internal_scan_id", internalScanID).Msg("domain scan: failed marking scan as failed")
	}
}

// finalizeDomainScan builds the scan report, stamps the terminal status, and notifies siblings
func (s domainScanSaga) finalizeDomainScan(ctx context.Context, organizationID, internalScanID string, siblingScanIDs []string, result *DomainScanPollResult) (enums.ScanStatus, error) {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	scanRecord, err := s.services.DB().Scan.Get(systemCtx, internalScanID)
	if err != nil {
		return enums.ScanStatusFailed, err
	}

	var enrichment domainscan.Enrichment
	if err := jsonx.RoundTrip(scanRecord.Metadata[domainScanEnrichmentMetadataKey], &enrichment); err != nil {
		return enums.ScanStatusFailed, err
	}

	status := enums.ScanStatusCompleted
	if result == nil && enrichment.IsEmpty() {
		status = enums.ScanStatusFailed
	}

	var resultJSON json.RawMessage

	if result != nil {
		resultJSON, err = json.Marshal(result.Result)
		if err != nil {
			return status, err
		}
	}

	config, err := json.Marshal(DomainScanBuildReport{
		InternalScanID: internalScanID,
		Result:         resultJSON,
		Enrichment:     enrichment,
	})
	if err != nil {
		return status, err
	}

	response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanBuildReportOp.Name(), config)
	if err != nil {
		return status, err
	}

	var enriched DomainScanBuildReportResult
	if err := json.Unmarshal(response, &enriched); err != nil {
		return status, err
	}

	enriched.Data = vendorenrich.EnrichVendors(systemCtx, s.services.DB(), enriched.Data)

	if err := s.services.DB().Scan.UpdateOneID(internalScanID).
		SetStatus(status).
		SetMetadata(enriched.Data).
		Exec(systemCtx); err != nil {
		return status, err
	}

	return status, s.maybeNotifyDomainScanGroup(ctx, organizationID, siblingScanIDs)
}

// maybeNotifyDomainScanGroup checks whether every sibling scan has reached a terminal state
func (s domainScanSaga) maybeNotifyDomainScanGroup(ctx context.Context, organizationID string, siblingScanIDs []string) error {
	if organizationID == "" {
		return nil
	}

	systemCtx := domainScanSystemContext(ctx, organizationID)

	siblings, err := s.services.DB().Scan.Query().Where(scan.IDIn(siblingScanIDs...), scan.OwnerID(organizationID)).All(systemCtx)
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
