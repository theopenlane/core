package cloudflare

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// DomainScanRequest queues a domain scan for a single domain by creating a pending Scan record which
// is then picked up to do the domain scan
type DomainScanRequest struct {
	// ScanID identifies the Scan record that triggered an internally dispatched request
	// this is needed beceause domain_scan previously just filtered the scans table by domain
	// and selected the first one. but this may end up picking out an older record
	// if no id is provided, default to picking out the first one as before ( coming in from onboarding/domain scan)
	ScanID string `json:"scanId,omitempty"`
	// OrganizationID is the organization the scan belongs to, only used when dispatched without an  Integration
	// customer-facing calls always derive the organization from their resolved Integration instead, ignoring this field
	OrganizationID string `json:"organizationId,omitempty"`
	// Domain is the domain to scan
	Domain string `json:"domain" jsonschema:"required,title=Domain,description=Domain to scan"`
	// ForceRefresh bypasses Cloudflare's Browser Rendering cache, forcing a fresh render
	// instead of reusing one from a previous scan of the same domain
	ForceRefresh bool `json:"forceRefresh,omitempty" jsonschema:"title=Force Refresh,description=Bypass the render cache and force a fresh scan"`
	// BrandDesignOnly extracts the brand design without running the full domain scan
	BrandDesignOnly bool `json:"brandDesignOnly,omitempty" jsonschema:"title=Brand Design Only,description=Extract and apply the brand design without building a full domain scan report"`
	// ApplyBrandDesign applies extracted brand design to Trust Center settings
	ApplyBrandDesign bool `json:"applyBrandDesign,omitempty" jsonschema:"title=Apply Brand Design,description=Apply extracted brand design to Trust Center settings"`
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

		var scanRecord *generated.Scan
		var err error

		if cfg.ScanID != "" {

			// if scan id exists, we need to make sure the domain matches what we expect
			// and is also a candidate for scanning
			scanRecord, err = request.DB.Scan.Query().Where(
				scan.ID(cfg.ScanID),
				scan.OwnerID(organizationID),
				scan.Target(cfg.Domain),
				scan.ScanTypeEQ(enums.ScanTypeDomain),
				scan.PerformedBy(DomainScanPerformedBy),
			).Only(ctx)
			if err != nil {
				return nil, err
			}

			if scanRecord.Status != enums.ScanStatusPending && scanRecord.Status != enums.ScanStatusProcessing {
				return providerkit.EncodeResult(DomainScanRequestResult{
					Message: "domain scan already processed",
					ScanID:  scanRecord.ID,
				}, ErrResultEncode)
			}

		} else {

			scanRecord, err = request.DB.Scan.Query().
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
		}

		if scanRecord == nil {
			metadata := map[string]any{"forceRefresh": cfg.ForceRefresh}
			if cfg.BrandDesignOnly {
				metadata[DomainScanBrandDesignOnlyMetadataKey] = true
			}
			if cfg.ApplyBrandDesign {
				metadata[DomainScanApplyBrandDesignMetadataKey] = true
			}
			if groupID != "" {
				metadata[DomainScanGroupMetadataKey] = groupID
			}

			// on the internal path (Integration == nil) we're about to run the saga ourselves
			// below, so skip the event emissions
			createCtx := ctx
			if request.Integration == nil {
				createCtx = entityops.WithEmissionVetoed(ctx)
			}

			scanRecord, err = request.DB.Scan.Create().
				SetOwnerID(organizationID).
				SetTarget(cfg.Domain).
				SetScanType(enums.ScanTypeDomain).
				SetPerformedBy(DomainScanPerformedBy).
				SetStatus(enums.ScanStatusPending).
				SetMetadata(metadata).
				Save(createCtx)
			if err != nil {
				return nil, err
			}
		} else if groupID != "" {
			metadata := map[string]any{DomainScanGroupMetadataKey: groupID}
			if cfg.ApplyBrandDesign {
				metadata[DomainScanApplyBrandDesignMetadataKey] = true
			}

			scanRecord, err = scanRecord.Update().SetMetadata(metadata).Save(ctx)
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

		if cfg.BrandDesignOnly {
			if err := saga.runBrandDesignScan(ctx, organizationID, scanRecord.ID, cfg.Domain, cfg.ApplyBrandDesign); err != nil {
				return nil, err
			}

			return providerkit.EncodeResult(DomainScanRequestResult{
				Message: "domain brand design scan completed",
				ScanID:  scanRecord.ID,
			}, ErrResultEncode)
		}

		if err := saga.submitAndScheduleDomainScan(ctx, organizationID, scanRecord.ID, cfg.Domain, cfg.ForceRefresh); err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(DomainScanRequestResult{
			Message: "domain scan submitted",
			ScanID:  scanRecord.ID,
		}, ErrResultEncode)
	}
}
