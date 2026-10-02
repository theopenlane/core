package cloudflare

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// DomainScanRequestOp is the operation ref for requesting a domain scan
//
//nolint:revive
var DomainScanRequestOp = types.OperationPayloadOf[DomainScanRequest]().
	HandlesRequest(runDomainScanRequest).
	Policy(types.ExecutionPolicy{Inline: true, SkipRunRecord: true}).
	RateLimit(types.RateLimitPolicy{Window: time.Hour}).
	CustomerSelectable(false).
	RequiresPaymentMethod()

// DomainScanRequest queues a domain scan by creating a pending Scan record
type DomainScanRequest struct {
	// ScanID identifies the Scan record that triggered an internally dispatched request
	ScanID string `json:"scanId,omitempty"`
	// OrganizationID is the organization the scan belongs to, used only without an Integration
	OrganizationID string `json:"organizationId,omitempty"`
	// Domain is the domain to scan
	Domain string `json:"domain" jsonschema:"required,title=Domain,description=Domain to scan"`
	// ForceRefresh bypasses Cloudflare's Browser Rendering cache, forcing a fresh render
	ForceRefresh bool `json:"forceRefresh,omitempty" jsonschema:"title=Force Refresh,description=Bypass the render cache and force a fresh scan"`
	// BrandDesignOnly extracts the brand design without running the full domain scan
	BrandDesignOnly bool `json:"brandDesignOnly,omitempty" jsonschema:"title=Brand Design Only,description=Extract and apply the brand design without building a full domain scan report"`

	// ApplyBrandDesignToPreview applies the extracted brand design to the preview trustcenter
	ApplyBrandDesignToPreview bool `json:"applyBrandDesignToPreview,omitempty" jsonschema:"title=Apply Brand Design to Preview,description=Apply extracted brand design to preview Trust Center settings"`

	// ApplyBrandDesignToLive applies the extracted brand design to the live trustcenter
	ApplyBrandDesignToLive bool `json:"applyBrandDesignToLive,omitempty" jsonschema:"title=Apply Brand Design to Live,description=Apply extracted brand design to live Trust Center settings"`

	// GroupID links this scan to sibling scans requested together for a combined notification
	GroupID string `json:"groupId,omitempty"`
}

// DomainScanRequestResult acknowledges that a domain scan was queued or run
type DomainScanRequestResult struct {
	// Message describes what happened
	Message string `json:"message"`
	// ScanID is the id of the Scan record for this request
	ScanID string `json:"scanId"`
}

// runDomainScanRequest queues a pending Scan, or runs the domain scan saga for internal dispatches
func runDomainScanRequest(ctx context.Context, request types.OperationRequest, cfg DomainScanRequest) (json.RawMessage, error) {
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
		metadata[DomainScanApplyBrandDesignToPreviewMetadataKey] = cfg.ApplyBrandDesignToPreview
		metadata[DomainScanApplyBrandDesignToLiveMetadataKey] = cfg.ApplyBrandDesignToLive
		if groupID != "" {
			metadata[DomainScanGroupMetadataKey] = groupID
		}

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
		metadata[DomainScanApplyBrandDesignToPreviewMetadataKey] = cfg.ApplyBrandDesignToPreview
		metadata[DomainScanApplyBrandDesignToLiveMetadataKey] = cfg.ApplyBrandDesignToLive

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
		applyToPreview, _ := scanRecord.Metadata[DomainScanApplyBrandDesignToPreviewMetadataKey].(bool)
		applyToLive, _ := scanRecord.Metadata[DomainScanApplyBrandDesignToLiveMetadataKey].(bool)
		if err := saga.runBrandDesignScan(ctx, brandDesignScanOpts{
			organizationID:            organizationID,
			scanID:                    scanRecord.ID,
			domain:                    cfg.Domain,
			applyBrandDesignToPreview: applyToPreview,
			applyBrandDesignToLive:    applyToLive,
		}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("scan_id", scanRecord.ID).Msg("domain scan: brand design scan failed")
			saga.markDomainScanFailed(ctx, organizationID, scanRecord.ID)

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
