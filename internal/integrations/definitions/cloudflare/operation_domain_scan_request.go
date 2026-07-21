package cloudflare

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/scan"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// DomainScanRequest queues a domain scan for a single domain by creating a pending Scan record which
// is then picked up to do the domain scan
type DomainScanRequest struct {
	// OrganizationID is the organization the scan belongs to, only used when dispatched without an  Integration
	// customer-facing calls always derive the organization from their resolved Integration instead, ignoring this field
	OrganizationID string `json:"organizationId,omitempty"`
	// Domain is the domain to scan
	Domain string `json:"domain" jsonschema:"required,title=Domain,description=Domain to scan"`
	// ForceRefresh bypasses Cloudflare's Browser Rendering cache, forcing a fresh render
	// instead of reusing one from a previous scan of the same domain
	ForceRefresh bool `json:"forceRefresh,omitempty" jsonschema:"title=Force Refresh,description=Bypass the render cache and force a fresh scan"`
	// GroupID links this scan to sibling scans requested together so they can be recombined into a
	// single notification once the whole group finishes
	GroupID string `json:"groupId,omitempty"`
}

// DomainScanRequestResult acknowledges that a domain scan was queued or run
type DomainScanRequestResult struct {
	// Message describes what happened
	Message string `json:"message"`
	// ScanID is the id of the Scan record for this request
	ScanID string `json:"scanId"`
}

// Handle adapts DomainScanRequest to the generic operation registration boundary
func (d DomainScanRequest) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		var cfg DomainScanRequest
		if err := json.Unmarshal(request.Config, &cfg); err != nil {
			return nil, ErrOperationConfigInvalid
		}

		organizationID := cfg.OrganizationID
		groupID := cfg.GroupID

		if request.Integration != nil {
			organizationID = request.Integration.OwnerID
			groupID = ""
		}

		if organizationID == "" {
			return nil, ErrInstallationRequired
		}

		scanRecord, err := request.DB.Scan.Query().
			Where(
				scan.OwnerID(organizationID),
				scan.Target(cfg.Domain),
				scan.ScanTypeEQ(enums.ScanTypeDomain),
				scan.PerformedBy(DomainScanPerformedBy),
				scan.StatusEQ(enums.ScanStatusPending),
			).
			First(ctx)
		if err != nil && !generated.IsNotFound(err) {
			return nil, err
		}

		if scanRecord == nil {
			metadata := map[string]any{"forceRefresh": cfg.ForceRefresh}
			if groupID != "" {
				metadata[DomainScanGroupMetadataKey] = groupID
			}

			scanRecord, err = request.DB.Scan.Create().
				SetOwnerID(organizationID).
				SetTarget(cfg.Domain).
				SetScanType(enums.ScanTypeDomain).
				SetPerformedBy(DomainScanPerformedBy).
				SetStatus(enums.ScanStatusPending).
				SetMetadata(metadata).
				Save(ctx)
			if err != nil {
				return nil, err
			}
		} else if groupID != "" {
			scanRecord, err = scanRecord.Update().SetMetadata(map[string]any{DomainScanGroupMetadataKey: groupID}).Save(ctx)
			if err != nil {
				return nil, err
			}
		}

		if request.Integration != nil {
			return providerkit.EncodeResult(DomainScanRequestResult{
				Message: "domain scan queued",
				ScanID:  scanRecord.ID,
			}, ErrResultEncode)
		}

		saga := domainScanSaga{services: request.Services}

		if err := saga.submitAndScheduleDomainScan(ctx, organizationID, scanRecord.ID, cfg.Domain, cfg.ForceRefresh); err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(DomainScanRequestResult{
			Message: "domain scan submitted",
			ScanID:  scanRecord.ID,
		}, ErrResultEncode)
	}
}
