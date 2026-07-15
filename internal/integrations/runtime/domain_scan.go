package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/riverqueue/river"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// errDomainScanTaskFailed is returned when Cloudflare reports the scan task itself failed
var errDomainScanTaskFailed = errors.New("domain scan: cloudflare scan task failed")

// errDomainScanMaxAttemptsReached is returned when a scan never completes within the poll budget
var errDomainScanMaxAttemptsReached = errors.New("domain scan: max poll attempts reached")

// domainScanForceRefreshMetadataKey is the Scan.Metadata key used to carry the ForceRefresh flag
const domainScanForceRefreshMetadataKey = "force_refresh"

// HandleDomainScanCreate submits an organization's domains to Cloudflare's URL Scanner
func (r *Runtime) HandleDomainScanCreate(ctx context.Context, envelope operations.DomainScanCreateEnvelope) error {
	logger := logx.FromContext(ctx).With().
		Str("organization_id", envelope.OrganizationID).
		Strs("domains", envelope.Domains).
		Logger()

	config, err := json.Marshal(cloudflare.DomainScanSubmit{
		Domains: envelope.Domains,
	})
	if err != nil {
		return err
	}

	response, err := r.ExecuteRuntimeOperation(ctx, cloudflare.DefinitionID.ID(), cloudflare.DomainScanSubmitOp.Name(), config)
	if err != nil {
		logger.Error().Err(err).Msg("domain scan: failed submitting scans to cloudflare")
		return err
	}

	var result cloudflare.DomainScanSubmitResult
	if err := json.Unmarshal(response, &result); err != nil {
		return err
	}

	systemCtx := domainScanSystemContext(ctx, envelope.OrganizationID)

	now, err := models.ToDateTime(time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	for _, scan := range result.Scans {
		scanRecord, err := r.DB().Scan.Create().
			SetOwnerID(envelope.OrganizationID).
			SetTarget(hostFromURL(scan.URL)).
			SetScanType(enums.ScanTypeDomain).
			SetScanDate(*now).
			SetPerformedBy("openlane_domain_scan").
			SetStatus(enums.ScanStatusProcessing).
			SetMetadata(map[string]any{domainScanForceRefreshMetadataKey: envelope.ForceRefresh}).
			Save(systemCtx)
		if err != nil {
			logger.Error().Err(err).Str("scan_id", scan.UUID).Msg("domain scan: failed creating scan record")
			return err
		}

		scheduledAt := time.Now().Add(operations.DomainScanInitialWait)

		receipt := r.Gala().EmitWithHeaders(ctx, operations.DomainScanPollTopic, operations.DomainScanPollEnvelope{
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

// HandleDomainScanPoll processes one poll cycle for a submitted scan: re-emitting itself for
// another attempt while the scan is still processing, giving up after the attempt budget is
// exhausted, and finalizing the scan once ready
func (r *Runtime) HandleDomainScanPoll(ctx context.Context, envelope operations.DomainScanPollEnvelope) (bool, error) {
	logger := logx.FromContext(ctx).With().
		Str("organization_id", envelope.OrganizationID).
		Str("scan_result_id", envelope.ScanResultID).
		Int("attempt", envelope.Attempt).
		Logger()

	config, err := json.Marshal(cloudflare.DomainScanPoll{
		ScanResultID: envelope.ScanResultID,
	})
	if err != nil {
		return true, river.JobCancel(err)
	}

	response, err := r.ExecuteRuntimeOperation(ctx, cloudflare.DefinitionID.ID(), cloudflare.DomainScanPollOp.Name(), config)
	if err != nil {
		logger.Error().Err(err).Msg("domain scan: failed polling cloudflare for scan result")
		return true, river.JobCancel(err)
	}

	var result cloudflare.DomainScanPollResult
	if err := json.Unmarshal(response, &result); err != nil {
		return true, river.JobCancel(err)
	}

	if len(result.TaskErrors) > 0 {
		taskErr := fmt.Errorf("%w: %s", errDomainScanTaskFailed, result.TaskErrors.Error())
		// logged as info because it could be a retryable error
		logger.Info().Err(taskErr).Msg("domain scan: cloudflare scan task failed")
		r.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

		return true, river.JobCancel(taskErr)
	}

	if !result.Result.Task.Success {
		if envelope.Attempt >= operations.DomainScanMaxAttempts {
			logger.Warn().Msg("domain scan: max poll attempts reached, giving up")
			r.markDomainScanFailed(ctx, envelope.OrganizationID, envelope.InternalScanID)

			return true, river.JobCancel(errDomainScanMaxAttemptsReached)
		}

		scheduledAt := time.Now().Add(operations.DomainScanPollBackoff(envelope.Attempt))

		receipt := r.Gala().EmitWithHeaders(ctx, operations.DomainScanPollTopic, operations.DomainScanPollEnvelope{
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

	if err := r.finalizeDomainScan(ctx, envelope.OrganizationID, envelope.InternalScanID, result); err != nil {
		logger.Error().Err(err).Msg("domain scan: failed finalizing scan")
		return true, err
	}

	logger.Info().Msg("domain scan: notification created successfully")

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

// finalizeDomainScan enriches the completed scan result and builds the structured scan report through
// ExecuteRuntimeOperation, updates the Scan record (created in "processing" status by HandleDomainScanCreate
// to completed), then notifies the organization
func (r *Runtime) finalizeDomainScan(ctx context.Context, organizationID, internalScanID string, result cloudflare.DomainScanPollResult) error {
	domain := hostFromURL(result.Result.Task.URL)

	systemCtx := domainScanSystemContext(ctx, organizationID)

	scanRecord, err := r.DB().Scan.Get(systemCtx, internalScanID)
	if err != nil {
		return err
	}

	forceRefresh, _ := scanRecord.Metadata[domainScanForceRefreshMetadataKey].(bool)

	resultJSON, err := json.Marshal(result.Result)
	if err != nil {
		return err
	}

	config, err := json.Marshal(cloudflare.DomainScanEnrich{
		Domain:         domain,
		InternalScanID: internalScanID,
		Result:         resultJSON,
		ForceRefresh:   forceRefresh,
	})
	if err != nil {
		return err
	}

	response, err := r.ExecuteRuntimeOperation(ctx, cloudflare.DefinitionID.ID(), cloudflare.DomainScanEnrichOp.Name(), config)
	if err != nil {
		return err
	}

	var enriched cloudflare.DomainScanEnrichResult
	if err := json.Unmarshal(response, &enriched); err != nil {
		return err
	}

	if err := r.DB().Scan.UpdateOneID(internalScanID).
		SetStatus(enums.ScanStatusCompleted).
		SetMetadata(enriched.Data).
		Exec(systemCtx); err != nil {
		return err
	}

	_, err = r.DB().Notification.Create().
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
