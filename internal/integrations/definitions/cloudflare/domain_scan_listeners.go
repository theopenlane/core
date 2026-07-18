package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog"
	"github.com/theopenlane/iam/auth"
	"golang.org/x/sync/errgroup"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
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

			return append(createIDs, pollIDs...), nil
		},
	}
}

// handleCreate submits an organization's domains to Cloudflare's URL Scanner
func (s domainScanSaga) handleCreate(ctx context.Context, envelope operations.DomainScanCreateEnvelope) error {
	logger := logx.FromContext(ctx).With().
		Str("organization_id", envelope.OrganizationID).
		Strs("domains", envelope.Domains).
		Logger()

	config, err := json.Marshal(DomainScanSubmit{
		Domains: envelope.Domains,
	})
	if err != nil {
		return err
	}

	response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanSubmitOp.Name(), config)
	if err != nil {
		logger.Error().Err(err).Msg("domain scan: failed submitting scans to cloudflare")
		return err
	}

	var result DomainScanSubmitResult
	if err := json.Unmarshal(response, &result); err != nil {
		return err
	}

	enrichments := s.gatherDomainScanEnrichments(ctx, result.Scans, envelope.ForceRefresh, logger)

	systemCtx := domainScanSystemContext(ctx, envelope.OrganizationID)

	now, err := models.ToDateTime(time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	for i, scan := range result.Scans {
		scanRecord, err := s.services.DB().Scan.Create().
			SetOwnerID(envelope.OrganizationID).
			SetTarget(hostFromURL(scan.URL)).
			SetScanType(enums.ScanTypeDomain).
			SetScanDate(*now).
			SetPerformedBy("openlane_domain_scan").
			SetStatus(enums.ScanStatusProcessing).
			SetMetadata(map[string]any{domainScanEnrichmentMetadataKey: enrichments[i]}).
			Save(systemCtx)
		if err != nil {
			logger.Error().Err(err).Str("scan_id", scan.UUID).Msg("domain scan: failed creating scan record")
			return err
		}

		scheduledAt := time.Now().Add(operations.DomainScanInitialWait)

		receipt := s.services.Gala().EmitWithHeaders(ctx, operations.DomainScanPollTopic, operations.DomainScanPollEnvelope{
			OrganizationID: envelope.OrganizationID,
			ScanResultID:   scan.UUID,
			InternalScanID: scanRecord.ID,
		}, gala.Headers{ScheduledAt: &scheduledAt})
		if receipt.Err != nil {
			logger.Error().Err(receipt.Err).Str("scan_id", scan.UUID).Msg("domain scan: failed scheduling poll cycle")
			return receipt.Err
		}
	}

	logger.Info().Int("scan_count", len(result.Scans)).Msg("domain scan: submission jobs scheduled")

	return nil
}

// domainScanSystemContext builds a context authorized to create/update Scan and Notification
// records for organizationID on behalf of the system, bypassing org filtering and FGA
func domainScanSystemContext(ctx context.Context, organizationID string) context.Context {
	return auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		OrganizationID: organizationID,
		Capabilities:   auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
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
func (s domainScanSaga) gatherDomainScanEnrichments(ctx context.Context, scans []url_scanner.ScanBulkNewResponse, forceRefresh bool, logger zerolog.Logger) []domainscan.Enrichment {
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
				logger.Error().Err(err).Str("domain", domain).Msg("domain scan: failed encoding enrichment gather config")
				return nil
			}

			response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanGatherEnrichmentOp.Name(), config)
			if err != nil {
				logger.Error().Err(err).Str("domain", domain).Msg("domain scan: failed gathering enrichment")
				return nil
			}

			var result DomainScanGatherEnrichmentResult
			if err := json.Unmarshal(response, &result); err != nil {
				logger.Error().Err(err).Str("domain", domain).Msg("domain scan: failed decoding gathered enrichment")
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
	logger := logx.FromContext(ctx).With().
		Str("organization_id", envelope.OrganizationID).
		Str("scan_result_id", envelope.ScanResultID).
		Int("attempt", envelope.Attempt).
		Logger()

	config, err := json.Marshal(DomainScanPoll{
		ScanResultID: envelope.ScanResultID,
	})
	if err != nil {
		return true, river.JobCancel(err)
	}

	response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), DomainScanPollOp.Name(), config)
	if err != nil {
		logger.Error().Err(err).Msg("domain scan: failed polling cloudflare for scan result")
		return true, river.JobCancel(err)
	}

	var result DomainScanPollResult
	if err := json.Unmarshal(response, &result); err != nil {
		return true, river.JobCancel(err)
	}

	if len(result.TaskErrors) > 0 {
		taskErr := fmt.Errorf("%w: %s", ErrDomainScanTaskFailed, result.TaskErrors.Error())
		// logged as info because it could be a retryable error
		logger.Info().Err(taskErr).Msg("domain scan: cloudflare scan task failed")
		s.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

		return true, river.JobCancel(taskErr)
	}

	if !result.Result.Task.Success {
		if envelope.Attempt >= operations.DomainScanMaxAttempts {
			logger.Warn().Msg("domain scan: max poll attempts reached, giving up")
			s.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

			return true, river.JobCancel(ErrDomainScanMaxAttemptsReached)
		}

		scheduledAt := time.Now().Add(operations.DomainScanPollBackoff(envelope.Attempt))

		receipt := s.services.Gala().EmitWithHeaders(ctx, operations.DomainScanPollTopic, operations.DomainScanPollEnvelope{
			OrganizationID: envelope.OrganizationID,
			ScanResultID:   envelope.ScanResultID,
			InternalScanID: envelope.InternalScanID,
			Attempt:        envelope.Attempt + 1,
		}, gala.Headers{ScheduledAt: &scheduledAt})
		if receipt.Err != nil {
			logger.Error().Err(receipt.Err).Msg("domain scan: failed scheduling next poll cycle")
			return true, receipt.Err
		}

		logger.Info().Msg("domain scan: result not ready, poll cycle scheduled")

		return false, nil
	}

	if err := s.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, result); err != nil {
		logger.Error().Err(err).Msg("domain scan: failed finalizing scan")
		return true, err
	}

	logger.Info().Msg("domain scan: notification created successfully")

	return true, nil
}

// markDomainScanFailed marks a Scan record as failed when its poll cycle gives up.
// Best-effort: failures updating the record are logged but not propagated, since the caller is already
// returning the original poll failure
func (s domainScanSaga) markDomainScanFailed(ctx context.Context, organizationID, internalScanID string) {
	systemCtx := domainScanSystemContext(ctx, organizationID)

	if err := s.services.DB().Scan.UpdateOneID(internalScanID).SetStatus(enums.ScanStatusFailed).Exec(systemCtx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("internal_scan_id", internalScanID).Msg("domain scan: failed marking scan as failed")
	}
}

// finalizeDomainScan enriches the completed scan result and builds the structured scan report through
// the runtime operation, updates the Scan record (created in "processing" status by handleCreate
// to completed), then notifies the organization
func (s domainScanSaga) finalizeDomainScan(ctx context.Context, organizationID, internalScanID string, result DomainScanPollResult) error {
	domain := hostFromURL(result.Result.Task.URL)

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

	resultJSON, err := json.Marshal(result.Result)
	if err != nil {
		return err
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

	if err := s.services.DB().Scan.UpdateOneID(internalScanID).
		SetStatus(enums.ScanStatusCompleted).
		SetMetadata(enriched.Data).
		Exec(systemCtx); err != nil {
		return err
	}

	_, err = s.services.DB().Notification.Create().
		SetOwnerID(organizationID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType("scan.created").
		SetTitle("Domain scan completed").
		SetBody(fmt.Sprintf("Scan completed successfully for %s, see the results to import your detected vendors, findings, and more", domain)).
		SetData(enriched.Data).
		SetTopic(enums.NotificationTopicDomainScan).
		Save(systemCtx)

	return err
}
